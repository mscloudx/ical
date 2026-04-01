package model

import "time"

// Todo представляет компонент VTODO — задачу (to-do).
//
// Задача может иметь срок выполнения (Due), дату начала (DTStart),
// процент готовности (PercentComplete), приоритет (Priority) и
// вложенные напоминания VALARM.
// RFC 5545 §3.6.2.
type Todo struct {
	// UID — глобально уникальный идентификатор задачи.
	// RFC 5545 §3.8.4.7.
	UID string
	// DTStamp — временная метка создания/изменения iCalendar-объекта.
	// RFC 5545 §3.8.7.2.
	DTStamp time.Time
	// DTStart — дата и время начала задачи (nil, если не задано).
	// RFC 5545 §3.8.2.4.
	DTStart *time.Time
	// AllDay — true, если DTStart имеет тип VALUE=DATE (задача на весь день).
	AllDay bool
	// Due — срок выполнения задачи (nil, если не задано).
	// Взаимоисключающее с Duration.
	// RFC 5545 §3.8.2.3.
	Due *time.Time
	// Duration — длительность задачи (nil, если не задана).
	// Взаимоисключающее с Due. При наличии требует DTStart.
	// RFC 5545 §3.8.2.5.
	Duration *time.Duration
	// Completed — дата и время фактического выполнения (nil, если не задано).
	// Должна быть в UTC.
	// RFC 5545 §3.8.2.1.
	Completed *time.Time

	// Priority — приоритет задачи (0–9).
	// 0 = не определён, 1 = высший, 9 = низший.
	// RFC 5545 §3.8.1.9.
	Priority int
	// PercentComplete — процент готовности задачи (0–100).
	// RFC 5545 §3.8.1.8.
	PercentComplete int

	// Summary — краткий заголовок задачи.
	// RFC 5545 §3.8.1.12.
	Summary string
	// Description — подробное описание задачи.
	// RFC 5545 §3.8.1.5.
	Description string
	// Location — место выполнения задачи.
	// RFC 5545 §3.8.1.7.
	Location string
	// URL — связанный URL.
	// RFC 5545 §3.8.4.6.
	URL string
	// Attach — вложения: URI или встроенные двоичные данные (BASE64).
	// RFC 5545 §3.8.1.1.
	Attach []Attachment

	// Status — статус задачи (NEEDS-ACTION, IN-PROCESS, COMPLETED, CANCELLED).
	// RFC 5545 §3.8.1.11.
	Status Status
	// Classification — уровень видимости (PUBLIC, PRIVATE, CONFIDENTIAL).
	// RFC 5545 §3.8.1.3.
	Classification Classification

	// Organizer — организатор / постановщик задачи.
	// RFC 5545 §3.8.4.3.
	Organizer *Attendee
	// Attendees — участники задачи.
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

	// RecurrenceID — идентификатор экземпляра повторяющейся задачи.
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
