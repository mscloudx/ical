package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"gitverse.ru/cloudcoder/ical/parser"
)

// benchExamplesDir — путь к примерам iCalendar.
var benchExamplesDir = filepath.Join("..", "examples", "ical")

// loadFile загружает файл в память для бенчмарка.
func loadFile(b *testing.B, relPath string) []byte {
	b.Helper()
	data, err := os.ReadFile(filepath.Join(benchExamplesDir, relPath))
	if err != nil {
		b.Fatal(err)
	}
	return data
}

// runParseBench — общий хелпер для бенчмарков ParseBytes.
func runParseBench(b *testing.B, relPath string) {
	b.Helper()
	data := loadFile(b, relPath)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = parser.ParseBytes(data)
	}
}

// --------------------------------------------------------------------------
// 01_basic
// --------------------------------------------------------------------------

func BenchmarkParse_EmptyVCalendar(b *testing.B) {
	runParseBench(b, filepath.Join("01_basic", "01_empty_vcalendar.ics"))
}

func BenchmarkParse_SingleVEvent(b *testing.B) {
	runParseBench(b, filepath.Join("01_basic", "02_single_vevent.ics"))
}

// --------------------------------------------------------------------------
// 02_timezones
// --------------------------------------------------------------------------

func BenchmarkParse_FloatingTime(b *testing.B) {
	runParseBench(b, filepath.Join("02_timezones", "01_floating_time.ics"))
}

func BenchmarkParse_UTCTime(b *testing.B) {
	runParseBench(b, filepath.Join("02_timezones", "02_utc_time.ics"))
}

func BenchmarkParse_LocalTimeWithVTimezone(b *testing.B) {
	runParseBench(b, filepath.Join("02_timezones", "03_local_time_with_vtimezone.ics"))
}

// --------------------------------------------------------------------------
// 03_recurrence
// --------------------------------------------------------------------------

func BenchmarkParse_DailyRecurrence(b *testing.B) {
	runParseBench(b, filepath.Join("03_recurrence", "01_daily.ics"))
}

func BenchmarkParse_WeeklyWithExDate(b *testing.B) {
	runParseBench(b, filepath.Join("03_recurrence", "02_weekly_with_exdate.ics"))
}

func BenchmarkParse_ComplexRRuleByMonthDay(b *testing.B) {
	runParseBench(b, filepath.Join("03_recurrence", "03_complex_rrule_bymonthday.ics"))
}

func BenchmarkParse_ComplexRRuleBySetPos(b *testing.B) {
	runParseBench(b, filepath.Join("03_recurrence", "04_complex_rrule_bysetpos.ics"))
}

func BenchmarkParse_RecurrenceIDOverride(b *testing.B) {
	runParseBench(b, filepath.Join("03_recurrence", "05_recurrence_id_override.ics"))
}

// --------------------------------------------------------------------------
// 04_edge_cases
// --------------------------------------------------------------------------

func BenchmarkParse_LongLinesFolded(b *testing.B) {
	runParseBench(b, filepath.Join("04_edge_cases", "01_long_lines_folded.ics"))
}

func BenchmarkParse_CRLFAndLFMixed(b *testing.B) {
	runParseBench(b, filepath.Join("04_edge_cases", "02_crlf_and_lf_mixed.ics"))
}

func BenchmarkParse_MissingDTEndAndDuration(b *testing.B) {
	runParseBench(b, filepath.Join("04_edge_cases", "02_missing_dtend_and_duration.ics"))
}

func BenchmarkParse_EscapedCharacters(b *testing.B) {
	runParseBench(b, filepath.Join("04_edge_cases", "03_escaped_characters.ics"))
}

func BenchmarkParse_HugeDescription(b *testing.B) {
	runParseBench(b, filepath.Join("04_edge_cases", "05_huge_description.ics"))
}

func BenchmarkParse_InvalidMultipleRRules(b *testing.B) {
	runParseBench(b, filepath.Join("04_edge_cases", "06_invalid_multiple_rrules.ics"))
}

// --------------------------------------------------------------------------
// 05_vendor_specific
// --------------------------------------------------------------------------

func BenchmarkParse_GoogleCalendar(b *testing.B) {
	runParseBench(b, filepath.Join("05_vendor_specific", "01_google_calendar.ics"))
}

func BenchmarkParse_AppleCalendar(b *testing.B) {
	runParseBench(b, filepath.Join("05_vendor_specific", "02_apple_calendar.ics"))
}

// --------------------------------------------------------------------------
// 06_components
// --------------------------------------------------------------------------

func BenchmarkParse_VTodo(b *testing.B) {
	runParseBench(b, filepath.Join("06_components", "01_vtodo.ics"))
}

func BenchmarkParse_VJournal(b *testing.B) {
	runParseBench(b, filepath.Join("06_components", "02_vjournal.ics"))
}

func BenchmarkParse_VFreeBusy(b *testing.B) {
	runParseBench(b, filepath.Join("06_components", "03_vfreebusy.ics"))
}

// --------------------------------------------------------------------------
// 07_alarms
// --------------------------------------------------------------------------

func BenchmarkParse_AudioAlarm(b *testing.B) {
	runParseBench(b, filepath.Join("07_alarms", "01_audio_alarm.ics"))
}

func BenchmarkParse_EmailAlarm(b *testing.B) {
	runParseBench(b, filepath.Join("07_alarms", "02_email_alarm.ics"))
}

// --------------------------------------------------------------------------
// 08_real_world
// --------------------------------------------------------------------------

func BenchmarkParse_ConferenceSchedule(b *testing.B) {
	runParseBench(b, filepath.Join("08_real_world", "01_conference_schedule.ics"))
}
