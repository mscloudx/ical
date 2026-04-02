// Package diff provides semantic comparison utilities for iCalendar models.
//
// Comparison is structural (field-by-field), not textual. Property ordering,
// CRLF vs LF, and encoding variants are ignored — only the model values matter.
//
// Usage:
//
//	issues := diff.Compare(src, dst)
//	if len(issues) > 0 { ... }
//
//	// or just:
//	if !diff.Equal(src, dst) { ... }
package diff

import (
	"fmt"
	"strings"
	"time"

	"github.com/mscloudx/ical/model"
)

// Issue describes a single field difference between two calendars.
type Issue struct {
	// Path is a dot-separated location of the differing field,
	// e.g. "Events[uid1@example.com].Summary".
	Path string
	// Got is the value found in the actual (second) calendar.
	Got string
	// Want is the value from the reference (first) calendar.
	Want string
}

func (i Issue) String() string {
	return fmt.Sprintf("%s: got %q, want %q", i.Path, i.Got, i.Want)
}

// Compare returns all semantic differences between src and dst.
// Returns nil if they are semantically equal.
func Compare(src, dst *model.Calendar) []Issue {
	c := &collector{}
	c.compareCalendar("", src, dst)
	return c.issues
}

// Equal reports whether src and dst are semantically equal.
func Equal(src, dst *model.Calendar) bool {
	return len(Compare(src, dst)) == 0
}

// --------------------------------------------------------------------------
// Internal collector
// --------------------------------------------------------------------------

type collector struct {
	issues []Issue
}

func (c *collector) add(path, got, want string) {
	c.issues = append(c.issues, Issue{Path: path, Got: got, Want: want})
}

func (c *collector) check(path, got, want string) {
	if got != want {
		c.add(path, got, want)
	}
}

func (c *collector) checkBool(path string, got, want bool) {
	if got != want {
		c.add(path, fmt.Sprintf("%v", got), fmt.Sprintf("%v", want))
	}
}

func (c *collector) checkInt(path string, got, want int) {
	if got != want {
		c.add(path, fmt.Sprintf("%d", got), fmt.Sprintf("%d", want))
	}
}

func (c *collector) checkTime(path string, got, want *time.Time) {
	switch {
	case got == nil && want == nil:
		return
	case got == nil:
		c.add(path, "<nil>", want.UTC().Format(time.RFC3339))
	case want == nil:
		c.add(path, got.UTC().Format(time.RFC3339), "<nil>")
	default:
		if !got.Equal(*want) {
			c.add(path, got.UTC().Format(time.RFC3339), want.UTC().Format(time.RFC3339))
		}
	}
}

func (c *collector) checkTimeVal(path string, got, want time.Time) {
	if !got.Equal(want) {
		c.add(path, got.UTC().Format(time.RFC3339), want.UTC().Format(time.RFC3339))
	}
}

func (c *collector) checkDuration(path string, got, want *time.Duration) {
	switch {
	case got == nil && want == nil:
		return
	case got == nil:
		c.add(path, "<nil>", want.String())
	case want == nil:
		c.add(path, got.String(), "<nil>")
	default:
		if *got != *want {
			c.add(path, got.String(), want.String())
		}
	}
}

func (c *collector) checkStrSlice(path string, got, want []string) {
	if strings.Join(got, ",") != strings.Join(want, ",") {
		c.add(path, fmt.Sprintf("%v", got), fmt.Sprintf("%v", want))
	}
}

func (c *collector) checkTimeSlice(path string, got, want []time.Time) {
	if len(got) != len(want) {
		c.add(path, fmt.Sprintf("len=%d", len(got)), fmt.Sprintf("len=%d", len(want)))
		return
	}
	for i := range got {
		if !got[i].Equal(want[i]) {
			c.add(fmt.Sprintf("%s[%d]", path, i), got[i].UTC().Format(time.RFC3339), want[i].UTC().Format(time.RFC3339))
		}
	}
}

