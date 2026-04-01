package model

import "time"

// Journal представляет компонент VJOURNAL — запись в журнале / дневнике.
//
// Запись привязана к конкретной дате и содержит текстовые заметки.
// Не занимает время в календаре (прозрачна для free/busy).
// Не поддерживает VALARM.
// RFC 5545 §3.6.3.
type Journal struct {
	// UID — глобально уникальный идентификатор записи.
	// RFC 5545 §3.8.4.7.
	UID string
	// DTStamp — временная метка создания/изменения iCalendar-объекта.
	// RFC 5545 §3.8.7.2.
	DTStamp time.Time
	// DTStart — дата записи (обычно VALUE=DATE).
	// RFC 5545 §3.8.2.4.
	DTStart *time.Time
	// AllDay — true, если DTStart имеет тип VALUE=DATE.
	AllDay bool

	// Summary — заголовок записи.
	// RFC 5545 §3.8.1.12.
	Summary string
	// Descriptions — текстовые блоки записи.
	// RFC 5545 допускает несколько DESCRIPTION в одном VJOURNAL (§3.6.3).
	// RFC 5545 §3.8.1.5.
	Descriptions []string
	// URL — связанный URL.
	// RFC 5545 §3.8.4.6.
	URL string
	// Attach — вложения: URI или встроенные двоичные данные (BASE64).
	// RFC 5545 §3.8.1.1.
	Attach []Attachment

	// Status — статус записи (DRAFT, FINAL, CANCELLED).
	// RFC 5545 §3.8.1.11.
	Status Status
	// Classification — уровень видимости (PUBLIC, PRIVATE, CONFIDENTIAL).
	// RFC 5545 §3.8.1.3.
	Classification Classification

	// Organizer — автор записи.
	// RFC 5545 §3.8.4.3.
	Organizer *Attendee
	// Attendees — участники (упомянутые в записи).
	// RFC 5545 §3.8.4.1.
	Attendees []Attendee

	// Created — дата создания компонента.
	// RFC 5545 §3.8.7.1.
	Created *time.Time
	// LastModified — дата последнего изменения.
	// RFC 5545 §3.8.7.3.
	LastModified *time.Time
	// Sequence — порядковый номер ревизии.
	// RFC 5545 §3.8.7.4.
	Sequence int

	// RecurrenceID — идентификатор экземпляра повторяющейся записи.
	// RFC 5545 §3.8.4.4.
	RecurrenceID *time.Time

	// RRules — правила повторения (RRULE).
	// RFC 5545 §3.8.5.3.
	RRules []RecurrenceRule
	// RDates — дополнительные даты повторения (RDATE) с типом DATE или DATE-TIME.
	// RFC 5545 §3.8.5.2.
	RDates []time.Time
	// RDatePeriods — дополнительные периоды повторения (RDATE;VALUE=PERIOD).
	// RFC 5545 §3.8.5.2.
	RDatePeriods []Period
	// ExDates — даты-исключения из повторений (EXDATE).
	// RFC 5545 §3.8.5.1.
	ExDates []time.Time

	// Related — связи с другими компонентами (RELATED-TO).
	// RFC 5545 §3.8.4.5.
	Related []Relation
	// Concepts — семантические категории (CONCEPT, RFC 9253).
	Concepts []string
	// Links — связи с внешними ресурсами (LINK, RFC 9253).
	Links []Link
	// RefIDs — идентификаторы связей (REFID, RFC 9253).
	RefIDs []string

	// XProps — нестандартные свойства (X-*).
	XProps []Property
	// IanaProps — зарегистрированные IANA свойства без префикса X-.
	// Например: COLOR, CONFERENCE (RFC 7986).
	IanaProps []Property

	// StyledDescriptions — форматированное описание (RFC 9073).
	StyledDescriptions []StyledDescription
	// StructuredData — структурированные данные (RFC 9073).
	StructuredData []StructuredData

	// Participants — участники (RFC 9073).
	Participants []Participant
	// Locations — локации (RFC 9073).
	Locations []LocationComponent
	// Resources — ресурсы (RFC 9073).
	Resources []ResourceComponent
}
