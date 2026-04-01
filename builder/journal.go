package builder

import (
	"time"

	"github.com/mscloudx/ical/model"
)

// JournalOption — функция настройки Journal.
type JournalOption func(*model.Journal)

// NewJournal создаёт новый VJOURNAL.
// UID и DTStamp обязательны.
// Возвращает ErrEmptyUID, если uid пуст.
func NewJournal(uid string, dtStamp time.Time, opts ...JournalOption) (model.Journal, error) {
	if uid == "" {
		return model.Journal{}, ErrEmptyUID
	}
	j := model.Journal{
		UID:     uid,
		DTStamp: dtStamp,
	}
	for _, opt := range opts {
		opt(&j)
	}
	return j, nil
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

// WithJournalIanaProp добавляет зарегистрированное IANA-свойство к Journal.
func WithJournalIanaProp(name, value string, params ...model.Param) JournalOption {
	return func(j *model.Journal) {
		j.IanaProps = append(j.IanaProps, model.Property{Name: name, Value: value, Params: params})
	}
}

// WithJournalParticipant добавляет PARTICIPANT (RFC 9073) к Journal.
func WithJournalParticipant(p ...model.Participant) JournalOption {
	return func(j *model.Journal) {
		j.Participants = append(j.Participants, p...)
	}
}

// WithJournalLocationComponent добавляет LOCATION (RFC 9073) к Journal.
func WithJournalLocationComponent(l ...model.LocationComponent) JournalOption {
	return func(j *model.Journal) {
		j.Locations = append(j.Locations, l...)
	}
}

// WithJournalResourceComponent добавляет RESOURCE (RFC 9073) к Journal.
func WithJournalResourceComponent(r ...model.ResourceComponent) JournalOption {
	return func(j *model.Journal) {
		j.Resources = append(j.Resources, r...)
	}
}

// WithJournalStructuredData добавляет STRUCTURED-DATA (RFC 9073) к Journal.
func WithJournalStructuredData(sd ...model.StructuredData) JournalOption {
	return func(j *model.Journal) {
		j.StructuredData = append(j.StructuredData, sd...)
	}
}

// WithJournalStyledDescription добавляет STYLED-DESCRIPTION (RFC 9073) к Journal.
func WithJournalStyledDescription(sd ...model.StyledDescription) JournalOption {
	return func(j *model.Journal) {
		j.StyledDescriptions = append(j.StyledDescriptions, sd...)
	}
}

// WithJournalAttachment добавляет вложение ATTACH к записи (RFC 5545 §3.8.1.1).
func WithJournalAttachment(a ...model.Attachment) JournalOption {
	return func(j *model.Journal) {
		j.Attach = append(j.Attach, a...)
	}
}