func (c *collector) checkProperties(path string, got, want []model.Property) {
	byName := func(props []model.Property) map[string]string {
		m := make(map[string]string, len(props))
		for _, p := range props {
			key := p.Name
			for _, param := range p.Params {
				key += ";" + param.Name + "=" + strings.Join(param.Values, ",")
			}
			m[key] = p.Value
		}
		return m
	}
	gm := byName(got)
	wm := byName(want)
	for k, wv := range wm {
		if gv, ok := gm[k]; !ok {
			c.add(path+"."+k, "<missing>", wv)
		} else if gv != wv {
			c.add(path+"."+k, gv, wv)
		}
	}
	for k := range gm {
		if _, ok := wm[k]; !ok {
			c.add(path+"."+k, gm[k], "<missing>")
		}
	}
}

// --------------------------------------------------------------------------
// Calendar
// --------------------------------------------------------------------------

func (c *collector) compareCalendar(path string, got, want *model.Calendar) {
	p := func(f string) string {
		if path == "" {
			return f
		}
		return path + "." + f
	}

	if got == nil && want == nil {
		return
	}
	if got == nil {
		c.add(path, "<nil>", "Calendar{...}")
		return
	}
	if want == nil {
		c.add(path, "Calendar{...}", "<nil>")
		return
	}

	c.check(p("Version"), got.Version, want.Version)
	c.check(p("ProdID"), got.ProdID, want.ProdID)
	c.check(p("CalScale"), got.CalScale, want.CalScale)
	c.check(p("Method"), got.Method.String(), want.Method.String())
	c.check(p("Name"), got.Name, want.Name)
	c.check(p("Description"), got.Description, want.Description)
	c.check(p("UID"), got.UID, want.UID)
	c.check(p("URL"), got.URL, want.URL)
	c.check(p("Color"), got.Color, want.Color)
	c.check(p("Source"), got.Source, want.Source)
	c.checkTime(p("LastModified"), got.LastModified, want.LastModified)
	c.checkDuration(p("RefreshInterval"), got.RefreshInterval, want.RefreshInterval)
	c.checkStrSlice(p("Categories"), got.Categories, want.Categories)

	c.checkProperties(p("XProps"), got.XProps, want.XProps)
	c.checkProperties(p("IanaProps"), got.IanaProps, want.IanaProps)

	// Events — match by UID.
	c.compareEventSlice(p("Events"), got.Events, want.Events)
	c.compareTodoSlice(p("Todos"), got.Todos, want.Todos)
	c.compareJournalSlice(p("Journals"), got.Journals, want.Journals)
	c.compareFreeBusySlice(p("FreeBusys"), got.FreeBusys, want.FreeBusys)
	c.compareTimezoneSlice(p("Timezones"), got.Timezones, want.Timezones)
}

// --------------------------------------------------------------------------
// Events
// --------------------------------------------------------------------------

func (c *collector) compareEventSlice(path string, got, want []model.Event) {
	// Index want by UID.
	wantByUID := make(map[string]*model.Event, len(want))
	for i := range want {
		wantByUID[want[i].UID] = &want[i]
	}
	gotByUID := make(map[string]*model.Event, len(got))
	for i := range got {
		gotByUID[got[i].UID] = &got[i]
	}
	for uid, we := range wantByUID {
		key := fmt.Sprintf("%s[%s]", path, uid)
		if ge, ok := gotByUID[uid]; ok {
			c.compareEvent(key, ge, we)
		} else {
			c.add(key, "<missing>", "Event{UID:"+uid+"}")
		}
	}
	for uid := range gotByUID {
		if _, ok := wantByUID[uid]; !ok {
			c.add(fmt.Sprintf("%s[%s]", path, uid), "Event{UID:"+uid+"}", "<missing>")
		}
	}
}

