# AGENTS.md — CalForge

## What This Project Is
CalForge is a Go CLI tool and library for processing iCalendar (RFC 5545) and vCard (RFC 6350) files.

## Project Structure
```
cmd/calforge/main.go     # CLI entry point
internal/ical/           # iCalendar parser, serializer, query engine
internal/vcard/          # vCard parser, serializer, validator
internal/convert/        # Format conversion (JSON, CSV)
internal/cmd/            # CLI command implementations
tests/                   # Test suites (ical, vcard, convert)
```

## Build & Test
```bash
go build ./cmd/calforge/
go test ./tests/... -v
go vet ./...
```

## Key Design Decisions
- Custom iCalendar parser (no external dependencies)
- Custom vCard parser with RFC 6350 support
- Line unfolding per RFC 5545 §3.1
- BEGIN/END component parsing with proper nesting
- Property parameters parsed separately from values

## Dependencies
- Zero external dependencies (stdlib only)

## Testing
- 18 tests across 3 test packages
- Covers parsing, serialization, validation, filtering, search, conversion
- Uses realistic test data with multiple events, todos, attendees
