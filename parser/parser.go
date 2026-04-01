// Package parser реализует высокопроизводительный парсер iCalendar (RFC 5545).
//
// Предоставляет три точки входа для различных сценариев:
//   - Parse(io.Reader) — для потокового чтения
//   - ParseBytes([]byte) — для данных в памяти
//   - ParseWithClose(io.ReadCloser) — для HTTP-ответов с автоматическим Close
package parser

import (
	"bytes"
	"encoding/base64"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mscloudx/ical/model"
	"github.com/pkg/errors"
)

// Константы для часто используемых имён свойств.
const (
	propUID           = "UID"
	propDTSTAMP       = "DTSTAMP"
	propDTSTART       = "DTSTART"
	propDTEND         = "DTEND"
	propDURATION      = "DURATION"
	propSUMMARY       = "SUMMARY"
	propDESCRIPTION   = "DESCRIPTION"
	propURL           = "URL"
	propSTATUS        = "STATUS"
	propCLASS         = "CLASS"
	propORGANIZER     = "ORGANIZER"
	propATTENDEE      = "ATTENDEE"
	propCREATED       = "CREATED"
	propLASTMODIFIED  = "LAST-MODIFIED"
	propSEQUENCE      = "SEQUENCE"
	propRECURRENCEID  = "RECURRENCE-ID"
	propRRULE         = "RRULE"
	propRDATE         = "RDATE"
	propEXDATE        = "EXDATE"
	propRELATEDTO     = "RELATED-TO"
	propRequestStatus = "REQUEST-STATUS"
	propATTACH        = "ATTACH"
	propTZOFFSETTO    = "TZOFFSETTO"
)

// --------------------------------------------------------------------------
// Публичный API
// --------------------------------------------------------------------------

// Parse парсит iCalendar-данные из io.Reader и возвращает Calendar.
func Parse(r io.Reader) (*model.Calendar, error) {
	if r == nil {
		return nil, ErrNilReader
	}
	scanner := newLineScanner(r)
	raw, err := scanComponentTree(scanner)
	if err != nil {
		return nil, errors.Wrap(err, "scan")
	}
	if scanner.Err() != nil {
		return nil, errors.Wrap(ErrScanFailed, scanner.Err().Error())
	}
	return buildCalendar(raw)
}

// ParseBytes парсит iCalendar-данные из среза байт.
func ParseBytes(data []byte) (*model.Calendar, error) {
	return Parse(bytes.NewReader(data))
}

// ParseWithClose парсит iCalendar-данные из io.ReadCloser
// и гарантирует вызов Close() после завершения чтения.
func ParseWithClose(rc io.ReadCloser) (*model.Calendar, error) {
	if rc == nil {
		return nil, ErrNilReader
	}
	defer rc.Close()
	return Parse(rc)
}

// --------------------------------------------------------------------------
// Промежуточное представление компонента
// --------------------------------------------------------------------------

// rawComponent — узел дерева компонентов при парсинге.
type rawComponent struct {
	name     string
	props    []model.Property
	children []*rawComponent
}

// scanComponentTree сканирует поток и строит дерево rawComponent.
func scanComponentTree(scanner *lineScanner) (*rawComponent, error) {
	root := &rawComponent{name: "ROOT"}
	stack := []*rawComponent{root}

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		prop := parseContentLine(line)
		current := stack[len(stack)-1]

		switch prop.Name {
		case "BEGIN":
			child := &rawComponent{name: strings.ToUpper(prop.Value)}
			current.children = append(current.children, child)
			stack = append(stack, child)

		case "END":
			if len(stack) <= 1 {
				return nil, errors.Wrapf(ErrUnexpectedEnd, "END:%s", prop.Value)
			}
			expected := stack[len(stack)-1].name
			got := strings.ToUpper(prop.Value)
			if expected != got {
				return nil, errors.Wrapf(ErrMismatchedEnd, "expected %s, got %s", expected, got)
			}
			stack = stack[:len(stack)-1]

		default:
			current.props = append(current.props, prop)
		}
	}

	if len(stack) > 1 {
		return nil, errors.Wrapf(ErrUnclosedComponent, "%s", stack[len(stack)-1].name)
	}

	return root, nil
}

// --------------------------------------------------------------------------
// Построение типизированных моделей из rawComponent
// --------------------------------------------------------------------------

