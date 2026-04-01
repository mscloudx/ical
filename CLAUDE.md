# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go library for parsing, constructing, and serializing iCalendar data (RFC 5545). Supports streaming via `io.Reader`/`io.Writer` with minimal allocations. The module path is `github.com/mscloudx/ical`.

## Commands

```bash
make test       # run all tests
make lint       # run golangci-lint --fix
make fmt        # format with gofumpt + gci

go test ./...                    # alternative: all tests
go test -run TestName ./parser/  # single test
go test -bench=. ./parser/       # benchmarks
```

## Architecture

Four packages with a strict one-way dependency chain: `model` ← `parser`/`builder`/`encoder` (no cross-dependencies between parser, builder, encoder).

- **`model/`** — pure data structures: `Calendar`, `Event`, `Todo`, `Journal`, `Alarm`, `Timezone`, `Availability`, `RRule`, and all enumerations (`Status`, `Transparency`, `Role`, `PartStat`, etc.)
- **`parser/`** — streaming parser. `Parse(io.Reader)` / `ParseBytes([]byte)` / `ParseWithClose(io.ReadCloser)`. Internally: `lineScanner` (with RFC 5545 line-folding and CRLF/LF handling) → `rawComponent` tree → final model. Complex value parsing (dates, RRULE, TZID, attendees) lives in `value.go`. Generated code is in `*_gen.go` files.
- **`builder/`** — Functional Options pattern. `NewCalendar(prodid, ...opts)`, `NewEvent(uid, dtstart, dtend, ...opts)`, etc. Each component type has `With*` option constructors.
- **`encoder/`** — `Encode(io.Writer, *model.Calendar)`. Handles line folding (max 75 octets per RFC), CRLF, text escaping. Formatting helpers in `format.go`.

## Code Style

- **Language:** all comments and docstrings messages are in **English**; code identifiers, stdlib and commit usage in English.
- **Commit format:** `<type>(<scope>): <description in English>` — types: `feat`, `fix`, `chore`, `bump`, `test`, `docs`, `refactor`.
- **Testing:** `testify/suite` with AAA pattern (Arrange, Act, Assert). Test data (real `.ics` files from various vendors) lives in `examples/ical/`.
- **Linting:** 26+ linters via golangci-lint. `exhaustive` requires handling all enum values (except `*Unspecified` members). `errcheck` checks type assertions and blank assignments.

## RFC Support

RFC 5545 (core) plus extensions: RFC 7986 (calendar metadata), RFC 7953 (`VAVAILABILITY`), RFC 9073 (publishing extensions), RFC 9253 (relationships), RFC 9074 (`VALARM` extensions), RFC 7529 (non-Gregorian RRULE via `RSCALE`/`SKIP`).

## Key Rules

- **Never merge directly to `main`** — always create a PR/MR.
- Generated files (`parser/*_gen.go`) have relaxed lint rules (`gocognit`, `gocyclo`, `mnd` excluded).
- Test files are excluded from `errcheck`, `funlen`, `gocognit`, `gocyclo`, `mnd`, `nilnil`.
