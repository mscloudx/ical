package model

// Calendar представляет корневой компонент VCALENDAR.
//
// Содержит метаданные календаря и вложенные компоненты.
// RFC 5545 §3.4, §3.6.
type Calendar struct {
	// Version — версия iCalendar (обычно "2.0").
	// RFC 5545 §3.7.4.
	Version string
	// ProdID — идентификатор продукта, создавшего объект.
	// RFC 5545 §3.7.3.
	ProdID string
	// CalScale — шкала календаря (по умолчанию "GREGORIAN").
	// RFC 5545 §3.7.1.
	CalScale string
	// Method — метод iTIP (REQUEST, REPLY, CANCEL и т.д.).
	// RFC 5545 §3.7.2.
	Method string

	// Events — вложенные компоненты VEVENT.
	Events []Event
	// Todos — вложенные компоненты VTODO.
	Todos []Todo
	// Journals — вложенные компоненты VJOURNAL.
	Journals []Journal
	// FreeBusys — вложенные компоненты VFREEBUSY.
	FreeBusys []FreeBusy
	// Timezones — вложенные компоненты VTIMEZONE.
	Timezones []Timezone

	// XProps — нестандартные и vendor-специфичные свойства
	// (X-WR-CALNAME, X-WR-TIMEZONE и др.).
	XProps []Property
	// IanaProps — зарегистрированные IANA свойства календаря без префикса X-.
	// Например: REFRESH-INTERVAL, SOURCE, COLOR (RFC 7986).
	IanaProps []Property
}
