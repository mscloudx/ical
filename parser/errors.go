package parser

import "github.com/pkg/errors"

// Sentinel-ошибки парсера iCalendar.
// Позволяют caller-у различать виды ошибок через errors.Is().
var (
	// ErrNoVCalendar возвращается, если в потоке не найден компонент VCALENDAR.
	ErrNoVCalendar = errors.New("VCALENDAR component not found")

	// ErrUnexpectedEnd возвращается при встрече END без открытого BEGIN.
	ErrUnexpectedEnd = errors.New("unexpected END without matching BEGIN")

	// ErrMismatchedEnd возвращается при несовпадении имён BEGIN и END.
	ErrMismatchedEnd = errors.New("END component name does not match BEGIN")

	// ErrInvalidDateTime возвращается при некорректном формате значения даты/времени.
	ErrInvalidDateTime = errors.New("invalid date-time value")

	// ErrInvalidDate возвращается при некорректном формате даты (YYYYMMDD).
	ErrInvalidDate = errors.New("invalid date value")

	// ErrInvalidTime возвращается при некорректном формате времени (HHMMSS).
	ErrInvalidTime = errors.New("invalid time value")

	// ErrInvalidDuration возвращается при некорректном значении DURATION (ISO 8601).
	ErrInvalidDuration = errors.New("invalid duration value")

	// ErrInvalidRRule возвращается при ошибке разбора RRULE.
	ErrInvalidRRule = errors.New("invalid RRULE value")

	// ErrUnknownFrequency возвращается при неизвестном значении FREQ.
	ErrUnknownFrequency = errors.New("unknown RRULE frequency")

	// ErrInvalidByDay возвращается при ошибке разбора BYDAY.
	ErrInvalidByDay = errors.New("invalid BYDAY value")

	// ErrUnknownWeekday возвращается при неизвестной аббревиатуре дня недели.
	ErrUnknownWeekday = errors.New("unknown weekday abbreviation")

	// ErrInvalidPeriod возвращается при некорректном значении PERIOD.
	ErrInvalidPeriod = errors.New("invalid PERIOD value")

	// ErrInvalidTrigger возвращается при ошибке разбора TRIGGER.
	ErrInvalidTrigger = errors.New("invalid TRIGGER value")

	// ErrInvalidValue возвращается при некорректном числовом значении свойства.
	ErrInvalidValue = errors.New("invalid property value")

	// ErrScanFailed возвращается при ошибке чтения iCalendar-потока.
	ErrScanFailed = errors.New("iCalendar stream read failed")
)
