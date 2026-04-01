# iCalendar (Parser & Encoder)

> [RU](README_RU.md)
> EN

![Go Version](docs/assets/go.svg)
[![RFC 5545](docs/assets/rfc.svg)](docs/ical_analysis.md#анализ-rfc-5545-icalendar-и-компонента-vevent)
[![RFC 7986](docs/assets/rfc-7986.svg)](docs/rfc_extensions_analysis.md#1-rfc-7986-новые-свойства-для-данных-icalendar)
[![RFC 7953](docs/assets/rfc-7953.svg)](docs/rfc_extensions_analysis.md#2-rfc-7953-доступность-календаря-vavailability)
[![RFC 9073](docs/assets/rfc-9073.svg)](docs/rfc_extensions_analysis.md#3-rfc-9073-расширения-для-публикации-событий)
[![RFC 9253](docs/assets/rfc-9253.svg)](docs/rfc_extensions_analysis.md#4-rfc-9253-поддержка-взаимосвязей-в-icalendar)
[![RFC 9074](docs/assets/rfc-9074.svg)](docs/rfc_extensions_analysis.md#5-rfc-9074-расширения-valarm)
[![RFC 7529](docs/assets/rfc-7529.svg)](docs/rfc_extensions_analysis.md#6-rfc-7529-негригорианские-rrule)
[![RFC 6868](docs/assets/rfc-6868.svg)](docs/rfc_extensions_analysis.md)

A high-performance Go library for working with the **iCalendar (RFC 5545)** format. Supports parsing, constructing, and serializing calendars with minimal allocations and strict type safety.

Works great for processing and generating large calendars (e.g., `.ics` files from Google Calendar, Apple Calendar, Microsoft Outlook, and other providers).

---

## Features

- ⚡️ **High performance:** Streaming parsing via `io.Reader` and writing via `io.Writer` with minimal allocations.
- 🛠 **Builder pattern:** Convenient calendar and event construction using Functional Options.
- 📤 **Encoder:** Full serialization of Go models back into valid `.ics` format (RFC 5545).
- 📆 **Time & timezone support:** Integration with the standard `time` package, `TZID` handling, and `VTIMEZONE` components.
- 🔁 **Complex recurrences:** Parsing and generating `RRULE`, `RDATE`, and `EXDATE` (daily, weekly, complex `BYDAY` combinations, etc.).
- 🛡 **Fault tolerance:** Correct handling of vendor-specific properties (X-Props), files with `LF` and mixed `CRLF`/`LF` line endings.
- 📦 **Clean object model:** Pure data structures independent of the source text. Enumerations for constant parameters.
- 🧩 **RFC extension support:** `RFC 6868`, `RFC 7529`, `RFC 7953`, `RFC 7986`, `RFC 9073`, `RFC 9074`, `RFC 9253`.

---

## RFC 5545 Support

The parser and encoder support the most commonly used components of the standard. Thanks to the extensible architecture, even non-standard properties are preserved and can be processed.

### Components

| Component   | Status | Description                                                                 |
| ----------- | :----: | --------------------------------------------------------------------------- |
| `VCALENDAR` |   🟢    | Root object, supports `VERSION`, `PRODID`, `CALSCALE`, `METHOD`             |
| `VEVENT`    |   🟢    | Calendar event: full support (dates, statuses, attendees, recurrence rules) |
| `VTIMEZONE` |   🟢    | Timezones: `STANDARD` and `DAYLIGHT` transition time                        |
| `VALARM`    |   🟢    | Alarms and reminders: triggers (`TRIGGER`), actions, durations              |
| `VFREEBUSY` |   🟢    | Free/busy information: `FREE`, `BUSY`, period lists                         |
| `VTODO`     |   🟢    | Tasks: deadlines (`DUE`), progress, priorities, nested alarms               |
| `VJOURNAL`  |   🟢    | Journal entries: full support for text descriptions and statuses            |

### VEVENT Properties

| Property           | Status | Notes                                                                           |
| ------------------ | :----: | ------------------------------------------------------------------------------- |
| **Dates & time**   |   🟢    | `DTSTART`, `DTEND`, `DURATION`, `DTSTAMP`, `CREATED`, `LAST-MODIFIED`           |
| **Identification** |   🟢    | `UID`, `SEQUENCE`, `RELATED-TO`, `RECURRENCE-ID`                                |
| **Description**    |   🟢    | `SUMMARY`, `DESCRIPTION`, `LOCATION`, `URL`                                     |
| **Attendees**      |   🟢    | `ORGANIZER`, `ATTENDEE` (with `ROLE`, `PARTSTAT`, `RSVP`, and other parameters) |
| **Recurrence**     |   🟢    | `RRULE`, `RDATE`, `EXDATE`                                                      |
| **Statuses**       |   🟢    | `STATUS`, `TRANSP` (`OPAQUE`/`TRANSPARENT`), `CLASS`                            |
| **Other**          |   🟢    | Vendor-specific extensions (via `XProps`)                                       |

---

## RFC Extension Support

Key RFC extensions for iCalendar are supported. Details and examples — in [`docs/rfc_extensions_analysis.md`](docs/rfc_extensions_analysis.md).

| RFC          | Status | Coverage                                                                                                                                                                    |
| ------------ | :----: | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **RFC 6868** |   🟢    | `^`-escaping in parameter values: `^n` → `\n`, `^^` → `^`, `^'` → `"` — supported in parser, encoder, and builder                                                           |
| **RFC 7529** |   🟢    | `RSCALE`, `SKIP` in `RRULE` rules                                                                                                                                           |
| **RFC 7953** |   🟢    | `VAVAILABILITY` and `AVAILABLE`, recurrent availability window support                                                                                                      |
| **RFC 7986** |   🟢    | `VCALENDAR` properties: `NAME`, `DESCRIPTION`, `UID`, `LAST-MODIFIED`, `URL`, `CATEGORIES`, `REFRESH-INTERVAL`, `COLOR`, `SOURCE`; RFC 7986 IANA properties via `IanaProps` |
| **RFC 9073** |   🟢    | `STRUCTURED-DATA`, `STYLED-DESCRIPTION`, `PARTICIPANT`, `LOCATION`, `RESOURCE` components                                                                                   |
| **RFC 9074** |   🟢    | `ACKNOWLEDGED`, `PROXIMITY`, `SNOOZE`, `VLOCATION` for `VALARM`                                                                                                             |
| **RFC 9253** |   🟢    | `LINK`, `CONCEPT`, `REFID`, `RELATED-TO` extensions                                                                                                                         |

---

## Installation

Use the standard `go get` toolchain:

```bash
go get github.com/mscloudx/ical
```

---

## Quick Start

The `parser` package provides convenient entry points: `Parse(io.Reader)`, `ParseBytes([]byte)`, and `ParseWithClose(io.ReadCloser)`.

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mscloudx/ical/parser"
)

func main() {
	f, err := os.Open("calendar.ics")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	// Parse the stream
	calendar, err := parser.Parse(f)
	if err != nil {
		log.Fatalf("Error parsing iCalendar: %v", err)
	}

	fmt.Printf("Calendar (Version: %s)\n", calendar.Version)

	// Print events
	for _, event := range calendar.Events {
		fmt.Printf("- Event: %s\n", event.Summary)
		fmt.Printf("  Start: %s\n", event.DTStart.Format("2006-01-02 15:04:05"))
	}
}
```

### Builder & Encoder

The library provides a convenient way to create new calendars from scratch and save them to a file.

```go
package main

import (
	"os"
	"time"

	"github.com/mscloudx/ical/builder"
	"github.com/mscloudx/ical/encoder"
)

func main() {
	// Create an event using the builder
	now := time.Now()
	event := builder.NewEvent("unique-id@example.com", now, now.Add(2*time.Hour),
		builder.WithEventSummary("Team meeting"),
		builder.WithEventDescription("Weekly planning discussion"),
		builder.WithEventLocation("Room 101"),
	)

	// Build the calendar
	cal := builder.NewCalendar("-//mscloudx//iCalendar//EN",
		builder.WithEvents(event),
	)

	// Serialize to file
	f, _ := os.Create("my_calendar.ics")
	defer f.Close()

	encoder.Encode(f, cal)
}
```

---

## Architecture & Documentation

The project follows **Clean Architecture** principles (Go-style). The library is split into logical packages to avoid circular dependencies and simplify testing:

- `model/` — Domain data structures (calendars, events, types, enumerations).
- `parser/` — iCalendar stream parsing logic, lexer, and complex property parsing.
- `builder/` — Fluent interface for constructing models via Functional Options.
- `encoder/` — Model serialization to RFC 5545 text format with line folding support.

See the [`docs/`](docs/) folder for in-depth standard details:
- [📖 Glossary (glossary.md)](docs/glossary.md)
- [⚙️ RFC 5545 Analysis (ical_analysis.md)](docs/ical_analysis.md)
- [📈 iCalendar Extensions (rfc_extensions_analysis.md)](docs/rfc_extensions_analysis.md)
- [🛠 Vendor Quirks (vendor_quirks.md)](docs/vendor_quirks.md)

---

## Testing & Code Quality

The library is covered by tests using the AAA (Arrange, Act, Assert) paradigm and `testify/suite`.
An extensive `examples/ical/` folder is included in the repository for testing against real-world calendars, covering both standard cases and "quirky" files from specific providers.

To run tests, linting, and formatting:
```bash
make test
make lint
make fmt
```

Direct commands are also available:
```bash
go test ./...
go test -bench=. ./parser/
```

The code meets strict standards: `golangci-lint`, formatting via `gofumpt`, and import organization via `gci`.

---

## License

This project is distributed under the **MIT License**. You are free to use, modify, and distribute the code for any (including commercial) purposes, provided that authorship is credited. See the [LICENSE](LICENSE) file for details.

---

*Note: AI was used for analysis, documentation writing, and partial code authoring in this project.*
