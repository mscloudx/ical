package parser_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gitverse.ru/cloudcoder/ical/model"
	"gitverse.ru/cloudcoder/ical/parser"
)

// ParserSuite — набор тестов парсера iCalendar.
type ParserSuite struct {
	suite.Suite
	examplesDir string
}

func TestParserSuite(t *testing.T) {
	suite.Run(t, new(ParserSuite))
}

func (s *ParserSuite) SetupSuite() {
	s.examplesDir = filepath.Join("..", "examples", "ical")
}

// parseFile — вспомогательный метод для чтения и парсинга .ics файла.
func (s *ParserSuite) parseFile(relPath string) *model.Calendar {
	s.T().Helper()
	path := filepath.Join(s.examplesDir, relPath)
	f, err := os.Open(path)
	s.Require().NoError(err, "open file: %s", path)
	defer f.Close()

	cal, err := parser.Parse(f)
	s.Require().NoError(err, "parse file: %s", path)
	s.Require().NotNil(cal)
	return cal
}

// --------------------------------------------------------------------------
// 01_basic
// --------------------------------------------------------------------------

func (s *ParserSuite) TestEmptyVCalendar() {
	// Arrange & Act
	cal := s.parseFile("01_basic/01_empty_vcalendar.ics")

	// Assert
	s.Equal("2.0", cal.Version)
	s.Contains(cal.ProdID, "CloudCoder")
	s.Empty(cal.Events)
}

func (s *ParserSuite) TestSingleVEvent() {
	// Arrange & Act
	cal := s.parseFile("01_basic/02_single_vevent.ics")

	// Assert
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]

	s.Equal("basic-event-001@example.com", e.UID)
	s.Equal("Встреча команды (Basic)", e.Summary)
	s.Equal("Обсуждение архитектуры парсера", e.Description)
	s.Equal("Переговорная 1", e.Location)
	s.False(e.AllDay)

	expected := time.Date(2023, 10, 25, 9, 0, 0, 0, time.UTC)
	s.True(e.DTStart.Equal(expected), "DTStart: want %v, got %v", expected, e.DTStart)

	s.Require().NotNil(e.DTEnd)
	expectedEnd := time.Date(2023, 10, 25, 10, 0, 0, 0, time.UTC)
	s.True(e.DTEnd.Equal(expectedEnd))

	expectedStamp := time.Date(2023, 10, 24, 12, 0, 0, 0, time.UTC)
	s.True(e.DTStamp.Equal(expectedStamp))
}

// --------------------------------------------------------------------------
// 02_timezones
// --------------------------------------------------------------------------

func (s *ParserSuite) TestFloatingTime() {
	// Arrange & Act
	cal := s.parseFile("02_timezones/01_floating_time.ics")

	// Assert
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]
	s.False(e.AllDay)
	s.Equal(9, e.DTStart.Hour())
}

func (s *ParserSuite) TestUTCTime() {
	// Arrange & Act
	cal := s.parseFile("02_timezones/02_utc_time.ics")

	// Assert
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]
	s.Equal(time.UTC, e.DTStart.Location())
}

func (s *ParserSuite) TestLocalTimeWithVTimezone() {
	// Arrange & Act
	cal := s.parseFile("02_timezones/03_local_time_with_vtimezone.ics")

	// Assert
	s.Require().Len(cal.Timezones, 1)
	tz := cal.Timezones[0]
	s.Equal("Europe/Moscow", tz.TZID)
	s.Require().Len(tz.Standards, 1)
	s.Equal("+0400", tz.Standards[0].OffsetFrom)
	s.Equal("+0300", tz.Standards[0].OffsetTo)
	s.Equal("MSK", tz.Standards[0].TZName)

	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]
	s.Equal("Europe/Moscow", e.DTStart.Location().String())
	s.Equal(9, e.DTStart.Hour())
}

// --------------------------------------------------------------------------
// 03_recurrence
// --------------------------------------------------------------------------

func (s *ParserSuite) TestDailyRecurrence() {
	// Arrange & Act
	cal := s.parseFile("03_recurrence/01_daily.ics")

	// Assert
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]
	s.Require().Len(e.RRules, 1)

	rule := e.RRules[0]
	s.Equal(model.FreqDaily, rule.Freq)
	s.Equal(5, rule.Count)
	s.Nil(rule.Until)
}