// buildCalendar строит Calendar из дерева rawComponent.
//
//nolint:gocyclo // Ветвление следует RFC: упрощение ухудшит читаемость.
func buildCalendar(root *rawComponent) (*model.Calendar, error) {
	// Ожидаем, что корень ROOT содержит один дочерний VCALENDAR.
	var vcal *rawComponent
	for _, child := range root.children {
		if child.name == "VCALENDAR" {
			vcal = child
			break
		}
	}
	if vcal == nil {
		return nil, ErrNoVCalendar
	}

	cal := &model.Calendar{}

	// Обрабатываем свойства VCALENDAR.
	for _, p := range vcal.props {
		switch strings.ToUpper(p.Name) {
		case "VERSION":
			cal.Version = p.Value
		case "PRODID":
			cal.ProdID = p.Value
		case "CALSCALE":
			cal.CalScale = p.Value
		case "METHOD":
			cal.Method = parseMethod(p.Value)
		case "NAME":
			cal.Name = unescapeText(p.Value)
		case "DESCRIPTION":
			cal.Description = unescapeText(p.Value)
		case "UID":
			cal.UID = p.Value
		case "LAST-MODIFIED":
			t, _, err := parseDateTime(p.Value, nil)
			if err != nil {
				return nil, errors.Wrap(err, "LAST-MODIFIED")
			}
			cal.LastModified = &t
		case "URL":
			cal.URL = p.Value
		case "CATEGORIES":
			cats := strings.Split(p.Value, ",")
			for i := range cats {
				cats[i] = strings.TrimSpace(cats[i])
			}
			cal.Categories = append(cal.Categories, cats...)
		case "REFRESH-INTERVAL":
			d, err := parseDuration(p.Value)
			if err != nil {
				return nil, errors.Wrap(err, "REFRESH-INTERVAL")
			}
			cal.RefreshInterval = &d
		case "COLOR":
			cal.Color = p.Value
		case "SOURCE":
			cal.Source = p.Value
		default:
			switch {
			case isXProp(p.Name):
				cal.XProps = append(cal.XProps, p)
			case isIanaProp(p.Name):
				cal.IanaProps = append(cal.IanaProps, p)
			}
		}
	}

	// Pre-pass: регистрируем часовые пояса из VTIMEZONE в per-parse контексте.
	// Это позволяет resolveTZIDCtx использовать их при построении событий и
	// гарантирует корректность при параллельном парсинге разных календарей
	// с одинаковым TZID, но разными правилами переходов.
	tzCtx := newTZContext()
	for _, child := range vcal.children {
		if child.name == "VTIMEZONE" {
			registerVTimezoneLocation(child, tzCtx)
		}
	}

	// Обрабатываем дочерние компоненты.
	for _, child := range vcal.children {
		switch child.name {
		case "VEVENT":
			event, err := buildEvent(child, tzCtx)
			if err != nil {
				return nil, errors.Wrap(err, "VEVENT")
			}
			cal.Events = append(cal.Events, event)
		case "VTODO":
			todo, err := buildTodo(child, tzCtx)
			if err != nil {
				return nil, errors.Wrap(err, "VTODO")
			}
			cal.Todos = append(cal.Todos, todo)
		case "VJOURNAL":
			journal, err := buildJournal(child, tzCtx)
			if err != nil {
				return nil, errors.Wrap(err, "VJOURNAL")
			}
			cal.Journals = append(cal.Journals, journal)
		case "VTIMEZONE":
			tz, err := buildTimezone(child)
			if err != nil {
				return nil, errors.Wrap(err, "VTIMEZONE")
			}
			cal.Timezones = append(cal.Timezones, tz)
		case "VAVAILABILITY":
			av, err := buildAvailability(child, tzCtx)
			if err != nil {
				return nil, errors.Wrap(err, "VAVAILABILITY")
			}
			cal.Availabilities = append(cal.Availabilities, av)
		case "VFREEBUSY":
			fb, err := buildFreeBusy(child, tzCtx)
			if err != nil {
				return nil, errors.Wrap(err, "VFREEBUSY")
			}
			cal.FreeBusys = append(cal.FreeBusys, fb)
		}
	}

	return cal, nil
}

// buildEvent строит Event из rawComponent.
func buildEvent(raw *rawComponent, tzCtx *perCalendarTZContext) (model.Event, error) {
	var e model.Event

	for _, p := range raw.props {
		if err := setEventProp(&e, p, tzCtx); err != nil {
			return e, err
		}
	}

	for _, child := range raw.children {
		switch child.name {
		case componentAlarm:
			alarm, err := buildAlarm(child, tzCtx)
			if err != nil {
				return e, errors.Wrap(err, componentAlarm)
			}
			e.Alarms = append(e.Alarms, alarm)
		case componentParticipant:
			part, err := buildParticipant(child)
			if err != nil {
				return e, errors.Wrap(err, componentParticipant)
			}
			e.Participants = append(e.Participants, part)
		case componentLocation:
			locComponent, err := buildLocationComponent(child)
			if err != nil {
				return e, errors.Wrap(err, "LOCATION component")
			}
			e.Locations = append(e.Locations, locComponent)
		case componentResource:
			resComponent, err := buildResourceComponent(child)
			if err != nil {
				return e, errors.Wrap(err, "RESOURCE component")
			}
			e.Resources = append(e.Resources, resComponent)
		}
	}

	if e.UID == "" {
		return e, ErrMissingUID
	}

	return e, nil
}