func (c *collector) compareEvent(path string, got, want *model.Event) {
	p := func(f string) string { return path + "." + f }

	c.check(p("UID"), got.UID, want.UID)
	c.checkTimeVal(p("DTStamp"), got.DTStamp, want.DTStamp)
	c.checkTimeVal(p("DTStart"), got.DTStart, want.DTStart)
	c.checkBool(p("AllDay"), got.AllDay, want.AllDay)
	c.checkTime(p("DTEnd"), got.DTEnd, want.DTEnd)
	c.checkDuration(p("Duration"), got.Duration, want.Duration)
	c.check(p("Summary"), got.Summary, want.Summary)
	c.check(p("Description"), got.Description, want.Description)
	c.check(p("Location"), got.Location, want.Location)
	c.check(p("URL"), got.URL, want.URL)
	c.check(p("Status"), got.Status.String(), want.Status.String())
	c.check(p("Transparency"), got.Transparency.String(), want.Transparency.String())
	c.check(p("Classification"), got.Classification.String(), want.Classification.String())
	c.checkTime(p("Created"), got.Created, want.Created)
	c.checkTime(p("LastModified"), got.LastModified, want.LastModified)
	c.checkInt(p("Sequence"), got.Sequence, want.Sequence)
	c.checkTime(p("RecurrenceID"), got.RecurrenceID, want.RecurrenceID)
	c.checkTimeSlice(p("RDates"), got.RDates, want.RDates)
	c.checkTimeSlice(p("ExDates"), got.ExDates, want.ExDates)
	c.checkStrSlice(p("Concepts"), got.Concepts, want.Concepts)
	c.checkStrSlice(p("RefIDs"), got.RefIDs, want.RefIDs)
	c.checkProperties(p("XProps"), got.XProps, want.XProps)
	c.checkProperties(p("IanaProps"), got.IanaProps, want.IanaProps)
	c.compareAlarmSlice(p("Alarms"), got.Alarms, want.Alarms)

	if len(got.RRules) != len(want.RRules) {
		c.add(p("RRules"), fmt.Sprintf("len=%d", len(got.RRules)), fmt.Sprintf("len=%d", len(want.RRules)))
	}
}

// --------------------------------------------------------------------------
// Todos
// --------------------------------------------------------------------------

func (c *collector) compareTodoSlice(path string, got, want []model.Todo) {
	wantByUID := make(map[string]*model.Todo, len(want))
	for i := range want {
		wantByUID[want[i].UID] = &want[i]
	}
	gotByUID := make(map[string]*model.Todo, len(got))
	for i := range got {
		gotByUID[got[i].UID] = &got[i]
	}
	for uid, wt := range wantByUID {
		key := fmt.Sprintf("%s[%s]", path, uid)
		if gt, ok := gotByUID[uid]; ok {
			c.compareTodo(key, gt, wt)
		} else {
			c.add(key, "<missing>", "Todo{UID:"+uid+"}")
		}
	}
	for uid := range gotByUID {
		if _, ok := wantByUID[uid]; !ok {
			c.add(fmt.Sprintf("%s[%s]", path, uid), "Todo{UID:"+uid+"}", "<missing>")
		}
	}
}

func (c *collector) compareTodo(path string, got, want *model.Todo) {
	p := func(f string) string { return path + "." + f }

	c.check(p("UID"), got.UID, want.UID)
	c.checkTimeVal(p("DTStamp"), got.DTStamp, want.DTStamp)
	c.checkTime(p("DTStart"), got.DTStart, want.DTStart)
	c.checkBool(p("AllDay"), got.AllDay, want.AllDay)
	c.checkTime(p("Due"), got.Due, want.Due)
	c.checkDuration(p("Duration"), got.Duration, want.Duration)
	c.checkTime(p("Completed"), got.Completed, want.Completed)
	c.checkInt(p("Priority"), got.Priority, want.Priority)
	c.checkInt(p("PercentComplete"), got.PercentComplete, want.PercentComplete)
	c.check(p("Summary"), got.Summary, want.Summary)
	c.check(p("Description"), got.Description, want.Description)
	c.check(p("Location"), got.Location, want.Location)
	c.check(p("URL"), got.URL, want.URL)
	c.check(p("Status"), got.Status.String(), want.Status.String())
	c.check(p("Classification"), got.Classification.String(), want.Classification.String())
	c.checkTime(p("Created"), got.Created, want.Created)
	c.checkTime(p("LastModified"), got.LastModified, want.LastModified)
	c.checkInt(p("Sequence"), got.Sequence, want.Sequence)
	c.checkTime(p("RecurrenceID"), got.RecurrenceID, want.RecurrenceID)
	c.checkTimeSlice(p("RDates"), got.RDates, want.RDates)
	c.checkTimeSlice(p("ExDates"), got.ExDates, want.ExDates)
	c.checkStrSlice(p("Concepts"), got.Concepts, want.Concepts)
	c.checkStrSlice(p("RefIDs"), got.RefIDs, want.RefIDs)
	c.checkProperties(p("XProps"), got.XProps, want.XProps)
	c.checkProperties(p("IanaProps"), got.IanaProps, want.IanaProps)
	c.compareAlarmSlice(p("Alarms"), got.Alarms, want.Alarms)
}

