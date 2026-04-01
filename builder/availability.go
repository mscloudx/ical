package builder

import (
	"time"

	"github.com/mscloudx/ical/model"
)

// AvailabilityOption — функция для настройки Availability.
type AvailabilityOption func(*model.Availability)

// NewAvailability создаёт новый компонент VAVAILABILITY.
func NewAvailability(uid string, dtStamp time.Time, opts ...AvailabilityOption) *model.Availability {
	a := &model.Availability{
		UID:     uid,
		DTStamp: dtStamp,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// WithAvailabilityDTStart задаёт время начала доступности.
func WithAvailabilityDTStart(dtStart time.Time) AvailabilityOption {
	return func(a *model.Availability) {
		a.DTStart = &dtStart
	}
}

// WithAvailabilityDTEnd задаёт время окончания доступности.
func WithAvailabilityDTEnd(dtEnd time.Time) AvailabilityOption {
	return func(a *model.Availability) {
		a.DTEnd = &dtEnd
		a.Duration = nil // DTEnd и Duration взаимоисключающие
	}
}

// WithAvailabilityDuration задаёт длительность доступности.
func WithAvailabilityDuration(duration time.Duration) AvailabilityOption {
	return func(a *model.Availability) {
		a.Duration = &duration
		a.DTEnd = nil // DTEnd и Duration взаимоисключающие
	}
}

// WithAvailabilityBusyType задаёт тип занятости (BUSY, BUSY-UNAVAILABLE, BUSY-TENTATIVE).
func WithAvailabilityBusyType(bt model.BusyType) AvailabilityOption {
	return func(a *model.Availability) {
		a.BusyType = bt
	}
}

// WithAvailabilitySummary задаёт заголовок доступности.
func WithAvailabilitySummary(summary string) AvailabilityOption {
	return func(a *model.Availability) {
		a.Summary = summary
	}
}

// WithAvailabilityDescription задаёт описание доступности.
func WithAvailabilityDescription(desc string) AvailabilityOption {
	return func(a *model.Availability) {
		a.Description = desc
	}
}

// WithAvailabilityAvailable добавляет окна доступности AVAILABLE.
func WithAvailabilityAvailable(available ...model.Available) AvailabilityOption {
	return func(a *model.Availability) {
		a.Available = append(a.Available, available...)
	}
}

// AvailableOption — функция для настройки Available.
type AvailableOption func(*model.Available)

// NewAvailable создаёт новый компонент AVAILABLE.
func NewAvailable(uid string, dtStamp, dtStart time.Time, opts ...AvailableOption) *model.Available {
	a := &model.Available{
		UID:     uid,
		DTStamp: dtStamp,
		DTStart: dtStart,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// WithAvailableDTEnd задаёт время окончания окна доступности.
func WithAvailableDTEnd(dtEnd time.Time) AvailableOption {
	return func(a *model.Available) {
		a.DTEnd = &dtEnd
		a.Duration = nil
	}
}

// WithAvailableDuration задаёт длительность окна доступности.
func WithAvailableDuration(duration time.Duration) AvailableOption {
	return func(a *model.Available) {
		a.Duration = &duration
		a.DTEnd = nil
	}
}

// WithAvailableSummary задаёт заголовок окна доступности.
func WithAvailableSummary(summary string) AvailableOption {
	return func(a *model.Available) {
		a.Summary = summary
	}
}

// WithAvailableDescription задаёт описание окна доступности.
func WithAvailableDescription(desc string) AvailableOption {
	return func(a *model.Available) {
		a.Description = desc
	}
}