// setEventProp устанавливает значение одного свойства в Event.
//
//nolint:gocyclo // switch по свойствам iCalendar неизбежно содержит много case-ов.
func setEventProp(e *model.Event, p model.Property, tzCtx *perCalendarTZContext) error {
	loc := resolveTZIDCtx(p, tzCtx)

	switch strings.ToUpper(p.Name) {
	case propUID:
		e.UID = p.Value
	case propDTSTAMP:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propDTSTAMP)
		}
		e.DTStamp = t
	case propDTSTART:
		t, isDate, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "DTSTART")
		}
		e.DTStart = t
		e.AllDay = isDate
	case propDTEND:
		t, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propDTEND)
		}
		e.DTEnd = &t
	case propDURATION:
		d, err := parseDuration(p.Value)
		if err != nil {
			return errors.Wrap(err, propDURATION)
		}
		e.Duration = &d
	case propSUMMARY:
		e.Summary = unescapeText(p.Value)
	case propDESCRIPTION:
		e.Description = unescapeText(p.Value)
	case propLocation:
		e.Location = unescapeText(p.Value)
	case propURL:
		e.URL = p.Value
	case propSTATUS:
		e.Status = parseStatus(p.Value)
	case "TRANSP":
		e.Transparency = parseTransparency(p.Value)
	case propCLASS:
		e.Classification = parseClassification(p.Value)
	case propORGANIZER:
		a := parseAttendee(p.Params, p.Value)
		e.Organizer = &a
	case propATTENDEE:
		e.Attendees = append(e.Attendees, parseAttendee(p.Params, p.Value))
	case propCREATED:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propCREATED)
		}
		e.Created = &t
	case propLASTMODIFIED:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propLASTMODIFIED)
		}
		e.LastModified = &t
	case propSEQUENCE:
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidRRule, "SEQUENCE=%s", p.Value)
		}
		e.Sequence = n
	case propRECURRENCEID:
		t, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propRECURRENCEID)
		}
		e.RecurrenceID = &t
	case propRRULE:
		rule, err := parseRRule(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propRRULE)
		}
		e.RRules = append(e.RRules, rule)
	case propRDATE:
		if strings.EqualFold(p.ParamValue("VALUE"), "PERIOD") {
			periods, err := parsePeriodList(p.Value, loc)
			if err != nil {
				return errors.Wrap(err, propRDATE)
			}
			e.RDatePeriods = append(e.RDatePeriods, periods...)
		} else {
			dates, err := parseDateTimeList(p.Value, loc)
			if err != nil {
				return errors.Wrap(err, propRDATE)
			}
			e.RDates = append(e.RDates, dates...)
		}
	case propEXDATE:
		dates, err := parseDateTimeList(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propEXDATE)
		}
		e.ExDates = append(e.ExDates, dates...)
	case propRELATEDTO:
		e.Related = append(e.Related, buildRelation(p))
	case propStructuredData:
		e.StructuredData = append(e.StructuredData, parseStructuredData(p))
	case propStyledDescription:
		e.StyledDescriptions = append(e.StyledDescriptions, parseStyledDescription(p))
	case propConcept:
		e.Concepts = append(e.Concepts, p.Value)
	case propRefID:
		e.RefIDs = append(e.RefIDs, p.Value)
	case propLink:
		e.Links = append(e.Links, parseLink(p))
	case propRequestStatus:
		e.RequestStatus = append(e.RequestStatus, parseRequestStatus(p.Value))
	case propATTACH:
		att, err := parseAttachment(p)
		if err != nil {
			return errors.Wrap(err, propATTACH)
		}
		e.Attach = append(e.Attach, att)
	default:
		switch {
		case isXProp(p.Name):
			e.XProps = append(e.XProps, p)
		case isIanaProp(p.Name):
			e.IanaProps = append(e.IanaProps, p)
		}
	}

	return nil
}

// registerVTimezoneLocation разбирает VTIMEZONE и сохраняет информацию
// о переходах в per-parse tzCtx.
//
// Для TZID, известных системной базе IANA, используется time.LoadLocation
// (результат кэшируется глобально). Для кастомных TZID парсятся правила
// STANDARD/DAYLIGHT и кэшируются по content-fingerprint, что позволяет:
//   - корректно выбирать смещение по дате события (DAYLIGHT vs STANDARD),
//   - не путать разные правила при параллельном парсинге (P1),
//   - переиспользовать разобранные правила без повторного разбора (эффективность).
func registerVTimezoneLocation(raw *rawComponent, tzCtx *perCalendarTZContext) {
	var tzid string
	for _, p := range raw.props {
		if strings.EqualFold(p.Name, "TZID") {
			tzid = p.Value
			break
		}
	}
	if tzid == "" {
		return
	}

	// Duplicate VTIMEZONE block in the same calendar (Thunderbird / Outlook pattern):
	// already registered this iteration — skip fingerprint computation entirely.
	if _, ok := tzCtx.iana[tzid]; ok {
		return
	}
	if _, ok := tzCtx.custom[tzid]; ok {
		return
	}

	// IANA-известный TZID — используем системную базу (глобально кэшируется).
	if loc, err := loadIANA(tzid); err == nil {
		tzCtx.iana[tzid] = loc
		return
	}

	// Кастомный TZID — ищем в content-addressed глобальном кэше.
	fp := vtimezoneFingerprint(tzid, raw)
	if vtz, ok := customTZCache.load(fp); ok {
		tzCtx.custom[tzid] = vtz
		return
	}

	// Разбираем переходы и сохраняем в кэш.
	vtz := parseVTimezoneTransitions(tzid, raw)
	vtz = customTZCache.loadOrStore(fp, vtz)
	tzCtx.custom[tzid] = vtz
}

// vtimezoneFingerprint строит детерминированный ключ для content-addressed кэша.
// Включает TZID и все свойства дочерних компонентов STANDARD/DAYLIGHT.
func vtimezoneFingerprint(tzid string, raw *rawComponent) string {
	var sb strings.Builder
	sb.WriteString(tzid)
	for _, child := range raw.children {
		sb.WriteByte('|')
		sb.WriteString(child.name)
		for _, p := range child.props {
			sb.WriteByte(';')
			sb.WriteString(strings.ToUpper(p.Name))
			sb.WriteByte('=')
			sb.WriteString(p.Value)
		}
	}
	return sb.String()
}

