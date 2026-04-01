package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mscloudx/ical/parser"
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

func BenchmarkParse_AllDayAndFloatingEvents(b *testing.B) {
	runParseBench(b, filepath.Join("01_basic", "03_all_day_and_floating.ics"))
}

func BenchmarkParse_EventWithCategoriesURL(b *testing.B) {
	runParseBench(b, filepath.Join("01_basic", "04_event_with_categories_url.ics"))
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

func BenchmarkParse_VTimezoneMultipleTransitions(b *testing.B) {
	runParseBench(b, filepath.Join("02_timezones", "04_vtimezone_multiple_transitions.ics"))
}

func BenchmarkParse_MixedTimeKinds(b *testing.B) {
	runParseBench(b, filepath.Join("02_timezones", "05_mixed_time_kinds.ics"))
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

func BenchmarkParse_RRuleUntilUTCWithExDate(b *testing.B) {
	runParseBench(b, filepath.Join("03_recurrence", "06_rrule_until_z_and_exdate.ics"))
}

func BenchmarkParse_RRuleByYearDayByWeekNo(b *testing.B) {
	runParseBench(b, filepath.Join("03_recurrence", "07_rrule_byyearday_byweekno.ics"))
}

// --------------------------------------------------------------------------
// 04_edge_cases
// --------------------------------------------------------------------------

func BenchmarkParse_LongLinesFolded(b *testing.B) {
	runParseBench(b, filepath.Join("04_edge_cases", "01_long_lines_folded.ics"))
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

func BenchmarkParse_ManyXPropsAndIana(b *testing.B) {
	runParseBench(b, filepath.Join("04_edge_cases", "07_many_xprops_and_iana.ics"))
}

func BenchmarkParse_MultipleRDateExDateLines(b *testing.B) {
	runParseBench(b, filepath.Join("04_edge_cases", "08_multiple_rdate_exdate_lines.ics"))
}

func BenchmarkParse_CRLFAndLFMixed(b *testing.B) {
	runParseBench(b, filepath.Join("04_edge_cases", "09_crlf_and_lf_mixed.ics"))
}

// --------------------------------------------------------------------------
// 06_vendor_specific
// --------------------------------------------------------------------------

func BenchmarkParse_GoogleCalendar(b *testing.B) {
	runParseBench(b, filepath.Join("06_vendor_specific", "01_google_calendar.ics"))
}

func BenchmarkParse_AppleCalendar(b *testing.B) {
	runParseBench(b, filepath.Join("06_vendor_specific", "02_apple_calendar.ics"))
}

func BenchmarkParse_OutlookLike(b *testing.B) {
	runParseBench(b, filepath.Join("06_vendor_specific", "03_outlook_like.ics"))
}

func BenchmarkParse_ExchangeLike(b *testing.B) {
	runParseBench(b, filepath.Join("06_vendor_specific", "04_exchange_like.ics"))
}

func BenchmarkParse_OutlookFull(b *testing.B) {
	runParseBench(b, filepath.Join("06_vendor_specific", "05_outlook_full.ics"))
}

func BenchmarkParse_AppleCalendarExtended(b *testing.B) {
	runParseBench(b, filepath.Join("06_vendor_specific", "06_apple_calendar_extended.ics"))
}

// --------------------------------------------------------------------------
// 05_extensions
// --------------------------------------------------------------------------

func BenchmarkParse_RFC7986Properties(b *testing.B) {
	runParseBench(b, filepath.Join("05_extensions", "01_rfc7986_properties.ics"))
}

func BenchmarkParse_RFC7953Availability(b *testing.B) {
	runParseBench(b, filepath.Join("05_extensions", "02_rfc7953_availability.ics"))
}

func BenchmarkParse_RFC9073Publishing(b *testing.B) {
	runParseBench(b, filepath.Join("05_extensions", "04_rfc9073_publishing.ics"))
}

func BenchmarkParse_RFC9253Relationships(b *testing.B) {
	runParseBench(b, filepath.Join("05_extensions", "05_rfc9253_relationships.ics"))
}

func BenchmarkParse_RFC7986ImageColor(b *testing.B) {
	runParseBench(b, filepath.Join("05_extensions", "06_rfc7986_image_color.ics"))
}

func BenchmarkParse_RFC7529RScale(b *testing.B) {
	runParseBench(b, filepath.Join("05_extensions", "07_rfc7529_rscale.ics"))
}

func BenchmarkParse_RFC9253RelatedGapRefID(b *testing.B) {
	runParseBench(b, filepath.Join("05_extensions", "08_rfc9253_related_gap_refid.ics"))
}

func BenchmarkParse_RFC9074Alarms(b *testing.B) {
	runParseBench(b, filepath.Join("05_extensions", "09_rfc9074_alarms.ics"))
}

func BenchmarkParse_RFC9074MultiLocations(b *testing.B) {
	runParseBench(b, filepath.Join("05_extensions", "10_rfc9074_multi_locations.ics"))
}

// --------------------------------------------------------------------------
// 07_components
// --------------------------------------------------------------------------

func BenchmarkParse_VTodo(b *testing.B) {
	runParseBench(b, filepath.Join("07_components", "01_vtodo.ics"))
}

func BenchmarkParse_VJournal(b *testing.B) {
	runParseBench(b, filepath.Join("07_components", "02_vjournal.ics"))
}

func BenchmarkParse_VFreeBusy(b *testing.B) {
	runParseBench(b, filepath.Join("07_components", "03_vfreebusy.ics"))
}

func BenchmarkParse_VTodoComplex(b *testing.B) {
	runParseBench(b, filepath.Join("07_components", "04_vtodo_complex.ics"))
}

func BenchmarkParse_VFreeBusyMultipleTypes(b *testing.B) {
	runParseBench(b, filepath.Join("07_components", "05_vfreebusy_multiple_types.ics"))
}

// --------------------------------------------------------------------------
// 08_alarms
// --------------------------------------------------------------------------

func BenchmarkParse_AudioAlarm(b *testing.B) {
	runParseBench(b, filepath.Join("08_alarms", "01_audio_alarm.ics"))
}

func BenchmarkParse_EmailAlarm(b *testing.B) {
	runParseBench(b, filepath.Join("08_alarms", "02_email_alarm.ics"))
}

func BenchmarkParse_DisplayAlarmWithRepeat(b *testing.B) {
	runParseBench(b, filepath.Join("08_alarms", "03_display_alarm_with_repeat.ics"))
}

func BenchmarkParse_AudioAlarmWithAttach(b *testing.B) {
	runParseBench(b, filepath.Join("08_alarms", "04_audio_alarm_with_attach.ics"))
}

// --------------------------------------------------------------------------
// 09_real_world
// --------------------------------------------------------------------------

func BenchmarkParse_ConferenceSchedule(b *testing.B) {
	runParseBench(b, filepath.Join("09_real_world", "01_conference_schedule.ics"))
}

func BenchmarkParse_RealWorldLargeMixed(b *testing.B) {
	runParseBench(b, filepath.Join("09_real_world", "03_large_mixed_calendar.ics"))
}

func BenchmarkParse_RealWorldCancelledEvent(b *testing.B) {
	runParseBench(b, filepath.Join("09_real_world", "04_cancelled_event.ics"))
}

// --------------------------------------------------------------------------
// 10_vtodo
// --------------------------------------------------------------------------

func BenchmarkParse_SimpleTodo(b *testing.B) {
	runParseBench(b, filepath.Join("10_vtodo", "01_simple_todo.ics"))
}

func BenchmarkParse_TodoWithAlarm(b *testing.B) {
	runParseBench(b, filepath.Join("10_vtodo", "02_todo_with_alarm.ics"))
}

func BenchmarkParse_VTodoRecurrenceRelations(b *testing.B) {
	runParseBench(b, filepath.Join("10_vtodo", "03_todo_with_recurrence_and_relations.ics"))
}

func BenchmarkParse_VTodoStructuredData(b *testing.B) {
	runParseBench(b, filepath.Join("10_vtodo", "04_todo_structured_data.ics"))
}

// --------------------------------------------------------------------------
// 11_vjournal
// --------------------------------------------------------------------------

func BenchmarkParse_SimpleJournal(b *testing.B) {
	runParseBench(b, filepath.Join("11_vjournal", "01_simple_journal.ics"))
}

func BenchmarkParse_VJournalMultipleDescriptions(b *testing.B) {
	runParseBench(b, filepath.Join("11_vjournal", "03_vjournal_multiple_desc_xalt.ics"))
}

func BenchmarkParse_VJournalRelatedTo(b *testing.B) {
	runParseBench(b, filepath.Join("11_vjournal", "04_vjournal_related_to.ics"))
}

// --------------------------------------------------------------------------
// 12_bench_complex
// --------------------------------------------------------------------------

func BenchmarkParse_ComplexMultiComponentMix(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "01_multi_component_mix.ics"))
}

func BenchmarkParse_ComplexManyEventsWithAlarms(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "02_many_events_with_alarms.ics"))
}

