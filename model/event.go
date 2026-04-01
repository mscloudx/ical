package model

import "time"

// Event представляет компонент VEVENT — календарное событие.
//
// Событие наступает в определённое время (DTStart), может иметь конец (DTEnd)
// или длительность (Duration), правила повторения и вложенные напоминания.
// RFC 5545 §3.6.1.
type Event struct {
	// UID — глобально уникальный идентификатор события.
	// RFC 5545 §3.8.4.7.
	UID string
	// DTStamp — временная метка создания/изменения iCalendar-объекта.
	// RFC 5545 §3.8.7.2.
	DTStamp time.Time
	// DTStart — дата и время начала события.
	// RFC 5545 §3.8.2.4.
	DTStart time.Time
	// DTEnd — дата и время окончания события (nil, если не задано).
	// Взаимоисключающее с Duration.
	// RFC 5545 §3.8.2.2.
	DTEnd *time.Time
	// Duration — длительность события (nil, если не задана).
	// Взаимоисключающее с DTEnd.
	// RFC 5545 §3.8.2.5.
	Duration *time.Duration

	// AllDay — true, если DTStart имеет тип VALUE=DATE (событие на весь день).
	AllDay bool

	// Summary — краткий заголовок события.
	// RFC 5545 §3.8.1.12.
	Summary string
	// Description — подробное описание события.
	// RFC 5545 §3.8.1.5.
	Description string
	// Location — место проведения.
	// RFC 5545 §3.8.1.7.
	Location string
	// URL — связанный URL.
	// RFC 5545 §3.8.4.6.
	URL string
	// Attach — вложения: URI или встроенные двоичные данные (BASE64).
	// RFC 5545 §3.8.1.1.
	Attach []Attachment

	// Status — статус события (TENTATIVE, CONFIRMED, CANCELLED).
	// RFC 5545 §3.8.1.11.
	Status Status
	// Transparency — прозрачность времени (OPAQUE, TRANSPARENT).
	// RFC 5545 §3.8.2.7.
	Transparency Transparency
	// Classification — уровень видимости (PUBLIC, PRIVATE, CONFIDENTIAL).
	// RFC 5545 §3.8.1.3.
	Classification Classification

	// Organizer — организатор события.
	// RFC 5545 §3.8.4.3.
	Organizer *Attendee
	// Attendees — участники события.
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

	// RecurrenceID — идентификатор экземпляра повторяющегося события.
	// Используется для изменения конкретного повторения.
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

	// Alarms — вложенные напоминания (VALARM).
	// RFC 5545 §3.6.6.
	Alarms []Alarm
	// Related — связи с другими компонентами (RELATED-TO).
	// RFC 5545 §3.8.4.5.
	Related []Relation
	// Concepts — семантические категории (CONCEPT, RFC 9253).
	Concepts []string
	// Links — связи с внешними ресурсами (LINK, RFC 9253).
	Links []Link
	// RefIDs — идентификаторы связей (REFID, RFC 9253).
	RefIDs []string

	// RequestStatus — статусы обработки iTIP-запроса (REQUEST-STATUS).
	// RFC 5545 §3.8.8.3, RFC 5546 §3.6.
	RequestStatus []RequestStatus

	// XProps — нестандартные свойства (X-*).
	XProps []Property
	// IanaProps — зарегистрированные IANA свойства без префикса X-.
	// Например: COLOR, IMAGE, CONFERENCE (RFC 7986).
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