// parseVTimezoneTransitions извлекает переходы STANDARD/DAYLIGHT из VTIMEZONE.
// Результат сортируется по DTSTART для корректного поиска в locationAt.
//
// После построения transitions функция предвычисляет:
//   - offsetLocs: map[offset]*time.Location со всеми уникальными FixedZone,
//     чтобы locationAt не вызывал time.FixedZone на каждый datetime-вызов.
//   - fixedLoc: ненулевой, когда зона имеет ровно один non-yearly переход
//     (постоянный offset), что позволяет locationAt вернуться без alloc вообще.
func parseVTimezoneTransitions(tzid string, raw *rawComponent) *vtimezoneTransitions {
	vtz := &vtimezoneTransitions{tzid: tzid}
	for _, child := range raw.children {
		if child.name != "STANDARD" && child.name != "DAYLIGHT" {
			continue
		}
		var tr tzTransition
		for _, p := range child.props {
			switch strings.ToUpper(p.Name) {
			case propDTSTART:
				// DTSTART в VTIMEZONE — локальное wall-clock время.
				// Парсим как UTC для сравнения (смещение не важно здесь).
				t, _, err := parseDateTime(p.Value, nil)
				if err == nil {
					tr.dtstart = t
				}
			case propTZOFFSETTO:
				secs, err := parseUTCOffset(p.Value)
				if err == nil {
					tr.offsetTo = secs
				}
			case propRRULE:
				if strings.Contains(strings.ToUpper(p.Value), "FREQ=YEARLY") {
					tr.yearly = true
				}
			}
		}
		vtz.transitions = append(vtz.transitions, tr)
	}
	sort.Slice(vtz.transitions, func(i, j int) bool {
		return vtz.transitions[i].dtstart.Before(vtz.transitions[j].dtstart)
	})

	// Pre-build one *time.Location per unique UTC offset.
	// vtimezoneTransitions is immutable after construction (stored in the global
	// content-addressed cache), so this map is safe to read concurrently.
	vtz.offsetLocs = make(map[int]*time.Location, len(vtz.transitions))
	for _, tr := range vtz.transitions {
		if _, ok := vtz.offsetLocs[tr.offsetTo]; !ok {
			vtz.offsetLocs[tr.offsetTo] = time.FixedZone(tzid, tr.offsetTo)
		}
	}

	// Fast path: single non-recurring transition → offset never changes.
	if len(vtz.transitions) == 1 && !vtz.transitions[0].yearly {
		vtz.fixedLoc = vtz.offsetLocs[vtz.transitions[0].offsetTo]
	}

	return vtz
}

// wallClockApprox парсит строку даты/времени как наивный UTC-момент.
// Используется для выбора нужного перехода в vtimezoneTransitions.locationAt.
func wallClockApprox(s string) time.Time {
	switch len(s) {
	case dateLen:
		y, m, d, err := parseDateDigits(s)
		if err != nil {
			return time.Time{}
		}
		return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	case dtLocalLen, dtUTCLen:
		if s[8] != 'T' {
			return time.Time{}
		}
		y, m, d, err := parseDateDigits(s[:8])
		if err != nil {
			return time.Time{}
		}
		h, mi, sec, err := parseTimeDigits(s[9:15])
		if err != nil {
			return time.Time{}
		}
		return time.Date(y, time.Month(m), d, h, mi, sec, 0, time.UTC)
	}
	return time.Time{}
}

// buildTimezone строит Timezone из rawComponent.
func buildTimezone(raw *rawComponent) (model.Timezone, error) {
	var tz model.Timezone

	for _, p := range raw.props {
		switch strings.ToUpper(p.Name) {
		case "TZID":
			tz.TZID = p.Value
		default:
			switch {
			case isXProp(p.Name):
				tz.XProps = append(tz.XProps, p)
			case isIanaProp(p.Name):
				tz.IanaProps = append(tz.IanaProps, p)
			}
		}
	}

	for _, child := range raw.children {
		switch child.name {
		case "STANDARD":
			tr, err := buildTzTransition(child)
			if err != nil {
				return tz, errors.Wrap(err, "STANDARD")
			}
			tz.Standards = append(tz.Standards, tr)
		case "DAYLIGHT":
			tr, err := buildTzTransition(child)
			if err != nil {
				return tz, errors.Wrap(err, "DAYLIGHT")
			}
			tz.Daylights = append(tz.Daylights, tr)
		}
	}

	return tz, nil
}

// buildTzTransition строит TzTransition из rawComponent.
func buildTzTransition(raw *rawComponent) (model.TzTransition, error) {
	var tr model.TzTransition

	for _, p := range raw.props {
		switch strings.ToUpper(p.Name) {
		case propDTSTART:
			t, _, err := parseDateTime(p.Value, nil)
			if err != nil {
				return tr, errors.Wrap(err, "DTSTART")
			}
			tr.DTStart = t
		case "TZOFFSETFROM":
			tr.OffsetFrom = p.Value
		case propTZOFFSETTO:
			tr.OffsetTo = p.Value
		case "TZNAME":
			tr.TZName = unescapeText(p.Value)
		case propRRULE:
			rule, err := parseRRule(p.Value, nil)
			if err != nil {
				return tr, errors.Wrap(err, propRRULE)
			}
			tr.RRules = append(tr.RRules, rule)
		case propRDATE:
			dates, err := parseDateTimeList(p.Value, nil)
			if err != nil {
				return tr, errors.Wrap(err, propRDATE)
			}
			tr.RDates = append(tr.RDates, dates...)
		default:
			switch {
			case isXProp(p.Name):
				tr.XProps = append(tr.XProps, p)
			case isIanaProp(p.Name):
				tr.IanaProps = append(tr.IanaProps, p)
			}
		}
	}

	return tr, nil
}

