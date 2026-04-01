// Package builder предоставляет удобный API для конструирования iCalendar-старуктур.
//
// Используется паттерн Functional Options: обязательные параметры передаются
// как аргументы в функцию New*, а опциональные — передаются через функции With*.
// Пример:
//
//	e := builder.NewEvent("uid", dtStart, builder.WithEventSummary("Title"))
package builder

import (
	"time"

	"github.com/mscloudx/ical/model"
)

// CalendarOption — функция для настройки Calendar.
type CalendarOption func(*model.Calendar)

// NewCalendar создаёт новый VCALENDAR.
// По умолчанию задаёт Version="2.0".
// Возвращает ErrEmptyProdID, если prodID пуст.
func NewCalendar(prodID string, opts ...CalendarOption) (*model.Calendar, error) {
	if prodID == "" {
		return nil, ErrEmptyProdID
	}
	cal := &model.Calendar{
		Version: "2.0",
		ProdID:  prodID,
	}

	for _, opt := range opts {
		opt(cal)
	}

	return cal, nil
}

// WithCalScale задаёт шкалу календаря (например, "GREGORIAN").
func WithCalScale(scale string) CalendarOption {
	return func(c *model.Calendar) {
		c.CalScale = scale
	}
}

// WithMethod задаёт метод iTIP (RFC 5546 §3.2).
func WithMethod(method model.Method) CalendarOption {
	return func(c *model.Calendar) {
		c.Method = method
	}
}

// WithCalendarName задаёт текстовое название календаря (RFC 7986).
func WithCalendarName(name string) CalendarOption {
	return func(c *model.Calendar) {
		c.Name = name
	}
}

// WithCalendarDescription задаёт описание календаря (RFC 7986).
func WithCalendarDescription(desc string) CalendarOption {
	return func(c *model.Calendar) {
		c.Description = desc
	}
}

// WithCalendarUID задаёт уникальный идентификатор календаря (RFC 7986).
func WithCalendarUID(uid string) CalendarOption {
	return func(c *model.Calendar) {
		c.UID = uid
	}
}

// WithCalendarLastModified задаёт дату последнего изменения календаря (RFC 7986).
func WithCalendarLastModified(t time.Time) CalendarOption {
	return func(c *model.Calendar) {
		c.LastModified = &t
	}
}

// WithCalendarURL задаёт ссылку на календарь (RFC 7986).
func WithCalendarURL(url string) CalendarOption {
	return func(c *model.Calendar) {
		c.URL = url
	}
}

// WithCalendarCategories задаёт категории календаря (RFC 7986).
func WithCalendarCategories(cats ...string) CalendarOption {
	return func(c *model.Calendar) {
		c.Categories = append(c.Categories, cats...)
	}
}

// WithCalendarRefreshInterval задаёт рекомендуемый интервал обновления (RFC 7986).
func WithCalendarRefreshInterval(d time.Duration) CalendarOption {
	return func(c *model.Calendar) {
		c.RefreshInterval = &d
	}
}

// WithCalendarColor задаёт цвет календаря (RFC 7986).
func WithCalendarColor(color string) CalendarOption {
	return func(c *model.Calendar) {
		c.Color = color
	}
}

// WithCalendarSource задаёт источник календаря URI (RFC 7986).
func WithCalendarSource(source string) CalendarOption {
	return func(c *model.Calendar) {
		c.Source = source
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

// WithAvailabilities добавляет VAVAILABILITY (RFC 7953) в календарь.
func WithAvailabilities(availabilities ...model.Availability) CalendarOption {
	return func(c *model.Calendar) {
		c.Availabilities = append(c.Availabilities, availabilities...)
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
