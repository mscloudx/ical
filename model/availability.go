package model

import "time"

// BusyType определяет тип занятости для VAVAILABILITY (RFC 7953).
type BusyType string

const (
	// BusyTypeBusy означает, что компонент занят (по умолчанию).
	BusyTypeBusy BusyType = "BUSY"
	// BusyTypeBusyUnavailable означает недоступность.
	BusyTypeBusyUnavailable BusyType = "BUSY-UNAVAILABLE"
	// BusyTypeBusyTentative означает вероятную занятость.
	BusyTypeBusyTentative BusyType = "BUSY-TENTATIVE"
)

// String возвращает строковое представление BusyType.
func (bt BusyType) String() string {
	return string(bt)
}

// Availability представляет компонент VAVAILABILITY (RFC 7953).
//
// Используется для описания периодов времени, когда ресурс или пользователь
// доступен для планирования.
type Availability struct {
	// UID — уникальный идентификатор.
	UID string
	// DTStamp — временная метка создания/изменения.
	DTStamp time.Time
	// DTStart — начало доступности (опционально).
	DTStart *time.Time
	// DTEnd — конец доступности (взаимоисключающее с Duration).
	DTEnd *time.Time
	// Duration — длительность доступности (взаимоисключающее с DTEnd).
	Duration *time.Duration

	// BusyType указывает тип времени (BUSY, BUSY-UNAVAILABLE, BUSY-TENTATIVE).
	BusyType BusyType

	// Created — дата создания.
	Created *time.Time
	// LastModified — дата последнего изменения.
	LastModified *time.Time
	// Sequence — номер ревизии.
	Sequence int

	// Summary — заголовок.
	Summary string
	// Description — текстовое описание.
	Description string
	// URL — ссылка.
	URL string
	// Categories — категории.
	Categories []string

	// Organizer — организатор.
	Organizer *Attendee

	// Available — вложенные "окна" доступного времени (AVAILABLE).
	Available []Available
	// Alarms — вложенные напоминания (VALARM).
	Alarms []Alarm

	// XProps — нестандартные свойства.
	XProps []Property
	// IanaProps — зарегистрированные свойства IANA.
	IanaProps []Property
}

// Available представляет компонент AVAILABLE вложенный в VAVAILABILITY (RFC 7953).
//
// Описывает конкретный период времени, когда ресурс доступен.
type Available struct {
	// UID — уникальный идентификатор.
	UID string
	// DTStamp — временная метка создания/изменения.
	DTStamp time.Time
	// DTStart — дата и время начала периода.
	DTStart time.Time
	// DTEnd — дата и время окончания периода (взаимоисключающее с Duration).
	DTEnd *time.Time
	// Duration — длительность периода (взаимоисключающее с DTEnd).
	Duration *time.Duration

	// Summary — заголовок.
	Summary string
	// Description — подробное описание.
	Description string

	// Created — дата создания.
	Created *time.Time
	// LastModified — дата последнего изменения.
	LastModified *time.Time
	// Sequence — номер ревизии.
	Sequence int

	// RecurrenceID — идентификатор экземпляра повторяющегося окна.
	RecurrenceID *time.Time

	// RRules — правила повторения (RRULE).
	RRules []RecurrenceRule
	// RDates — доп. даты повторения (RDATE).
	RDates []time.Time
	// ExDates — даты-исключения из повторений (EXDATE).
	ExDates []time.Time

	// Alarms — вложенные напоминания (VALARM).
	Alarms []Alarm

	// XProps — нестандартные свойства.
	XProps []Property
	// IanaProps — зарегистрированные свойства IANA.
	IanaProps []Property
}