// buildAlarm строит Alarm из rawComponent.
func buildAlarm(raw *rawComponent, tzCtx *perCalendarTZContext) (model.Alarm, error) {
	var a model.Alarm

	for _, p := range raw.props {
		switch strings.ToUpper(p.Name) {
		case propUID:
			a.UID = p.Value
		case "ACTION":
			a.Action = parseAlarmAction(p.Value)
		case "TRIGGER":
			t, err := parseTrigger(p.Params, p.Value)
			if err != nil {
				return a, errors.Wrap(err, "TRIGGER")
			}
			a.Trigger = t
		case propDESCRIPTION:
			a.Description = unescapeText(p.Value)
		case propSUMMARY:
			a.Summary = unescapeText(p.Value)
		case propATTENDEE:
			a.Attendees = append(a.Attendees, parseAttendee(p.Params, p.Value))
		case propDURATION:
			d, err := parseDuration(p.Value)
			if err != nil {
				return a, errors.Wrap(err, propDURATION)
			}
			a.Duration = &d
		case "ACKNOWLEDGED":
			loc := resolveTZIDCtx(p, tzCtx)
			t, _, err := parseDateTime(p.Value, loc)
			if err != nil {
				return a, errors.Wrap(err, "ACKNOWLEDGED")
			}
			a.Acknowledged = &t
		case "PROXIMITY":
			a.Proximity = parseProximity(p.Value)
		case "REPEAT":
			n, err := strconv.Atoi(p.Value)
			if err != nil {
				return a, errors.Wrapf(ErrInvalidRRule, "REPEAT=%s", p.Value)
			}
			a.Repeat = n
		case propRELATEDTO:
			a.Related = append(a.Related, buildRelation(p))
		case propATTACH:
			att, err := parseAttachment(p)
			if err != nil {
				return a, errors.Wrap(err, propATTACH)
			}
			a.Attach = append(a.Attach, att)
		default:
			switch {
			case isXProp(p.Name):
				a.XProps = append(a.XProps, p)
			case isIanaProp(p.Name):
				a.IanaProps = append(a.IanaProps, p)
			}
		}
	}

	for _, child := range raw.children {
		if child.name != componentLocation {
			continue
		}
		loc, err := buildLocationComponent(child)
		if err != nil {
			return a, errors.Wrap(err, componentLocation)
		}
		a.Locations = append(a.Locations, loc)
	}

	return a, nil
}

// buildTodo строит Todo из rawComponent.
func buildTodo(raw *rawComponent, tzCtx *perCalendarTZContext) (model.Todo, error) {
	var t model.Todo

	for _, p := range raw.props {
		if err := setTodoProp(&t, p, tzCtx); err != nil {
			return t, err
		}
	}

	for _, child := range raw.children {
		switch child.name {
		case componentAlarm:
			alarm, err := buildAlarm(child, tzCtx)
			if err != nil {
				return t, errors.Wrap(err, componentAlarm)
			}
			t.Alarms = append(t.Alarms, alarm)
		case componentParticipant:
			part, err := buildParticipant(child)
			if err != nil {
				return t, errors.Wrap(err, componentParticipant)
			}
			t.Participants = append(t.Participants, part)
		case componentLocation:
			locComponent, err := buildLocationComponent(child)
			if err != nil {
				return t, errors.Wrap(err, "LOCATION component")
			}
			t.Locations = append(t.Locations, locComponent)
		case componentResource:
			resComponent, err := buildResourceComponent(child)
			if err != nil {
				return t, errors.Wrap(err, "RESOURCE component")
			}
			t.Resources = append(t.Resources, resComponent)
		}
	}

	return t, nil
}

// setTodoProp устанавливает значение одного свойства в Todo.
//
//nolint:gocyclo // switch по свойствам iCalendar неизбежно содержит много case-ов.
func setTodoProp(t *model.Todo, p model.Property, tzCtx *perCalendarTZContext) error {
	loc := resolveTZIDCtx(p, tzCtx)

	switch strings.ToUpper(p.Name) {
	case propUID:
		t.UID = p.Value
	case propDTSTAMP:
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propDTSTAMP)
		}
		t.DTStamp = v
	case propDTSTART:
		v, isDate, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "DTSTART")
		}
		t.DTStart = &v
		t.AllDay = isDate
	case "DUE":
		v, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "DUE")
		}
		t.Due = &v
	case propDURATION:
		d, err := parseDuration(p.Value)
		if err != nil {
			return errors.Wrap(err, propDURATION)
		}
		t.Duration = &d
	case "COMPLETED":
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, "COMPLETED")
		}
		t.Completed = &v
	case "PRIORITY":
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidValue, "PRIORITY=%s", p.Value)
		}
		t.Priority = n
	case "PERCENT-COMPLETE":
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidValue, "PERCENT-COMPLETE=%s", p.Value)
		}
		t.PercentComplete = n
	case propSUMMARY:
		t.Summary = unescapeText(p.Value)
	case propDESCRIPTION:
		t.Description = unescapeText(p.Value)
	case propLocation:
		t.Location = unescapeText(p.Value)
	case propURL:
		t.URL = p.Value
	case propSTATUS:
		t.Status = parseStatus(p.Value)
	case propCLASS:
		t.Classification = parseClassification(p.Value)
	case propORGANIZER:
		a := parseAttendee(p.Params, p.Value)
		t.Organizer = &a
	case propATTENDEE:
		t.Attendees = append(t.Attendees, parseAttendee(p.Params, p.Value))
	case propCREATED:
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propCREATED)
		}
		t.Created = &v
	case propLASTMODIFIED:
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propLASTMODIFIED)
		}
		t.LastModified = &v
	case propSEQUENCE:
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidRRule, "SEQUENCE=%s", p.Value)
		}
		t.Sequence = n
	case propRECURRENCEID:
		v, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propRECURRENCEID)
		}
		t.RecurrenceID = &v
	case propRRULE:
		rule, err := parseRRule(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propRRULE)
		}
		t.RRules = append(t.RRules, rule)
	case propRDATE:
		if strings.EqualFold(p.ParamValue("VALUE"), "PERIOD") {
			periods, err := parsePeriodList(p.Value, loc)
			if err != nil {
				return errors.Wrap(err, propRDATE)
			}
			t.RDatePeriods = append(t.RDatePeriods, periods...)
		} else {
			dates, err := parseDateTimeList(p.Value, loc)
			if err != nil {
				return errors.Wrap(err, propRDATE)
			}
			t.RDates = append(t.RDates, dates...)
		}
	case propEXDATE:
		dates, err := parseDateTimeList(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propEXDATE)
		}
		t.ExDates = append(t.ExDates, dates...)
	case propRELATEDTO:
		t.Related = append(t.Related, buildRelation(p))
	case propStructuredData:
		t.StructuredData = append(t.StructuredData, parseStructuredData(p))
	case propStyledDescription:
		t.StyledDescriptions = append(t.StyledDescriptions, parseStyledDescription(p))
	case propConcept:
		t.Concepts = append(t.Concepts, p.Value)
	case propRefID:
		t.RefIDs = append(t.RefIDs, p.Value)
	case propLink:
		t.Links = append(t.Links, parseLink(p))
	case propRequestStatus:
		t.RequestStatus = append(t.RequestStatus, parseRequestStatus(p.Value))
	case propATTACH:
		att, err := parseAttachment(p)
		if err != nil {
			return errors.Wrap(err, propATTACH)
		}
		t.Attach = append(t.Attach, att)
	default:
		switch {
		case isXProp(p.Name):
			t.XProps = append(t.XProps, p)
		case isIanaProp(p.Name):
			t.IanaProps = append(t.IanaProps, p)
		}
	}

	return nil
}

