package parser

import (
	"strconv"
	"strings"
	"time"

	"github.com/mscloudx/ical/model"
	"github.com/pkg/errors"
)

// Константы длин форматов даты/времени в iCalendar.
const (
	dateLen     = 8  // "20231024"
	dtLocalLen  = 15 // "20231024T120000"
	dtUTCLen    = 16 // "20231024T120000Z"
	minDateLen  = 8  // минимальное кол-во символов для даты
	minTimeLen  = 6  // минимальное кол-во символов для времени
	minByDayLen = 2  // минимальное кол-во символов для BYDAY

	// Константы для преобразования единиц времени.
	hoursPerDay = 24
	daysPerWeek = 7
)

// --------------------------------------------------------------------------
// Парсинг даты и времени
// --------------------------------------------------------------------------

// parseDateTime парсит значение даты/времени iCalendar.
//
// Поддерживаемые форматы:
//   - DATE:      "20231024"          (8 символов)
//   - DATE-TIME: "20231024T120000"   (15 символов, floating/local)
//   - DATE-TIME: "20231024T120000Z"  (16 символов, UTC)
//
// loc используется для локального времени (из TZID). Если nil — UTC.
// Возвращает isDate=true для формата VALUE=DATE.
func parseDateTime(s string, loc *time.Location) (time.Time, bool, error) {
	if loc == nil {
		loc = time.UTC
	}

	switch len(s) {
	case dateLen: // DATE: 20231024
		y, m, d, err := parseDateDigits(s)
		if err != nil {
			return time.Time{}, false, err
		}
		t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, loc)
		if t.Year() != y || int(t.Month()) != m || t.Day() != d {
			return time.Time{}, false, errors.Wrapf(ErrInvalidDate, "day out of range for month: %q", s)
		}
		return t, true, nil

	case dtLocalLen: // DATE-TIME local: 20231024T120000
		if s[8] != 'T' {
			return time.Time{}, false, errors.Wrapf(ErrInvalidDateTime, "expected 'T' at position 8: %q", s)
		}
		y, m, d, err := parseDateDigits(s[:8])
		if err != nil {
			return time.Time{}, false, err
		}
		h, mi, sec, err := parseTimeDigits(s[9:15])
		if err != nil {
			return time.Time{}, false, err
		}
		t := time.Date(y, time.Month(m), d, h, mi, sec, 0, loc)
		if t.Year() != y || int(t.Month()) != m || t.Day() != d {
			return time.Time{}, false, errors.Wrapf(ErrInvalidDate, "day out of range for month: %q", s)
		}
		return t, false, nil

	case dtUTCLen: // DATE-TIME UTC: 20231024T120000Z
		if s[8] != 'T' || s[15] != 'Z' {
			return time.Time{}, false, errors.Wrapf(ErrInvalidDateTime, "expected 'T' at offset 8 and 'Z' suffix: %q", s)
		}
		y, m, d, err := parseDateDigits(s[:8])
		if err != nil {
			return time.Time{}, false, err
		}
		h, mi, sec, err := parseTimeDigits(s[9:15])
		if err != nil {
			return time.Time{}, false, err
		}
		t := time.Date(y, time.Month(m), d, h, mi, sec, 0, time.UTC)
		if t.Year() != y || int(t.Month()) != m || t.Day() != d {
			return time.Time{}, false, errors.Wrapf(ErrInvalidDate, "day out of range for month: %q", s)
		}
		return t, false, nil

	default:
		return time.Time{}, false, errors.Wrapf(ErrInvalidDateTime, "unknown format (%d chars): %q", len(s), s)
	}
}

