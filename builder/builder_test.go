package builder_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gitverse.ru/cloudcoder/ical/builder"
	"gitverse.ru/cloudcoder/ical/encoder"
	"gitverse.ru/cloudcoder/ical/model"
)

// BuilderSuite проверяет корректность конструирования моделей через builder.
type BuilderSuite struct {
	suite.Suite
}

func TestBuilderSuite(t *testing.T) {
	suite.Run(t, new(BuilderSuite))
}

func (s *BuilderSuite) TestBuilder_FullCalendarWithEvent() {
	// Arrange
	dtStart := time.Date(2023, 10, 24, 12, 0, 0, 0, time.UTC)
	dtStamp := dtStart.Add(-24 * time.Hour)

	// Act
	// Конструируем Alarm
	alarm := builder.NewDisplayAlarm(
		builder.TriggerBefore(15*time.Minute),
		"Reminder: Meeting starts soon",
		builder.WithAlarmDuration(5*time.Minute),
		builder.WithAlarmRepeat(2),
	)

	// Конструируем Event
	event := builder.NewEvent(
		"test-event-uid",
		dtStamp,
		dtStart,
		builder.WithEventDuration(1*time.Hour),
		builder.WithEventSummary("Team Sync"),
		builder.WithEventLocation("Room 101"),
		builder.WithEventStatus(model.StatusConfirmed),
		builder.WithEventOrganizer("John Doe", "mailto:john@example.com"),
		builder.WithEventAttendee(&model.Attendee{
			Address: "mailto:jane@example.com",
			Name:    "Jane Doe",
			Role:    model.RoleReqParticipant,
			Status:  model.PartStatAccepted,
		}),
		builder.WithEventAlarm(&alarm),
		builder.WithEventRRule(&model.RecurrenceRule{
			Freq:  model.FreqWeekly,
			Count: 10,
			ByDay: []model.WeekdayNum{{Day: model.Monday}},
		}),
	)

	// Конструируем Calendar
	cal := builder.NewCalendar(
		"-//Builder Test//EN",
		builder.WithEvents(event),
		builder.WithMethod("PUBLISH"),
	)

	// Assert - проверяем саму модель
	s.Equal("2.0", cal.Version)
	s.Equal("-//Builder Test//EN", cal.ProdID)
	s.Equal("PUBLISH", cal.Method)

	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]
	s.Equal("test-event-uid", e.UID)
	s.Equal(dtStart, e.DTStart)
	s.Require().NotNil(e.Duration)
	s.Equal(1*time.Hour, *e.Duration)
	s.Equal("Team Sync", e.Summary)
	s.Equal(model.StatusConfirmed, e.Status)

	s.Require().NotNil(e.Organizer)
	s.Equal("John Doe", e.Organizer.Name)

	s.Require().Len(e.Attendees, 1)
	s.Equal(model.RoleReqParticipant, e.Attendees[0].Role)

	s.Require().Len(e.Alarms, 1)
	s.Equal(model.ActionDisplay, e.Alarms[0].Action)
	s.Require().NotNil(e.Alarms[0].Trigger.Duration)
	s.Equal(-15*time.Minute, *e.Alarms[0].Trigger.Duration)

	// Assert - проверяем кодирование (Round-Trip)
	encoded, err := encoder.MarshalString(cal)
	s.Require().NoError(err)
	s.Contains(encoded, "BEGIN:VCALENDAR")
	s.Contains(encoded, "BEGIN:VEVENT")
	s.Contains(encoded, "BEGIN:VALARM")
	s.Contains(encoded, "TRIGGER:-PT15M")
}

func (s *BuilderSuite) TestBuilder_TodoAndJournal() {
	// Arrange
	dtStamp := time.Date(2023, 10, 24, 12, 0, 0, 0, time.UTC)

	// Act
	todo := builder.NewTodo(
		"todo-uid",
		dtStamp,
		builder.WithTodoSummary("Buy milk"),
		builder.WithTodoPriority(1),
		builder.WithTodoPercentComplete(50),
		builder.WithTodoStatus(model.StatusInProcess),
	)

	journal := builder.NewJournal(
		"journal-uid",
		dtStamp,
		builder.WithJournalSummary("Captain's log"),
		builder.WithJournalDescription("Stardate 41153.7. Our destination is planet Deneb IV."),
		builder.WithJournalStatus(model.StatusFinal),
	)

	cal := builder.NewCalendar(
		"-//Builder Test//EN",
		builder.WithTodos(todo),
		builder.WithJournals(journal),
	)

	// Assert
	s.Require().Len(cal.Todos, 1)
	s.Equal("Buy milk", cal.Todos[0].Summary)
	s.Equal(1, cal.Todos[0].Priority)
	s.Equal(50, cal.Todos[0].PercentComplete)

	s.Require().Len(cal.Journals, 1)
	s.Equal("Captain's log", cal.Journals[0].Summary)
	s.Require().Len(cal.Journals[0].Descriptions, 1)
	s.Contains(cal.Journals[0].Descriptions[0], "Deneb IV")

	// Encode check
	encoded, err := encoder.MarshalString(cal)
	s.Require().NoError(err)
	s.Contains(encoded, "BEGIN:VTODO")
	s.Contains(encoded, "PERCENT-COMPLETE:50")
	s.Contains(encoded, "STATUS:IN-PROCESS")
	s.Contains(encoded, "BEGIN:VJOURNAL")
	s.Contains(encoded, "STATUS:FINAL")
}
