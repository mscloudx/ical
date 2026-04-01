package encoder_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mscloudx/ical/encoder"
	"github.com/mscloudx/ical/model"
	"github.com/mscloudx/ical/parser"
	"github.com/stretchr/testify/suite"
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
	s.assertRoundTrip("10_vtodo/02_todo_with_alarm.ics")
}

func (s *EncoderSuite) TestRoundTrip_SimpleJournal() {
	s.assertRoundTrip("11_vjournal/01_simple_journal.ics")
}

func (s *EncoderSuite) TestRoundTrip_RFC7986Properties() {
	s.assertRoundTrip("05_extensions/01_rfc7986_properties.ics")
}

func (s *EncoderSuite) TestRoundTrip_RFC7953Availability() {
	s.assertRoundTrip("05_extensions/02_rfc7953_availability.ics")
}

func (s *EncoderSuite) TestRoundTrip_RFC9073Publishing() {
	s.assertRoundTrip("05_extensions/04_rfc9073_publishing.ics")
}

func (s *EncoderSuite) TestRoundTrip_RFC9253Relationships() {
	s.assertRoundTrip("05_extensions/05_rfc9253_relationships.ics")
}

func (s *EncoderSuite) TestRoundTrip_RFC9074Alarms() {
	s.assertRoundTrip("05_extensions/09_rfc9074_alarms.ics")
}

func (s *EncoderSuite) TestRoundTrip_RFC7529RScale() {
	s.assertRoundTrip("05_extensions/07_rfc7529_rscale.ics")
}

