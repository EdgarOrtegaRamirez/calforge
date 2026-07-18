// Package ical implements iCalendar (RFC 5545) parsing and generation.
package ical

import (
	"fmt"
	"strings"
	"time"
)

// Calendar represents a complete iCalendar object.
type Calendar struct {
	ProdID      string
	Version     string
	CalName     string
	Description string
	Events      []*Event
	Todos       []*Todo
	Journals    []*Journal
	FreeBusy    []*FreeBusy
	TimeZones   []*TimeZone
	Properties  []Property
}

// Event represents a VEVENT component.
type Event struct {
	UID            string
	Summary        string
	Description    string
	Location       string
	URL            string
	Status         string // TENTATIVE, CONFIRMED, CANCELLED
	Transp         string // OPAQUE, TRANSPARENT
	Organizer      *Attendee
	Attendees      []*Attendee
	StartTime      *time.Time
	EndTime        *time.Time
	Due            *time.Time
	Duration       string
	Created        *time.Time
	LastModified   *time.Time
	Sequence       int
	RecurrenceRule string
	ExDate         []time.Time
	RDate          []time.Time
	Alarms         []*Alarm
	Categories     []string
	Attachments    []Attachment
	Properties     []Property
}

// Todo represents a VTODO component.
type Todo struct {
	UID             string
	Summary         string
	Description     string
	Status          string // NEEDS-ACTION, COMPLETED, IN-PROCESS, CANCELLED
	Priority        int
	StartTime       *time.Time
	Due             *time.Time
	Duration        string
	Completed       *time.Time
	PercentComplete int
	Created         *time.Time
	LastModified    *time.Time
	RecurrenceRule  string
	Alarms          []*Alarm
	Properties      []Property
}

// Journal represents a VJOURNAL component.
type Journal struct {
	UID          string
	Summary      string
	Description  string
	Status       string
	StartTime    *time.Time
	Created      *time.Time
	LastModified *time.Time
	Properties   []Property
}

// FreeBusy represents a VFREEBUSY component.
type FreeBusy struct {
	UID          string
	StartTime    *time.Time
	EndTime      *time.Time
	Organizer    string
	FreeBusyList []FreeBusyInterval
	Properties   []Property
}

// FreeBusyInterval represents a free/busy time interval.
type FreeBusyInterval struct {
	Start time.Time
	End   time.Time
	Type  string // FREE, BUSY, BUSY-UNAVAILABLE, BUSY-TENTATIVE
}

// TimeZone represents a VTIMEZONE component.
type TimeZone struct {
	ID         string
	StartDate  *time.Time
	Standard   *TimeZoneObservance
	Daylight   *TimeZoneObservance
	Properties []Property
}

// TimeZoneObservance represents STANDARD or DAYLIGHT sub-component.
type TimeZoneObservance struct {
	OffsetFrom     string
	OffsetTo       string
	StartTime      *time.Time
	RecurrenceRule string
	Abbreviation   string
	Properties     []Property
}

// Alarm represents a VALARM sub-component.
type Alarm struct {
	Action      string
	Trigger     string
	Description string
	Repeat      int
	Duration    string
	Attachments []Attachment
	Properties  []Property
}

// Attendee represents an ATTENDEE property.
type Attendee struct {
	CN            string
	Email         string
	Role          string // CHAIR, REQ-PARTICIPANT, OPT-PARTICIPANT, NON-PARTICIPANT
	PartStat      string // NEEDS-ACTION, ACCEPTED, DECLINED, TENTATIVE, DELEGATED
	Rsvp          bool
	DelegatedTo   string
	DelegatedFrom string
	SentBy        string
	Dir           string
	Properties    []Property
}

// Attachment represents an ATTACH property.
type Attachment struct {
	URI        string
	Binary     string
	Format     string
	Properties []Property
}

// Property represents an iCalendar property with parameters.
type Property struct {
	Name       string
	Value      string
	Parameters map[string][]string
}

