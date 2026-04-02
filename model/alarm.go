package model

import "time"

// Alarm представляет компонент VALARM — напоминание.
//
// Вкладывается внутрь VEVENT или VTODO. Поддерживает три типа действий:
// DISPLAY (визуальное), AUDIO (звуковое) и EMAIL (email-уведомление).
// RFC 5545 §3.6.6.
type Alarm struct {
	// UID — уникальный идентификатор напоминания (RFC 9074).
	UID string
	// Action — тип действия напоминания (DISPLAY, AUDIO, EMAIL).
	// RFC 5545 §3.8.6.1.
	Action AlarmAction
	// Trigger — момент срабатывания напоминания.
	// RFC 5545 §3.8.6.3.
	Trigger Trigger
	// Description — текст напоминания (для DISPLAY и EMAIL).
	// RFC 5545 §3.8.1.5.
	Description string
	// Summary — тема email (только для ACTION:EMAIL).
	// RFC 5545 §3.8.1.12.
	Summary string

	// Attendees — получатели email (только для ACTION:EMAIL).
	// RFC 5545 §3.8.4.1.
	Attendees []Attendee
	// Attach — вложения (только для ACTION:EMAIL).
	// RFC 5545 §3.8.1.1.
	Attach []Attachment

	// Duration — интервал повтора напоминания (nil, если не повторяется).
	// Используется совместно с Repeat.
	// RFC 5545 §3.8.6.2.
	Duration *time.Duration
	// Repeat — количество повторов напоминания.
	// RFC 5545 §3.8.6.2.
	Repeat int

	// Acknowledged — время подтверждения напоминания (RFC 9074).
	Acknowledged *time.Time
	// Proximity — proximity-триггер (RFC 9074).
	Proximity Proximity
	// Related — связи с другими VALARM (RELATED-TO;RELTYPE=SNOOZE, RFC 9074).
	Related []Relation
	// Locations — proximity-локации (VLOCATION, RFC 9074 + RFC 9073).
	Locations []LocationComponent

	// XProps — нестандартные свойства (X-*).
	XProps []Property
	// IanaProps — зарегистрированные IANA свойства без префикса X-.
	IanaProps []Property
}

// Trigger — триггер срабатывания напоминания.
//
// Задаётся либо как Duration (смещение относительно события),
// либо как абсолютное DateTime.
// Отрицательный Duration означает «до начала события».
// RFC 5545 §3.8.6.3.
type Trigger struct {
	// Duration — смещение относительно начала/конца компонента.
	// Отрицательное значение = до события, положительное = после.
	Duration *time.Duration
	// DateTime — абсолютная дата и время срабатывания.
	DateTime *time.Time
	// Related — к чему привязан Duration: "START" (по умолчанию) или "END".
	// Задаётся параметром RELATED=END. Пустая строка означает START.
	// RFC 5545 §3.2.14.
	Related string
}
