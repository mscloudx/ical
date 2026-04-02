// Package encoder реализует сериализацию iCalendar-моделей в формат .ics (RFC 5545).
//
// Предоставляет три точки входа (зеркально пакету parser):
//   - Encode(io.Writer, *Calendar)   — потоковая запись
//   - Marshal(*Calendar) ([]byte)    — в срез байт
//   - MarshalString(*Calendar) string — в строку
package encoder

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mscloudx/ical/model"
)

// maxLineLen — максимальная длина content-line до фолдинга (RFC 5545 §3.1).
const maxLineLen = 75

// --------------------------------------------------------------------------
// Публичный API
// --------------------------------------------------------------------------

// Encode записывает Calendar в io.Writer в формате iCalendar (RFC 5545).
func Encode(w io.Writer, cal *model.Calendar) error {
	if w == nil {
		return ErrNilWriter
	}
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
// Разрыв выбирается на границе UTF-8 символа, чтобы не разрезать
// многобайтный символ (кириллица, emoji и т.д.).
func (w *icsWriter) writeFolded(line string) {
	if w.err != nil {
		return
	}
	b := []byte(line)
	// Первая строка: до maxLineLen октетов.
	// Continuation-строки: до maxLineLen-1 октетов контента, потому что ведущий
	// пробел (fold-индикатор, RFC 5545 §3.1) занимает 1 из 75 допустимых октетов.
	limit := maxLineLen
	for len(b) > limit {
		// Находим безопасную границу UTF-8 символа ≤ limit байт.
		// utf8.RuneStart(b) истинно для ASCII (<0x80) или начального байта
		// многобайтного символа (≥0xC0). Откат максимум на 3 байта.
		cut := limit
		for cut > 0 && !utf8.RuneStart(b[cut]) {
			cut--
		}
		_, w.err = w.w.Write(b[:cut])
		if w.err != nil {
			return
		}
		_, w.err = w.w.Write([]byte("\r\n "))
		if w.err != nil {
			return
		}
		b = b[cut:]
		limit = maxLineLen - 1
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
			encoded := encodeParamValueRFC6868(v)
			// Если значение содержит спецсимволы — оборачиваем в кавычки.
			if needsQuoting(encoded) {
				sb.WriteByte('"')
				sb.WriteString(encoded)
				sb.WriteByte('"')
			} else {
				sb.WriteString(encoded)
			}
		}
	}
	sb.WriteByte(':')
	sb.WriteString(p.Value)
	w.writeLine(sb.String())
}

// writePropStr записывает простое свойство NAME:VALUE (без экранирования).
// Используется для URI, перечислений, дат и других не-TEXT типов.
func (w *icsWriter) writePropStr(name, value string) {
	if value == "" {
		return
	}
	w.writeLine(name + ":" + value)
}

// writePropText записывает TEXT-свойство NAME:VALUE с экранированием (RFC 5545 §3.3.11).
// Используется для SUMMARY, DESCRIPTION, LOCATION, NAME и аналогичных полей.
func (w *icsWriter) writePropText(name, value string) {
	if value == "" {
		return
	}
	w.writeLine(name + ":" + escapeText(value))
}

// writePropInt записывает числовое свойство NAME:VALUE.
func (w *icsWriter) writePropInt(name string, value int) {
	w.writeLine(fmt.Sprintf("%s:%d", name, value))
}

// writeDTStamp записывает DTSTAMP, только если значение не нулевое.
func (w *icsWriter) writeDTStamp(t time.Time) {
	if !t.IsZero() {
		w.writePropStr("DTSTAMP", formatDateTime(t, false))
	}
}