// Parse parses an iCalendar string into a Calendar object.
func Parse(input string) (*Calendar, error) {
	c := &Calendar{
		Events:     make([]*Event, 0),
		Todos:      make([]*Todo, 0),
		Journals:   make([]*Journal, 0),
		FreeBusy:   make([]*FreeBusy, 0),
		TimeZones:  make([]*TimeZone, 0),
		Properties: make([]Property, 0),
	}

	lines := unfoldAndParse(input)
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	// Find VCALENDAR boundaries
	if !strings.HasPrefix(lines[0].name, "BEGIN:VCALENDAR") {
		return nil, fmt.Errorf("expected BEGIN:VCALENDAR, got %s", lines[0].name)
	}

	// Parse content lines
	currentComponent := ""
	var currentEvent *Event
	var currentTodo *Todo
	var currentJournal *Journal
	var currentFreeBusy *FreeBusy
	var currentTz *TimeZone
	var currentAlarm *Alarm
	var currentObservance *TimeZoneObservance
	inAlarm := false
	inObservance := false

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		name := strings.ToUpper(line.name)

		switch {
		case name == "BEGIN:VEVENT":
			currentComponent = "VEVENT"
			currentEvent = &Event{Attendees: make([]*Attendee, 0), Properties: make([]Property, 0)}
		case name == "END:VEVENT":
			if currentEvent != nil {
				c.Events = append(c.Events, currentEvent)
			}
			currentEvent = nil
			currentComponent = ""
		case name == "BEGIN:VTODO":
			currentComponent = "VTODO"
			currentTodo = &Todo{Properties: make([]Property, 0)}
		case name == "END:VTODO":
			if currentTodo != nil {
				c.Todos = append(c.Todos, currentTodo)
			}
			currentTodo = nil
			currentComponent = ""
		case name == "BEGIN:VJOURNAL":
			currentComponent = "VJOURNAL"
			currentJournal = &Journal{Properties: make([]Property, 0)}
		case name == "END:VJOURNAL":
			if currentJournal != nil {
				c.Journals = append(c.Journals, currentJournal)
			}
			currentJournal = nil
			currentComponent = ""
		case name == "BEGIN:VFREEBUSY":
			currentComponent = "VFREEBUSY"
			currentFreeBusy = &FreeBusy{FreeBusyList: make([]FreeBusyInterval, 0), Properties: make([]Property, 0)}
		case name == "END:VFREEBUSY":
			if currentFreeBusy != nil {
				c.FreeBusy = append(c.FreeBusy, currentFreeBusy)
			}
			currentFreeBusy = nil
			currentComponent = ""
		case name == "BEGIN:VTIMEZONE":
			currentComponent = "VTIMEZONE"
			currentTz = &TimeZone{Properties: make([]Property, 0)}
		case name == "END:VTIMEZONE":
			if currentTz != nil {
				c.TimeZones = append(c.TimeZones, currentTz)
			}
			currentTz = nil
			currentComponent = ""
		case name == "BEGIN:VALARM":
			inAlarm = true
			currentAlarm = &Alarm{Properties: make([]Property, 0)}
		case name == "END:VALARM":
			inAlarm = false
			if currentAlarm != nil {
				switch currentComponent {
				case "VEVENT":
					if currentEvent != nil {
						currentEvent.Alarms = append(currentEvent.Alarms, currentAlarm)
					}
				case "VTODO":
					if currentTodo != nil {
						currentTodo.Alarms = append(currentTodo.Alarms, currentAlarm)
					}
				}
			}
			currentAlarm = nil
		case name == "BEGIN:STANDARD", name == "BEGIN:DAYLIGHT":
			inObservance = true
			currentObservance = &TimeZoneObservance{Properties: make([]Property, 0)}
		case name == "END:STANDARD", name == "END:DAYLIGHT":
			inObservance = false
			if currentTz != nil && currentObservance != nil {
				if strings.HasSuffix(name, "STANDARD") {
					currentTz.Standard = currentObservance
				} else {
					currentTz.Daylight = currentObservance
				}
			}
			currentObservance = nil
		default:
			if name == "END:VCALENDAR" {
				continue
			}

			prop := parseProperty(line)

			// Dispatch property to appropriate component
			if inAlarm && currentAlarm != nil {
				applyAlarmProperty(currentAlarm, prop)
			} else if inObservance && currentObservance != nil {
				applyObservanceProperty(currentObservance, prop)
			} else {
				switch currentComponent {
				case "VEVENT":
					if currentEvent != nil {
						applyEventProperty(currentEvent, prop)
					}
				case "VTODO":
					if currentTodo != nil {
						applyTodoProperty(currentTodo, prop)
					}
				case "VJOURNAL":
					if currentJournal != nil {
						applyJournalProperty(currentJournal, prop)
					}
				case "VFREEBUSY":
					if currentFreeBusy != nil {
						applyFreeBusyProperty(currentFreeBusy, prop)
					}
				case "VTIMEZONE":
					if currentTz != nil {
						applyTimeZoneProperty(currentTz, prop)
					}
				default:
					// Top-level calendar properties
					applyCalendarProperty(c, prop)
				}
			}
		}
	}

	return c, nil
}

