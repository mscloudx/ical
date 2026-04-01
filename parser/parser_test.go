package parser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mscloudx/ical/model"
	"github.com/mscloudx/ical/parser"
	"github.com/stretchr/testify/suite"
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

// readFileBytes — вспомогательный метод для чтения .ics файла как []byte.
func (s *ParserSuite) readFileBytes(relPath string) []byte {
	s.T().Helper()
	path := filepath.Join(s.examplesDir, relPath)
	data, err := os.ReadFile(path)
	s.Require().NoError(err, "read file: %s", path)
	return data
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
	s.Contains(cal.ProdID, "mscloudx")
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

	// Assert: RFC 5545 §3.3.11 — экранирование должно быть снято парсером.
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]

	// \; → ;  и  \, → ,
	s.Equal("Встреча; Важное, Очень важное", e.Summary)

	// \n → реальный перенос строки, \, → запятая, \; → точка с запятой, \\ → \
	s.Contains(e.Description, "\n")
	s.Contains(e.Description, ",")
	s.Contains(e.Description, ";")
	s.Contains(e.Description, `\`)

	// \, → запятая в LOCATION
	s.Contains(e.Location, ",")
}

// --------------------------------------------------------------------------
// 06_vendor_specific
// --------------------------------------------------------------------------

func (s *ParserSuite) TestGoogleCalendar() {
	// Arrange & Act
	cal := s.parseFile("06_vendor_specific/01_google_calendar.ics")

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
	cal := s.parseFile("06_vendor_specific/02_apple_calendar.ics")

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
	data := s.readFileBytes("00_testing/parse_bytes_basic.ics")

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
	data := s.readFileBytes("00_testing/parse_bytes_lf_only.ics")

	// Act
	cal, err := parser.ParseBytes(data)

	// Assert
	s.Require().NoError(err)
	s.Require().Len(cal.Events, 1)
	s.Equal("LF only", cal.Events[0].Summary)
}

func (s *ParserSuite) TestSentinelErrors_NoVCalendar() {
	// Arrange
	data := s.readFileBytes("00_testing/parse_bytes_no_vcalendar.ics")

	// Act
	_, err := parser.ParseBytes(data)

	// Assert — должна вернуться ошибка ErrNoVCalendar.
	s.Error(err)
}

// --------------------------------------------------------------------------
// 10_vtodo
// --------------------------------------------------------------------------

func (s *ParserSuite) TestSimpleTodo() {
	// Arrange & Act
	cal := s.parseFile("10_vtodo/01_simple_todo.ics")

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
	cal := s.parseFile("10_vtodo/02_todo_with_alarm.ics")

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
// 11_vjournal
// --------------------------------------------------------------------------

func (s *ParserSuite) TestSimpleJournal() {
	// Arrange & Act
	cal := s.parseFile("11_vjournal/01_simple_journal.ics")

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

// --------------------------------------------------------------------------
// 10_extensions (RFC 7986, etc)
// --------------------------------------------------------------------------

func (s *ParserSuite) TestRFC7986Properties() {
	// Arrange & Act
	cal := s.parseFile("05_extensions/01_rfc7986_properties.ics")

	// Assert
	s.Equal("Рабочий календарь", cal.Name)
	s.Equal("Мой основной рабочий календарь", cal.Description)
	s.Equal("user123@example.com", cal.UID)

	s.Require().NotNil(cal.LastModified)
	expectedLM := time.Date(2023, 10, 15, 12, 0, 0, 0, time.UTC)
	s.True(cal.LastModified.Equal(expectedLM))

	s.Equal("https://example.com/calendar.ics", cal.URL)

	s.Require().Len(cal.Categories, 2)
	s.Equal("WORK", cal.Categories[0])
	s.Equal("IMPORTANT", cal.Categories[1])

	s.Require().NotNil(cal.RefreshInterval)
	s.Equal(1*time.Hour, *cal.RefreshInterval)

	s.Equal("blue", cal.Color)
	s.Equal("https://example.com/calendar.ics", cal.Source)

	// Проверяем, что вложенные события тоже парсятся
	s.Require().Len(cal.Events, 1)
	s.Equal("event1@example.com", cal.Events[0].UID)
}

func (s *ParserSuite) TestRFC7953Availability() {
	// Arrange & Act
	cal := s.parseFile("05_extensions/02_rfc7953_availability.ics")

	// Assert
	s.Require().Len(cal.Availabilities, 1)
	av := cal.Availabilities[0]

	s.Equal("avail1@example.com", av.UID)
	s.Equal(model.BusyTypeBusyUnavailable, av.BusyType)
	s.Equal("Ноябрь недоступен, кроме особых часов", av.Summary)

	s.Require().NotNil(av.DTStart)
	expectedStart := time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC)
	s.True(av.DTStart.Equal(expectedStart))

	s.Require().NotNil(av.DTEnd)
	expectedEnd := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)
	s.True(av.DTEnd.Equal(expectedEnd))

	// Вложенные AVAILABLE окна
	s.Require().Len(av.Available, 2)

	// Окно 1 (Дни: MO,WE,FR; Длительность: PT3H)
	w1 := av.Available[0]
	s.Equal("avail-window1@example.com", w1.UID)
	s.Equal("Утренние окна", w1.Summary)

	s.Require().NotNil(w1.Duration)
	s.Equal(3*time.Hour, *w1.Duration)

	s.Require().Len(w1.RRules, 1)
	s.Equal(model.FreqWeekly, w1.RRules[0].Freq)
	s.Len(w1.RRules[0].ByDay, 3)

	// Окно 2 (Дни: TU,TH; Длительность: PT2H)
	w2 := av.Available[1]
	s.Equal("avail-window2@example.com", w2.UID)
	s.Equal("Послеобеденные окна", w2.Summary)

	s.Require().NotNil(w2.Duration)
	s.Equal(2*time.Hour, *w2.Duration)

	s.Require().Len(w2.RRules, 1)
	s.Len(w2.RRules[0].ByDay, 2)
}

func (s *ParserSuite) TestRFC7953AvailabilityInvalid() {
	// Arrange
	path := filepath.Join(s.examplesDir, "05_extensions", "03_rfc7953_availability_invalid.ics")
	f, err := os.Open(path)
	s.Require().NoError(err, "open file: %s", path)
	defer f.Close()

	// Act
	cal, err := parser.Parse(f)

	// Assert
	s.Require().NoError(err, "parsing should succeed by resolving the conflict")
	s.Require().Len(cal.Availabilities, 1)

	av := cal.Availabilities[0]
	// Since DURATION comes after DTEND in the 03_rfc7953_availability_invalid.ics file,
	// DTEND should have been cleared, and Duration set.
	s.Nil(av.DTEnd)
	s.Require().NotNil(av.Duration)
	s.Equal(3*time.Hour, *av.Duration)
}

func (s *ParserSuite) TestRFC9073Publishing() {
	// Arrange & Act
	cal := s.parseFile("05_extensions/04_rfc9073_publishing.ics")

	// Assert
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]

	// Свойства STYLED-DESCRIPTION и STRUCTURED-DATA
	s.Require().Len(e.StyledDescriptions, 1)
	s.Equal("https://example.com/desc.html", e.StyledDescriptions[0].Value)
	s.Equal("URI", e.StyledDescriptions[0].ValueType)

	s.Require().Len(e.StructuredData, 1)
	s.Contains(e.StructuredData[0].Value, "Product Launch")
	s.Equal("application/json", e.StructuredData[0].FmtType)
	s.Equal("https://schema.org/Event", e.StructuredData[0].Schema)

	// Компонент PARTICIPANT
	s.Require().Len(e.Participants, 1)
	part := e.Participants[0]
	s.Equal("part1@example.com", part.UID)
	s.Equal("INDIVIDUAL", part.ParticipantType)
	s.Equal("mailto:john@example.com", part.CalendarAddress)
	s.Equal("Джон Доу (Спикер)", part.Summary)

	// Компонент LOCATION
	s.Require().Len(e.Locations, 1)
	loc := e.Locations[0]
	s.Equal("loc1@example.com", loc.UID)
	s.Equal("Main Hall", loc.Name)
	s.Equal("ROOM", loc.Type[0])

	// Компонент RESOURCE
	s.Require().Len(e.Resources, 1)
	res := e.Resources[0]
	s.Equal("res1@example.com", res.UID)
	s.Equal("Projector", res.Name)
	s.Equal("PROJECTOR", res.ResourceType[0])
}

func (s *ParserSuite) TestRFC9253Relationships() {
	// Arrange & Act
	cal := s.parseFile("05_extensions/05_rfc9253_relationships.ics")

	// Assert: VEVENT
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]
	s.Require().Len(e.Concepts, 1)
	s.Equal("https://example.com/concept/meeting", e.Concepts[0])
	s.Require().Len(e.RefIDs, 1)
	s.Equal("ref-1", e.RefIDs[0])
	s.Require().Len(e.Links, 2)
	s.Equal("https://example.com/event", e.Links[0].Value)
	s.Equal("URI", e.Links[0].ValueType)
	s.Equal("describedby", e.Links[0].Rel)
	s.Equal("text/html", e.Links[0].FmtType)
	s.Equal("Event Page", e.Links[0].Label)
	s.Equal("ru", e.Links[0].Language)
	s.Equal("https://example.com/multi", e.Links[1].Value)
	s.Require().Len(e.Links[1].Params, 2)
	s.Equal("LINKREL", e.Links[1].Params[0].Name)
	s.ElementsMatch([]string{"related", "alternate"}, e.Links[1].Params[0].Values)
	s.Equal("LABEL", e.Links[1].Params[1].Name)
	s.ElementsMatch([]string{"Primary", "Secondary"}, e.Links[1].Params[1].Values)

	s.Require().Len(e.Related, 1)
	s.Equal(model.RelTypeDependsOn, e.Related[0].Type)
	s.Require().NotNil(e.Related[0].Gap)
	s.Equal(15*time.Minute, *e.Related[0].Gap)

	// Assert: PARTICIPANT
	s.Require().Len(e.Participants, 1)
	part := e.Participants[0]
	s.Require().Len(part.Concepts, 1)
	s.Equal("https://example.com/concept/speaker", part.Concepts[0])
	s.Require().Len(part.RefIDs, 1)
	s.Equal("part-ref-1", part.RefIDs[0])
	s.Require().Len(part.Links, 1)
	s.Equal("https://example.com/people/john", part.Links[0].Value)
	s.Equal("profile", part.Links[0].Rel)

	// Assert: LOCATION
	s.Require().Len(e.Locations, 1)
	loc := e.Locations[0]
	s.Require().Len(loc.Concepts, 1)
	s.Equal("https://example.com/concept/location", loc.Concepts[0])
	s.Require().Len(loc.RefIDs, 1)
	s.Equal("loc-ref-1", loc.RefIDs[0])
	s.Require().Len(loc.Links, 1)
	s.Equal("https://example.com/location", loc.Links[0].Value)
	s.Equal("info", loc.Links[0].Rel)

	// Assert: RESOURCE
	s.Require().Len(e.Resources, 1)
	res := e.Resources[0]
	s.Require().Len(res.Concepts, 1)
	s.Equal("https://example.com/concept/resource", res.Concepts[0])
	s.Require().Len(res.RefIDs, 1)
	s.Equal("res-ref-1", res.RefIDs[0])
	s.Require().Len(res.Links, 1)
	s.Equal("https://example.com/resource", res.Links[0].Value)
	s.Equal("spec", res.Links[0].Rel)

	// Assert: VTODO
	s.Require().Len(cal.Todos, 1)
	todo := cal.Todos[0]
	s.Require().Len(todo.Concepts, 1)
	s.Equal("https://example.com/concept/todo", todo.Concepts[0])
	s.Require().Len(todo.RefIDs, 1)
	s.Equal("todo-ref-1", todo.RefIDs[0])
	s.Require().Len(todo.Links, 1)
	s.Equal("https://example.com/todo", todo.Links[0].Value)
	s.Equal("related", todo.Links[0].Rel)
	s.Require().Len(todo.Related, 1)
	s.Equal(model.RelTypeFinishToStart, todo.Related[0].Type)
	s.Require().NotNil(todo.Related[0].Gap)
	s.Equal(1*time.Hour, *todo.Related[0].Gap)

	// Assert: VJOURNAL
	s.Require().Len(cal.Journals, 1)
	j := cal.Journals[0]
	s.Require().Len(j.Concepts, 1)
	s.Equal("https://example.com/concept/journal", j.Concepts[0])
	s.Require().Len(j.RefIDs, 1)
	s.Equal("journal-ref-1", j.RefIDs[0])
	s.Require().Len(j.Links, 1)
	s.Equal("https://example.com/journal", j.Links[0].Value)
	s.Equal("alternate", j.Links[0].Rel)
	s.Require().Len(j.Related, 1)
	s.Equal(model.RelTypeFirst, j.Related[0].Type)
}

func (s *ParserSuite) TestRFC9074AlarmExtensions() {
	// Arrange
	data := s.readFileBytes("05_extensions/09_rfc9074_alarms.ics")

	// Act
	cal, err := parser.ParseBytes(data)

	// Assert
	s.Require().NoError(err)
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]
	s.Require().Len(e.Alarms, 3)

	// Alarm 1: ACKNOWLEDGED + UID
	a1 := e.Alarms[0]
	s.Equal("alarm-1", a1.UID)
	s.Require().NotNil(a1.Acknowledged)
	s.True(a1.Acknowledged.Equal(time.Date(2024, 1, 2, 9, 45, 0, 0, time.UTC)))

	// Alarm 2: RELATED-TO;RELTYPE=SNOOZE
	a2 := e.Alarms[1]
	s.Equal("alarm-2", a2.UID)
	s.Require().Len(a2.Related, 1)
	s.Equal("alarm-1", a2.Related[0].UID)
	s.Equal(model.RelTypeSnooze, a2.Related[0].Type)

	// Alarm 3: PROXIMITY + VLOCATION
	a3 := e.Alarms[2]
	s.Equal("alarm-3", a3.UID)
	s.Equal(model.ProximityArrive, a3.Proximity)
	s.Require().Len(a3.Locations, 1)
	s.Equal("loc-1", a3.Locations[0].UID)
	s.Equal("geo:37.386013,-122.082932", a3.Locations[0].URL)
}

func (s *ParserSuite) TestRFC7529RScaleAndSkip() {
	// Arrange
	data := s.readFileBytes("05_extensions/07_rfc7529_rscale.ics")

	// Act
	cal, err := parser.ParseBytes(data)

	// Assert
	s.Require().NoError(err)
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]
	s.Require().Len(e.RRules, 1)
	r := e.RRules[0]
	s.Equal(model.RecurrenceScale("HEBREW"), r.RScale)
	s.Equal(model.SkipForward, r.Skip)
}

func (s *ParserSuite) TestVendorSpecific_OutlookFull() {
	// Arrange
	data := s.readFileBytes("06_vendor_specific/05_outlook_full.ics")

	// Act
	cal, err := parser.ParseBytes(data)

	// Assert
	s.Require().NoError(err)
	s.Require().Len(cal.Events, 1)
	s.Equal("outlook-full-1", cal.Events[0].UID)
}

func (s *ParserSuite) TestVendorSpecific_AppleCalendarExtended() {
	// Arrange
	data := s.readFileBytes("06_vendor_specific/06_apple_calendar_extended.ics")

	// Act
	cal, err := parser.ParseBytes(data)

	// Assert
	s.Require().NoError(err)
	s.Require().Len(cal.Events, 1)
	s.Equal("apple-extended-001@example.com", cal.Events[0].UID)
	s.Require().Len(cal.Events[0].Alarms, 1)
}

func (s *ParserSuite) TestExamples_AllICSParse() {
	// Arrange
	var files []string
	root := filepath.Join("..", "examples", "ical")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == "00_testing" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".ics") {
			return nil
		}
		if strings.Contains(d.Name(), "invalid") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(files)

	// Act & Assert
	for _, path := range files {
		data, readErr := os.ReadFile(path)
		s.Require().NoError(readErr, "read file %s", path)
		_, parseErr := parser.ParseBytes(data)
		s.Require().NoError(parseErr, "parse file %s", path)
	}
}

func (s *ParserSuite) TestMandatoryFields() {
	s.Run("MissingUID in Participant", func() {
		data := s.readFileBytes("00_testing/mandatory_participant_missing_uid.ics")
		_, err := parser.ParseBytes(data)
		s.Require().Error(err)
		s.Contains(err.Error(), "PARTICIPANT: missing UID")
	})

	s.Run("MissingType in Participant", func() {
		data := s.readFileBytes("00_testing/mandatory_participant_missing_type.ics")
		_, err := parser.ParseBytes(data)
		s.Require().Error(err)
		s.Contains(err.Error(), "PARTICIPANT: missing PARTICIPANT-TYPE")
	})

	s.Run("MissingUID in Location", func() {
		data := s.readFileBytes("00_testing/mandatory_location_missing_uid.ics")
		_, err := parser.ParseBytes(data)
		s.Require().Error(err)
		s.Contains(err.Error(), "LOCATION: missing UID")
	})
}

func (s *ParserSuite) TestPropertyNameCaseInsensitivity() {
	// Arrange - use lowercase property names (begin, uid, summary, etc.)
	data := s.readFileBytes("00_testing/property_name_case_insensitive.ics")

	// Act
	cal, err := parser.ParseBytes(data)

	// Assert
	s.Require().NoError(err)
	s.Require().Len(cal.Events, 1)
	s.Equal("Case Test", cal.Events[0].Summary)
	s.Equal("case-test@example.com", cal.Events[0].UID)
}

func (s *ParserSuite) TestParameterCaseInsensitivity() {
	// Arrange - use lowercase parameter names
	data := s.readFileBytes("00_testing/parameter_case_insensitive.ics")

	// Act
	cal, err := parser.ParseBytes(data)

	// Assert
	s.Require().NoError(err)
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]
	s.Require().Len(e.StructuredData, 1)
	sd := e.StructuredData[0]

	// Parameter names were normalized to uppercase, so parsing should work.
	s.Equal("URI", sd.ValueType)
	s.Equal("application/json", sd.FmtType)
	s.Equal("https://schema.org/Event", sd.Schema)
}

func (s *ParserSuite) TestRFC6868ParameterValueDecoding() {
	// Arrange
	data := s.readFileBytes("00_testing/parameter_value_rfc6868.ics")

	// Act
	cal, err := parser.ParseBytes(data)

	// Assert
	s.Require().NoError(err)
	s.Require().Len(cal.Events, 1)
	e := cal.Events[0]
	s.Require().Len(e.Attendees, 1)
	a := e.Attendees[0]
	s.Equal("George Herman \"Babe\" Ruth", a.Name)

	var note, carets, raw string
	for _, param := range a.Params {
		if len(param.Values) == 0 {
			continue
		}
		switch param.Name {
		case "X-NOTE":
			note = param.Values[0]
		case "X-CARETS":
			carets = param.Values[0]
		case "X-RAW":
			raw = param.Values[0]
		}
	}
	s.Equal("Line1\nLine2", note)
	s.Equal("Hat ^Carets^", carets)
	s.Equal("Keep ^x here", raw)
}

// --- Additional Explicit Tests for Coverage --- //

func (s *ParserSuite) TestAllDayAndFloating() {
	cal := s.parseFile("01_basic/03_all_day_and_floating.ics")
	s.Require().Len(cal.Events, 2)
}

func (s *ParserSuite) TestEventWithCategoriesURL() {
	cal := s.parseFile("01_basic/04_event_with_categories_url.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestVTimezoneMultipleTransitions() {
	cal := s.parseFile("02_timezones/04_vtimezone_multiple_transitions.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestMixedTimeKinds() {
	cal := s.parseFile("02_timezones/05_mixed_time_kinds.ics")
	s.Require().Len(cal.Events, 3)
}

func (s *ParserSuite) TestComplexRRuleByMonthDay() {
	cal := s.parseFile("03_recurrence/03_complex_rrule_bymonthday.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestComplexRRuleBySetPos() {
	cal := s.parseFile("03_recurrence/04_complex_rrule_bysetpos.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestRecurrenceIDOverride() {
	cal := s.parseFile("03_recurrence/05_recurrence_id_override.ics")
	s.Require().Len(cal.Events, 2)
}

func (s *ParserSuite) TestRRuleUntilUTCWithExDate() {
	cal := s.parseFile("03_recurrence/06_rrule_until_z_and_exdate.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestRRuleByYearDayByWeekNo() {
	cal := s.parseFile("03_recurrence/07_rrule_byyearday_byweekno.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestCRLFAndLFMixed() {
	cal := s.parseFile("04_edge_cases/09_crlf_and_lf_mixed.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestHugeDescription() {
	cal := s.parseFile("04_edge_cases/05_huge_description.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestInvalidMultipleRRules() {
	cal := s.parseFile("04_edge_cases/06_invalid_multiple_rrules.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestManyXPropsAndIana() {
	cal := s.parseFile("04_edge_cases/07_many_xprops_and_iana.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestMultipleRDateExDateLines() {
	cal := s.parseFile("04_edge_cases/08_multiple_rdate_exdate_lines.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestVendorOutlookLike() {
	cal := s.parseFile("06_vendor_specific/03_outlook_like.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestVendorExchangeLike() {
	cal := s.parseFile("06_vendor_specific/04_exchange_like.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestRFC7986ImageColor() {
	cal := s.parseFile("05_extensions/06_rfc7986_image_color.ics")
	s.Require().NotNil(cal)
}

func (s *ParserSuite) TestRFC9074MultiLocations() {
	cal := s.parseFile("05_extensions/10_rfc9074_multi_locations.ics")
	s.Require().NotNil(cal)
}

func (s *ParserSuite) TestRFC9253RelatedGapRefID() {
	cal := s.parseFile("05_extensions/08_rfc9253_related_gap_refid.ics")
	s.Require().Len(cal.Events, 2)
}

func (s *ParserSuite) TestComponentsVTodo() {
	cal := s.parseFile("07_components/01_vtodo.ics")
	s.Require().Len(cal.Todos, 1)
}

func (s *ParserSuite) TestComponentsVJournal() {
	cal := s.parseFile("07_components/02_vjournal.ics")
	s.Require().Len(cal.Journals, 1)
}

func (s *ParserSuite) TestComponentsVFreeBusy() {
	cal := s.parseFile("07_components/03_vfreebusy.ics")
	s.Require().Len(cal.FreeBusys, 1)
}

func (s *ParserSuite) TestComponentsVTodoComplex() {
	cal := s.parseFile("07_components/04_vtodo_complex.ics")
	s.Require().Len(cal.Todos, 1)
}

func (s *ParserSuite) TestComponentsVFreeBusyMultipleTypes() {
	cal := s.parseFile("07_components/05_vfreebusy_multiple_types.ics")
	s.Require().Len(cal.FreeBusys, 1)
}

func (s *ParserSuite) TestAlarmsAudio() {
	cal := s.parseFile("08_alarms/01_audio_alarm.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestAlarmsEmail() {
	cal := s.parseFile("08_alarms/02_email_alarm.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestAlarmsDisplayRepeat() {
	cal := s.parseFile("08_alarms/03_display_alarm_with_repeat.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestAlarmsAudioAttach() {
	cal := s.parseFile("08_alarms/04_audio_alarm_with_attach.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestRealWorldConference() {
	cal := s.parseFile("09_real_world/01_conference_schedule.ics")
	s.Require().Len(cal.Events, 3)
}

func (s *ParserSuite) TestRealWorldLargeMixed() {
	cal := s.parseFile("09_real_world/03_large_mixed_calendar.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestRealWorldCancelled() {
	cal := s.parseFile("09_real_world/04_cancelled_event.ics")
	s.Require().Len(cal.Events, 1)
}

func (s *ParserSuite) TestVTodoRecurrenceRelations() {
	cal := s.parseFile("10_vtodo/03_todo_with_recurrence_and_relations.ics")
	s.Require().Len(cal.Todos, 1)
}

func (s *ParserSuite) TestVTodoStructuredData() {
	cal := s.parseFile("10_vtodo/04_todo_structured_data.ics")
	s.Require().Len(cal.Todos, 1)
}

func (s *ParserSuite) TestVJournalMultipleDescXAlt() {
	cal := s.parseFile("11_vjournal/03_vjournal_multiple_desc_xalt.ics")
	s.Require().Len(cal.Journals, 1)
}

func (s *ParserSuite) TestVJournalRelatedTo() {
	cal := s.parseFile("11_vjournal/04_vjournal_related_to.ics")
	s.Require().Len(cal.Journals, 1)
}

func (s *ParserSuite) TestBenchComplexMultiComponentMix() {
	cal := s.parseFile("12_bench_complex/01_multi_component_mix.ics")
	s.Require().NotNil(cal)
}

func (s *ParserSuite) TestBenchComplexManyEventsWithAlarms() {
	cal := s.parseFile("12_bench_complex/02_many_events_with_alarms.ics")
	s.Require().NotNil(cal)
}

func (s *ParserSuite) TestBenchComplexTimezonesRDateExDate() {
	cal := s.parseFile("12_bench_complex/03_timezones_rdate_exdate.ics")
	s.Require().NotNil(cal)
}

func (s *ParserSuite) TestBenchComplexPublishingComponents() {
	cal := s.parseFile("12_bench_complex/04_publishing_components.ics")
	s.Require().NotNil(cal)
}

func (s *ParserSuite) TestBenchComplexRelationshipsLinksConcepts() {
	cal := s.parseFile("12_bench_complex/05_relationships_links_concepts.ics")
	s.Require().NotNil(cal)
}

func (s *ParserSuite) TestBenchComplexAvailabilityVAvailability() {
	cal := s.parseFile("12_bench_complex/06_availability_vavailability.ics")
	s.Require().NotNil(cal)
}

func (s *ParserSuite) TestBenchComplexRScaleSkipRRule() {
	cal := s.parseFile("12_bench_complex/07_rscale_skip_rrule.ics")
	s.Require().NotNil(cal)
}

func (s *ParserSuite) TestBenchComplexParamEncodingRFC6868() {
	cal := s.parseFile("12_bench_complex/08_param_encoding_rfc6868.ics")
	s.Require().NotNil(cal)
}

func (s *ParserSuite) TestBenchComplexLongFoldedText() {
	cal := s.parseFile("12_bench_complex/09_long_folded_text.ics")
	s.Require().NotNil(cal)
}

func (s *ParserSuite) TestBenchComplexVendorQuirksXProps() {
	cal := s.parseFile("12_bench_complex/10_vendor_quirks_xprops.ics")
	s.Require().NotNil(cal)
}

// --------------------------------------------------------------------------
// 05_extensions: RFC 5546 (iTIP) и RFC 6638 (SCHEDULE-*)
// --------------------------------------------------------------------------

// TestRFC5546iTIPRequest проверяет парсинг METHOD:REQUEST с SCHEDULE-AGENT.
func (s *ParserSuite) TestRFC5546iTIPRequest() {
	// Arrange & Act
	cal := s.parseFile("05_extensions/12_rfc5546_itip_request.ics")

	// Assert
	s.Equal(model.MethodRequest, cal.Method)
	s.Require().Len(cal.Events, 1)

	e := cal.Events[0]
	s.Equal("itip-request-001@example.com", e.UID)
	s.Equal("Team Standup (iTIP REQUEST)", e.Summary)
	s.Equal(model.StatusConfirmed, e.Status)

	// Проверяем организатора.
	s.Require().NotNil(e.Organizer)
	s.Equal("mailto:alice@example.com", e.Organizer.Address)
	s.Equal("Alice Manager", e.Organizer.Name)

	// Проверяем участников и их SCHEDULE-AGENT.
	s.Require().Len(e.Attendees, 3)

	bob := e.Attendees[0]
	s.Equal("mailto:bob@example.com", bob.Address)
	s.Equal(model.ScheduleAgentServer, bob.ScheduleAgent)
	s.Equal(model.RoleReqParticipant, bob.Role)
	s.True(bob.RSVP)

	carol := e.Attendees[1]
	s.Equal("mailto:carol@example.com", carol.Address)
	s.Equal(model.ScheduleAgentServer, carol.ScheduleAgent)
	s.Equal(model.RoleOptParticipant, carol.Role)

	dave := e.Attendees[2]
	s.Equal("mailto:dave@external.org", dave.Address)
	s.Equal(model.ScheduleAgentClient, dave.ScheduleAgent)
	s.False(dave.RSVP)
}

// TestRFC5546iTIPReply проверяет парсинг METHOD:REPLY с REQUEST-STATUS.
func (s *ParserSuite) TestRFC5546iTIPReply() {
	// Arrange & Act
	cal := s.parseFile("05_extensions/13_rfc5546_itip_reply.ics")

	// Assert
	s.Equal(model.MethodReply, cal.Method)
	s.Require().Len(cal.Events, 1)

	e := cal.Events[0]
	s.Equal("itip-request-001@example.com", e.UID)

	// Проверяем REQUEST-STATUS.
	s.Require().Len(e.RequestStatus, 2)
	s.Equal("2.0", e.RequestStatus[0].Code)
	s.Equal("Success", e.RequestStatus[0].Description)
	s.Empty(e.RequestStatus[0].ExtraData)
	s.Equal("2.3", e.RequestStatus[1].Code)

	// Проверяем SCHEDULE-AGENT и SCHEDULE-STATUS на участнике.
	s.Require().Len(e.Attendees, 1)
	bob := e.Attendees[0]
	s.Equal(model.ScheduleAgentServer, bob.ScheduleAgent)
	s.Equal(model.PartStatAccepted, bob.Status)
	s.Require().Len(bob.ScheduleStatus, 1)
	s.Equal("2.0", bob.ScheduleStatus[0])
}

// TestRFC5546iTIPCancel проверяет парсинг METHOD:CANCEL.
func (s *ParserSuite) TestRFC5546iTIPCancel() {
	// Arrange & Act
	cal := s.parseFile("05_extensions/14_rfc5546_itip_cancel.ics")

	// Assert
	s.Equal(model.MethodCancel, cal.Method)
	s.Require().Len(cal.Events, 1)

	e := cal.Events[0]
	s.Equal(model.StatusCancelled, e.Status)
	s.Equal(1, e.Sequence)
	s.Require().Len(e.Attendees, 2)
}

// TestRFC6638ScheduleParams проверяет полный набор параметров RFC 6638.
func (s *ParserSuite) TestRFC6638ScheduleParams() {
	// Arrange & Act
	cal := s.parseFile("05_extensions/15_rfc6638_schedule_params.ics")

	// Assert
	s.Equal(model.MethodRequest, cal.Method)
	s.Require().Len(cal.Events, 1)

	e := cal.Events[0]
	s.Equal("rfc6638-schedule-001@example.com", e.UID)

	// Проверяем REQUEST-STATUS.
	s.Require().Len(e.RequestStatus, 2)
	rs0 := e.RequestStatus[0]
	s.Equal("1.2", rs0.Code)
	s.Equal("Scheduling message sent", rs0.Description)
	rs1 := e.RequestStatus[1]
	s.Equal("3.7", rs1.Code)
	s.Equal("Invalid calendar user", rs1.Description)
	s.Contains(rs1.ExtraData, "ATTENDEE")

	// Проверяем организатора.
	s.Require().NotNil(e.Organizer)
	s.Equal(model.ScheduleAgentServer, e.Organizer.ScheduleAgent)

	// Проверяем участников.
	s.Require().Len(e.Attendees, 5)

	// SERVER agent.
	s.Equal(model.ScheduleAgentServer, e.Attendees[0].ScheduleAgent)
	// CLIENT agent.
	s.Equal(model.ScheduleAgentClient, e.Attendees[1].ScheduleAgent)
	// NONE agent.
	s.Equal(model.ScheduleAgentNone, e.Attendees[2].ScheduleAgent)
	// SCHEDULE-FORCE-SEND=REQUEST.
	s.Equal("REQUEST", e.Attendees[3].ScheduleForceSend)
	s.Require().Len(e.Attendees[3].ScheduleStatus, 1)
	s.Equal("1.2", e.Attendees[3].ScheduleStatus[0])
	// SCHEDULE-STATUS=3.7.
	s.Require().Len(e.Attendees[4].ScheduleStatus, 1)
	s.Equal("3.7", e.Attendees[4].ScheduleStatus[0])
}

// --------------------------------------------------------------------------
// 12_bench_complex (smoke-тест)
// --------------------------------------------------------------------------

// TestBenchComplexITIPScheduling проверяет базовую читаемость bench-файла.
func (s *ParserSuite) TestBenchComplexITIPScheduling() {
	cal := s.parseFile("12_bench_complex/11_itip_scheduling.ics")
	s.Require().NotNil(cal)
	s.Equal(model.MethodRequest, cal.Method)
	s.Require().Len(cal.Events, 3)
	s.Require().Len(cal.Todos, 1)
	// Проверяем REQUEST-STATUS первого события.
	s.Require().Len(cal.Events[0].RequestStatus, 2)
	s.Equal("2.0", cal.Events[0].RequestStatus[0].Code)
}