// buildJournal строит Journal из rawComponent.
func buildJournal(raw *rawComponent, tzCtx *perCalendarTZContext) (model.Journal, error) {
	var j model.Journal

	for _, p := range raw.props {
		if err := setJournalProp(&j, p, tzCtx); err != nil {
			return j, err
		}
	}

	for _, child := range raw.children {
		switch child.name {
		case componentParticipant:
			part, err := buildParticipant(child)
			if err != nil {
				return j, errors.Wrap(err, componentParticipant)
			}
			j.Participants = append(j.Participants, part)
		case componentLocation:
			locComponent, err := buildLocationComponent(child)
			if err != nil {
				return j, errors.Wrap(err, "LOCATION component")
			}
			j.Locations = append(j.Locations, locComponent)
		case componentResource:
			resComponent, err := buildResourceComponent(child)
			if err != nil {
				return j, errors.Wrap(err, "RESOURCE component")
			}
			j.Resources = append(j.Resources, resComponent)
		}
	}

	return j, nil
}

// setJournalProp устанавливает значение одного свойства в Journal.
//
//nolint:gocyclo // Ветвление следует RFC: упрощение ухудшит читаемость.
func setJournalProp(j *model.Journal, p model.Property, tzCtx *perCalendarTZContext) error {
	loc := resolveTZIDCtx(p, tzCtx)

	switch strings.ToUpper(p.Name) {
	case propUID:
		j.UID = p.Value
	case propDTSTAMP:
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propDTSTAMP)
		}
		j.DTStamp = v
	case propDTSTART:
		v, isDate, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "DTSTART")
		}
		j.DTStart = &v
		j.AllDay = isDate
	case propSUMMARY:
		j.Summary = unescapeText(p.Value)
	case propDESCRIPTION:
		// VJOURNAL допускает несколько DESCRIPTION (RFC 5545 §3.6.3).
		j.Descriptions = append(j.Descriptions, unescapeText(p.Value))
	case propURL:
		j.URL = p.Value
	case propSTATUS:
		j.Status = parseStatus(p.Value)
	case propCLASS:
		j.Classification = parseClassification(p.Value)
	case propORGANIZER:
		a := parseAttendee(p.Params, p.Value)
		j.Organizer = &a
	case propATTENDEE:
		j.Attendees = append(j.Attendees, parseAttendee(p.Params, p.Value))
	case propCREATED:
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propCREATED)
		}
		j.Created = &v
	case propLASTMODIFIED:
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propLASTMODIFIED)
		}
		j.LastModified = &v
	case propSEQUENCE:
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidRRule, "SEQUENCE=%s", p.Value)
		}
		j.Sequence = n
	case propRECURRENCEID:
		v, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propRECURRENCEID)
		}
		j.RecurrenceID = &v
	case propRRULE:
		rule, err := parseRRule(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propRRULE)
		}
		j.RRules = append(j.RRules, rule)
	case propRDATE:
		if strings.EqualFold(p.ParamValue("VALUE"), "PERIOD") {
			periods, err := parsePeriodList(p.Value, loc)
			if err != nil {
				return errors.Wrap(err, propRDATE)
			}
			j.RDatePeriods = append(j.RDatePeriods, periods...)
		} else {
			dates, err := parseDateTimeList(p.Value, loc)
			if err != nil {
				return errors.Wrap(err, propRDATE)
			}
			j.RDates = append(j.RDates, dates...)
		}
	case propEXDATE:
		dates, err := parseDateTimeList(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propEXDATE)
		}
		j.ExDates = append(j.ExDates, dates...)
	case propRELATEDTO:
		j.Related = append(j.Related, buildRelation(p))
	case propStructuredData:
		j.StructuredData = append(j.StructuredData, parseStructuredData(p))
	case propStyledDescription:
		j.StyledDescriptions = append(j.StyledDescriptions, parseStyledDescription(p))
	case propConcept:
		j.Concepts = append(j.Concepts, p.Value)
	case propRefID:
		j.RefIDs = append(j.RefIDs, p.Value)
	case propLink:
		j.Links = append(j.Links, parseLink(p))
	case propATTACH:
		att, err := parseAttachment(p)
		if err != nil {
			return errors.Wrap(err, propATTACH)
		}
		j.Attach = append(j.Attach, att)
	default:
		switch {
		case isXProp(p.Name):
			j.XProps = append(j.XProps, p)
		case isIanaProp(p.Name):
			j.IanaProps = append(j.IanaProps, p)
		}
	}

	return nil
}

