package builder

import (
	"time"

	"gitverse.ru/cloudcoder/ical/model"
)

// JournalOption — функция настройки Journal.
type JournalOption func(*model.Journal)

// NewJournal создаёт новый VJOURNAL.
// UID и DTStamp обязательны.
func NewJournal(uid string, dtStamp time.Time, opts ...JournalOption) model.Journal {
	j := model.Journal{
		UID:     uid,
		DTStamp: dtStamp,
	}
	for _, opt := range opts {
		opt(&j)
	}
	return j
}

func WithJournalDTStart(t time.Time) JournalOption {
	return func(j *model.Journal) {
		j.DTStart = &t
		j.AllDay = model.IsDateOnly(t)
	}
}

func WithJournalSummary(s string) JournalOption {
	return func(j *model.Journal) {
		j.Summary = s
	}
}

func WithJournalDescription(s string) JournalOption {
	return func(j *model.Journal) {
		j.Descriptions = append(j.Descriptions, s)
	}
}

func WithJournalURL(s string) JournalOption {
	return func(j *model.Journal) {
		j.URL = s
	}
}

func WithJournalStatus(s model.Status) JournalOption {
	return func(j *model.Journal) {
		j.Status = s
	}
}

func WithJournalClassification(c model.Classification) JournalOption {
	return func(j *model.Journal) {
		j.Classification = c
	}
}

func WithJournalOrganizer(name, addr string) JournalOption {
	return func(j *model.Journal) {
		j.Organizer = &model.Attendee{
			Name:    name,
			Address: addr,
		}
	}
}

func WithJournalAttendee(a *model.Attendee) JournalOption {
	return func(j *model.Journal) {
		if a != nil {
			j.Attendees = append(j.Attendees, *a)
		}
	}
}

func WithJournalRRule(r *model.RecurrenceRule) JournalOption {
	return func(j *model.Journal) {
		if r != nil {
			j.RRules = append(j.RRules, *r)
		}
	}
}

func WithJournalRDate(dates ...time.Time) JournalOption {
	return func(j *model.Journal) {
		j.RDates = append(j.RDates, dates...)
	}
}

func WithJournalExDate(dates ...time.Time) JournalOption {
	return func(j *model.Journal) {
		j.ExDates = append(j.ExDates, dates...)
	}
}

func WithJournalCreated(t time.Time) JournalOption {
	return func(j *model.Journal) {
		j.Created = &t
	}
}

func WithJournalLastModified(t time.Time) JournalOption {
	return func(j *model.Journal) {
		j.LastModified = &t
	}
}

func WithJournalSequence(seq int) JournalOption {
	return func(j *model.Journal) {
		j.Sequence = seq
	}
}

func WithJournalRelatedTo(uid string, relType model.RelationshipType) JournalOption {
	return func(j *model.Journal) {
		j.Related = append(j.Related, model.Relation{UID: uid, Type: relType})
	}
}

func WithJournalXProp(name, value string, params ...model.Param) JournalOption {
	return func(j *model.Journal) {
		j.XProps = append(j.XProps, model.Property{Name: name, Value: value, Params: params})
	}
}