// parsedLine represents a parsed content line.
type parsedLine struct {
	name       string
	params     map[string][]string
	value      string
	paramOrder []string
}

// unfoldAndParse handles RFC 5545 line unfolding and parses lines.
func unfoldAndParse(input string) []parsedLine {
	// Unfold: replace CRLF followed by space/tab with empty string
	input = strings.ReplaceAll(input, "\r\n ", "")
	input = strings.ReplaceAll(input, "\r\n\t", "")
	input = strings.ReplaceAll(input, "\n ", "")
	input = strings.ReplaceAll(input, "\n\t", "")

	lines := strings.Split(input, "\n")
	result := make([]parsedLine, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pl := parseLine(line)
		result = append(result, pl)
	}
	return result
}

// parseLine parses a single iCalendar content line.
func parseLine(line string) parsedLine {
	pl := parsedLine{
		params: make(map[string][]string),
	}

	// Find the property name (before first ':')
	colonIdx := strings.Index(line, ":")
	if colonIdx < 0 {
		pl.name = line
		return pl
	}

	nameAndParams := line[:colonIdx]
	pl.value = line[colonIdx+1:]

	// For BEGIN/END lines, include the component name in the property name
	upperLine := strings.ToUpper(line)
	if strings.HasPrefix(upperLine, "BEGIN:") || strings.HasPrefix(upperLine, "END:") {
		pl.name = upperLine
		return pl
	}

	// Parse property name and parameters
	// Format: PROPNAME;param1=val1;param2=val2
	parts := strings.SplitN(nameAndParams, ";", 2)
	pl.name = strings.ToUpper(parts[0])

	if len(parts) > 1 {
		parseParameters(parts[1], pl.params)
	}

	return pl
}

// parseParameters parses property parameters.
func parseParameters(paramStr string, params map[string][]string) {
	// Split by ';' but respect quoted values
	segments := splitParams(paramStr)
	for _, seg := range segments {
		eqIdx := strings.Index(seg, "=")
		if eqIdx < 0 {
			continue
		}
		key := strings.TrimSpace(seg[:eqIdx])
		val := strings.TrimSpace(seg[eqIdx+1:])
		// Remove quotes if present
		val = strings.Trim(val, "\"")
		params[key] = append(params[key], val)
	}
}

// splitParams splits parameter string by semicolons, respecting quotes.
func splitParams(s string) []string {
	var result []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			inQuote = !inQuote
			current.WriteByte(s[i])
		case ';':
			if !inQuote {
				result = append(result, current.String())
				current.Reset()
				continue
			}
			current.WriteByte(s[i])
		default:
			current.WriteByte(s[i])
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}

// parseProperty converts a parsedLine to a Property.
func parseProperty(line parsedLine) Property {
	return Property{
		Name:       line.name,
		Value:      line.value,
		Parameters: line.params,
	}
}

// parseTime attempts to parse an iCalendar datetime string.
func parseTime(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// Try various formats
	formats := []string{
		"20060102T150405Z",
		"20060102T150405",
		"20060102",
	}

	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return &t
		}
	}
	return nil
}

