// Package encoder реализует сериализацию iCalendar-моделей в формат .ics (RFC 5545).
//
// Предоставляет три точки входа (зеркально пакету parser):
//   - Encode(io.Writer, *Calendar)   — потоковая запись
//   - Marshal(*Calendar) ([]byte)    — в срез байт
//   - MarshalString(*Calendar) string — в строку
package encoder

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"gitverse.ru/cloudcoder/ical/model"
)

// maxLineLen — максимальная длина content-line до фолдинга (RFC 5545 §3.1).
const maxLineLen = 75

// --------------------------------------------------------------------------
// Публичный API
// --------------------------------------------------------------------------

// Encode записывает Calendar в io.Writer в формате iCalendar (RFC 5545).
func Encode(w io.Writer, cal *model.Calendar) error {
	if cal == nil {
		return ErrNilCalendar
	}
	e := &icsWriter{w: w}
	e.writeCalendar(cal)
	return e.err
}

// Marshal сериализует Calendar в []byte.
func Marshal(cal *model.Calendar) ([]byte, error) {
	var buf bytes.Buffer
	if err := Encode(&buf, cal); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalString сериализует Calendar в строку.
func MarshalString(cal *model.Calendar) (string, error) {
	var buf bytes.Buffer
	if err := Encode(&buf, cal); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// --------------------------------------------------------------------------
// icsWriter — обёртка над io.Writer с отслеживанием ошибки.
// --------------------------------------------------------------------------

// icsWriter пишет content-lines в io.Writer, отслеживая ошибку.
type icsWriter struct {
	w   io.Writer
	err error
}

// writeLine записывает одну content-line с фолдингом и CRLF.
func (w *icsWriter) writeLine(line string) {
	if w.err != nil {
		return
	}
	w.writeFolded(line)
}

// writeFolded записывает строку с фолдингом (RFC 5545 §3.1):
// строки длиннее 75 октетов разбиваются переносом CRLF + пробел.
func (w *icsWriter) writeFolded(line string) {
	if w.err != nil {
		return
	}
	b := []byte(line)
	for len(b) > maxLineLen {
		_, w.err = w.w.Write(b[:maxLineLen])
		if w.err != nil {
			return
		}
		_, w.err = w.w.Write([]byte("\r\n "))
		if w.err != nil {
			return
		}
		b = b[maxLineLen:]
	}
	_, w.err = w.w.Write(b)
	if w.err != nil {
		return
	}
	_, w.err = w.w.Write([]byte("\r\n"))
}

// writeProperty записывает Property как content-line.
func (w *icsWriter) writeProperty(p model.Property) {
	if w.err != nil {
		return
	}
	var sb strings.Builder
	sb.WriteString(p.Name)
	for _, param := range p.Params {
		sb.WriteByte(';')
		sb.WriteString(param.Name)
		sb.WriteByte('=')
		for i, v := range param.Values {
			if i > 0 {
				sb.WriteByte(',')
			}
			// Если значение содержит спецсимволы — оборачиваем в кавычки.
			if needsQuoting(v) {
				sb.WriteByte('"')
				sb.WriteString(v)
				sb.WriteByte('"')
			} else {
				sb.WriteString(v)
			}
		}
	}
	sb.WriteByte(':')
	sb.WriteString(p.Value)
	w.writeLine(sb.String())
}

// writePropStr записывает простое свойство NAME:VALUE.
func (w *icsWriter) writePropStr(name, value string) {
	if value == "" {
		return
	}
	w.writeLine(name + ":" + value)
}

// writePropInt записывает числовое свойство NAME:VALUE.
func (w *icsWriter) writePropInt(name string, value int) {
	w.writeLine(fmt.Sprintf("%s:%d", name, value))
}

// writeBegin записывает BEGIN:COMPONENT.
func (w *icsWriter) writeBegin(component string) {
	w.writeLine("BEGIN:" + component)
}

// writeEnd записывает END:COMPONENT.
func (w *icsWriter) writeEnd(component string) {
	w.writeLine("END:" + component)
}

// --------------------------------------------------------------------------
// Сериализация компонентов
// --------------------------------------------------------------------------

// writeCalendar записывает полный VCALENDAR.
func (w *icsWriter) writeCalendar(cal *model.Calendar) {
	w.writeBegin("VCALENDAR")

	// Свойства VCALENDAR.
	version := cal.Version
	if version == "" {
		version = "2.0"
	}
	w.writePropStr("VERSION", version)
	w.writePropStr("PRODID", cal.ProdID)
	w.writePropStr("CALSCALE", cal.CalScale)
	w.writePropStr("METHOD", cal.Method)

	// X- и IANA свойства календаря.
	w.writeProperties(cal.XProps)
	w.writeProperties(cal.IanaProps)

	// Вложенные компоненты.
	for i := range cal.Timezones {
		w.writeTimezone(&cal.Timezones[i])
	}
	for i := range cal.Events {
		w.writeEvent(&cal.Events[i])
	}
	for i := range cal.Todos {
		w.writeTodo(&cal.Todos[i])
	}
	for i := range cal.Journals {
		w.writeJournal(&cal.Journals[i])
	}
	for i := range cal.FreeBusys {
		w.writeFreeBusy(&cal.FreeBusys[i])
	}

	w.writeEnd("VCALENDAR")
}

// writeEvent записывает VEVENT.
func (w *icsWriter) writeEvent(e *model.Event) {
	w.writeBegin("VEVENT")

	w.writePropStr("UID", e.UID)
	w.writePropStr("DTSTAMP", formatDateTime(e.DTStamp, false))
	w.writePropStr("DTSTART", formatDateTime(e.DTStart, e.AllDay))

	if e.DTEnd != nil {
		w.writePropStr("DTEND", formatDateTime(*e.DTEnd, false))
	}
	if e.Duration != nil {
		w.writePropStr("DURATION", formatDuration(*e.Duration))
	}

	w.writePropStr("SUMMARY", e.Summary)
	w.writePropStr("DESCRIPTION", e.Description)
	w.writePropStr("LOCATION", e.Location)
	w.writePropStr("URL", e.URL)

	if s := e.Status.String(); s != "" {
		w.writePropStr("STATUS", s)
	}
	if s := e.Transparency.String(); s != "" {
		w.writePropStr("TRANSP", s)
	}
	if s := e.Classification.String(); s != "" {
		w.writePropStr("CLASS", s)
	}

	if e.Organizer != nil {
		w.writeProperty(formatAttendee("ORGANIZER", e.Organizer))
	}
	for i := range e.Attendees {
		w.writeProperty(formatAttendee("ATTENDEE", &e.Attendees[i]))
	}

	if e.Created != nil {
		w.writePropStr("CREATED", formatDateTime(*e.Created, false))
	}
	if e.LastModified != nil {
		w.writePropStr("LAST-MODIFIED", formatDateTime(*e.LastModified, false))
	}
	if e.Sequence != 0 {
		w.writePropInt("SEQUENCE", e.Sequence)
	}
	if e.RecurrenceID != nil {
		w.writePropStr("RECURRENCE-ID", formatDateTime(*e.RecurrenceID, false))
	}

	// Recurrence.
	for i := range e.RRules {
		w.writePropStr("RRULE", formatRRule(&e.RRules[i]))
	}
	w.writeDateTimeList("RDATE", e.RDates)
	w.writeDateTimeList("EXDATE", e.ExDates)

	// Relations.
	for _, rel := range e.Related {
		w.writeProperty(formatRelation(rel))
	}

	// Alarms.
	for i := range e.Alarms {
		w.writeAlarm(&e.Alarms[i])
	}

	// X- и IANA.
	w.writeProperties(e.XProps)
	w.writeProperties(e.IanaProps)

	w.writeEnd("VEVENT")
}

// writeTodo записывает VTODO.
func (w *icsWriter) writeTodo(t *model.Todo) {
	w.writeBegin("VTODO")

	w.writePropStr("UID", t.UID)
	w.writePropStr("DTSTAMP", formatDateTime(t.DTStamp, false))

	if t.DTStart != nil {
		w.writePropStr("DTSTART", formatDateTime(*t.DTStart, t.AllDay))
	}
	if t.Due != nil {
		w.writePropStr("DUE", formatDateTime(*t.Due, model.IsDateOnly(*t.Due)))
	}
	if t.Duration != nil {
		w.writePropStr("DURATION", formatDuration(*t.Duration))
	}
	if t.Completed != nil {
		w.writePropStr("COMPLETED", formatDateTime(*t.Completed, false))
	}

	if t.Priority != 0 {
		w.writePropInt("PRIORITY", t.Priority)
	}
	if t.PercentComplete != 0 {
		w.writePropInt("PERCENT-COMPLETE", t.PercentComplete)
	}

	w.writePropStr("SUMMARY", t.Summary)
	w.writePropStr("DESCRIPTION", t.Description)
	w.writePropStr("LOCATION", t.Location)
	w.writePropStr("URL", t.URL)

	if s := t.Status.String(); s != "" {
		w.writePropStr("STATUS", s)
	}
	if s := t.Classification.String(); s != "" {
		w.writePropStr("CLASS", s)
	}

	if t.Organizer != nil {
		w.writeProperty(formatAttendee("ORGANIZER", t.Organizer))
	}
	for i := range t.Attendees {
		w.writeProperty(formatAttendee("ATTENDEE", &t.Attendees[i]))
	}

	if t.Created != nil {
		w.writePropStr("CREATED", formatDateTime(*t.Created, false))
	}
	if t.LastModified != nil {
		w.writePropStr("LAST-MODIFIED", formatDateTime(*t.LastModified, false))
	}
	if t.Sequence != 0 {
		w.writePropInt("SEQUENCE", t.Sequence)
	}
	if t.RecurrenceID != nil {
		w.writePropStr("RECURRENCE-ID", formatDateTime(*t.RecurrenceID, false))
	}

	for i := range t.RRules {
		w.writePropStr("RRULE", formatRRule(&t.RRules[i]))
	}
	w.writeDateTimeList("RDATE", t.RDates)
	w.writeDateTimeList("EXDATE", t.ExDates)

	for _, rel := range t.Related {
		w.writeProperty(formatRelation(rel))
	}

	for i := range t.Alarms {
		w.writeAlarm(&t.Alarms[i])
	}

	w.writeProperties(t.XProps)
	w.writeProperties(t.IanaProps)

	w.writeEnd("VTODO")
}

// writeJournal записывает VJOURNAL.
func (w *icsWriter) writeJournal(j *model.Journal) {
	w.writeBegin("VJOURNAL")

	w.writePropStr("UID", j.UID)
	w.writePropStr("DTSTAMP", formatDateTime(j.DTStamp, false))

	if j.DTStart != nil {
		w.writePropStr("DTSTART", formatDateTime(*j.DTStart, j.AllDay))
	}

	w.writePropStr("SUMMARY", j.Summary)

	// VJOURNAL допускает несколько DESCRIPTION.
	for _, desc := range j.Descriptions {
		w.writePropStr("DESCRIPTION", desc)
	}

	w.writePropStr("URL", j.URL)

	if s := j.Status.String(); s != "" {
		w.writePropStr("STATUS", s)
	}
	if s := j.Classification.String(); s != "" {
		w.writePropStr("CLASS", s)
	}

	if j.Organizer != nil {
		w.writeProperty(formatAttendee("ORGANIZER", j.Organizer))
	}
	for i := range j.Attendees {
		w.writeProperty(formatAttendee("ATTENDEE", &j.Attendees[i]))
	}

	if j.Created != nil {
		w.writePropStr("CREATED", formatDateTime(*j.Created, false))
	}
	if j.LastModified != nil {
		w.writePropStr("LAST-MODIFIED", formatDateTime(*j.LastModified, false))
	}
	if j.Sequence != 0 {
		w.writePropInt("SEQUENCE", j.Sequence)
	}
	if j.RecurrenceID != nil {
		w.writePropStr("RECURRENCE-ID", formatDateTime(*j.RecurrenceID, false))
	}

	for i := range j.RRules {
		w.writePropStr("RRULE", formatRRule(&j.RRules[i]))
	}
	w.writeDateTimeList("RDATE", j.RDates)
	w.writeDateTimeList("EXDATE", j.ExDates)

	for _, rel := range j.Related {
		w.writeProperty(formatRelation(rel))
	}

	w.writeProperties(j.XProps)
	w.writeProperties(j.IanaProps)

	w.writeEnd("VJOURNAL")
}

// writeFreeBusy записывает VFREEBUSY.
func (w *icsWriter) writeFreeBusy(fb *model.FreeBusy) {
	w.writeBegin("VFREEBUSY")

	w.writePropStr("UID", fb.UID)
	w.writePropStr("DTSTAMP", formatDateTime(fb.DTStamp, false))

	if fb.DTStart != nil {
		w.writePropStr("DTSTART", formatDateTime(*fb.DTStart, false))
	}
	if fb.DTEnd != nil {
		w.writePropStr("DTEND", formatDateTime(*fb.DTEnd, false))
	}

	if fb.Organizer != nil {
		w.writeProperty(formatAttendee("ORGANIZER", fb.Organizer))
	}
	for i := range fb.Attendees {
		w.writeProperty(formatAttendee("ATTENDEE", &fb.Attendees[i]))
	}

	w.writePropStr("URL", fb.URL)

	// Группируем периоды по типу для компактности.
	for _, p := range fb.Periods {
		prop := model.Property{
			Name:  "FREEBUSY",
			Value: formatPeriod(p.Period),
		}
		if s := p.Type.String(); s != "" {
			prop.Params = append(prop.Params, model.Param{
				Name:   "FBTYPE",
				Values: []string{s},
			})
		}
		w.writeProperty(prop)
	}

	w.writeProperties(fb.XProps)
	w.writeProperties(fb.IanaProps)

	w.writeEnd("VFREEBUSY")
}

// writeTimezone записывает VTIMEZONE.
func (w *icsWriter) writeTimezone(tz *model.Timezone) {
	w.writeBegin("VTIMEZONE")

	w.writePropStr("TZID", tz.TZID)

	w.writeProperties(tz.XProps)
	w.writeProperties(tz.IanaProps)

	for i := range tz.Standards {
		w.writeTzTransition("STANDARD", &tz.Standards[i])
	}
	for i := range tz.Daylights {
		w.writeTzTransition("DAYLIGHT", &tz.Daylights[i])
	}

	w.writeEnd("VTIMEZONE")
}

// writeTzTransition записывает STANDARD или DAYLIGHT.
func (w *icsWriter) writeTzTransition(name string, tr *model.TzTransition) {
	w.writeBegin(name)

	w.writePropStr("DTSTART", formatDateTimeLocal(tr.DTStart))
	w.writePropStr("TZOFFSETFROM", tr.OffsetFrom)
	w.writePropStr("TZOFFSETTO", tr.OffsetTo)
	w.writePropStr("TZNAME", tr.TZName)

	for i := range tr.RRules {
		w.writePropStr("RRULE", formatRRule(&tr.RRules[i]))
	}
	w.writeDateTimeList("RDATE", tr.RDates)

	w.writeProperties(tr.XProps)
	w.writeProperties(tr.IanaProps)

	w.writeEnd(name)
}

// writeAlarm записывает VALARM.
func (w *icsWriter) writeAlarm(a *model.Alarm) {
	w.writeBegin("VALARM")

	if s := a.Action.String(); s != "" {
		w.writePropStr("ACTION", s)
	}

	w.writeProperty(formatTrigger(a.Trigger))
	w.writePropStr("DESCRIPTION", a.Description)
	w.writePropStr("SUMMARY", a.Summary)

	for i := range a.Attendees {
		w.writeProperty(formatAttendee("ATTENDEE", &a.Attendees[i]))
	}

	if a.Duration != nil {
		w.writePropStr("DURATION", formatDuration(*a.Duration))
	}
	if a.Repeat != 0 {
		w.writePropInt("REPEAT", a.Repeat)
	}

	w.writeProperties(a.XProps)
	w.writeProperties(a.IanaProps)

	w.writeEnd("VALARM")
}

// --------------------------------------------------------------------------
// Вспомогательные методы
// --------------------------------------------------------------------------

// writeProperties записывает список свойств.
func (w *icsWriter) writeProperties(props []model.Property) {
	for _, p := range props {
		w.writeProperty(p)
	}
}

// writeDateTimeList записывает список дат через запятую.
func (w *icsWriter) writeDateTimeList(name string, dates []time.Time) {
	if len(dates) == 0 {
		return
	}
	parts := make([]string, len(dates))
	for i, d := range dates {
		parts[i] = formatDateTime(d, false)
	}
	w.writePropStr(name, strings.Join(parts, ","))
}

// needsQuoting проверяет, нужно ли оборачивать значение параметра в кавычки.
func needsQuoting(s string) bool {
	return strings.ContainsAny(s, ";:, ")
}