// writePropDateTime записывает свойство даты/времени с поддержкой TZID.
//
// Если allDay=true — формат VALUE=DATE без TZID.
// Если время в UTC (loc==time.UTC) — формат с суффиксом 'Z' без TZID.
// Если loc — именованный не-UTC пояс — добавляет параметр TZID и локальный формат.
func (w *icsWriter) writePropDateTime(name string, t time.Time, allDay bool) {
	if allDay {
		w.writeLine(name + ";VALUE=DATE:" + t.Format("20060102"))
		return
	}
	loc := t.Location()
	if loc == nil || loc == time.UTC {
		w.writePropStr(name, t.Format("20060102T150405Z"))
		return
	}
	tzName := loc.String()
	if tzName == "UTC" {
		w.writePropStr(name, t.Format("20060102T150405Z"))
		return
	}
	// Именованный не-UTC пояс: DTSTART;TZID=America/New_York:20231025T090000
	w.writeLine(name + ";TZID=" + tzName + ":" + t.Format("20060102T150405"))
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
	w.writePropStr("METHOD", cal.Method.String())

	w.writePropText("NAME", cal.Name)
	w.writePropText("DESCRIPTION", cal.Description)
	w.writePropStr("UID", cal.UID)
	if cal.LastModified != nil {
		w.writePropStr("LAST-MODIFIED", formatDateTime(*cal.LastModified, false))
	}
	w.writePropStr("URL", cal.URL)
	if len(cal.Categories) > 0 {
		w.writePropStr("CATEGORIES", strings.Join(cal.Categories, ","))
	}
	if cal.RefreshInterval != nil {
		w.writePropStr("REFRESH-INTERVAL", formatDuration(*cal.RefreshInterval))
	}
	w.writePropStr("COLOR", cal.Color)
	w.writePropStr("SOURCE", cal.Source)

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
	for i := range cal.Availabilities {
		w.writeAvailability(&cal.Availabilities[i])
	}

	w.writeEnd("VCALENDAR")
}

// writeEvent записывает VEVENT.
func (w *icsWriter) writeEvent(e *model.Event) {
	w.writeBegin("VEVENT")

	w.writePropStr("UID", e.UID)
	w.writeDTStamp(e.DTStamp)
	w.writePropDateTime("DTSTART", e.DTStart, e.AllDay)

	if e.DTEnd != nil {
		w.writePropDateTime("DTEND", *e.DTEnd, e.AllDay)
	}
	if e.Duration != nil {
		w.writePropStr("DURATION", formatDuration(*e.Duration))
	}

	w.writePropText("SUMMARY", e.Summary)
	w.writePropText("DESCRIPTION", e.Description)
	w.writePropText("LOCATION", e.Location)
	w.writePropStr("URL", e.URL)
	w.writeAttachments(e.Attach)

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
		w.writePropDateTime("RECURRENCE-ID", *e.RecurrenceID, false)
	}

	// Recurrence.
	for i := range e.RRules {
		w.writePropStr("RRULE", formatRRule(&e.RRules[i]))
	}
	w.writeDateTimeList("RDATE", e.RDates)
	w.writePeriodList(e.RDatePeriods)
	w.writeDateTimeList("EXDATE", e.ExDates)

	// Relations.
	for _, rel := range e.Related {
		w.writeProperty(formatRelation(rel))
	}
	for i := range e.Concepts {
		w.writePropStr("CONCEPT", e.Concepts[i])
	}
	for i := range e.RefIDs {
		w.writePropStr("REFID", e.RefIDs[i])
	}
	for i := range e.Links {
		w.writeLink(&e.Links[i])
	}
	for i := range e.RequestStatus {
		w.writePropStr("REQUEST-STATUS", formatRequestStatus(e.RequestStatus[i]))
	}

	// Alarms.
	for i := range e.Alarms {
		w.writeAlarm(&e.Alarms[i])
	}

	for i := range e.StyledDescriptions {
		w.writeStyledDescription(&e.StyledDescriptions[i])
	}
	for i := range e.StructuredData {
		w.writeStructuredData(&e.StructuredData[i])
	}
	for i := range e.Participants {
		w.writeParticipant(&e.Participants[i])
	}
	for i := range e.Locations {
		w.writeLocationComponent(&e.Locations[i])
	}
	for i := range e.Resources {
		w.writeResourceComponent(&e.Resources[i])
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
	w.writeDTStamp(t.DTStamp)

	if t.DTStart != nil {
		w.writePropDateTime("DTSTART", *t.DTStart, t.AllDay)
	}
	if t.Due != nil {
		w.writePropDateTime("DUE", *t.Due, model.IsDateOnly(*t.Due))
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

	w.writePropText("SUMMARY", t.Summary)
	w.writePropText("DESCRIPTION", t.Description)
	w.writePropText("LOCATION", t.Location)
	w.writePropStr("URL", t.URL)
	w.writeAttachments(t.Attach)

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
		w.writePropDateTime("RECURRENCE-ID", *t.RecurrenceID, false)
	}

	for i := range t.RRules {
		w.writePropStr("RRULE", formatRRule(&t.RRules[i]))
	}
	w.writeDateTimeList("RDATE", t.RDates)
	w.writePeriodList(t.RDatePeriods)
	w.writeDateTimeList("EXDATE", t.ExDates)

	for _, rel := range t.Related {
		w.writeProperty(formatRelation(rel))
	}
	for i := range t.Concepts {
		w.writePropStr("CONCEPT", t.Concepts[i])
	}
	for i := range t.RefIDs {
		w.writePropStr("REFID", t.RefIDs[i])
	}
	for i := range t.Links {
		w.writeLink(&t.Links[i])
	}
	for i := range t.RequestStatus {
		w.writePropStr("REQUEST-STATUS", formatRequestStatus(t.RequestStatus[i]))
	}

	for i := range t.Alarms {
		w.writeAlarm(&t.Alarms[i])
	}

	for i := range t.StyledDescriptions {
		w.writeStyledDescription(&t.StyledDescriptions[i])
	}
	for i := range t.StructuredData {
		w.writeStructuredData(&t.StructuredData[i])
	}
	for i := range t.Participants {
		w.writeParticipant(&t.Participants[i])
	}
	for i := range t.Locations {
		w.writeLocationComponent(&t.Locations[i])
	}
	for i := range t.Resources {
		w.writeResourceComponent(&t.Resources[i])
	}

	w.writeProperties(t.XProps)
	w.writeProperties(t.IanaProps)

	w.writeEnd("VTODO")
}