func (s *EncoderSuite) TestEncode_ParticipantPropertyOrder() {
	// Arrange
	cal := &model.Calendar{
		ProdID: "-//Test//EN",
		Events: []model.Event{
			{
				UID: "event-1",
				Participants: []model.Participant{
					{
						UID:             "part-1",
						ParticipantType: "INDIVIDUAL",
						DTStamp:         time.Date(2023, 10, 24, 12, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	}

	// Act
	actual, err := encoder.MarshalString(cal)

	// Assert
	s.Require().NoError(err)

	// Check order: BEGIN:PARTICIPANT followed by UID then DTSTAMP
	lines := strings.Split(actual, "\r\n")
	foundBegin := false
	for i, line := range lines {
		if line != "BEGIN:PARTICIPANT" {
			continue
		}
		foundBegin = true
		s.Require().Greater(len(lines), i+2)
		s.Equal("UID:part-1", lines[i+1])
		s.Equal("DTSTAMP:20231024T120000Z", lines[i+2])
		break
	}
	s.True(foundBegin, "PARTICIPANT block not found")
}

func (s *EncoderSuite) TestEncode_ParticipantOrder_ABNF() {
	// Arrange
	cal := &model.Calendar{
		ProdID: "-//Test//EN",
		Events: []model.Event{
			{
				UID: "event-1",
				Participants: []model.Participant{
					{
						UID:             "part-1",
						ParticipantType: "INDIVIDUAL",
						DTStamp:         time.Date(2023, 10, 24, 12, 0, 0, 0, time.UTC),
						XProps: []model.Property{
							{Name: "X-TEST", Value: "test-value"},
						},
						Locations: []model.LocationComponent{
							{UID: "loc-1", Name: "Room 1"},
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

	// Check order: X-TEST must be BEFORE BEGIN:LOCATION
	lines := strings.Split(actual, "\r\n")
	idxXProp := -1
	idxBeginLoc := -1
	for i, line := range lines {
		if line == "X-TEST:test-value" {
			idxXProp = i
		}
		if line == "BEGIN:LOCATION" {
			idxBeginLoc = i
		}
	}
	s.NotEqual(-1, idxXProp, "X-TEST not found")
	s.NotEqual(-1, idxBeginLoc, "BEGIN:LOCATION not found")
	s.Less(idxXProp, idxBeginLoc, "X-TEST must be before BEGIN:LOCATION")
}

func (s *EncoderSuite) TestEncode_RFC6868ParameterValueEncoding() {
	// Arrange
	cal := &model.Calendar{
		ProdID: "-//Test//EN",
		Events: []model.Event{
			{
				UID:     "event-1",
				DTStamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				DTStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Attendees: []model.Attendee{
					{
						Address: "mailto:babe@example.com",
						Name:    "George Herman \"Babe\" Ruth",
						Params: []model.Param{
							{Name: "X-NOTE", Values: []string{"Line1\nLine2"}},
							{Name: "X-CARETS", Values: []string{"Hat ^Carets^"}},
							{Name: "X-RAW", Values: []string{"Keep ^x here"}},
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
	s.Contains(actual, "CN=\"George Herman ^'Babe^' Ruth\"")
	s.Contains(actual, "X-NOTE=Line1^nLine2")
	s.Contains(actual, "X-CARETS=\"Hat")
	s.Contains(actual, "^^Carets^^\"")
	s.Contains(actual, "X-RAW=\"Keep ^^x here\"")
}

func (s *EncoderSuite) TestEncode_RFC7529RScaleAndSkip() {
	// Arrange
	cal := &model.Calendar{
		ProdID: "-//Test//EN",
		Events: []model.Event{
			{
				UID:     "event-rrule-1",
				DTStamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				DTStart: time.Date(2024, 1, 5, 9, 0, 0, 0, time.UTC),
				RRules: []model.RecurrenceRule{
					{
						Freq:       model.FreqMonthly,
						RScale:     model.RecurrenceScale("HEBREW"),
						Skip:       model.SkipForward,
						Count:      3,
						ByMonthDay: []int{30},
					},
				},
			},
		},
	}

	// Act
	actual, err := encoder.MarshalString(cal)

	// Assert
	s.Require().NoError(err)
	s.Contains(actual, "RRULE:FREQ=MONTHLY;RSCALE=HEBREW;SKIP=FORWARD;COUNT=3;BYMONTHDAY=30")
}

func (s *EncoderSuite) TestRoundTrip_RFC6868ParamEncoding() {
	s.assertRoundTrip("05_extensions/11_rfc6868_param_encoding.ics")
}

// --------------------------------------------------------------------------
// TEXT escaping / unescaping (RFC 5545 §3.3.11)
// --------------------------------------------------------------------------

func (s *EncoderSuite) TestEncode_TextEscaping() {
	// Arrange: SUMMARY и DESCRIPTION содержат спецсимволы TEXT-типа.
	cal := &model.Calendar{
		ProdID: "-//Test//EN",
		Events: []model.Event{
			{
				UID:         "escape-test-001",
				Summary:     `Meeting; Important, Very important`,
				Description: "Line one\nLine two\nBackslash: \\",
				Location:    `Room "Alpha", Floor 3`,
			},
		},
	}

	// Act
	actual, err := encoder.MarshalString(cal)

	// Assert: спецсимволы должны быть экранированы в выводе.
	s.Require().NoError(err)
	s.Contains(actual, `SUMMARY:Meeting\; Important\, Very important`)
	s.Contains(actual, `Line one\nLine two`)
	s.Contains(actual, `Backslash: \\`)
	// Запятая в LOCATION экранируется.
	s.Contains(actual, `Room "Alpha"\, Floor 3`)
}

func (s *EncoderSuite) TestRoundTrip_TextEscaping() {
	// Arrange: данные со спецсимволами.
	original := &model.Calendar{
		ProdID: "-//Test//EN",
		Events: []model.Event{
			{
				UID:         "round-trip-escape-001",
				Summary:     `Status; Done, Archived`,
				Description: "First line\nSecond line\nSlash: \\",
				Location:    `Main hall, Building A`,
			},
		},
	}

	// Act: кодируем → парсим обратно.
	encoded, err := encoder.MarshalString(original)
	s.Require().NoError(err)

	parsed, err := parser.ParseBytes([]byte(encoded))
	s.Require().NoError(err)
	s.Require().Len(parsed.Events, 1)

	// Assert: значения должны совпасть после round-trip.
	e := parsed.Events[0]
	s.Equal(original.Events[0].Summary, e.Summary)
	s.Equal(original.Events[0].Description, e.Description)
	s.Equal(original.Events[0].Location, e.Location)
}

func (s *EncoderSuite) TestEncode_UTF8LineFolding() {
	// Arrange: SUMMARY из кириллических символов длиной > 75 байт.
	// Каждый символ кириллицы — 2 байта в UTF-8.
	// 40 символов × 2 байта = 80 байт → должен быть фолдинг.
	cyrillicSummary := strings.Repeat("Б", 40) // 80 байт

	cal := &model.Calendar{
		ProdID: "-//Test//EN",
		Events: []model.Event{
			{
				UID:     "utf8-fold-001",
				Summary: cyrillicSummary,
			},
		},
	}

	// Act
	actual, err := encoder.MarshalString(cal)

	// Assert: фолдинг произошёл, но round-trip должен восстановить оригинал.
	s.Require().NoError(err)
	s.Contains(actual, "\r\n ")

	parsed, parseErr := parser.ParseBytes([]byte(actual))
	s.Require().NoError(parseErr)
	s.Require().Len(parsed.Events, 1)
	s.Equal(cyrillicSummary, parsed.Events[0].Summary)
}

func (s *EncoderSuite) TestRoundTrip_EscapedCharactersFile() {
	// Файл содержит \; \, \n \\ в SUMMARY, DESCRIPTION, LOCATION.
	// После parse→encode→parse значения должны совпасть.
	path := filepath.Join(s.examplesDir, "04_edge_cases", "03_escaped_characters.ics")
	f, err := os.Open(path)
	s.Require().NoError(err)
	defer f.Close()

	original, err := parser.Parse(f)
	s.Require().NoError(err)
	s.Require().Len(original.Events, 1)

	encoded, err := encoder.MarshalString(original)
	s.Require().NoError(err)

	roundTripped, err := parser.ParseBytes([]byte(encoded))
	s.Require().NoError(err)
	s.Require().Len(roundTripped.Events, 1)

	e1 := original.Events[0]
	e2 := roundTripped.Events[0]
	s.Equal(e1.Summary, e2.Summary)
	s.Equal(e1.Description, e2.Description)
	s.Equal(e1.Location, e2.Location)
}
