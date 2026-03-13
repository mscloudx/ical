package model

import "time"

// Timezone представляет компонент VTIMEZONE — описание часового пояса.
//
// Содержит набор правил перехода между стандартным и летним временем,
// используемых для разрешения локального времени в UTC.
// RFC 5545 §3.6.5.
type Timezone struct {
	// TZID — идентификатор часового пояса (например, "Europe/Moscow").
	// RFC 5545 §3.8.3.1.
	TZID string
	// Standards — правила стандартного времени (STANDARD).
	Standards []TzTransition
	// Daylights — правила летнего времени (DAYLIGHT).
	Daylights []TzTransition
	// XProps — нестандартные свойства (X-*).
	XProps []Property
	// IanaProps — зарегистрированные IANA свойства без префикса X-.
	IanaProps []Property
}

// TzTransition представляет одно правило перехода STANDARD или DAYLIGHT.
// RFC 5545 §3.6.5.
type TzTransition struct {
	// DTStart — дата и время перехода (локальное время).
	// RFC 5545 §3.8.2.4.
	DTStart time.Time
	// OffsetFrom — смещение UTC до перехода (например, "+0400").
	// RFC 5545 §3.8.3.3.
	OffsetFrom string
	// OffsetTo — смещение UTC после перехода (например, "+0300").
	// RFC 5545 §3.8.3.4.
	OffsetTo string
	// TZName — краткое название часового пояса (например, "MSK", "MSD").
	// RFC 5545 §3.8.3.2.
	TZName string

	// RRules — правила повторения перехода (для ежегодных переходов).
	RRules []RecurrenceRule
	// RDates — конкретные даты дополнительных переходов.
	RDates []time.Time

	// XProps — нестандартные свойства (X-*).
	XProps []Property
	// IanaProps — зарегистрированные IANA свойства без префикса X-.
	IanaProps []Property
}