// writeJournal записывает VJOURNAL.
func (w *icsWriter) writeJournal(j *model.Journal) {
	w.writeBegin("VJOURNAL")

	w.writePropStr("UID", j.UID)
	w.writeDTStamp(j.DTStamp)

	if j.DTStart != nil {
		w.writePropDateTime("DTSTART", *j.DTStart, j.AllDay)
	}

	w.writePropText("SUMMARY", j.Summary)

	// VJOURNAL допускает несколько DESCRIPTION.
	for _, desc := range j.Descriptions {
		w.writePropText("DESCRIPTION", desc)
	}

	w.writePropStr("URL", j.URL)
	w.writeAttachments(j.Attach)

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
	w.writePeriodList(j.RDatePeriods)
	w.writeDateTimeList("EXDATE", j.ExDates)

	for _, rel := range j.Related {
		w.writeProperty(formatRelation(rel))
	}
	for i := range j.Concepts {
		w.writePropStr("CONCEPT", j.Concepts[i])
	}
	for i := range j.RefIDs {
		w.writePropStr("REFID", j.RefIDs[i])
	}
	for i := range j.Links {
		w.writeLink(&j.Links[i])
	}

	for i := range j.StyledDescriptions {
		w.writeStyledDescription(&j.StyledDescriptions[i])
	}
	for i := range j.StructuredData {
		w.writeStructuredData(&j.StructuredData[i])
	}
	for i := range j.Participants {
		w.writeParticipant(&j.Participants[i])
	}
	for i := range j.Locations {
		w.writeLocationComponent(&j.Locations[i])
	}
	for i := range j.Resources {
		w.writeResourceComponent(&j.Resources[i])
	}

	w.writeProperties(j.XProps)
	w.writeProperties(j.IanaProps)

	w.writeEnd("VJOURNAL")
}

