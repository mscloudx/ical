package encoder

import (
	"fmt"
	"strings"
	"time"

	"gitverse.ru/cloudcoder/ical/model"
)

// --------------------------------------------------------------------------
// Форматирование даты и времени
// --------------------------------------------------------------------------

// formatDateTime форматирует time.Time в строку iCalendar.
//
// Если allDay=true — формат VALUE=DATE: "20231024".
// Если время в UTC — формат с суффиксом 'Z': "20231024T120000Z".
// Иначе — локальное время: "20231024T120000".
func formatDateTime(t time.Time, allDay bool) string {
	if allDay {
		return t.Format("20060102")
	}
	if t.Location() == time.UTC {
		return t.Format("20060102T150405Z")
	}
	return t.Format("20060102T150405")
}

// formatDateTimeLocal форматирует время как локальное (без 'Z'),
// используется для DTSTART в VTIMEZONE.
func formatDateTimeLocal(t time.Time) string {
	return t.Format("20060102T150405")
}

// --------------------------------------------------------------------------
// Форматирование DURATION (ISO 8601 подмножество)
// --------------------------------------------------------------------------

// formatDuration форматирует time.Duration в строку iCalendar.
//
// Примеры: PT1H, PT30M, P1DT2H, -PT10M, P1W.
func formatDuration(d time.Duration) string {
	var sb strings.Builder

	if d < 0 {
		sb.WriteByte('-')
		d = -d
	}
	sb.WriteByte('P')

	totalSeconds := int(d.Seconds())

	const (
		secondsPerMinute = 60
		minutesPerHour   = 60
		hoursPerDay      = 24
		daysPerWeek      = 7
		secondsPerHour   = secondsPerMinute * minutesPerHour
		secondsPerDay    = secondsPerHour * hoursPerDay
		secondsPerWeek   = secondsPerDay * daysPerWeek
	)

	weeks := totalSeconds / secondsPerWeek
	if weeks > 0 && totalSeconds%secondsPerWeek == 0 {
		fmt.Fprintf(&sb, "%dW", weeks)
		return sb.String()
	}

	days := totalSeconds / secondsPerDay
	totalSeconds %= secondsPerDay
	hours := totalSeconds / secondsPerHour
	totalSeconds %= secondsPerHour

	minutes := totalSeconds / secondsPerMinute
	seconds := totalSeconds % secondsPerMinute

	if days > 0 {
		fmt.Fprintf(&sb, "%dD", days)
	}

	if hours > 0 || minutes > 0 || seconds > 0 || days == 0 {
		sb.WriteByte('T')
		if hours > 0 {
			fmt.Fprintf(&sb, "%dH", hours)
		}
		if minutes > 0 {
			fmt.Fprintf(&sb, "%dM", minutes)
		}
		if seconds > 0 {
			fmt.Fprintf(&sb, "%dS", seconds)
		}
		// Если только T без вложенных — пишем 0S.
		if hours == 0 && minutes == 0 && seconds == 0 {
			sb.WriteString("0S")
		}
	}

	return sb.String()
}

// --------------------------------------------------------------------------
// Форматирование RRULE
// --------------------------------------------------------------------------

// formatRRule форматирует RecurrenceRule в строку значения RRULE.
//
// Пример: FREQ=WEEKLY;BYDAY=MO,WE,FR;COUNT=10
func formatRRule(rule *model.RecurrenceRule) string {
	var parts []string

	if s := rule.Freq.String(); s != "" {
		parts = append(parts, "FREQ="+s)
	}
	if rule.Until != nil {
		parts = append(parts, "UNTIL="+formatDateTime(*rule.Until, false))
	}
	if rule.Count > 0 {
		parts = append(parts, fmt.Sprintf("COUNT=%d", rule.Count))
	}
	if rule.Interval > 1 {
		parts = append(parts, fmt.Sprintf("INTERVAL=%d", rule.Interval))
	}
	if len(rule.BySecond) > 0 {
		parts = append(parts, "BYSECOND="+formatIntList(rule.BySecond))
	}
	if len(rule.ByMinute) > 0 {
		parts = append(parts, "BYMINUTE="+formatIntList(rule.ByMinute))
	}
	if len(rule.ByHour) > 0 {
		parts = append(parts, "BYHOUR="+formatIntList(rule.ByHour))
	}
	if len(rule.ByDay) > 0 {
		parts = append(parts, "BYDAY="+formatByDay(rule.ByDay))
	}
	if len(rule.ByMonthDay) > 0 {
		parts = append(parts, "BYMONTHDAY="+formatIntList(rule.ByMonthDay))
	}
	if len(rule.ByYearDay) > 0 {
		parts = append(parts, "BYYEARDAY="+formatIntList(rule.ByYearDay))
	}
	if len(rule.ByWeekNo) > 0 {
		parts = append(parts, "BYWEEKNO="+formatIntList(rule.ByWeekNo))
	}
	if len(rule.ByMonth) > 0 {
		parts = append(parts, "BYMONTH="+formatIntList(rule.ByMonth))
	}
	if len(rule.BySetPos) > 0 {
		parts = append(parts, "BYSETPOS="+formatIntList(rule.BySetPos))
	}
	if rule.WkSt != model.WeekdayUnspecified && rule.WkSt != model.Monday {
		parts = append(parts, "WKST="+rule.WkSt.String())
	}

	return strings.Join(parts, ";")
}