// --------------------------------------------------------------------------
// Journals
// --------------------------------------------------------------------------

func (c *collector) compareJournalSlice(path string, got, want []model.Journal) {
	wantByUID := make(map[string]*model.Journal, len(want))
	for i := range want {
		wantByUID[want[i].UID] = &want[i]
	}
	gotByUID := make(map[string]*model.Journal, len(got))
	for i := range got {
		gotByUID[got[i].UID] = &got[i]
	}
	for uid, wj := range wantByUID {
		key := fmt.Sprintf("%s[%s]", path, uid)
		if gj, ok := gotByUID[uid]; ok {
			c.compareJournal(key, gj, wj)
		} else {
			c.add(key, "<missing>", "Journal{UID:"+uid+"}")
		}
	}
	for uid := range gotByUID {
		if _, ok := wantByUID[uid]; !ok {
			c.add(fmt.Sprintf("%s[%s]", path, uid), "Journal{UID:"+uid+"}", "<missing>")
		}
	}
}

func (c *collector) compareJournal(path string, got, want *model.Journal) {
	p := func(f string) string { return path + "." + f }

	c.check(p("UID"), got.UID, want.UID)
	c.checkTimeVal(p("DTStamp"), got.DTStamp, want.DTStamp)
	c.checkTime(p("DTStart"), got.DTStart, want.DTStart)
	c.checkBool(p("AllDay"), got.AllDay, want.AllDay)
	c.check(p("Summary"), got.Summary, want.Summary)
	c.checkStrSlice(p("Descriptions"), got.Descriptions, want.Descriptions)
	c.check(p("URL"), got.URL, want.URL)
	c.check(p("Status"), got.Status.String(), want.Status.String())
	c.check(p("Classification"), got.Classification.String(), want.Classification.String())
	c.checkTime(p("Created"), got.Created, want.Created)
	c.checkTime(p("LastModified"), got.LastModified, want.LastModified)
	c.checkInt(p("Sequence"), got.Sequence, want.Sequence)
	c.checkTimeSlice(p("RDates"), got.RDates, want.RDates)
	c.checkTimeSlice(p("ExDates"), got.ExDates, want.ExDates)
	c.checkStrSlice(p("Concepts"), got.Concepts, want.Concepts)
	c.checkStrSlice(p("RefIDs"), got.RefIDs, want.RefIDs)
	c.checkProperties(p("XProps"), got.XProps, want.XProps)
	c.checkProperties(p("IanaProps"), got.IanaProps, want.IanaProps)
}

// --------------------------------------------------------------------------
// FreeBusy
// --------------------------------------------------------------------------

func (c *collector) compareFreeBusySlice(path string, got, want []model.FreeBusy) {
	if len(got) != len(want) {
		c.add(path, fmt.Sprintf("len=%d", len(got)), fmt.Sprintf("len=%d", len(want)))
		return
	}
	for i := range got {
		c.compareFreeBusy(fmt.Sprintf("%s[%d]", path, i), &got[i], &want[i])
	}
}

