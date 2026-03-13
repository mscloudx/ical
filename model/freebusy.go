package model

import "time"

// FreeBusy представляет компонент VFREEBUSY — информация о занятости.
//
// Используется для запроса или публикации свободного/занятого времени
// без раскрытия деталей конкретных событий.
// RFC 5545 §3.6.4.
type FreeBusy struct {
	// UID — уникальный идентификатор компонента.
	// RFC 5545 §3.8.4.7.
	UID string
	// DTStamp — временная метка создания объекта.
	// RFC 5545 §3.8.7.2.
	DTStamp time.Time
	// DTStart — начало интересующего периода (nil, если не задано).
	// RFC 5545 §3.8.2.4.
	DTStart *time.Time
	// DTEnd — конец интересующего периода (nil, если не задано).
	// RFC 5545 §3.8.2.2.
	DTEnd *time.Time

	// Organizer — организатор запроса.
	// RFC 5545 §3.8.4.3.
	Organizer *Attendee
	// Attendees — участники, чья занятость запрашивается.
	// RFC 5545 §3.8.4.1.
	Attendees []Attendee

	// URL — связанный URL.
	// RFC 5545 §3.8.4.6.
	URL string

	// Periods — периоды свободного/занятого времени (FREEBUSY).
	// RFC 5545 §3.8.2.6.
	Periods []FreeBusyPeriod

	// XProps — нестандартные свойства (X-*).
	XProps []Property
	// IanaProps — зарегистрированные IANA свойства без префикса X-.
	IanaProps []Property
}

// FreeBusyPeriod — отдельный период занятости с типом.
type FreeBusyPeriod struct {
	// Type — тип занятости (FREE, BUSY, BUSY-UNAVAILABLE, BUSY-TENTATIVE).
	Type FreeBusyType
	// Period — временной период.
	Period Period
}
