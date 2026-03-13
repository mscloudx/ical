package encoder_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gitverse.ru/cloudcoder/ical/encoder"
	"gitverse.ru/cloudcoder/ical/model"
	"gitverse.ru/cloudcoder/ical/parser"
)

// EncoderSuite проверяет корректность сериализации Calendar → .ics.
type EncoderSuite struct {
	suite.Suite
	examplesDir string
}

func TestEncoderSuite(t *testing.T) {
	suite.Run(t, new(EncoderSuite))
}

func (s *EncoderSuite) SetupSuite() {
	s.examplesDir = filepath.Join("..", "examples", "ical")
}

// --------------------------------------------------------------------------
// Unit Тесты: Форматирование и сериализация
// --------------------------------------------------------------------------

func (s *EncoderSuite) TestEncode_NilCalendar() {
	// Act
	err := encoder.Encode(new(bytes.Buffer), nil)

	// Assert
	s.ErrorIs(err, encoder.ErrNilCalendar)
}

func (s *EncoderSuite) TestEncode_EmptyCalendar() {
	// Arrange
	cal := &model.Calendar{
		Version: "2.0",
		ProdID:  "-//Test//EN",
	}

	// Act
	actual, err := encoder.MarshalString(cal)

	// Assert
	s.Require().NoError(err)
	expected := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//Test//EN\r\nEND:VCALENDAR\r\n"
	s.Equal(expected, actual)
}

func (s *EncoderSuite) TestEncode_LongLineFolding() {
	// Arrange
	longDesc := strings.Repeat("A", 100)
	cal := &model.Calendar{
		ProdID: "-//Test//EN",
		Events: []model.Event{
			{
				UID:         "123",
				Description: longDesc,
			},
		},
	}

	// Act
	actual, err := encoder.MarshalString(cal)

	// Assert
	s.Require().NoError(err)
	// Должно быть разбито на несколько строк с CRLF + пробел
	s.Contains(actual, "\r\n ")

	// Убедимся, что парсер сможет это обратно склеить
	parsed, parseErr := parser.ParseBytes([]byte(actual))
	s.Require().NoError(parseErr)
	s.Require().Len(parsed.Events, 1)
	s.Equal(longDesc, parsed.Events[0].Description)
}

func (s *EncoderSuite) TestEncode_EventWithAllFields() {
	// Arrange
	dtStart := time.Date(2023, 10, 24, 12, 0, 0, 0, time.UTC)
	dtEnd := dtStart.Add(2 * time.Hour)
	dtStamp := dtStart.Add(-24 * time.Hour)

	cal := &model.Calendar{
		ProdID: "-//Test//EN",
		Events: []model.Event{
			{
				UID:      "test-uid",
				DTStamp:  dtStamp,
				DTStart:  dtStart,
				DTEnd:    &dtEnd,
				Summary:  "Test Event",
				Location: "Test Location",
				Status:   model.StatusConfirmed,
				Organizer: &model.Attendee{
					Address: "mailto:org@example.com",
					Name:    "Organizer Name",
				},
				Attendees: []model.Attendee{
					{
						Address: "mailto:att@example.com",
						Role:    model.RoleReqParticipant,
						Status:  model.PartStatAccepted,
					},
				},
				RRules: []model.RecurrenceRule{
					{
						Freq:  model.FreqWeekly,
						Count: 5,
						ByDay: []model.WeekdayNum{
							{Day: model.Monday},
						},
					},
				},
			},
		},
	}

	// Act
	actual, err := encoder.MarshalString(cal)

	// Assert
	s.Require().NoError(err)
	s.Contains(actual, "BEGIN:VCALENDAR")
	s.Contains(actual, "BEGIN:VEVENT")
	s.Contains(actual, "UID:test-uid")
	s.Contains(actual, "DTSTART:20231024T120000Z")
	s.Contains(actual, "DTEND:20231024T140000Z")
	s.Contains(actual, "STATUS:CONFIRMED")
	s.Contains(actual, "ORGANIZER;CN=\"Organizer Name\":mailto:org@example.com")
	s.Contains(actual, "ATTENDEE;ROLE=REQ-PARTICIPANT;PARTSTAT=ACCEPTED:mailto:att@example.com")
	s.Contains(actual, "RRULE:FREQ=WEEKLY;COUNT=5;BYDAY=MO")
	s.Contains(actual, "END:VEVENT")
	s.Contains(actual, "END:VCALENDAR")
}

