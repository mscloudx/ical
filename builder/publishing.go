package builder

import (
	"time"

	"github.com/mscloudx/ical/model"
)

// ParticipantOption — функция для настройки Participant.
type ParticipantOption func(*model.Participant)

// NewParticipant создаёт новый PARTICIPANT (RFC 9073).
func NewParticipant(uid, calendarAddress string, dtStamp time.Time, opts ...ParticipantOption) *model.Participant {
	p := &model.Participant{
		UID:             uid,
		CalendarAddress: calendarAddress,
		DTStamp:         dtStamp,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// WithParticipantType задаёт тип участника (Например, INDIVIDUAL, GROUP, RESOURCE).
func WithParticipantType(kind string) ParticipantOption {
	return func(p *model.Participant) {
		p.ParticipantType = kind
	}
}

// WithParticipantSummary задаёт заголовок участника.
func WithParticipantSummary(summary string) ParticipantOption {
	return func(p *model.Participant) {
		p.Summary = summary
	}
}

// WithParticipantLocationString задаёт свойство LOCATION для участника.
func WithParticipantLocationString(location string) ParticipantOption {
	return func(p *model.Participant) {
		p.Location = location
	}
}

// LocationComponentOption — функция для настройки LocationComponent.
type LocationComponentOption func(*model.LocationComponent)

// NewLocationComponent создаёт новый LOCATION (RFC 9073).
func NewLocationComponent(uid string, dtStamp time.Time, opts ...LocationComponentOption) *model.LocationComponent {
	l := &model.LocationComponent{
		UID:     uid,
		DTStamp: dtStamp,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// WithLocationComponentName задаёт имя локации.
func WithLocationComponentName(name string) LocationComponentOption {
	return func(l *model.LocationComponent) {
		l.Name = name
	}
}

// WithLocationComponentType задаёт тип локации.
func WithLocationComponentType(types ...string) LocationComponentOption {
	return func(l *model.LocationComponent) {
		l.Type = append(l.Type, types...)
	}
}

// ResourceComponentOption — функция для настройки ResourceComponent.
type ResourceComponentOption func(*model.ResourceComponent)

// NewResourceComponent создаёт новый RESOURCE (RFC 9073).
func NewResourceComponent(uid string, dtStamp time.Time, opts ...ResourceComponentOption) *model.ResourceComponent {
	r := &model.ResourceComponent{
		UID:     uid,
		DTStamp: dtStamp,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// WithResourceComponentName задаёт имя ресурса.
func WithResourceComponentName(name string) ResourceComponentOption {
	return func(r *model.ResourceComponent) {
		r.Name = name
	}
}

// WithResourceComponentType задаёт тип ресурса.
func WithResourceComponentType(types ...string) ResourceComponentOption {
	return func(r *model.ResourceComponent) {
		r.ResourceType = append(r.ResourceType, types...)
	}
}

// NewStructuredData создаёт новое свойство STRUCTURED-DATA.
func NewStructuredData(value, fmtType, schema string) model.StructuredData {
	return model.StructuredData{
		Value:   value,
		FmtType: fmtType,
		Schema:  schema,
	}
}

// NewStyledDescription создаёт новое свойство STYLED-DESCRIPTION.
func NewStyledDescription(value, fmtType, valueType string) model.StyledDescription {
	return model.StyledDescription{
		Value:     value,
		FmtType:   fmtType,
		ValueType: valueType,
	}
}
