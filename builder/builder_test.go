package builder_test

import (
	"testing"
	"time"

	"github.com/mscloudx/ical/builder"
	"github.com/mscloudx/ical/encoder"
	"github.com/mscloudx/ical/model"
	"github.com/mscloudx/ical/parser"
	"github.com/stretchr/testify/suite"
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
	event, err := builder.NewEvent(
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
	s.Require().NoError(err)

	// Конструируем Calendar
	cal, err := builder.NewCalendar(
		"-//Builder Test//EN",
		builder.WithEvents(event),
		builder.WithMethod(model.MethodPublish),
	)
	s.Require().NoError(err)

	// Assert - проверяем саму модель
	s.Equal("2.0", cal.Version)
	s.Equal("-//Builder Test//EN", cal.ProdID)
	s.Equal(model.MethodPublish, cal.Method)

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
	todo, err := builder.NewTodo(
		"todo-uid",
		dtStamp,
		builder.WithTodoSummary("Buy milk"),
		builder.WithTodoPriority(1),
		builder.WithTodoPercentComplete(50),
		builder.WithTodoStatus(model.StatusInProcess),
	)
	s.Require().NoError(err)

	journal, err := builder.NewJournal(
		"journal-uid",
		dtStamp,
		builder.WithJournalSummary("Captain's log"),
		builder.WithJournalDescription("Stardate 41153.7. Our destination is planet Deneb IV."),
		builder.WithJournalStatus(model.StatusFinal),
	)
	s.Require().NoError(err)

	cal, err := builder.NewCalendar(
		"-//Builder Test//EN",
		builder.WithTodos(todo),
		builder.WithJournals(journal),
	)
	s.Require().NoError(err)

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

// TestBuilder_RFC6868ParameterEncoding проверяет полный цикл (builder → encoder → parser)
// для параметров с RFC 6868 ^-escape: кавычки, переносы строк, каретки.
func (s *BuilderSuite) TestBuilder_RFC6868ParameterEncoding() {
	// Arrange
	dtStamp := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	dtStart := time.Date(2024, 3, 15, 14, 0, 0, 0, time.UTC)

	event, err := builder.NewEvent(
		"rfc6868-builder@example.com",
		dtStamp,
		dtStart,
		builder.WithEventSummary("RFC 6868 Builder Test"),
		builder.WithEventAttendee(&model.Attendee{
			Address: "mailto:babe@example.com",
			Name:    `George Herman "Babe" Ruth`,
			Params: []model.Param{
				{Name: "X-NOTE", Values: []string{"Line1\nLine2"}},
				{Name: "X-CARETS", Values: []string{"Hat ^Carets^"}},
			},
		}),
	)
	s.Require().NoError(err)

	cal, err := builder.NewCalendar("-//Builder RFC6868 Test//EN", builder.WithEvents(event))
	s.Require().NoError(err)

	// Act: encode
	encoded, err := encoder.MarshalString(cal)

	// Assert: encoder применил RFC 6868 escape.
	// Строка может быть разбита фолдингом, поэтому проверяем отдельные фрагменты.
	s.Require().NoError(err)
	s.Contains(encoded, "CN=\"George Herman ^'Babe^' Ruth\"")
	s.Contains(encoded, "X-NOTE=Line1^nLine2")
	s.Contains(encoded, "^^Carets^^")

	// Act: round-trip — парсим обратно
	roundTripped, err := parser.ParseBytes([]byte(encoded))

	// Assert: декодирование восстанавливает оригинальные значения
	s.Require().NoError(err)
	s.Require().Len(roundTripped.Events, 1)
	s.Require().Len(roundTripped.Events[0].Attendees, 1)

	a := roundTripped.Events[0].Attendees[0]
	s.Equal(`George Herman "Babe" Ruth`, a.Name)

	paramMap := make(map[string]string, len(a.Params))
	for _, p := range a.Params {
		if len(p.Values) > 0 {
			paramMap[p.Name] = p.Values[0]
		}
	}
	s.Equal("Line1\nLine2", paramMap["X-NOTE"])
	s.Equal("Hat ^Carets^", paramMap["X-CARETS"])
}

// TestBuilder_ValidationErrors проверяет, что builder возвращает ошибки при пустых обязательных полях.
func (s *BuilderSuite) TestBuilder_ValidationErrors() {
	dtStamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := builder.NewCalendar("")
	s.Require().ErrorIs(err, builder.ErrEmptyProdID)

	_, err = builder.NewEvent("", dtStamp, dtStamp)
	s.Require().ErrorIs(err, builder.ErrEmptyUID)

	_, err = builder.NewTodo("", dtStamp)
	s.Require().ErrorIs(err, builder.ErrEmptyUID)

	_, err = builder.NewJournal("", dtStamp)
	s.Require().ErrorIs(err, builder.ErrEmptyUID)
}

// TestBuilder_AttachmentRoundTrip проверяет полный цикл для ATTACH: URI и BASE64.
func (s *BuilderSuite) TestBuilder_AttachmentRoundTrip() {
	// Arrange
	dtStamp := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	dtStart := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)

	pdfData := []byte("fake-pdf-content")

	event, err := builder.NewEvent(
		"attach-test-uid",
		dtStamp,
		dtStart,
		builder.WithEventSummary("Attachment Test"),
		builder.WithEventAttachment(
			model.Attachment{URI: "https://example.com/doc.pdf", MIMEType: "application/pdf"},
			model.Attachment{Data: pdfData, MIMEType: "application/pdf"},
		),
	)
	s.Require().NoError(err)

	cal, err := builder.NewCalendar("-//Attach Test//EN", builder.WithEvents(event))
	s.Require().NoError(err)

	// Act: encode
	encoded, err := encoder.MarshalString(cal)
	s.Require().NoError(err)

	// Assert: оба ATTACH-свойства присутствуют
	s.Contains(encoded, "ATTACH;FMTTYPE=application/pdf:https://example.com/doc.pdf")
	s.Contains(encoded, "ENCODING=BASE64")

	// Round-trip: парсим обратно
	roundTripped, err := parser.ParseBytes([]byte(encoded))
	s.Require().NoError(err)
	s.Require().Len(roundTripped.Events, 1)

	attachments := roundTripped.Events[0].Attach
	s.Require().Len(attachments, 2)

	// URI вложение
	s.Equal("https://example.com/doc.pdf", attachments[0].URI)
	s.Equal("application/pdf", attachments[0].MIMEType)
	s.Empty(attachments[0].Data)

	// BASE64 вложение
	s.Equal(pdfData, attachments[1].Data)
	s.Equal("application/pdf", attachments[1].MIMEType)
	s.Empty(attachments[1].URI)
}