// buildFreeBusy строит FreeBusy из rawComponent.
func buildFreeBusy(raw *rawComponent, tzCtx *perCalendarTZContext) (model.FreeBusy, error) {
	var fb model.FreeBusy

	for _, p := range raw.props {
		loc := resolveTZIDCtx(p, tzCtx)

		switch strings.ToUpper(p.Name) {
		case propUID:
			fb.UID = p.Value
		case propDTSTAMP:
			t, _, err := parseDateTime(p.Value, nil)
			if err != nil {
				return fb, errors.Wrap(err, propDTSTAMP)
			}
			fb.DTStamp = t
		case propDTSTART:
			t, _, err := parseDateTime(p.Value, loc)
			if err != nil {
				return fb, errors.Wrap(err, "DTSTART")
			}
			fb.DTStart = &t
		case propDTEND:
			t, _, err := parseDateTime(p.Value, loc)
			if err != nil {
				return fb, errors.Wrap(err, propDTEND)
			}
			fb.DTEnd = &t
		case propORGANIZER:
			a := parseAttendee(p.Params, p.Value)
			fb.Organizer = &a
		case propATTENDEE:
			fb.Attendees = append(fb.Attendees, parseAttendee(p.Params, p.Value))
		case propURL:
			fb.URL = p.Value
		case "FREEBUSY":
			fbType := model.FBTypeBusy // по умолчанию BUSY
			for _, param := range p.Params {
				if strings.EqualFold(param.Name, "FBTYPE") && len(param.Values) > 0 {
					fbType = parseFreeBusyType(param.Values[0])
				}
			}
			// Значение может содержать несколько периодов через запятую.
			for _, periodStr := range strings.Split(p.Value, ",") {
				period, err := parsePeriod(strings.TrimSpace(periodStr), loc)
				if err != nil {
					return fb, errors.Wrap(err, "FREEBUSY")
				}
				fb.Periods = append(fb.Periods, model.FreeBusyPeriod{
					Type:   fbType,
					Period: period,
				})
			}
		default:
			switch {
			case isXProp(p.Name):
				fb.XProps = append(fb.XProps, p)
			case isIanaProp(p.Name):
				fb.IanaProps = append(fb.IanaProps, p)
			}
		}
	}

	return fb, nil
}

// buildAvailability строит Availability из rawComponent.
func buildAvailability(raw *rawComponent, tzCtx *perCalendarTZContext) (model.Availability, error) {
	var a model.Availability

	for _, p := range raw.props {
		if err := setAvailabilityProp(&a, p, tzCtx); err != nil {
			return a, err
		}
	}

	for _, child := range raw.children {
		switch child.name {
		case "AVAILABLE":
			av, err := buildAvailable(child, tzCtx)
			if err != nil {
				return a, errors.Wrap(err, "AVAILABLE")
			}
			a.Available = append(a.Available, av)
		case componentAlarm:
			alarm, err := buildAlarm(child, tzCtx)
			if err != nil {
				return a, errors.Wrap(err, componentAlarm)
			}
			a.Alarms = append(a.Alarms, alarm)
		}
	}

	return a, nil
}

// setAvailabilityProp устанавливает значение одного свойства в Availability.
func setAvailabilityProp(a *model.Availability, p model.Property, tzCtx *perCalendarTZContext) error {
	loc := resolveTZIDCtx(p, tzCtx)

	switch strings.ToUpper(p.Name) {
	case propUID:
		a.UID = p.Value
	case propDTSTAMP:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propDTSTAMP)
		}
		a.DTStamp = t
	case propDTSTART:
		t, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "DTSTART")
		}
		a.DTStart = &t
	case propDTEND:
		t, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propDTEND)
		}
		if a.Duration != nil {
			a.Duration = nil
		}
		a.DTEnd = &t
	case propDURATION:
		d, err := parseDuration(p.Value)
		if err != nil {
			return errors.Wrap(err, propDURATION)
		}
		if a.DTEnd != nil {
			a.DTEnd = nil
		}
		a.Duration = &d
	case "BUSYTYPE":
		a.BusyType = model.BusyType(strings.ToUpper(p.Value))
	case propCREATED:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propCREATED)
		}
		a.Created = &t
	case propLASTMODIFIED:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propLASTMODIFIED)
		}
		a.LastModified = &t
	case propSEQUENCE:
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidValue, "SEQUENCE=%s", p.Value)
		}
		a.Sequence = n
	case propSUMMARY:
		a.Summary = unescapeText(p.Value)
	case propDESCRIPTION:
		a.Description = unescapeText(p.Value)
	case propURL:
		a.URL = p.Value
	case "CATEGORIES":
		cats := strings.Split(p.Value, ",")
		for i := range cats {
			cats[i] = strings.TrimSpace(cats[i])
		}
		a.Categories = append(a.Categories, cats...)
	case propORGANIZER:
		org := parseAttendee(p.Params, p.Value)
		a.Organizer = &org
	default:
		switch {
		case isXProp(p.Name):
			a.XProps = append(a.XProps, p)
		case isIanaProp(p.Name):
			a.IanaProps = append(a.IanaProps, p)
		}
	}

	return nil
}

// buildAvailable строит Available из rawComponent.
func buildAvailable(raw *rawComponent, tzCtx *perCalendarTZContext) (model.Available, error) {
	var a model.Available

	for _, p := range raw.props {
		if err := setAvailableProp(&a, p, tzCtx); err != nil {
			return a, err
		}
	}

	for _, child := range raw.children {
		if child.name != componentAlarm {
			continue
		}
		alarm, err := buildAlarm(child, tzCtx)
		if err != nil {
			return a, errors.Wrap(err, componentAlarm)
		}
		a.Alarms = append(a.Alarms, alarm)
	}

	return a, nil
}