// writeFreeBusy записывает VFREEBUSY.
func (w *icsWriter) writeFreeBusy(fb *model.FreeBusy) {
	w.writeBegin("VFREEBUSY")

	w.writePropStr("UID", fb.UID)
	w.writeDTStamp(fb.DTStamp)

	if fb.DTStart != nil {
		w.writePropDateTime("DTSTART", *fb.DTStart, false)
	}
	if fb.DTEnd != nil {
		w.writePropDateTime("DTEND", *fb.DTEnd, false)
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

// writeAvailability записывает VAVAILABILITY (RFC 7953).
func (w *icsWriter) writeAvailability(a *model.Availability) {
	w.writeBegin("VAVAILABILITY")

	w.writePropStr("UID", a.UID)
	w.writeDTStamp(a.DTStamp)

	if a.DTStart != nil {
		w.writePropDateTime("DTSTART", *a.DTStart, false)
	}

	if a.DTEnd != nil && a.Duration != nil {
		w.err = fmt.Errorf("DTEND and DURATION must not be used simultaneously in VAVAILABILITY")
		return
	}

	if a.DTEnd != nil {
		w.writePropDateTime("DTEND", *a.DTEnd, false)
	} else if a.Duration != nil {
		w.writePropStr("DURATION", formatDuration(*a.Duration))
	}

	if s := a.BusyType.String(); s != "" {
		w.writePropStr("BUSYTYPE", s)
	}

	if a.Created != nil {
		w.writePropStr("CREATED", formatDateTime(*a.Created, false))
	}
	if a.LastModified != nil {
		w.writePropStr("LAST-MODIFIED", formatDateTime(*a.LastModified, false))
	}
	if a.Sequence != 0 {
		w.writePropInt("SEQUENCE", a.Sequence)
	}

	w.writePropText("SUMMARY", a.Summary)
	w.writePropText("DESCRIPTION", a.Description)
	w.writePropStr("URL", a.URL)

	if len(a.Categories) > 0 {
		w.writePropStr("CATEGORIES", strings.Join(a.Categories, ","))
	}

	if a.Organizer != nil {
		w.writeProperty(formatAttendee("ORGANIZER", a.Organizer))
	}

	for i := range a.Available {
		w.writeAvailable(&a.Available[i])
	}
	for i := range a.Alarms {
		w.writeAlarm(&a.Alarms[i])
	}

	w.writeProperties(a.XProps)
	w.writeProperties(a.IanaProps)

	w.writeEnd("VAVAILABILITY")
}

// writeAvailable записывает AVAILABLE (RFC 7953).
func (w *icsWriter) writeAvailable(a *model.Available) {
	w.writeBegin("AVAILABLE")

	w.writePropStr("UID", a.UID)
	w.writeDTStamp(a.DTStamp)
	w.writePropDateTime("DTSTART", a.DTStart, false)

	if a.DTEnd != nil {
		w.writePropDateTime("DTEND", *a.DTEnd, false)
	} else if a.Duration != nil {
		w.writePropStr("DURATION", formatDuration(*a.Duration))
	}

	w.writePropText("SUMMARY", a.Summary)
	w.writePropText("DESCRIPTION", a.Description)

	if a.Created != nil {
		w.writePropStr("CREATED", formatDateTime(*a.Created, false))
	}
	if a.LastModified != nil {
		w.writePropStr("LAST-MODIFIED", formatDateTime(*a.LastModified, false))
	}
	if a.Sequence != 0 {
		w.writePropInt("SEQUENCE", a.Sequence)
	}
	if a.RecurrenceID != nil {
		w.writePropStr("RECURRENCE-ID", formatDateTime(*a.RecurrenceID, false))
	}

	for i := range a.RRules {
		w.writePropStr("RRULE", formatRRule(&a.RRules[i]))
	}
	w.writeDateTimeList("RDATE", a.RDates)
	w.writeDateTimeList("EXDATE", a.ExDates)

	for i := range a.Alarms {
		w.writeAlarm(&a.Alarms[i])
	}

	w.writeProperties(a.XProps)
	w.writeProperties(a.IanaProps)

	w.writeEnd("AVAILABLE")
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
	w.writePropText("TZNAME", tr.TZName)

	for i := range tr.RRules {
		w.writePropStr("RRULE", formatRRule(&tr.RRules[i]))
	}
	// VTIMEZONE RDATEs must be local time (no 'Z'), per RFC 5545 §3.6.5.
	if len(tr.RDates) > 0 {
		parts := make([]string, len(tr.RDates))
		for i, d := range tr.RDates {
			parts[i] = formatDateTimeLocal(d)
		}
		w.writePropStr("RDATE", strings.Join(parts, ","))
	}

	w.writeProperties(tr.XProps)
	w.writeProperties(tr.IanaProps)

	w.writeEnd(name)
}

// writeAlarm записывает VALARM.
func (w *icsWriter) writeAlarm(a *model.Alarm) {
	w.writeBegin("VALARM")

	w.writePropStr("UID", a.UID)

	if s := a.Action.String(); s != "" {
		w.writePropStr("ACTION", s)
	}

	w.writeProperty(formatTrigger(a.Trigger))
	for _, rel := range a.Related {
		w.writeProperty(formatRelation(rel))
	}
	w.writePropText("DESCRIPTION", a.Description)
	w.writePropText("SUMMARY", a.Summary)

	for i := range a.Attendees {
		w.writeProperty(formatAttendee("ATTENDEE", &a.Attendees[i]))
	}
	w.writeAttachments(a.Attach)

	if a.Duration != nil {
		w.writePropStr("DURATION", formatDuration(*a.Duration))
	}
	if a.Repeat != 0 {
		w.writePropInt("REPEAT", a.Repeat)
	}
	if a.Acknowledged != nil {
		w.writePropStr("ACKNOWLEDGED", formatDateTime(*a.Acknowledged, false))
	}
	if s := a.Proximity.String(); s != "" {
		w.writePropStr("PROXIMITY", s)
	}

	w.writeProperties(a.XProps)
	w.writeProperties(a.IanaProps)

	for i := range a.Locations {
		w.writeLocationComponent(&a.Locations[i])
	}

	w.writeEnd("VALARM")
}

// --------------------------------------------------------------------------
// Вспомогательные методы
// --------------------------------------------------------------------------

// writeAttachments записывает список вложений ATTACH.
func (w *icsWriter) writeAttachments(attachments []model.Attachment) {
	for i := range attachments {
		w.writeAttachment(&attachments[i])
	}
}

// writeAttachment записывает одно свойство ATTACH (RFC 5545 §3.8.1.1).
// При наличии Data — кодирует в BASE64 (ENCODING=BASE64;VALUE=BINARY).
// При наличии URI — записывает как обычный URI.
func (w *icsWriter) writeAttachment(a *model.Attachment) {
	if w.err != nil {
		return
	}
	var sb strings.Builder
	sb.WriteString("ATTACH")
	if a.MIMEType != "" {
		sb.WriteString(";FMTTYPE=")
		sb.WriteString(a.MIMEType)
	}
	for _, param := range a.Params {
		sb.WriteByte(';')
		sb.WriteString(param.Name)
		sb.WriteByte('=')
		for i, v := range param.Values {
			if i > 0 {
				sb.WriteByte(',')
			}
			if needsQuoting(v) {
				sb.WriteByte('"')
				sb.WriteString(v)
				sb.WriteByte('"')
			} else {
				sb.WriteString(v)
			}
		}
	}
	if len(a.Data) > 0 {
		sb.WriteString(";ENCODING=BASE64;VALUE=BINARY:")
		sb.WriteString(base64.StdEncoding.EncodeToString(a.Data))
	} else {
		sb.WriteByte(':')
		sb.WriteString(a.URI)
	}
	w.writeLine(sb.String())
}

// writeProperties записывает список свойств.
func (w *icsWriter) writeProperties(props []model.Property) {
	for _, p := range props {
		w.writeProperty(p)
	}
}

// writeDateTimeList записывает список дат через запятую.
// Если даты имеют не-UTC именованный часовой пояс — добавляет параметр TZID.
// Все даты в списке должны иметь одинаковый часовой пояс (RFC 5545 §3.8.5.2).
func (w *icsWriter) writeDateTimeList(name string, dates []time.Time) {
	if len(dates) == 0 {
		return
	}
	// Определяем часовой пояс по первой дате.
	loc := dates[0].Location()
	if loc != nil && loc != time.UTC && loc.String() != "UTC" {
		// Именованный не-UTC пояс: NAME;TZID=<tz>:date1,date2,...
		parts := make([]string, len(dates))
		for i, d := range dates {
			parts[i] = d.Format("20060102T150405")
		}
		w.writeLine(name + ";TZID=" + loc.String() + ":" + strings.Join(parts, ","))
		return
	}
	parts := make([]string, len(dates))
	for i, d := range dates {
		parts[i] = formatDateTime(d, false)
	}
	w.writePropStr(name, strings.Join(parts, ","))
}

// writePeriodList записывает список RDATE;VALUE=PERIOD (RFC 5545 §3.8.5.2).
// Каждый период — отдельная content-line вида "RDATE;VALUE=PERIOD:start/end".
func (w *icsWriter) writePeriodList(periods []model.Period) {
	for i := range periods {
		w.writeLine("RDATE;VALUE=PERIOD:" + formatPeriod(periods[i]))
	}
}

// needsQuoting проверяет, нужно ли оборачивать значение параметра в кавычки.
//
// RFC 5545 §3.1 требует кавычки только при наличии DQUOTE-unsafe символов:
// точка с запятой, двоеточие, запятая. Пробел (WSP) является SAFE-CHAR
// и кавычек не требует.
func needsQuoting(s string) bool {
	return strings.ContainsAny(s, ";:,")
}

// encodeParamValueRFC6868 кодирует параметр в ^-escape (RFC 6868).
// Переносы строк -> ^n, ^ -> ^^, \" -> ^'.
func encodeParamValueRFC6868(value string) string {
	if !strings.ContainsAny(value, "^\n\r\"") {
		return value
	}
	var sb strings.Builder
	sb.Grow(len(value))
	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '^':
			sb.WriteString("^^")
		case '"':
			sb.WriteString("^'")
		case '\r':
			if i+1 < len(value) && value[i+1] == '\n' {
				i++
			}
			sb.WriteString("^n")
		case '\n':
			sb.WriteString("^n")
		default:
			sb.WriteByte(value[i])
		}
	}
	return sb.String()
}
