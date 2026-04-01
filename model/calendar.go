package model

import "time"

// Calendar представляет корневой компонент VCALENDAR.
//
// Содержит метаданные календаря и вложенные компоненты.
// Базовая спецификация: RFC 5545 §3.4, §3.6.
// Новые свойства: RFC 7986.
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
	// RFC 5545 §3.7.2, RFC 5546 §3.2.
	Method Method

	// Name — текстовое название календаря (опционально, RFC 7986).
	Name string
	// Description — описание календаря (опционально, RFC 7986).
	Description string
	// UID — уникальный идентификатор календаря (опционально, RFC 7986).
	UID string
	// LastModified — дата последнего изменения календаря (опционально, RFC 7986).
	LastModified *time.Time
	// URL — ссылка на связанный ресурс календаря (опционально, RFC 7986).
	URL string
	// Categories — категории календаря (опционально, RFC 7986).
	Categories []string
	// RefreshInterval — интервал обновления для клиентов (опционально, RFC 7986).
	RefreshInterval *time.Duration
	// Color — спецификация цвета (например, CSS3) для календаря (опционально, RFC 7986).
	Color string
	// Source — источник (URI) откуда можно получить актуальную версию календаря (опционально, RFC 7986).
	Source string

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
	// Availabilities — вложенные компоненты VAVAILABILITY (RFC 7953).
	Availabilities []Availability

	// XProps — нестандартные и vendor-специфичные свойства
	// (X-WR-CALNAME, X-WR-TIMEZONE и др.).
	XProps []Property
	// IanaProps — зарегистрированные IANA свойства календаря без префикса X-.
	// Например: REFRESH-INTERVAL, SOURCE, COLOR (RFC 7986).
	IanaProps []Property
}