// applyCalendarProperty applies a property to the Calendar.
func applyCalendarProperty(c *Calendar, p Property) {
	switch p.Name {
	case "PRODID":
		c.ProdID = p.Value
	case "VERSION":
		c.Version = p.Value
	case "X-WR-CALNAME":
		c.CalName = p.Value
	case "X-WR-CALDESC":
		c.Description = p.Value
	default:
		c.Properties = append(c.Properties, p)
	}
}

// applyEventProperty applies a property to an Event.
func applyEventProperty(e *Event, p Property) {
	switch p.Name {
	case "UID":
		e.UID = p.Value
	case "SUMMARY":
		e.Summary = p.Value
	case "DESCRIPTION":
		e.Description = p.Value
	case "LOCATION":
		e.Location = p.Value
	case "URL":
		e.URL = p.Value
	case "STATUS":
		e.Status = p.Value
	case "TRANSP":
		e.Transp = p.Value
	case "DTSTART":
		e.StartTime = parseTime(p.Value)
	case "DTEND":
		e.EndTime = parseTime(p.Value)
	case "DUE":
		e.Due = parseTime(p.Value)
	case "DURATION":
		e.Duration = p.Value
	case "CREATED":
		e.Created = parseTime(p.Value)
	case "LAST-MODIFIED":
		e.LastModified = parseTime(p.Value)
	case "SEQUENCE":
		fmt.Sscanf(p.Value, "%d", &e.Sequence)
	case "RRULE":
		e.RecurrenceRule = p.Value
	case "CATEGORIES":
		cats := strings.Split(p.Value, ",")
		for i := range cats {
			cats[i] = strings.TrimSpace(cats[i])
		}
		e.Categories = append(e.Categories, cats...)
	case "ORGANIZER":
		email := extractMailto(p.Value)
		cn := ""
		if cnVals, ok := p.Parameters["CN"]; ok && len(cnVals) > 0 {
			cn = cnVals[0]
		}
		e.Organizer = &Attendee{Email: email, CN: cn}
	case "ATTENDEE":
		a := parseAttendee(p)
		e.Attendees = append(e.Attendees, a)
	default:
		e.Properties = append(e.Properties, p)
	}
}

// applyTodoProperty applies a property to a Todo.
func applyTodoProperty(t *Todo, p Property) {
	switch p.Name {
	case "UID":
		t.UID = p.Value
	case "SUMMARY":
		t.Summary = p.Value
	case "DESCRIPTION":
		t.Description = p.Value
	case "STATUS":
		t.Status = p.Value
	case "PRIORITY":
		fmt.Sscanf(p.Value, "%d", &t.Priority)
	case "DTSTART":
		t.StartTime = parseTime(p.Value)
	case "DUE":
		t.Due = parseTime(p.Value)
	case "DURATION":
		t.Duration = p.Value
	case "COMPLETED":
		t.Completed = parseTime(p.Value)
	case "PERCENT-COMPLETE":
		fmt.Sscanf(p.Value, "%d", &t.PercentComplete)
	case "CREATED":
		t.Created = parseTime(p.Value)
	case "LAST-MODIFIED":
		t.LastModified = parseTime(p.Value)
	case "RRULE":
		t.RecurrenceRule = p.Value
	default:
		t.Properties = append(t.Properties, p)
	}
}

// applyJournalProperty applies a property to a Journal.
func applyJournalProperty(j *Journal, p Property) {
	switch p.Name {
	case "UID":
		j.UID = p.Value
	case "SUMMARY":
		j.Summary = p.Value
	case "DESCRIPTION":
		j.Description = p.Value
	case "STATUS":
		j.Status = p.Value
	case "DTSTART":
		j.StartTime = parseTime(p.Value)
	case "CREATED":
		j.Created = parseTime(p.Value)
	case "LAST-MODIFIED":
		j.LastModified = parseTime(p.Value)
	default:
		j.Properties = append(j.Properties, p)
	}
}

