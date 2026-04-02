package parser

import (
	"fmt"
	"sync"
	"time"
)

// --------------------------------------------------------------------------
// Global timezone caches
// --------------------------------------------------------------------------

// ianaZoneCache stores IANA system timezone locations, keyed by TZID.
// System zones are immutable so sharing across concurrent Parse calls is safe.
var ianaZoneCache = &syncZoneCache{
	zones:   make(map[string]*time.Location),
	missing: make(map[string]struct{}),
}

// customTZCache stores parsed custom VTIMEZONE rules, keyed by content
// fingerprint (NOT by TZID). Two VTIMEZONEs with identical transition rules
// share a single *vtimezoneTransitions regardless of TZID or calendar origin.
//
// This eliminates two problems at once:
//   - Thunderbird sometimes sends ≥20 identical VTIMEZONE blocks per calendar:
//     the second and subsequent blocks are free (O(1) hash lookup).
//   - Concurrent Parse calls with the same TZID but different rules no longer
//     collide, because each Parse builds its own TZID→transitions map from
//     per-calendar fingerprint lookups.
var customTZCache = &syncCustomTZCache{zones: make(map[string]*vtimezoneTransitions)}

// --------------------------------------------------------------------------
// syncZoneCache — thread-safe *time.Location cache
// --------------------------------------------------------------------------

// syncZoneCache caches both successful and failed time.LoadLocation calls.
//
// zones stores known-good IANA locations.
// missing is a negative cache: TZIDs confirmed absent from the system IANA
// database are stored here so time.LoadLocation (which reads a zip archive)
// is never called twice for the same unknown TZID. Without this, every Parse
// call for a calendar with custom VTIMEZONEs pays the full zip-read cost.
type syncZoneCache struct {
	mu      sync.RWMutex
	zones   map[string]*time.Location
	missing map[string]struct{}
}

// loadIANA returns the *time.Location for tzid from the system IANA database,
// caching both hits and misses so the zip archive is read at most once per TZID.
func loadIANA(tzid string) (*time.Location, error) {
	ianaZoneCache.mu.RLock()
	loc, found := ianaZoneCache.zones[tzid]
	_, absent := ianaZoneCache.missing[tzid]
	ianaZoneCache.mu.RUnlock()

	if found {
		return loc, nil
	}
	if absent {
		return nil, errNotIANA
	}

	loc, err := time.LoadLocation(tzid)

	ianaZoneCache.mu.Lock()
	defer ianaZoneCache.mu.Unlock()
	if err != nil {
		// Negative cache: don't read the zip again for this TZID.
		ianaZoneCache.missing[tzid] = struct{}{}
		return nil, err
	}
	// Double-check: another goroutine may have stored it while we were loading.
	if existing, ok := ianaZoneCache.zones[tzid]; ok {
		return existing, nil
	}
	ianaZoneCache.zones[tzid] = loc
	return loc, nil
}

// errNotIANA is returned by loadIANA for TZIDs confirmed absent from the system
// IANA database. It is a sentinel — callers only check err != nil.
var errNotIANA = fmt.Errorf("TZID not found in system IANA database")

// --------------------------------------------------------------------------
// syncCustomTZCache — content-addressed cache for custom VTIMEZONE rules
// --------------------------------------------------------------------------

type syncCustomTZCache struct {
	mu    sync.RWMutex
	zones map[string]*vtimezoneTransitions
}

// load returns the cached *vtimezoneTransitions for the given content fingerprint.
func (c *syncCustomTZCache) load(fp string) (*vtimezoneTransitions, bool) {
	c.mu.RLock()
	vtz, ok := c.zones[fp]
	c.mu.RUnlock()
	return vtz, ok
}

