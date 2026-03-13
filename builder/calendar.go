// Package builder предоставляет удобный API для конструирования iCalendar-старуктур.
//
// Используется паттерн Functional Options: обязательные параметры передаются
// как аргументы в функцию New*, а опциональные — передаются через функции With*.
// Пример:
//
//	e := builder.NewEvent("uid", dtStart, builder.WithEventSummary("Title"))
package builder

import (
	"gitverse.ru/cloudcoder/ical/model"
)

// CalendarOption — функция для настройки Calendar.
type CalendarOption func(*model.Calendar)

// NewCalendar создаёт новый VCALENDAR.
// По умолчанию задаёт Version="2.0".
func NewCalendar(prodID string, opts ...CalendarOption) *model.Calendar {
	cal := &model.Calendar{
		Version: "2.0",
		ProdID:  prodID,
	}

	for _, opt := range opts {
		opt(cal)
	}

	return cal
}

// WithCalScale задаёт шкалу календаря (например, "GREGORIAN").
func WithCalScale(scale string) CalendarOption {
	return func(c *model.Calendar) {
		c.CalScale = scale
	}
}

// WithMethod задаёт метод iTIP.
func WithMethod(method string) CalendarOption {
	return func(c *model.Calendar) {
		c.Method = method
	}
}

// WithEvents добавляет VEVENT в календарь.
func WithEvents(events ...model.Event) CalendarOption {
	return func(c *model.Calendar) {
		c.Events = append(c.Events, events...)
	}
}

// WithTodos добавляет VTODO в календарь.
func WithTodos(todos ...model.Todo) CalendarOption {
	return func(c *model.Calendar) {
		c.Todos = append(c.Todos, todos...)
	}
}

// WithJournals добавляет VJOURNAL в календарь.
func WithJournals(journals ...model.Journal) CalendarOption {
	return func(c *model.Calendar) {
		c.Journals = append(c.Journals, journals...)
	}
}

// WithTimezones добавляет VTIMEZONE в календарь.
func WithTimezones(timezones ...model.Timezone) CalendarOption {
	return func(c *model.Calendar) {
		c.Timezones = append(c.Timezones, timezones...)
	}
}

// WithFreeBusys добавляет VFREEBUSY в календарь.
func WithFreeBusys(fbs ...model.FreeBusy) CalendarOption {
	return func(c *model.Calendar) {
		c.FreeBusys = append(c.FreeBusys, fbs...)
	}
}

// WithCalendarXProp добавляет X-свойство.
func WithCalendarXProp(name, value string, params ...model.Param) CalendarOption {
	return func(c *model.Calendar) {
		c.XProps = append(c.XProps, model.Property{Name: name, Value: value, Params: params})
	}
}

// WithCalendarIanaProp добавляет зарегистрированное IANA-свойство.
func WithCalendarIanaProp(name, value string, params ...model.Param) CalendarOption {
	return func(c *model.Calendar) {
		c.IanaProps = append(c.IanaProps, model.Property{Name: name, Value: value, Params: params})
	}
}