// applyFreeBusyProperty applies a property to a FreeBusy.
func applyFreeBusyProperty(fb *FreeBusy, p Property) {
	switch p.Name {
	case "UID":
		fb.UID = p.Value
	case "DTSTART":
		fb.StartTime = parseTime(p.Value)
	case "DTEND":
		fb.EndTime = parseTime(p.Value)
	case "ORGANIZER":
		fb.Organizer = extractMailto(p.Value)
	case "FREEBUSY":
		// Parse FBTIME:FREE/BUSY/...
		parts := strings.SplitN(p.Value, "/", 2)
		if len(parts) == 2 {
			interval := FreeBusyInterval{}
			interval.Start = safeParseTime(parts[0])
			interval.End = safeParseTime(parts[1])
			if fbType, ok := p.Parameters["FBTYPE"]; ok && len(fbType) > 0 {
				interval.Type = fbType[0]
			} else {
				interval.Type = "BUSY"
			}
			fb.FreeBusyList = append(fb.FreeBusyList, interval)
		}
	default:
		fb.Properties = append(fb.Properties, p)
	}
}

// applyTimeZoneProperty applies a property to a TimeZone.
func applyTimeZoneProperty(tz *TimeZone, p Property) {
	switch p.Name {
	case "TZID":
		tz.ID = p.Value
	case "DTSTART":
		tz.StartDate = parseTime(p.Value)
	default:
		tz.Properties = append(tz.Properties, p)
	}
}

// applyObservanceProperty applies a property to a TimeZoneObservance.
func applyObservanceProperty(o *TimeZoneObservance, p Property) {
	switch p.Name {
	case "TZOFFSETFROM":
		o.OffsetFrom = p.Value
	case "TZOFFSETTO":
		o.OffsetTo = p.Value
	case "DTSTART":
		o.StartTime = parseTime(p.Value)
	case "RRULE":
		o.RecurrenceRule = p.Value
	case "TZNAME":
		o.Abbreviation = p.Value
	default:
		o.Properties = append(o.Properties, p)
	}
}

// applyAlarmProperty applies a property to an Alarm.
func applyAlarmProperty(a *Alarm, p Property) {
	switch p.Name {
	case "ACTION":
		a.Action = p.Value
	case "TRIGGER":
		a.Trigger = p.Value
	case "DESCRIPTION":
		a.Description = p.Value
	case "REPEAT":
		fmt.Sscanf(p.Value, "%d", &a.Repeat)
	case "DURATION":
		a.Duration = p.Value
	case "ATTACH":
		a.Attachments = append(a.Attachments, Attachment{URI: p.Value})
	default:
		a.Properties = append(a.Properties, p)
	}
}

// parseAttendee parses an ATTENDEE property.
func parseAttendee(p Property) *Attendee {
	a := &Attendee{
		Email: extractMailto(p.Value),
	}

	if cn, ok := p.Parameters["CN"]; ok && len(cn) > 0 {
		a.CN = cn[0]
	}
	if role, ok := p.Parameters["ROLE"]; ok && len(role) > 0 {
		a.Role = role[0]
	}
	if ps, ok := p.Parameters["PARTSTAT"]; ok && len(ps) > 0 {
		a.PartStat = ps[0]
	}
	if rsvp, ok := p.Parameters["RSVP"]; ok && len(rsvp) > 0 {
		a.Rsvp = strings.ToUpper(rsvp[0]) == "TRUE"
	}
	if dTo, ok := p.Parameters["DELEGATED-TO"]; ok && len(dTo) > 0 {
		a.DelegatedTo = extractMailto(dTo[0])
	}
	if dFrom, ok := p.Parameters["DELEGATED-FROM"]; ok && len(dFrom) > 0 {
		a.DelegatedFrom = extractMailto(dFrom[0])
	}
	if sentBy, ok := p.Parameters["SENT-BY"]; ok && len(sentBy) > 0 {
		a.SentBy = extractMailto(sentBy[0])
	}

	return a
}

// extractMailto extracts the email from a mailto: URI.
func extractMailto(s string) string {
	return strings.TrimPrefix(strings.ToLower(s), "mailto:")
}

// safeParseTime parses a time string, returning zero time on error.
func safeParseTime(s string) time.Time {
	if t := parseTime(s); t != nil {
		return *t
	}
	return time.Time{}
}