// parseDateDigits парсит "20231024" → year=2023, month=10, day=24.
// Прямая конвертация цифр без strconv — быстрее time.Parse.
func parseDateDigits(s string) (year, month, day int, err error) {
	if len(s) < minDateLen {
		return 0, 0, 0, errors.Wrapf(ErrInvalidDate, "too short: %q", s)
	}
	year = digit4(s[0], s[1], s[2], s[3])
	month = digit2(s[4], s[5])
	day = digit2(s[6], s[7])
	if year < 0 || month < 1 || month > 12 || day < 1 || day > 31 {
		return 0, 0, 0, errors.Wrapf(ErrInvalidDate, "out-of-range values: %q", s)
	}
	return year, month, day, nil
}

// parseTimeDigits парсит "120000" → hour=12, min=0, sec=0.
func parseTimeDigits(s string) (hour, minute, sec int, err error) {
	if len(s) < minTimeLen {
		return 0, 0, 0, errors.Wrapf(ErrInvalidTime, "too short: %q", s)
	}
	hour = digit2(s[0], s[1])
	minute = digit2(s[2], s[3])
	sec = digit2(s[4], s[5])
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 || sec < 0 || sec > 60 {
		return 0, 0, 0, errors.Wrapf(ErrInvalidTime, "out-of-range values: %q", s)
	}
	return hour, minute, sec, nil
}

// digit2 конвертирует два ASCII-символа в число (00–99).
// Возвращает -1 при невалидных символах.
func digit2(a, b byte) int {
	if a < '0' || a > '9' || b < '0' || b > '9' {
		return -1
	}
	return int(a-'0')*10 + int(b-'0')
}

// digit4 конвертирует четыре ASCII-символа в число (0000–9999).
func digit4(a, b, c, d byte) int {
	if a < '0' || a > '9' || b < '0' || b > '9' || c < '0' || c > '9' || d < '0' || d > '9' {
		return -1
	}
	return int(a-'0')*1000 + int(b-'0')*100 + int(c-'0')*10 + int(d-'0')
}

// --------------------------------------------------------------------------
// Парсинг UTC-смещения (TZOFFSETFROM / TZOFFSETTO)
// --------------------------------------------------------------------------

// parseUTCOffset парсит UTC-смещение в формате RFC 5545: [+-]HHMM или [+-]HHMMSS.
// Возвращает смещение в секундах.
func parseUTCOffset(s string) (int, error) {
	if s == "" {
		return 0, errors.Wrap(ErrInvalidValue, "empty UTC offset")
	}

	sign := 1
	switch s[0] {
	case '+':
		s = s[1:]
	case '-':
		sign = -1
		s = s[1:]
	}

	const utcOffsetWithSecs = 6 // len("[+-]HHMMSS") без знака
	if len(s) != 4 && len(s) != utcOffsetWithSecs {
		return 0, errors.Wrapf(ErrInvalidValue, "UTC offset %q: expected [+-]HHMM or [+-]HHMMSS", s)
	}

	hh := digit2(s[0], s[1])
	mm := digit2(s[2], s[3])
	ss := 0
	if len(s) == utcOffsetWithSecs {
		ss = digit2(s[4], s[5])
	}

	if hh < 0 || hh > 23 || mm < 0 || mm > 59 || ss < 0 || ss > 59 {
		return 0, errors.Wrapf(ErrInvalidValue, "UTC offset out of range: %q", s)
	}

	const (
		secsPerMinute = 60
		secsPerHour   = 3600
	)
	return sign * (hh*secsPerHour + mm*secsPerMinute + ss), nil
}

// --------------------------------------------------------------------------
// Парсинг длительности (DURATION)
// --------------------------------------------------------------------------

