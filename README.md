# CalForge

A comprehensive iCalendar (RFC 5545) and vCard (RFC 6350) processing toolkit for Go.

## Features

- **Parse** iCalendar (.ics) and vCard (.vcf) files
- **Validate** calendar and contact data for common issues
- **Filter** events by date range, status, category, or organizer
- **Search** events and contacts by text
- **Convert** between formats (ICS↔JSON, ICS↔CSV, VCF↔JSON, VCF↔CSV)
- **Merge** multiple calendar files into one
- **Generate** sample calendars and contacts
- **Statistics** for calendar analysis

## Quick Start

```bash
# Install
go install github.com/EdgarOrtegaRamirez/calforge/cmd/calforge@latest

# Parse an iCalendar file
calforge parse events.ics

# Validate a calendar
calforge validate events.ics

# Show statistics
calforge stats events.ics

# Convert to JSON
calforge convert --from ics --to json events.ics

# Convert contacts to CSV
calforge convert --from vcf --to csv contacts.vcf

# Merge calendars
calforge merge cal1.ics cal2.ics -o merged.ics

# Create sample data
calforge create --type ics --events 5
calforge create --type vcf --contacts 3
```

## Commands

| Command | Description |
|---------|-------------|
| `parse` | Parse and display iCalendar or vCard files |
| `validate` | Validate files for common issues |
| `stats` | Show calendar statistics |
| `filter` | Filter events by date, status, category |
| `search` | Search events by text |
| `convert` | Convert between formats |
| `merge` | Merge multiple calendar files |
| `create` | Generate sample files |
| `vcard` | vCard-specific operations |

## Supported Formats

### Input
- `.ics` / `.ical` / `.ifb` — iCalendar (RFC 5545)
- `.vcf` / `.vcard` — vCard (RFC 6350)

### Output
- JSON
- CSV
- iCalendar format
- vCard format

## Architecture

```
calforge/
├── cmd/calforge/       # CLI entry point
├── internal/
│   ├── ical/          # iCalendar parser, serializer, query engine
│   ├── vcard/         # vCard parser, serializer, validator
│   ├── convert/       # Format conversion (JSON, CSV)
│   └── cmd/           # CLI command implementations
└── tests/             # Test suites
```

## Library Usage

```go
import "github.com/EdgarOrtegaRamirez/calforge/internal/ical"

// Parse an iCalendar file
cal, err := ical.Parse(icsData)

// Filter events by date range
events := cal.FilterEventsByDateRange(start, end)

// Search events
events := cal.SearchEvents("meeting")

// Get statistics
stats := cal.Stats()
fmt.Print(stats.StatsText())

// Serialize back to iCalendar format
output := cal.Serialize()
```

## License

MIT