// setAvailableProp устанавливает значение одного свойства в Available.
func setAvailableProp(a *model.Available, p model.Property, tzCtx *perCalendarTZContext) error {
	loc := resolveTZIDCtx(p, tzCtx)

	switch strings.ToUpper(p.Name) {
	case propUID:
		a.UID = p.Value
	case propDTSTAMP:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propDTSTAMP)
		}
		a.DTStamp = t
	case propDTSTART:
		t, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "DTSTART")
		}
		a.DTStart = t
	case propDTEND:
		t, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propDTEND)
		}
		if a.Duration != nil {
			a.Duration = nil
		}
		a.DTEnd = &t
	case propDURATION:
		d, err := parseDuration(p.Value)
		if err != nil {
			return errors.Wrap(err, propDURATION)
		}
		if a.DTEnd != nil {
			a.DTEnd = nil
		}
		a.Duration = &d
	case propSUMMARY:
		a.Summary = unescapeText(p.Value)
	case propDESCRIPTION:
		a.Description = unescapeText(p.Value)
	case propCREATED:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propCREATED)
		}
		a.Created = &t
	case propLASTMODIFIED:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propLASTMODIFIED)
		}
		a.LastModified = &t
	case propSEQUENCE:
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidValue, "SEQUENCE=%s", p.Value)
		}
		a.Sequence = n
	case propRECURRENCEID:
		t, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propRECURRENCEID)
		}
		a.RecurrenceID = &t
	case propRRULE:
		rule, err := parseRRule(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propRRULE)
		}
		a.RRules = append(a.RRules, rule)
	case propRDATE:
		dates, err := parseDateTimeList(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propRDATE)
		}
		a.RDates = append(a.RDates, dates...)
	case propEXDATE:
		dates, err := parseDateTimeList(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, propEXDATE)
		}
		a.ExDates = append(a.ExDates, dates...)
	default:
		switch {
		case isXProp(p.Name):
			a.XProps = append(a.XProps, p)
		case isIanaProp(p.Name):
			a.IanaProps = append(a.IanaProps, p)
		}
	}

	return nil
}

// --------------------------------------------------------------------------
// Вспомогательные функции
// --------------------------------------------------------------------------

// --------------------------------------------------------------------------
// Per-parse timezone context
// --------------------------------------------------------------------------

// perCalendarTZContext хранит информацию о часовых поясах для одного Parse-вызова.
// Изолирует кастомные VTIMEZONE от других параллельных Parse-вызовов (P1).
type perCalendarTZContext struct {
	// iana: TZID → *time.Location для IANA-известных зон из VTIMEZONE.
	iana map[string]*time.Location
	// custom: TZID → *vtimezoneTransitions для кастомных зон.
	custom map[string]*vtimezoneTransitions
}

func newTZContext() *perCalendarTZContext {
	return &perCalendarTZContext{
		iana:   make(map[string]*time.Location),
		custom: make(map[string]*vtimezoneTransitions),
	}
}

// resolveTZIDCtx извлекает *time.Location из параметра TZID свойства p.
// Для кастомных VTIMEZONE выбирает смещение по wall-clock времени события,
// учитывая переходы STANDARD/DAYLIGHT (P2 улучшенный fallback).
// Если TZID отсутствует или не найден — возвращает nil (будет использован UTC).
func resolveTZIDCtx(p model.Property, tzCtx *perCalendarTZContext) *time.Location {
	tzid := p.ParamValue("TZID")
	if tzid == "" {
		return nil
	}

	// 1. Per-parse IANA map (приоритет — calendar-declared VTIMEZONE).
	if loc, ok := tzCtx.iana[tzid]; ok {
		return loc
	}

	// 2. Кастомный VTIMEZONE с transition-aware выбором смещения.
	if vtz, ok := tzCtx.custom[tzid]; ok {
		return vtz.locationAt(wallClockApprox(p.Value))
	}

	// 3. TZID не объявлен в VTIMEZONE — пробуем системную базу напрямую.
	loc, err := loadIANA(tzid)
	if err != nil {
		return nil
	}
	return loc
}

// buildRelation строит Relation из Property (RELATED-TO).
func buildRelation(p model.Property) model.Relation {
	rel := model.Relation{
		UID: p.Value,
	}
	relType := p.ParamValue("RELTYPE")
	if relType != "" {
		rel.Type = parseRelType(relType)
	}
	gap := p.ParamValue("GAP")
	if gap != "" {
		if d, err := parseDuration(gap); err == nil {
			rel.Gap = &d
		}
	}
	return rel
}

// parseDateTimeList парсит список дат через запятую.
func parseDateTimeList(s string, loc *time.Location) ([]time.Time, error) {
	parts := strings.Split(s, ",")
	result := make([]time.Time, 0, len(parts))
	for _, part := range parts {
		t, _, err := parseDateTime(strings.TrimSpace(part), loc)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

// parseAttachment разбирает свойство ATTACH (RFC 5545 §3.8.1.1).
// При ENCODING=BASE64 декодирует встроенные данные; иначе сохраняет значение как URI.
func parseAttachment(p model.Property) (model.Attachment, error) {
	a := model.Attachment{
		MIMEType: p.ParamValue("FMTTYPE"),
	}
	if strings.EqualFold(p.ParamValue("ENCODING"), "BASE64") {
		// Удаляем пробелы, возможные после line-unfolding.
		raw := strings.ReplaceAll(p.Value, " ", "")
		data, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return a, errors.Wrap(err, "ATTACH BASE64")
		}
		a.Data = data
	} else {
		a.URI = p.Value
	}
	return a, nil
}

// isXProp проверяет, является ли имя свойства нестандартным (X-*).
func isXProp(name string) bool {
	return len(name) > 2 && (name[0] == 'X' || name[0] == 'x') && name[1] == '-'
}

// isIanaProp проверяет, является ли имя свойства потенциальным IANA-расширением.
//
// IANA-свойства не имеют префикса X-, но остаются неизвестными парсеру
// (не обрабатываются в фиксированных case-ветках).
// Пример: COLOR, IMAGE, CONFERENCE (RFC 7986).
func isIanaProp(name string) bool {
	// Имя должно быть непустым и не являться X-свойством.
	return name != "" && !isXProp(name)
}