// loadOrStore returns the cached entry for fp, or stores and returns vtz.
// The returned pointer may differ from vtz if a concurrent goroutine stored first.
func (c *syncCustomTZCache) loadOrStore(fp string, vtz *vtimezoneTransitions) *vtimezoneTransitions {
	c.mu.RLock()
	existing, ok := c.zones[fp]
	c.mu.RUnlock()
	if ok {
		return existing
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if existing, ok = c.zones[fp]; ok {
		return existing
	}
	c.zones[fp] = vtz
	return vtz
}

// --------------------------------------------------------------------------
// vtimezoneTransitions — parsed VTIMEZONE rules
// --------------------------------------------------------------------------

// vtimezoneTransitions holds the parsed STANDARD/DAYLIGHT transition rules
// extracted from a VTIMEZONE component.
//
// fixedLoc is non-nil when the zone has a single non-recurring transition
// (offset is constant for all time). locationAt returns it directly with
// zero allocations — this covers the majority of real-world custom VTIMEZONEs.
//
// offsetLocs maps each unique UTC-offset value (seconds) to a pre-built
// *time.Location. locationAt uses this map instead of calling time.FixedZone
// on every datetime, eliminating per-call Location allocations for the
// YEARLY/DST path as well.
//
// Both fields are populated once in parseVTimezoneTransitions and then never
// mutated, so they are safe to read from multiple goroutines without locking.
type vtimezoneTransitions struct {
	tzid        string
	transitions []tzTransition         // sorted ascending by dtstart
	fixedLoc    *time.Location         // non-nil when offset is constant for all time
	offsetLocs  map[int]*time.Location // offset (seconds) → pre-built FixedZone
}

// tzTransition represents one STANDARD or DAYLIGHT rule.
type tzTransition struct {
	dtstart  time.Time // wall-clock start, parsed as naïve UTC for comparison
	offsetTo int       // seconds east of UTC after this transition
	yearly   bool      // true when RRULE contains FREQ=YEARLY
}

// tzCandidate is used by locationAt to evaluate active offset candidates.
// Defined at package level so a fixed-size stack buffer can be used.
type tzCandidate struct {
	dtstart  time.Time
	offsetTo int
}

// locationAt returns the *time.Location (a FixedZone) that was active at the
// given wall-clock time. wallTime must be the datetime string from the
// iCalendar property parsed as a naïve UTC moment (same digits, no zone shift).
//
// Fast path: if fixedLoc is set (single non-recurring transition), it is
// returned immediately with zero allocations.
//
// Algorithm for the general path:
//  1. For every YEARLY transition, generate occurrences for wallTime.Year()-1
//     through wallTime.Year()+1 to cover DST boundaries across year edges.
//  2. Among all candidates whose dtstart ≤ wallTime, pick the latest one.
//  3. If wallTime precedes all transitions, use the earliest known offset
//     (conservative: apply the first rule before it officially started).
//
// The candidates slice is stack-allocated for the common case of ≤2 transitions
// (capacity ≤ 8), avoiding heap pressure for STANDARD+DAYLIGHT VTIMEZONEs.
// The resulting *time.Location is looked up from the pre-built offsetLocs map
// instead of calling time.FixedZone on every invocation.
func (v *vtimezoneTransitions) locationAt(wallTime time.Time) *time.Location {
	if len(v.transitions) == 0 {
		return time.UTC
	}

	// Fast path: constant offset — no allocation needed.
	if v.fixedLoc != nil {
		return v.fixedLoc
	}

	year := wallTime.Year()

	// Stack buffer covers the typical STANDARD+DAYLIGHT case (2 transitions × 4 = 8).
	// For larger VTIMEZONEs the slice escapes to the heap, same as before.
	var stackBuf [8]tzCandidate
	var candidates []tzCandidate
	if need := len(v.transitions) * 4; need <= len(stackBuf) { //nolint:mnd // 4 candidates per transition: STANDARD+DAYLIGHT × start/end
		candidates = stackBuf[:0]
	} else {
		candidates = make([]tzCandidate, 0, need)
	}

	for _, tr := range v.transitions {
		candidates = append(candidates, tzCandidate{tr.dtstart, tr.offsetTo})
		if tr.yearly {
			for _, y := range [3]int{year - 1, year, year + 1} {
				dt := time.Date(
					y,
					tr.dtstart.Month(), tr.dtstart.Day(),
					tr.dtstart.Hour(), tr.dtstart.Minute(), tr.dtstart.Second(),
					0, time.UTC,
				)
				candidates = append(candidates, tzCandidate{dt, tr.offsetTo})
			}
		}
	}

	// Find the latest candidate whose dtstart ≤ wallTime.
	best := -1
	for i := range candidates {
		c := &candidates[i]
		if !wallTime.Before(c.dtstart) {
			if best < 0 || c.dtstart.After(candidates[best].dtstart) {
				best = i
			}
		}
	}
	if best >= 0 {
		return v.offsetLocs[candidates[best].offsetTo]
	}

	// wallTime precedes all transitions — use the earliest known offset.
	earliest := 0
	for i := range candidates[1:] {
		if candidates[i+1].dtstart.Before(candidates[earliest].dtstart) {
			earliest = i + 1
		}
	}
	return v.offsetLocs[candidates[earliest].offsetTo]
}