// parseDuration парсит значение DURATION (подмножество ISO 8601).
//
// Формат: [+/-]P[nW | [nD][T[nH][nM][nS]]]
// Примеры: PT1H, P1DT2H30M, P1W, -PT10M, -P0DT0H10M0S
func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, ErrInvalidDuration
	}

	neg := false
	i := 0

	// Опциональный знак.
	switch s[i] {
	case '+':
		i++
	case '-':
		neg = true
		i++
	}

	// 'P' обязателен.
	if i >= len(s) || s[i] != 'P' {
		return 0, errors.Wrapf(ErrInvalidDuration, "missing 'P' designator: %q", s)
	}
	i++

	var d time.Duration
	inTime := false

	for i < len(s) {
		if s[i] == 'T' {
			inTime = true
			i++
			continue
		}

		// Считываем число.
		numStart := i
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		if i == numStart || i >= len(s) {
			return 0, errors.Wrapf(ErrInvalidDuration, "missing digit or designator: %q", s)
		}
		n, atoiErr := strconv.Atoi(s[numStart:i])
		if atoiErr != nil {
			return 0, errors.Wrapf(ErrInvalidDuration, "invalid number: %q", s[numStart:i])
		}

		switch s[i] {
		case 'W':
			d += time.Duration(n) * daysPerWeek * hoursPerDay * time.Hour
		case 'D':
			d += time.Duration(n) * hoursPerDay * time.Hour
		case 'H':
			if !inTime {
				return 0, errors.Wrapf(ErrInvalidDuration, "'H' without 'T': %q", s)
			}
			d += time.Duration(n) * time.Hour
		case 'M':
			if !inTime {
				return 0, errors.Wrapf(ErrInvalidDuration, "'M' without 'T': %q", s)
			}
			d += time.Duration(n) * time.Minute
		case 'S':
			if !inTime {
				return 0, errors.Wrapf(ErrInvalidDuration, "'S' without 'T': %q", s)
			}
			d += time.Duration(n) * time.Second
		default:
			return 0, errors.Wrapf(ErrInvalidDuration, "unknown designator %q: %q", string(s[i]), s)
		}
		i++
	}

	if neg {
		d = -d
	}
	return d, nil
}

// --------------------------------------------------------------------------
// Парсинг RRULE (самая сложная часть)
// --------------------------------------------------------------------------

// parseRRule парсит строку значения RRULE.
//
// Формат: FREQ=WEEKLY;BYDAY=MO,WE,FR;UNTIL=20231231T235959Z
// Все части опциональны, кроме FREQ.
func parseRRule(s string, loc *time.Location) (model.RecurrenceRule, error) {
	var rule model.RecurrenceRule

	parts := strings.Split(s, ";")
	for _, part := range parts {
		eqIdx := strings.IndexByte(part, '=')
		if eqIdx < 0 {
			continue
		}
		key := part[:eqIdx]
		val := part[eqIdx+1:]

		var err error
		switch key {
		case "FREQ":
			rule.Freq = parseFrequency(val)
			if rule.Freq == model.FreqUnspecified {
				return rule, errors.Wrapf(ErrUnknownFrequency, "FREQ=%s", val)
			}
		case "RSCALE":
			rule.RScale = parseRScale(val)
		case "SKIP":
			rule.Skip = parseRSkip(val)
		case "UNTIL":
			ut, _, parseErr := parseDateTime(val, loc)
			if parseErr != nil {
				return rule, errors.Wrap(parseErr, "UNTIL")
			}
			rule.Until = &ut
		case "COUNT":
			rule.Count, err = strconv.Atoi(val)
			if err != nil {
				return rule, errors.Wrapf(ErrInvalidRRule, "COUNT=%s: %s", val, err.Error())
			}
		case "INTERVAL":
			rule.Interval, err = strconv.Atoi(val)
			if err != nil {
				return rule, errors.Wrapf(ErrInvalidRRule, "INTERVAL=%s: %s", val, err.Error())
			}
		case "BYSECOND":
			rule.BySecond, err = parseIntList(val)
		case "BYMINUTE":
			rule.ByMinute, err = parseIntList(val)
		case "BYHOUR":
			rule.ByHour, err = parseIntList(val)
		case "BYDAY":
			rule.ByDay, err = parseByDay(val)
		case "BYMONTHDAY":
			rule.ByMonthDay, err = parseIntList(val)
		case "BYYEARDAY":
			rule.ByYearDay, err = parseIntList(val)
		case "BYWEEKNO":
			rule.ByWeekNo, err = parseIntList(val)
		case "BYMONTH":
			rule.ByMonth, err = parseIntList(val)
		case "BYSETPOS":
			rule.BySetPos, err = parseIntList(val)
		case "WKST":
			rule.WkSt = parseWeekday(val)
		}
		if err != nil {
			return rule, errors.Wrapf(ErrInvalidRRule, "%s: %s", key, err.Error())
		}
	}

	return rule, nil
}

