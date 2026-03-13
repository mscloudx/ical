// Package parser реализует высокопроизводительный парсер iCalendar (RFC 5545).
//
// Предоставляет три точки входа для различных сценариев:
//   - Parse(io.Reader) — для потокового чтения
//   - ParseBytes([]byte) — для данных в памяти
//   - ParseWithClose(io.ReadCloser) — для HTTP-ответов с автоматическим Close
package parser

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitverse.ru/cloudcoder/ical/model"
)

// Константы для часто используемых имён свойств.
const (
	propDTSTART  = "DTSTART"
	propATTENDEE = "ATTENDEE"
)

// --------------------------------------------------------------------------
// Публичный API
// --------------------------------------------------------------------------

// Parse парсит iCalendar-данные из io.Reader и возвращает Calendar.
func Parse(r io.Reader) (*model.Calendar, error) {
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

	return root, nil
}

// --------------------------------------------------------------------------
// Построение типизированных моделей из rawComponent
// --------------------------------------------------------------------------

// buildCalendar строит Calendar из дерева rawComponent.
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
			cal.Method = p.Value
		default:
			switch {
			case isXProp(p.Name):
				cal.XProps = append(cal.XProps, p)
			case isIanaProp(p.Name):
				cal.IanaProps = append(cal.IanaProps, p)
			}
		}
	}

	// Обрабатываем дочерние компоненты.
	for _, child := range vcal.children {
		switch child.name {
		case "VEVENT":
			event, err := buildEvent(child)
			if err != nil {
				return nil, errors.Wrap(err, "VEVENT")
			}
			cal.Events = append(cal.Events, event)
		case "VTODO":
			todo, err := buildTodo(child)
			if err != nil {
				return nil, errors.Wrap(err, "VTODO")
			}
			cal.Todos = append(cal.Todos, todo)
		case "VJOURNAL":
			journal, err := buildJournal(child)
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
		case "VFREEBUSY":
			fb, err := buildFreeBusy(child)
			if err != nil {
				return nil, errors.Wrap(err, "VFREEBUSY")
			}
			cal.FreeBusys = append(cal.FreeBusys, fb)
		}
	}

	return cal, nil
}

// buildEvent строит Event из rawComponent.
func buildEvent(raw *rawComponent) (model.Event, error) {
	var e model.Event

	for _, p := range raw.props {
		if err := setEventProp(&e, p); err != nil {
			return e, err
		}
	}

	// Обрабатываем вложенные VALARM.
	for _, child := range raw.children {
		if child.name == "VALARM" {
			alarm, err := buildAlarm(child)
			if err != nil {
				return e, errors.Wrap(err, "VALARM")
			}
			e.Alarms = append(e.Alarms, alarm)
		}
	}

	return e, nil
}