// --------------------------------------------------------------------------
// Интеграционные Round-Trip Тесты
// --------------------------------------------------------------------------

// assertRoundTrip читает файл, парсит его, кодирует обратно,
// снова парсит и сравнивает структуру (с некоторыми ожидаемыми расхождениями).
func (s *EncoderSuite) assertRoundTrip(relPath string) {
	s.T().Helper()
	path := filepath.Join(s.examplesDir, relPath)
	f, err := os.Open(path)
	s.Require().NoError(err, "open file %s", path)
	defer f.Close()

	// 1. Парсим исходный файл
	original, err := parser.Parse(f)
	s.Require().NoError(err, "parse original %s", path)

	// 2. Сериализуем
	encodedStr, err := encoder.MarshalString(original)
	s.Require().NoError(err, "encode %s", path)

	// 3. Парсим сериализованное
	roundTripped, err := parser.ParseBytes([]byte(encodedStr))
	s.Require().NoError(err, "parse encoded %s", path)

	// 4. Сравниваем ключевые поля
	s.Equal(original.Version, roundTripped.Version)
	s.Equal(original.ProdID, roundTripped.ProdID)

	s.Len(roundTripped.Events, len(original.Events), "events count mismatch")
	for i := range original.Events {
		e1 := original.Events[i]
		e2 := roundTripped.Events[i]
		s.Equal(e1.UID, e2.UID)
		s.Equal(e1.Summary, e2.Summary)
		s.Equal(e1.Description, e2.Description)
		s.Equal(e1.Location, e2.Location)
		s.Equal(e1.Status, e2.Status)

		s.True(e1.DTStart.Equal(e2.DTStart), "DTStart mismatch: %v vs %v", e1.DTStart, e2.DTStart)
		if e1.DTEnd != nil {
			s.Require().NotNil(e2.DTEnd)
			s.True(e1.DTEnd.Equal(*e2.DTEnd), "DTEnd mismatch")
		}
	}

	s.Len(roundTripped.Todos, len(original.Todos), "todos count mismatch")
	for i := range original.Todos {
		t1 := original.Todos[i]
		t2 := roundTripped.Todos[i]
		s.Equal(t1.UID, t2.UID)
		s.Equal(t1.Summary, t2.Summary)
		s.Equal(t1.Status, t2.Status)
	}

	s.Len(roundTripped.Journals, len(original.Journals), "journals count mismatch")
	for i := range original.Journals {
		j1 := original.Journals[i]
		j2 := roundTripped.Journals[i]
		s.Equal(j1.UID, j2.UID)
		s.Equal(j1.Summary, j2.Summary)
		s.ElementsMatch(j1.Descriptions, j2.Descriptions)
	}
}

func (s *EncoderSuite) TestRoundTrip_BasicEvent() {
	s.assertRoundTrip("01_basic/02_single_vevent.ics")
}

func (s *EncoderSuite) TestRoundTrip_RecurrenceDaily() {
	s.assertRoundTrip("03_recurrence/01_daily.ics")
}

func (s *EncoderSuite) TestRoundTrip_TodoWithAlarm() {
	s.assertRoundTrip("08_vtodo/02_todo_with_alarm.ics")
}

func (s *EncoderSuite) TestRoundTrip_SimpleJournal() {
	s.assertRoundTrip("09_vjournal/01_simple_journal.ics")
}