// parseRScale конвертирует строку в RecurrenceScale.
// Нестандартные значения возвращаются как есть.
func parseRScale(s string) model.RecurrenceScale {
	upper := strings.ToUpper(s)
	switch upper {
	case "GREGORIAN":
		return model.RScaleGregorian
	default:
		return model.RecurrenceScale(s)
	}
}

// parseRSkip конвертирует строку в RecurrenceSkip.
// Нестандартные значения возвращаются как есть.
func parseRSkip(s string) model.RecurrenceSkip {
	upper := strings.ToUpper(s)
	switch upper {
	case "OMIT":
		return model.SkipOmit
	case "BACKWARD":
		return model.SkipBackward
	case "FORWARD":
		return model.SkipForward
	default:
		return model.RecurrenceSkip(s)
	}
}

// parseByDay парсит список BYDAY: "MO,WE,FR" или "1MO,-1FR".
func parseByDay(s string) ([]model.WeekdayNum, error) {
	parts := strings.Split(s, ",")
	result := make([]model.WeekdayNum, 0, len(parts))

	for _, part := range parts {
		wdn, err := parseWeekdayNum(part)
		if err != nil {
			return nil, err
		}
		result = append(result, wdn)
	}
	return result, nil
}

// parseWeekdayNum парсит один элемент BYDAY: "MO", "+1MO", "-1FR", "2TU".
func parseWeekdayNum(s string) (model.WeekdayNum, error) {
	if len(s) < minByDayLen {
		return model.WeekdayNum{}, errors.Wrapf(ErrInvalidByDay, "too short: %q", s)
	}

	// Последние 2 символа — всегда аббревиатура дня недели.
	dayStr := s[len(s)-2:]
	day := parseWeekday(dayStr)
	if day == model.WeekdayUnspecified {
		return model.WeekdayNum{}, errors.Wrapf(ErrUnknownWeekday, "%q", dayStr)
	}

	ordinal := 0
	if len(s) > minByDayLen {
		ordStr := s[:len(s)-minByDayLen]
		// Убираем опциональный '+'.
		if ordStr != "" && ordStr[0] == '+' {
			ordStr = ordStr[1:]
		}
		n, err := strconv.Atoi(ordStr)
		if err != nil {
			return model.WeekdayNum{}, errors.Wrapf(ErrInvalidByDay, "invalid ordinal in %q", s)
		}
		ordinal = n
	}

	return model.WeekdayNum{Ordinal: ordinal, Day: day}, nil
}

// --------------------------------------------------------------------------
// Парсинг Attendee / Organizer
// --------------------------------------------------------------------------

// parseAttendee собирает Attendee из параметров и значения свойства.
func parseAttendee(params []model.Param, value string) model.Attendee {
	a := model.Attendee{
		Address: value,
	}
	var extra []model.Param

	for _, param := range params {
		if len(param.Values) == 0 {
			continue
		}
		v := param.Values[0]
		switch strings.ToUpper(param.Name) {
		case "CN":
			a.Name = v
		case "ROLE":
			a.Role = parseRole(v)
		case "PARTSTAT":
			a.Status = parsePartStat(v)
		case "RSVP":
			a.RSVP = strings.EqualFold(v, "TRUE")
		case "SCHEDULE-AGENT":
			a.ScheduleAgent = parseScheduleAgent(v)
		case "SCHEDULE-FORCE-SEND":
			a.ScheduleForceSend = strings.ToUpper(v)
		case "SCHEDULE-STATUS":
			// Параметр может содержать несколько кодов через запятую.
			for _, code := range param.Values {
				code = strings.TrimSpace(code)
				if code != "" {
					a.ScheduleStatus = append(a.ScheduleStatus, code)
				}
			}
		default:
			extra = append(extra, param)
		}
	}
	a.Params = extra

	return a
}