func (s *ParserSuite) TestWeeklyWithExDate() {
	// Arrange & Act
	cal := s.parseFile("03_recurrence/02_weekly_with_exdate.ics")

	// Assert
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]

	s.Require().Len(e.RRules, 1)
	rule := e.RRules[0]
	s.Equal(model.FreqWeekly, rule.Freq)
	s.Require().Len(rule.ByDay, 1)
	s.Equal(model.Wednesday, rule.ByDay[0].Day)
	s.Equal(0, rule.ByDay[0].Ordinal)
	s.Require().NotNil(rule.Until)

	s.Len(e.ExDates, 2)
	expectedEx1 := time.Date(2023, 11, 1, 12, 0, 0, 0, time.UTC)
	s.True(e.ExDates[0].Equal(expectedEx1))
}

// --------------------------------------------------------------------------
// 04_edge_cases
// --------------------------------------------------------------------------

func (s *ParserSuite) TestLongLinesFolded() {
	// Arrange & Act
	cal := s.parseFile("04_edge_cases/01_long_lines_folded.ics")

	// Assert
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]
	s.NotContains(e.Description, "\n")
	s.Contains(e.Description, "RFC 5545")
}

func (s *ParserSuite) TestMissingDTEndAndDuration() {
	// Arrange & Act
	cal := s.parseFile("04_edge_cases/02_missing_dtend_and_duration.ics")

	// Assert
	s.Require().Len(cal.Events, 2)

	e1 := cal.Events[0]
	s.True(e1.AllDay)
	s.Nil(e1.DTEnd)
	s.Nil(e1.Duration)

	e2 := cal.Events[1]
	s.False(e2.AllDay)
	s.Nil(e2.DTEnd)
}

func (s *ParserSuite) TestEscapedCharacters() {
	// Arrange & Act
	cal := s.parseFile("04_edge_cases/03_escaped_characters.ics")

	// Assert
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]
	s.Contains(e.Summary, "\\;")
	s.Contains(e.Summary, "\\,")
}

// --------------------------------------------------------------------------
// 05_vendor_specific
// --------------------------------------------------------------------------

func (s *ParserSuite) TestGoogleCalendar() {
	// Arrange & Act
	cal := s.parseFile("05_vendor_specific/01_google_calendar.ics")

	// Assert
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]

	s.Equal("google-specific-001@google.com", e.UID)
	s.Equal(model.StatusConfirmed, e.Status)
	s.Equal(0, e.Sequence)

	s.Require().Len(e.XProps, 1)
	s.Equal("X-GOOGLE-CONFERENCE", e.XProps[0].Name)
	s.Contains(e.XProps[0].Value, "meet.google.com")

	s.Require().Len(e.Alarms, 1)
	alarm := e.Alarms[0]
	s.Equal(model.ActionDisplay, alarm.Action)
	s.Require().NotNil(alarm.Trigger.Duration)
	s.Equal(-10*time.Minute, *alarm.Trigger.Duration)
}

func (s *ParserSuite) TestAppleCalendar() {
	// Arrange & Act
	cal := s.parseFile("05_vendor_specific/02_apple_calendar.ics")

	// Assert
	s.Equal("GREGORIAN", cal.CalScale)
	s.Require().Len(cal.Timezones, 1)
	s.Require().Len(cal.Events, 1)

	e := cal.Events[0]
	s.Equal(model.TranspOpaque, e.Transparency)
	s.Equal(0, e.Sequence)
	s.GreaterOrEqual(len(e.XProps), 1)

	s.Require().Len(e.Alarms, 1)
	alarm := e.Alarms[0]
	s.Equal(model.ActionDisplay, alarm.Action)
	s.Require().NotNil(alarm.Trigger.Duration)
	s.Equal(-15*time.Minute, *alarm.Trigger.Duration)
}

// --------------------------------------------------------------------------
// API: ParseBytes, ParseWithClose, mixed line endings
// --------------------------------------------------------------------------

func (s *ParserSuite) TestParseBytes() {
	// Arrange
	data := []byte("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//Test//EN\r\nEND:VCALENDAR\r\n")

	// Act
	cal, err := parser.ParseBytes(data)

	// Assert
	s.Require().NoError(err)
	s.Equal("2.0", cal.Version)
	s.Equal("-//Test//EN", cal.ProdID)
}

func (s *ParserSuite) TestParseWithClose() {
	// Arrange
	path := filepath.Join(s.examplesDir, "01_basic", "02_single_vevent.ics")
	f, err := os.Open(path)
	s.Require().NoError(err)

	// Act
	cal, err := parser.ParseWithClose(f)

	// Assert
	s.Require().NoError(err)
	s.Len(cal.Events, 1)

	// File must be closed after ParseWithClose.
	_, readErr := f.Read(make([]byte, 1))
	s.Error(readErr, "file should be closed after ParseWithClose")
}