func BenchmarkParse_ComplexTimezonesRDateExDate(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "03_timezones_rdate_exdate.ics"))
}

func BenchmarkParse_ComplexPublishingComponents(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "04_publishing_components.ics"))
}

func BenchmarkParse_ComplexRelationshipsLinksConcepts(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "05_relationships_links_concepts.ics"))
}

func BenchmarkParse_ComplexAvailabilityVAvailability(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "06_availability_vavailability.ics"))
}

func BenchmarkParse_ComplexRScaleSkipRRule(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "07_rscale_skip_rrule.ics"))
}

func BenchmarkParse_ComplexParamEncodingRFC6868(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "08_param_encoding_rfc6868.ics"))
}

func BenchmarkParse_ComplexLongFoldedText(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "09_long_folded_text.ics"))
}

func BenchmarkParse_ComplexVendorQuirksXProps(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "10_vendor_quirks_xprops.ics"))
}

// BenchmarkParse_ITIPScheduling — RFC 5546 iTIP + RFC 6638 SCHEDULE-* параметры.
func BenchmarkParse_ITIPScheduling(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "11_itip_scheduling.ics"))
}

func BenchmarkParse_ComplexManyAttendees(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "12_many_attendees.ics"))
}

func BenchmarkParse_ComplexManyEvents(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "13_many_events.ics"))
}

func BenchmarkParse_ComplexManyVTimezonesCustomReady(b *testing.B) {
	runParseBench(b, filepath.Join("12_bench_complex", "14_many_vtimezones_custom_ready.ics"))
}

// BenchmarkParse_RFC5546Request — iTIP REQUEST с SCHEDULE-AGENT.
func BenchmarkParse_RFC5546Request(b *testing.B) {
	runParseBench(b, filepath.Join("05_extensions", "12_rfc5546_itip_request.ics"))
}

// BenchmarkParse_RFC6638ScheduleParams — полный набор SCHEDULE-* параметров.
func BenchmarkParse_RFC6638ScheduleParams(b *testing.B) {
	runParseBench(b, filepath.Join("05_extensions", "15_rfc6638_schedule_params.ics"))
}
