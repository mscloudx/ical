package builder

import (
	"time"

	"gitverse.ru/cloudcoder/ical/model"
)

// EventOption — функция настройки Event.
type EventOption func(*model.Event)

// NewEvent создаёт новый VEVENT.
// UID, DTStamp и DTStart обязательны.
func NewEvent(uid string, dtStamp, dtStart time.Time, opts ...EventOption) model.Event {
	e := model.Event{
		UID:     uid,
		DTStamp: dtStamp,
		DTStart: dtStart,
		AllDay:  model.IsDateOnly(dtStart),
	}
	for _, opt := range opts {
		opt(&e)
	}
	return e
}

func WithEventDTEnd(t time.Time) EventOption {
	return func(e *model.Event) {
		e.DTEnd = &t
	}
}

func WithEventDuration(d time.Duration) EventOption {
	return func(e *model.Event) {
		e.Duration = &d
	}
}

func WithEventSummary(s string) EventOption {
	return func(e *model.Event) {
		e.Summary = s
	}
}

func WithEventDescription(s string) EventOption {
	return func(e *model.Event) {
		e.Description = s
	}
}

func WithEventLocation(s string) EventOption {
	return func(e *model.Event) {
		e.Location = s
	}
}

func WithEventURL(s string) EventOption {
	return func(e *model.Event) {
		e.URL = s
	}
}

func WithEventStatus(s model.Status) EventOption {
	return func(e *model.Event) {
		e.Status = s
	}
}

func WithEventTransparency(t model.Transparency) EventOption {
	return func(e *model.Event) {
		e.Transparency = t
	}
}

func WithEventClassification(c model.Classification) EventOption {
	return func(e *model.Event) {
		e.Classification = c
	}
}

func WithEventOrganizer(name, addr string) EventOption {
	return func(e *model.Event) {
		e.Organizer = &model.Attendee{
			Name:    name,
			Address: addr,
		}
	}
}

func WithEventAttendee(a *model.Attendee) EventOption {
	return func(e *model.Event) {
		if a != nil {
			e.Attendees = append(e.Attendees, *a)
		}
	}
}

func WithEventRRule(r *model.RecurrenceRule) EventOption {
	return func(e *model.Event) {
		if r != nil {
			e.RRules = append(e.RRules, *r)
		}
	}
}

func WithEventRDate(dates ...time.Time) EventOption {
	return func(e *model.Event) {
		e.RDates = append(e.RDates, dates...)
	}
}

func WithEventExDate(dates ...time.Time) EventOption {
	return func(e *model.Event) {
		e.ExDates = append(e.ExDates, dates...)
	}
}

func WithEventAlarm(a *model.Alarm) EventOption {
	return func(e *model.Event) {
		if a != nil {
			e.Alarms = append(e.Alarms, *a)
		}
	}
}

func WithEventCreated(t time.Time) EventOption {
	return func(e *model.Event) {
		e.Created = &t
	}
}

func WithEventLastModified(t time.Time) EventOption {
	return func(e *model.Event) {
		e.LastModified = &t
	}
}

func WithEventSequence(seq int) EventOption {
	return func(e *model.Event) {
		e.Sequence = seq
	}
}

func WithEventRelatedTo(uid string, relType model.RelationshipType) EventOption {
	return func(e *model.Event) {
		e.Related = append(e.Related, model.Relation{UID: uid, Type: relType})
	}
}

func WithEventXProp(name, value string, params ...model.Param) EventOption {
	return func(e *model.Event) {
		e.XProps = append(e.XProps, model.Property{Name: name, Value: value, Params: params})
	}
}