// --------------------------------------------------------------------------
// Парсинг iTIP / RFC 6638
// --------------------------------------------------------------------------

// parseMethod конвертирует строку в Method (RFC 5546 §3.2).
func parseMethod(s string) model.Method {
	switch strings.ToUpper(s) {
	case "PUBLISH":
		return model.MethodPublish
	case "REQUEST":
		return model.MethodRequest
	case "REPLY":
		return model.MethodReply
	case "ADD":
		return model.MethodAdd
	case "CANCEL":
		return model.MethodCancel
	case "REFRESH":
		return model.MethodRefresh
	case "COUNTER":
		return model.MethodCounter
	case "DECLINECOUNTER":
		return model.MethodDeclineCounter
	default:
		return model.MethodUnspecified
	}
}

// parseRequestStatus парсит значение REQUEST-STATUS (RFC 5545 §3.8.8.3).
//
// Формат: <код>;<описание>[;<дополнительные данные>]
// Пример: "2.0;Success" или "3.7;Invalid calendar user;ATTENDEE:mailto:jdoe@example.com"
func parseRequestStatus(s string) model.RequestStatus {
	var rs model.RequestStatus
	// Ищем первую точку с запятой для разделения кода и описания.
	firstSemi := strings.IndexByte(s, ';')
	if firstSemi < 0 {
		rs.Code = s
		return rs
	}
	rs.Code = s[:firstSemi]
	rest := s[firstSemi+1:]

	// Ищем вторую точку с запятой для разделения описания и доп. данных.
	secondSemi := strings.IndexByte(rest, ';')
	if secondSemi < 0 {
		rs.Description = rest
		return rs
	}
	rs.Description = rest[:secondSemi]
	rs.ExtraData = rest[secondSemi+1:]
	return rs
}

// parseScheduleAgent конвертирует строку в ScheduleAgent (RFC 6638 §7.1).
func parseScheduleAgent(s string) model.ScheduleAgent {
	switch strings.ToUpper(s) {
	case "SERVER":
		return model.ScheduleAgentServer
	case "CLIENT":
		return model.ScheduleAgentClient
	case "NONE":
		return model.ScheduleAgentNone
	default:
		return model.ScheduleAgentUnspecified
	}
}

// --------------------------------------------------------------------------
// Парсинг Trigger (VALARM)
// --------------------------------------------------------------------------

// parseTrigger парсит значение TRIGGER с учётом параметров.
// По умолчанию — Duration. Если VALUE=DATE-TIME — абсолютное время.
func parseTrigger(params []model.Param, value string) (model.Trigger, error) {
	// Проверяем, задан ли VALUE=DATE-TIME.
	for _, param := range params {
		if strings.EqualFold(param.Name, "VALUE") && len(param.Values) > 0 {
			if strings.EqualFold(param.Values[0], "DATE-TIME") {
				t, _, err := parseDateTime(value, nil)
				if err != nil {
					return model.Trigger{}, errors.Wrap(err, "TRIGGER absolute datetime")
				}
				return model.Trigger{DateTime: &t}, nil
			}
		}
	}

	// По умолчанию — Duration.
	d, err := parseDuration(value)
	if err != nil {
		return model.Trigger{}, errors.Wrap(err, "TRIGGER duration")
	}
	return model.Trigger{Duration: &d}, nil
}

// --------------------------------------------------------------------------
// Lookup-функции для перечислений
// --------------------------------------------------------------------------

