package builder

import (
	"time"

	"github.com/mscloudx/ical/model"
)

// TodoOption — функция настройки Todo.
type TodoOption func(*model.Todo)

// NewTodo создаёт новый VTODO.
// UID и DTStamp обязательны.
// Возвращает ErrEmptyUID, если uid пуст.
func NewTodo(uid string, dtStamp time.Time, opts ...TodoOption) (model.Todo, error) {
	if uid == "" {
		return model.Todo{}, ErrEmptyUID
	}
	t := model.Todo{
		UID:     uid,
		DTStamp: dtStamp,
	}
	for _, opt := range opts {
		opt(&t)
	}
	return t, nil
}

func WithTodoDTStart(t time.Time) TodoOption {
	return func(td *model.Todo) {
		td.DTStart = &t
		td.AllDay = model.IsDateOnly(t)
	}
}

func WithTodoDue(t time.Time) TodoOption {
	return func(td *model.Todo) {
		td.Due = &t
	}
}

func WithTodoDuration(d time.Duration) TodoOption {
	return func(td *model.Todo) {
		td.Duration = &d
	}
}

func WithTodoCompleted(t time.Time) TodoOption {
	return func(td *model.Todo) {
		td.Completed = &t
	}
}

func WithTodoPriority(p int) TodoOption {
	return func(td *model.Todo) {
		td.Priority = p
	}
}

func WithTodoPercentComplete(p int) TodoOption {
	return func(td *model.Todo) {
		td.PercentComplete = p
	}
}

func WithTodoSummary(s string) TodoOption {
	return func(td *model.Todo) {
		td.Summary = s
	}
}

func WithTodoDescription(s string) TodoOption {
	return func(td *model.Todo) {
		td.Description = s
	}
}

func WithTodoLocation(s string) TodoOption {
	return func(td *model.Todo) {
		td.Location = s
	}
}

func WithTodoURL(s string) TodoOption {
	return func(td *model.Todo) {
		td.URL = s
	}
}

func WithTodoStatus(s model.Status) TodoOption {
	return func(td *model.Todo) {
		td.Status = s
	}
}

func WithTodoClassification(c model.Classification) TodoOption {
	return func(td *model.Todo) {
		td.Classification = c
	}
}

func WithTodoOrganizer(name, addr string) TodoOption {
	return func(td *model.Todo) {
		td.Organizer = &model.Attendee{
			Name:    name,
			Address: addr,
		}
	}
}

func WithTodoAttendee(a *model.Attendee) TodoOption {
	return func(td *model.Todo) {
		if a != nil {
			td.Attendees = append(td.Attendees, *a)
		}
	}
}

func WithTodoRRule(r *model.RecurrenceRule) TodoOption {
	return func(td *model.Todo) {
		if r != nil {
			td.RRules = append(td.RRules, *r)
		}
	}
}

func WithTodoRDate(dates ...time.Time) TodoOption {
	return func(td *model.Todo) {
		td.RDates = append(td.RDates, dates...)
	}
}

func WithTodoExDate(dates ...time.Time) TodoOption {
	return func(td *model.Todo) {
		td.ExDates = append(td.ExDates, dates...)
	}
}

func WithTodoAlarm(a *model.Alarm) TodoOption {
	return func(td *model.Todo) {
		if a != nil {
			td.Alarms = append(td.Alarms, *a)
		}
	}
}

func WithTodoCreated(t time.Time) TodoOption {
	return func(td *model.Todo) {
		td.Created = &t
	}
}

func WithTodoLastModified(t time.Time) TodoOption {
	return func(td *model.Todo) {
		td.LastModified = &t
	}
}

func WithTodoSequence(seq int) TodoOption {
	return func(td *model.Todo) {
		td.Sequence = seq
	}
}

func WithTodoRelatedTo(uid string, relType model.RelationshipType) TodoOption {
	return func(td *model.Todo) {
		td.Related = append(td.Related, model.Relation{UID: uid, Type: relType})
	}
}

func WithTodoXProp(name, value string, params ...model.Param) TodoOption {
	return func(td *model.Todo) {
		td.XProps = append(td.XProps, model.Property{Name: name, Value: value, Params: params})
	}
}

// WithTodoIanaProp добавляет зарегистрированное IANA-свойство к Todo.
func WithTodoIanaProp(name, value string, params ...model.Param) TodoOption {
	return func(td *model.Todo) {
		td.IanaProps = append(td.IanaProps, model.Property{Name: name, Value: value, Params: params})
	}
}

// WithTodoParticipant добавляет PARTICIPANT (RFC 9073) к Todo.
func WithTodoParticipant(p ...model.Participant) TodoOption {
	return func(td *model.Todo) {
		td.Participants = append(td.Participants, p...)
	}
}

// WithTodoLocationComponent добавляет LOCATION (RFC 9073) к Todo.
func WithTodoLocationComponent(l ...model.LocationComponent) TodoOption {
	return func(td *model.Todo) {
		td.Locations = append(td.Locations, l...)
	}
}

// WithTodoResourceComponent добавляет RESOURCE (RFC 9073) к Todo.
func WithTodoResourceComponent(r ...model.ResourceComponent) TodoOption {
	return func(td *model.Todo) {
		td.Resources = append(td.Resources, r...)
	}
}

// WithTodoStructuredData добавляет STRUCTURED-DATA (RFC 9073) к Todo.
func WithTodoStructuredData(sd ...model.StructuredData) TodoOption {
	return func(td *model.Todo) {
		td.StructuredData = append(td.StructuredData, sd...)
	}
}

// WithTodoStyledDescription добавляет STYLED-DESCRIPTION (RFC 9073) к Todo.
func WithTodoStyledDescription(sd ...model.StyledDescription) TodoOption {
	return func(td *model.Todo) {
		td.StyledDescriptions = append(td.StyledDescriptions, sd...)
	}
}

// WithTodoRequestStatus добавляет REQUEST-STATUS к задаче (RFC 5546 §3.6).
func WithTodoRequestStatus(rs ...model.RequestStatus) TodoOption {
	return func(td *model.Todo) {
		td.RequestStatus = append(td.RequestStatus, rs...)
	}
}

// WithTodoAttachment добавляет вложение ATTACH к задаче (RFC 5545 §3.8.1.1).
func WithTodoAttachment(a ...model.Attachment) TodoOption {
	return func(td *model.Todo) {
		td.Attach = append(td.Attach, a...)
	}
}