// formatByDay форматирует список WeekdayNum: "MO,WE,FR" или "1MO,-1FR".
func formatByDay(days []model.WeekdayNum) string {
	parts := make([]string, len(days))
	for i, d := range days {
		if d.Ordinal != 0 {
			parts[i] = fmt.Sprintf("%d%s", d.Ordinal, d.Day.String())
		} else {
			parts[i] = d.Day.String()
		}
	}
	return strings.Join(parts, ",")
}

// formatIntList форматирует список целых чисел через запятую.
func formatIntList(nums []int) string {
	parts := make([]string, len(nums))
	for i, n := range nums {
		parts[i] = fmt.Sprintf("%d", n)
	}
	return strings.Join(parts, ",")
}

// --------------------------------------------------------------------------
// Форматирование Attendee / Organizer
// --------------------------------------------------------------------------

// formatAttendee собирает Property из Attendee.
func formatAttendee(propName string, a *model.Attendee) model.Property {
	p := model.Property{
		Name:  propName,
		Value: a.Address,
	}

	if a.Name != "" {
		p.Params = append(p.Params, model.Param{Name: "CN", Values: []string{a.Name}})
	}
	if s := a.Role.String(); s != "" {
		p.Params = append(p.Params, model.Param{Name: "ROLE", Values: []string{s}})
	}
	if s := a.Status.String(); s != "" {
		p.Params = append(p.Params, model.Param{Name: "PARTSTAT", Values: []string{s}})
	}
	if a.RSVP {
		p.Params = append(p.Params, model.Param{Name: "RSVP", Values: []string{"TRUE"}})
	}

	// Дополнительные параметры, которые парсер не обработал.
	p.Params = append(p.Params, a.Params...)

	return p
}

// --------------------------------------------------------------------------
// Форматирование Trigger (VALARM)
// --------------------------------------------------------------------------

// formatTrigger собирает Property для TRIGGER.
func formatTrigger(t model.Trigger) model.Property {
	p := model.Property{Name: "TRIGGER"}

	if t.DateTime != nil {
		p.Params = append(p.Params, model.Param{
			Name:   "VALUE",
			Values: []string{"DATE-TIME"},
		})
		p.Value = formatDateTime(*t.DateTime, false)
	} else if t.Duration != nil {
		p.Value = formatDuration(*t.Duration)
	}

	return p
}

// --------------------------------------------------------------------------
// Форматирование Relation
// --------------------------------------------------------------------------

// formatRelation собирает Property для RELATED-TO.
func formatRelation(rel model.Relation) model.Property {
	p := model.Property{
		Name:  "RELATED-TO",
		Value: rel.UID,
	}
	if s := rel.Type.String(); s != "" {
		p.Params = append(p.Params, model.Param{
			Name:   "RELTYPE",
			Values: []string{s},
		})
	}
	return p
}

// --------------------------------------------------------------------------
// Форматирование Period
// --------------------------------------------------------------------------

// formatPeriod форматирует Period в строку "start/end" или "start/duration".
func formatPeriod(p model.Period) string {
	start := formatDateTime(p.Start, false)
	if p.Duration > 0 {
		return start + "/" + formatDuration(p.Duration)
	}
	return start + "/" + formatDateTime(p.End, false)
}