// parseFrequency конвертирует строку в Frequency.
func parseFrequency(s string) model.Frequency {
	switch strings.ToUpper(s) {
	case "SECONDLY":
		return model.FreqSecondly
	case "MINUTELY":
		return model.FreqMinutely
	case "HOURLY":
		return model.FreqHourly
	case "DAILY":
		return model.FreqDaily
	case "WEEKLY":
		return model.FreqWeekly
	case "MONTHLY":
		return model.FreqMonthly
	case "YEARLY":
		return model.FreqYearly
	default:
		return model.FreqUnspecified
	}
}

// parseWeekday конвертирует строку в Weekday.
func parseWeekday(s string) model.Weekday {
	switch strings.ToUpper(s) {
	case "MO":
		return model.Monday
	case "TU":
		return model.Tuesday
	case "WE":
		return model.Wednesday
	case "TH":
		return model.Thursday
	case "FR":
		return model.Friday
	case "SA":
		return model.Saturday
	case "SU":
		return model.Sunday
	default:
		return model.WeekdayUnspecified
	}
}

// parseStatus конвертирует строку в Status.
func parseStatus(s string) model.Status {
	switch strings.ToUpper(s) {
	case "TENTATIVE":
		return model.StatusTentative
	case "CONFIRMED":
		return model.StatusConfirmed
	case "CANCELLED":
		return model.StatusCancelled
	case "COMPLETED":
		return model.StatusCompleted
	case "NEEDS-ACTION":
		return model.StatusNeedsAction
	case "IN-PROCESS":
		return model.StatusInProcess
	case "DRAFT":
		return model.StatusDraft
	case "FINAL":
		return model.StatusFinal
	default:
		return model.StatusUnspecified
	}
}

// parseTransparency конвертирует строку в Transparency.
func parseTransparency(s string) model.Transparency {
	switch strings.ToUpper(s) {
	case "OPAQUE":
		return model.TranspOpaque
	case "TRANSPARENT":
		return model.TranspTransparent
	default:
		return model.TranspUnspecified
	}
}

// parseClassification конвертирует строку в Classification.
func parseClassification(s string) model.Classification {
	switch strings.ToUpper(s) {
	case "PUBLIC":
		return model.ClassPublic
	case "PRIVATE":
		return model.ClassPrivate
	case "CONFIDENTIAL":
		return model.ClassConfidential
	default:
		return model.ClassUnspecified
	}
}

// parsePartStat конвертирует строку в ParticipationStatus.
func parsePartStat(s string) model.ParticipationStatus {
	switch strings.ToUpper(s) {
	case "NEEDS-ACTION":
		return model.PartStatNeedsAction
	case "ACCEPTED":
		return model.PartStatAccepted
	case "DECLINED":
		return model.PartStatDeclined
	case "TENTATIVE":
		return model.PartStatTentative
	case "DELEGATED":
		return model.PartStatDelegated
	default:
		return model.PartStatUnspecified
	}
}

// parseRole конвертирует строку в Role.
func parseRole(s string) model.Role {
	switch strings.ToUpper(s) {
	case "CHAIR":
		return model.RoleChair
	case "REQ-PARTICIPANT":
		return model.RoleReqParticipant
	case "OPT-PARTICIPANT":
		return model.RoleOptParticipant
	case "NON-PARTICIPANT":
		return model.RoleNonParticipant
	default:
		return model.RoleUnspecified
	}
}

// parseAlarmAction конвертирует строку в AlarmAction.
func parseAlarmAction(s string) model.AlarmAction {
	switch strings.ToUpper(s) {
	case "AUDIO":
		return model.ActionAudio
	case "DISPLAY":
		return model.ActionDisplay
	case "EMAIL":
		return model.ActionEmail
	default:
		return model.ActionUnspecified
	}
}

