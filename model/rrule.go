package model

import "time"

// --------------------------------------------------------------------------
// Frequency — частота повторения (FREQ).
// RFC 5545 §3.3.10.
// --------------------------------------------------------------------------

// Frequency определяет частоту повторения правила.
type Frequency int

const (
	// FreqUnspecified — частота не задана (нулевое значение).
	FreqUnspecified Frequency = iota
	// FreqSecondly — каждую секунду.
	FreqSecondly
	// FreqMinutely — каждую минуту.
	FreqMinutely
	// FreqHourly — каждый час.
	FreqHourly
	// FreqDaily — каждый день.
	FreqDaily
	// FreqWeekly — каждую неделю.
	FreqWeekly
	// FreqMonthly — каждый месяц.
	FreqMonthly
	// FreqYearly — каждый год.
	FreqYearly
)

// --------------------------------------------------------------------------
// Weekday — день недели (для BYDAY и WKST).
// RFC 5545 §3.3.10.
// --------------------------------------------------------------------------

// Weekday определяет день недели.
type Weekday int

const (
	// WeekdayUnspecified — день не задан (нулевое значение).
	WeekdayUnspecified Weekday = iota
	// Monday — понедельник (MO).
	Monday
	// Tuesday — вторник (TU).
	Tuesday
	// Wednesday — среда (WE).
	Wednesday
	// Thursday — четверг (TH).
	Thursday
	// Friday — пятница (FR).
	Friday
	// Saturday — суббота (SA).
	Saturday
	// Sunday — воскресенье (SU).
	Sunday
)

// --------------------------------------------------------------------------
// WeekdayNum — день недели с порядковым номером для BYDAY.
// RFC 5545 §3.3.10.
// --------------------------------------------------------------------------

// WeekdayNum представляет день недели с опциональным порядковым номером.
//
// Примеры:
//   - {Ordinal: 0, Day: Monday}  → каждый понедельник (MO)
//   - {Ordinal: 1, Day: Monday}  → первый понедельник (+1MO)
//   - {Ordinal: -1, Day: Friday} → последняя пятница (-1FR)
type WeekdayNum struct {
	// Ordinal — порядковый номер (0 = все вхождения, +1 = первый, -1 = последний).
	Ordinal int
	// Day — день недели.
	Day Weekday
}

// --------------------------------------------------------------------------
// RecurrenceRule — правило повторения (RRULE).
// RFC 5545 §3.3.10, §3.8.5.3.
// --------------------------------------------------------------------------

// RecurrenceRule описывает правило повторения события или перехода.
//
// Поля соответствуют частям RRULE:
//
//	FREQ=WEEKLY;BYDAY=MO,WE,FR;UNTIL=20231231T235959Z
type RecurrenceRule struct {
	// Freq — частота повторения (обязательное поле).
	Freq Frequency
	// Until — дата окончания повторений (nil, если не ограничено по дате).
	// Взаимоисключающее с Count.
	Until *time.Time
	// Count — максимальное число повторений (0 = не ограничено).
	// Взаимоисключающее с Until.
	Count int
	// Interval — интервал между повторениями (0 или 1 = каждый раз).
	Interval int

	// BySecond — секунды (0–60).
	BySecond []int
	// ByMinute — минуты (0–59).
	ByMinute []int
	// ByHour — часы (0–23).
	ByHour []int
	// ByDay — дни недели с опциональным порядком.
	ByDay []WeekdayNum
	// ByMonthDay — дни месяца (-31..31, исключая 0).
	ByMonthDay []int
	// ByYearDay — дни года (-366..366, исключая 0).
	ByYearDay []int
	// ByWeekNo — номера недель (-53..53, исключая 0).
	ByWeekNo []int
	// ByMonth — месяцы (1–12).
	ByMonth []int
	// BySetPos — позиции в наборе (-366..366, исключая 0).
	BySetPos []int
	// WkSt — первый день недели (по умолчанию Monday).
	WkSt Weekday
}

// --------------------------------------------------------------------------
// Методы String() для enum-типов RRULE.
// --------------------------------------------------------------------------

// String возвращает RFC 5545 строку для частоты повторения.
func (f Frequency) String() string {
	switch f {
	case FreqSecondly:
		return "SECONDLY"
	case FreqMinutely:
		return "MINUTELY"
	case FreqHourly:
		return "HOURLY"
	case FreqDaily:
		return "DAILY"
	case FreqWeekly:
		return "WEEKLY"
	case FreqMonthly:
		return "MONTHLY"
	case FreqYearly:
		return "YEARLY"
	default:
		return ""
	}
}

// String возвращает RFC 5545 аббревиатуру дня недели.
func (w Weekday) String() string {
	switch w {
	case Monday:
		return "MO"
	case Tuesday:
		return "TU"
	case Wednesday:
		return "WE"
	case Thursday:
		return "TH"
	case Friday:
		return "FR"
	case Saturday:
		return "SA"
	case Sunday:
		return "SU"
	default:
		return ""
	}
}