func (s *ParserSuite) TestParseLFOnlyLineEndings() {
	// Arrange — LF only, no CRLF.
	data := []byte("BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//Test//EN\nBEGIN:VEVENT\nUID:test-lf@example.com\nDTSTAMP:20231024T120000Z\nDTSTART:20231025T090000Z\nSUMMARY:LF only\nEND:VEVENT\nEND:VCALENDAR\n")

	// Act
	cal, err := parser.ParseBytes(data)

	// Assert
	s.Require().NoError(err)
	s.Require().Len(cal.Events, 1)
	s.Equal("LF only", cal.Events[0].Summary)
}

func (s *ParserSuite) TestSentinelErrors_NoVCalendar() {
	// Arrange
	data := []byte("BEGIN:VEVENT\nEND:VEVENT\n")

	// Act
	_, err := parser.ParseBytes(data)

	// Assert — должна вернуться ошибка ErrNoVCalendar.
	s.Error(err)
}

// --------------------------------------------------------------------------
// 08_vtodo
// --------------------------------------------------------------------------

func (s *ParserSuite) TestSimpleTodo() {
	// Arrange & Act
	cal := s.parseFile("08_vtodo/01_simple_todo.ics")

	// Assert
	s.Require().Len(cal.Todos, 1)
	todo := cal.Todos[0]

	s.Equal("20070313T123432Z-456553@example.com", todo.UID)
	s.Equal("Submit Quebec Income Tax Return for 2006", todo.Summary)
	s.Equal(model.StatusNeedsAction, todo.Status)
	s.Equal(model.ClassConfidential, todo.Classification)
	s.Equal(0, todo.PercentComplete)
	s.Empty(cal.Events)

	// DUE — дата без времени.
	s.Require().NotNil(todo.Due)
	expectedDue := time.Date(2007, 5, 1, 0, 0, 0, 0, time.UTC)
	s.True(todo.Due.Equal(expectedDue), "Due: want %v, got %v", expectedDue, *todo.Due)
}

func (s *ParserSuite) TestTodoWithAlarm() {
	// Arrange & Act
	cal := s.parseFile("08_vtodo/02_todo_with_alarm.ics")

	// Assert
	s.Require().Len(cal.Todos, 1)
	todo := cal.Todos[0]

	s.Equal("Submit Revised Internet-Draft", todo.Summary)
	s.Equal(model.StatusCompleted, todo.Status)
	s.Equal(1, todo.Priority)
	s.Equal(100, todo.PercentComplete)

	// DTSTART.
	s.Require().NotNil(todo.DTStart)
	expectedStart := time.Date(2007, 5, 14, 11, 0, 0, 0, time.UTC)
	s.True(todo.DTStart.Equal(expectedStart))

	// COMPLETED.
	s.Require().NotNil(todo.Completed)
	expectedCompleted := time.Date(2007, 7, 7, 10, 0, 0, 0, time.UTC)
	s.True(todo.Completed.Equal(expectedCompleted))

	// VALARM.
	s.Require().Len(todo.Alarms, 1)
	alarm := todo.Alarms[0]
	s.Equal(model.ActionDisplay, alarm.Action)
	s.Require().NotNil(alarm.Trigger.Duration)
	s.Equal(-30*time.Minute, *alarm.Trigger.Duration)
}

// --------------------------------------------------------------------------
// 09_vjournal
// --------------------------------------------------------------------------

func (s *ParserSuite) TestSimpleJournal() {
	// Arrange & Act
	cal := s.parseFile("09_vjournal/01_simple_journal.ics")

	// Assert
	s.Require().Len(cal.Journals, 1)
	j := cal.Journals[0]

	s.Equal("19970901T130000Z-123405@example.com", j.UID)
	s.Equal("Staff meeting minutes", j.Summary)
	s.Equal(model.StatusFinal, j.Status)
	s.True(j.AllDay)
	s.Empty(cal.Events)

	// DTSTART — дата без времени.
	s.Require().NotNil(j.DTStart)
	expectedDate := time.Date(1997, 3, 17, 0, 0, 0, 0, time.UTC)
	s.True(j.DTStart.Equal(expectedDate))

	// Два блока DESCRIPTION (RFC 5545 §3.6.3 допускает несколько).
	s.Require().Len(j.Descriptions, 2)
	s.Contains(j.Descriptions[0], "Staff meeting")
	s.Contains(j.Descriptions[1], "Telephone Conference")
}