// parseFreeBusyType конвертирует строку в FreeBusyType.
func parseFreeBusyType(s string) model.FreeBusyType {
	switch strings.ToUpper(s) {
	case "FREE":
		return model.FBTypeFree
	case "BUSY":
		return model.FBTypeBusy
	case "BUSY-UNAVAILABLE":
		return model.FBTypeBusyUnavailable
	case "BUSY-TENTATIVE":
		return model.FBTypeBusyTentative
	default:
		return model.FBTypeUnspecified
	}
}

// parseRelType конвертирует строку в RelationshipType.
func parseRelType(s string) model.RelationshipType {
	switch strings.ToUpper(s) {
	case "PARENT":
		return model.RelTypeParent
	case "CHILD":
		return model.RelTypeChild
	case "SIBLING":
		return model.RelTypeSibling
	case "FINISHTOSTART":
		return model.RelTypeFinishToStart
	case "FINISHTOFINISH":
		return model.RelTypeFinishToFinish
	case "STARTTOFINISH":
		return model.RelTypeStartToFinish
	case "STARTTOSTART":
		return model.RelTypeStartToStart
	case "FIRST":
		return model.RelTypeFirst
	case "NEXT":
		return model.RelTypeNext
	case "DEPENDS-ON":
		return model.RelTypeDependsOn
	case "REFID":
		return model.RelTypeRefID
	case "CONCEPT":
		return model.RelTypeConcept
	case "SNOOZE":
		return model.RelTypeSnooze
	default:
		return model.RelTypeUnspecified
	}
}

// parseProximity конвертирует строку в Proximity.
// Если значение не стандартное — возвращает исходную строку как Proximity.
func parseProximity(s string) model.Proximity {
	upper := strings.ToUpper(s)
	switch upper {
	case "ARRIVE":
		return model.ProximityArrive
	case "DEPART":
		return model.ProximityDepart
	case "CONNECT":
		return model.ProximityConnect
	case "DISCONNECT":
		return model.ProximityDisconnect
	default:
		return model.Proximity(s)
	}
}

// --------------------------------------------------------------------------
// Вспомогательные функции
// --------------------------------------------------------------------------

// parseIntList парсит список целых чисел через запятую: "1,2,3".
func parseIntList(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, errors.Wrapf(ErrInvalidRRule, "invalid integer %q", p)
		}
		result = append(result, n)
	}
	return result, nil
}

// parsePeriod парсит значение PERIOD: "start/end" или "start/duration".
func parsePeriod(s string, loc *time.Location) (model.Period, error) {
	slashIdx := strings.IndexByte(s, '/')
	if slashIdx < 0 {
		return model.Period{}, errors.Wrapf(ErrInvalidPeriod, "missing '/' separator: %q", s)
	}

	start, _, err := parseDateTime(s[:slashIdx], loc)
	if err != nil {
		return model.Period{}, errors.Wrap(err, "PERIOD start")
	}

	endOrDur := s[slashIdx+1:]

	// Если начинается с 'P' — это Duration, иначе — DateTime.
	if endOrDur != "" && (endOrDur[0] == 'P' || endOrDur[0] == '+' || endOrDur[0] == '-') {
		durVal, durErr := parseDuration(endOrDur)
		if durErr != nil {
			return model.Period{}, errors.Wrap(durErr, "PERIOD duration")
		}
		return model.Period{Start: start, Duration: durVal}, nil
	}

	end, _, err := parseDateTime(endOrDur, loc)
	if err != nil {
		return model.Period{}, errors.Wrap(err, "PERIOD end")
	}
	return model.Period{Start: start, End: end}, nil
}

// parsePeriodList парсит список периодов через запятую (RDATE;VALUE=PERIOD).
func parsePeriodList(s string, loc *time.Location) ([]model.Period, error) {
	parts := strings.Split(s, ",")
	result := make([]model.Period, 0, len(parts))
	for _, part := range parts {
		p, err := parsePeriod(strings.TrimSpace(part), loc)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}