// setEventProp устанавливает значение одного свойства в Event.
//
//nolint:gocyclo // switch по свойствам iCalendar неизбежно содержит много case-ов.
func setEventProp(e *model.Event, p model.Property) error {
	loc := resolveTZID(p)

	switch strings.ToUpper(p.Name) {
	case "UID":
		e.UID = p.Value
	case "DTSTAMP":
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, "DTSTAMP")
		}
		e.DTStamp = t
	case propDTSTART:
		t, isDate, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "DTSTART")
		}
		e.DTStart = t
		e.AllDay = isDate
	case "DTEND":
		t, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "DTEND")
		}
		e.DTEnd = &t
	case "DURATION":
		d, err := parseDuration(p.Value)
		if err != nil {
			return errors.Wrap(err, "DURATION")
		}
		e.Duration = &d
	case "SUMMARY":
		e.Summary = p.Value
	case "DESCRIPTION":
		e.Description = p.Value
	case "LOCATION":
		e.Location = p.Value
	case "URL":
		e.URL = p.Value
	case "STATUS":
		e.Status = parseStatus(p.Value)
	case "TRANSP":
		e.Transparency = parseTransparency(p.Value)
	case "CLASS":
		e.Classification = parseClassification(p.Value)
	case "ORGANIZER":
		a := parseAttendee(p.Params, p.Value)
		e.Organizer = &a
	case propATTENDEE:
		e.Attendees = append(e.Attendees, parseAttendee(p.Params, p.Value))
	case "CREATED":
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, "CREATED")
		}
		e.Created = &t
	case "LAST-MODIFIED":
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, "LAST-MODIFIED")
		}
		e.LastModified = &t
	case "SEQUENCE":
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidRRule, "SEQUENCE=%s", p.Value)
		}
		e.Sequence = n
	case "RECURRENCE-ID":
		t, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "RECURRENCE-ID")
		}
		e.RecurrenceID = &t
	case "RRULE":
		rule, err := parseRRule(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "RRULE")
		}
		e.RRules = append(e.RRules, rule)
	case "RDATE":
		dates, err := parseDateTimeList(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "RDATE")
		}
		e.RDates = append(e.RDates, dates...)
	case "EXDATE":
		dates, err := parseDateTimeList(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "EXDATE")
		}
		e.ExDates = append(e.ExDates, dates...)
	case "RELATED-TO":
		e.Related = append(e.Related, buildRelation(p))
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
		case "TZOFFSETTO":
			tr.OffsetTo = p.Value
		case "TZNAME":
			tr.TZName = p.Value
		case "RRULE":
			rule, err := parseRRule(p.Value, nil)
			if err != nil {
				return tr, errors.Wrap(err, "RRULE")
			}
			tr.RRules = append(tr.RRules, rule)
		case "RDATE":
			dates, err := parseDateTimeList(p.Value, nil)
			if err != nil {
				return tr, errors.Wrap(err, "RDATE")
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
func buildAlarm(raw *rawComponent) (model.Alarm, error) {
	var a model.Alarm

	for _, p := range raw.props {
		switch strings.ToUpper(p.Name) {
		case "ACTION":
			a.Action = parseAlarmAction(p.Value)
		case "TRIGGER":
			t, err := parseTrigger(p.Params, p.Value)
			if err != nil {
				return a, errors.Wrap(err, "TRIGGER")
			}
			a.Trigger = t
		case "DESCRIPTION":
			a.Description = p.Value
		case "SUMMARY":
			a.Summary = p.Value
		case propATTENDEE:
			a.Attendees = append(a.Attendees, parseAttendee(p.Params, p.Value))
		case "DURATION":
			d, err := parseDuration(p.Value)
			if err != nil {
				return a, errors.Wrap(err, "DURATION")
			}
			a.Duration = &d
		case "REPEAT":
			n, err := strconv.Atoi(p.Value)
			if err != nil {
				return a, errors.Wrapf(ErrInvalidRRule, "REPEAT=%s", p.Value)
			}
			a.Repeat = n
		default:
			switch {
			case isXProp(p.Name):
				a.XProps = append(a.XProps, p)
			case isIanaProp(p.Name):
				a.IanaProps = append(a.IanaProps, p)
			}
		}
	}

	return a, nil
}

// buildTodo строит Todo из rawComponent.
func buildTodo(raw *rawComponent) (model.Todo, error) {
	var t model.Todo

	for _, p := range raw.props {
		if err := setTodoProp(&t, p); err != nil {
			return t, err
		}
	}

	// Обрабатываем вложенные VALARM.
	for _, child := range raw.children {
		if child.name == "VALARM" {
			alarm, err := buildAlarm(child)
			if err != nil {
				return t, errors.Wrap(err, "VALARM")
			}
			t.Alarms = append(t.Alarms, alarm)
		}
	}

	return t, nil
}

// setTodoProp устанавливает значение одного свойства в Todo.
//
//nolint:gocyclo // switch по свойствам iCalendar неизбежно содержит много case-ов.
func setTodoProp(t *model.Todo, p model.Property) error {
	loc := resolveTZID(p)

	switch strings.ToUpper(p.Name) {
	case "UID":
		t.UID = p.Value
	case "DTSTAMP":
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, "DTSTAMP")
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
	case "DURATION":
		d, err := parseDuration(p.Value)
		if err != nil {
			return errors.Wrap(err, "DURATION")
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
	case "SUMMARY":
		t.Summary = p.Value
	case "DESCRIPTION":
		t.Description = p.Value
	case "LOCATION":
		t.Location = p.Value
	case "URL":
		t.URL = p.Value
	case "STATUS":
		t.Status = parseStatus(p.Value)
	case "CLASS":
		t.Classification = parseClassification(p.Value)
	case "ORGANIZER":
		a := parseAttendee(p.Params, p.Value)
		t.Organizer = &a
	case propATTENDEE:
		t.Attendees = append(t.Attendees, parseAttendee(p.Params, p.Value))
	case "CREATED":
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, "CREATED")
		}
		t.Created = &v
	case "LAST-MODIFIED":
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, "LAST-MODIFIED")
		}
		t.LastModified = &v
	case "SEQUENCE":
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidRRule, "SEQUENCE=%s", p.Value)
		}
		t.Sequence = n
	case "RECURRENCE-ID":
		v, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "RECURRENCE-ID")
		}
		t.RecurrenceID = &v
	case "RRULE":
		rule, err := parseRRule(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "RRULE")
		}
		t.RRules = append(t.RRules, rule)
	case "RDATE":
		dates, err := parseDateTimeList(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "RDATE")
		}
		t.RDates = append(t.RDates, dates...)
	case "EXDATE":
		dates, err := parseDateTimeList(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "EXDATE")
		}
		t.ExDates = append(t.ExDates, dates...)
	case "RELATED-TO":
		t.Related = append(t.Related, buildRelation(p))
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
func buildJournal(raw *rawComponent) (model.Journal, error) {
	var j model.Journal

	for _, p := range raw.props {
		if err := setJournalProp(&j, p); err != nil {
			return j, err
		}
	}

	return j, nil
}