func (c *collector) compareFreeBusy(path string, got, want *model.FreeBusy) {
	p := func(f string) string { return path + "." + f }
	c.check(p("UID"), got.UID, want.UID)
	c.checkTimeVal(p("DTStamp"), got.DTStamp, want.DTStamp)
	c.checkTime(p("DTStart"), got.DTStart, want.DTStart)
	c.checkTime(p("DTEnd"), got.DTEnd, want.DTEnd)
	c.check(p("URL"), got.URL, want.URL)
	c.checkProperties(p("XProps"), got.XProps, want.XProps)
	c.checkProperties(p("IanaProps"), got.IanaProps, want.IanaProps)
}

// --------------------------------------------------------------------------
// Timezones
// --------------------------------------------------------------------------

func (c *collector) compareTimezoneSlice(path string, got, want []model.Timezone) {
	wantByTZID := make(map[string]*model.Timezone, len(want))
	for i := range want {
		wantByTZID[want[i].TZID] = &want[i]
	}
	gotByTZID := make(map[string]*model.Timezone, len(got))
	for i := range got {
		gotByTZID[got[i].TZID] = &got[i]
	}
	for tzid, wtz := range wantByTZID {
		key := fmt.Sprintf("%s[%s]", path, tzid)
		if gtz, ok := gotByTZID[tzid]; ok {
			c.compareTimezone(key, gtz, wtz)
		} else {
			c.add(key, "<missing>", "Timezone{TZID:"+tzid+"}")
		}
	}
	for tzid := range gotByTZID {
		if _, ok := wantByTZID[tzid]; !ok {
			c.add(fmt.Sprintf("%s[%s]", path, tzid), "Timezone{TZID:"+tzid+"}", "<missing>")
		}
	}
}

func (c *collector) compareTimezone(path string, got, want *model.Timezone) {
	p := func(f string) string { return path + "." + f }
	c.check(p("TZID"), got.TZID, want.TZID)
	if len(got.Standards) != len(want.Standards) {
		c.add(p("Standards"), fmt.Sprintf("len=%d", len(got.Standards)), fmt.Sprintf("len=%d", len(want.Standards)))
	}
	if len(got.Daylights) != len(want.Daylights) {
		c.add(p("Daylights"), fmt.Sprintf("len=%d", len(got.Daylights)), fmt.Sprintf("len=%d", len(want.Daylights)))
	}
}

// --------------------------------------------------------------------------
// Alarms
// --------------------------------------------------------------------------

func (c *collector) compareAlarmSlice(path string, got, want []model.Alarm) {
	if len(got) != len(want) {
		c.add(path, fmt.Sprintf("len=%d", len(got)), fmt.Sprintf("len=%d", len(want)))
		return
	}
	for i := range got {
		c.compareAlarm(fmt.Sprintf("%s[%d]", path, i), &got[i], &want[i])
	}
}

func (c *collector) compareAlarm(path string, got, want *model.Alarm) {
	p := func(f string) string { return path + "." + f }

	c.check(p("Action"), got.Action.String(), want.Action.String())
	c.compareTrigger(p("Trigger"), got.Trigger, want.Trigger)
	c.check(p("Description"), got.Description, want.Description)
	c.check(p("Summary"), got.Summary, want.Summary)
	c.checkInt(p("Repeat"), got.Repeat, want.Repeat)
	c.checkDuration(p("Duration"), got.Duration, want.Duration)
	c.checkProperties(p("XProps"), got.XProps, want.XProps)
	c.checkProperties(p("IanaProps"), got.IanaProps, want.IanaProps)
}

func (c *collector) compareTrigger(path string, got, want model.Trigger) {
	p := func(f string) string { return path + "." + f }
	c.checkDuration(p("Duration"), got.Duration, want.Duration)
	c.checkTime(p("DateTime"), got.DateTime, want.DateTime)
	c.check(p("Related"), got.Related, want.Related)
}