// setJournalProp устанавливает значение одного свойства в Journal.
func setJournalProp(j *model.Journal, p model.Property) error {
	loc := resolveTZID(p)

	switch strings.ToUpper(p.Name) {
	case "UID":
		j.UID = p.Value
	case "DTSTAMP":
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, "DTSTAMP")
		}
		j.DTStamp = v
	case propDTSTART:
		v, isDate, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "DTSTART")
		}
		j.DTStart = &v
		j.AllDay = isDate
	case "SUMMARY":
		j.Summary = p.Value
	case "DESCRIPTION":
		// VJOURNAL допускает несколько DESCRIPTION (RFC 5545 §3.6.3).
		j.Descriptions = append(j.Descriptions, p.Value)
	case "URL":
		j.URL = p.Value
	case "STATUS":
		j.Status = parseStatus(p.Value)
	case "CLASS":
		j.Classification = parseClassification(p.Value)
	case "ORGANIZER":
		a := parseAttendee(p.Params, p.Value)
		j.Organizer = &a
	case propATTENDEE:
		j.Attendees = append(j.Attendees, parseAttendee(p.Params, p.Value))
	case "CREATED":
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, "CREATED")
		}
		j.Created = &v
	case "LAST-MODIFIED":
		v, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, "LAST-MODIFIED")
		}
		j.LastModified = &v
	case "SEQUENCE":
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidRRule, "SEQUENCE=%s", p.Value)
		}
		j.Sequence = n
	case "RECURRENCE-ID":
		v, _, err := parseDateTime(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "RECURRENCE-ID")
		}
		j.RecurrenceID = &v
	case "RRULE":
		rule, err := parseRRule(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "RRULE")
		}
		j.RRules = append(j.RRules, rule)
	case "RDATE":
		dates, err := parseDateTimeList(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "RDATE")
		}
		j.RDates = append(j.RDates, dates...)
	case "EXDATE":
		dates, err := parseDateTimeList(p.Value, loc)
		if err != nil {
			return errors.Wrap(err, "EXDATE")
		}
		j.ExDates = append(j.ExDates, dates...)
	case "RELATED-TO":
		j.Related = append(j.Related, buildRelation(p))
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
func buildFreeBusy(raw *rawComponent) (model.FreeBusy, error) {
	var fb model.FreeBusy

	for _, p := range raw.props {
		loc := resolveTZID(p)

		switch strings.ToUpper(p.Name) {
		case "UID":
			fb.UID = p.Value
		case "DTSTAMP":
			t, _, err := parseDateTime(p.Value, nil)
			if err != nil {
				return fb, errors.Wrap(err, "DTSTAMP")
			}
			fb.DTStamp = t
		case propDTSTART:
			t, _, err := parseDateTime(p.Value, loc)
			if err != nil {
				return fb, errors.Wrap(err, "DTSTART")
			}
			fb.DTStart = &t
		case "DTEND":
			t, _, err := parseDateTime(p.Value, loc)
			if err != nil {
				return fb, errors.Wrap(err, "DTEND")
			}
			fb.DTEnd = &t
		case "ORGANIZER":
			a := parseAttendee(p.Params, p.Value)
			fb.Organizer = &a
		case propATTENDEE:
			fb.Attendees = append(fb.Attendees, parseAttendee(p.Params, p.Value))
		case "URL":
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

// --------------------------------------------------------------------------
// Вспомогательные функции
// --------------------------------------------------------------------------

// resolveTZID извлекает *time.Location из параметра TZID.
// Если параметр отсутствует — возвращает nil (будет использован UTC).
func resolveTZID(p model.Property) *time.Location {
	tzid := p.ParamValue("TZID")
	if tzid == "" {
		return nil
	}
	loc, err := globalTZCache.load(tzid)
	if err != nil {
		// Если timezone не найден в системе — используем UTC.
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
