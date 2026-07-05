package ical

import (
	"fmt"
	"strings"
	"time"
)

// Serialize converts a Calendar to iCalendar format string.
func (c *Calendar) Serialize() string {
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	if c.ProdID != "" {
		b.WriteString(fmt.Sprintf("PRODID:%s\r\n", c.ProdID))
	} else {
		b.WriteString("PRODID:-//CalForge//calforge//EN\r\n")
	}
	if c.CalName != "" {
		b.WriteString(fmt.Sprintf("X-WR-CALNAME:%s\r\n", c.CalName))
	}
	if c.Description != "" {
		b.WriteString(fmt.Sprintf("X-WR-CALDESC:%s\r\n", c.Description))
	}

	// Write properties
	for _, p := range c.Properties {
		writeProperty(&b, p)
	}

	// Write timezones
	for _, tz := range c.TimeZones {
		writeTimeZone(&b, tz)
	}

	// Write events
	for _, e := range c.Events {
		writeEvent(&b, e)
	}

	// Write todos
	for _, t := range c.Todos {
		writeTodo(&b, t)
	}

	// Write journals
	for _, j := range c.Journals {
		writeJournal(&b, j)
	}

	b.WriteString("END:VCALENDAR\r\n")
	return b.String()
}

// writeProperty writes a property to the builder.
func writeProperty(b *strings.Builder, p Property) {
	if len(p.Parameters) > 0 {
		params := make([]string, 0, len(p.Parameters))
		for k, vals := range p.Parameters {
			for _, v := range vals {
				params = append(params, fmt.Sprintf("%s=%s", k, v))
			}
		}
		b.WriteString(fmt.Sprintf("%s;%s:%s\r\n", p.Name, strings.Join(params, ";"), p.Value))
	} else {
		b.WriteString(fmt.Sprintf("%s:%s\r\n", p.Name, p.Value))
	}
}

// writeEvent writes an event to the builder.
func writeEvent(b *strings.Builder, e *Event) {
	b.WriteString("BEGIN:VEVENT\r\n")

	if e.UID != "" {
		b.WriteString(fmt.Sprintf("UID:%s\r\n", e.UID))
	}
	if e.Summary != "" {
		b.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", e.Summary))
	}
	if e.Description != "" {
		b.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", e.Description))
	}
	if e.Location != "" {
		b.WriteString(fmt.Sprintf("LOCATION:%s\r\n", e.Location))
	}
	if e.URL != "" {
		b.WriteString(fmt.Sprintf("URL:%s\r\n", e.URL))
	}
	if e.Status != "" {
		b.WriteString(fmt.Sprintf("STATUS:%s\r\n", e.Status))
	}
	if e.Transp != "" {
		b.WriteString(fmt.Sprintf("TRANSP:%s\r\n", e.Transp))
	}
	if e.StartTime != nil {
		b.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatTime(*e.StartTime)))
	}
	if e.EndTime != nil {
		b.WriteString(fmt.Sprintf("DTEND:%s\r\n", formatTime(*e.EndTime)))
	}
	if e.Duration != "" {
		b.WriteString(fmt.Sprintf("DURATION:%s\r\n", e.Duration))
	}
	if e.Created != nil {
		b.WriteString(fmt.Sprintf("CREATED:%s\r\n", formatTime(*e.Created)))
	}
	if e.LastModified != nil {
		b.WriteString(fmt.Sprintf("LAST-MODIFIED:%s\r\n", formatTime(*e.LastModified)))
	}
	if e.Sequence > 0 {
		b.WriteString(fmt.Sprintf("SEQUENCE:%d\r\n", e.Sequence))
	}
	if e.RecurrenceRule != "" {
		b.WriteString(fmt.Sprintf("RRULE:%s\r\n", e.RecurrenceRule))
	}
	if len(e.Categories) > 0 {
		b.WriteString(fmt.Sprintf("CATEGORIES:%s\r\n", strings.Join(e.Categories, ",")))
	}

	// Write organizer
	if e.Organizer != nil {
		b.WriteString("ORGANIZER")
		if e.Organizer.CN != "" {
			b.WriteString(fmt.Sprintf(";CN=%s", e.Organizer.CN))
		}
		b.WriteString(fmt.Sprintf(":mailto:%s\r\n", e.Organizer.Email))
	}

	// Write attendees
	for _, a := range e.Attendees {
		writeAttendee(b, a)
	}

	// Write properties
	for _, p := range e.Properties {
		writeProperty(b, p)
	}

	// Write alarms
	for _, alarm := range e.Alarms {
		writeAlarm(b, alarm)
	}

	b.WriteString("END:VEVENT\r\n")
}

// writeTodo writes a todo to the builder.
func writeTodo(b *strings.Builder, t *Todo) {
	b.WriteString("BEGIN:VTODO\r\n")

	if t.UID != "" {
		b.WriteString(fmt.Sprintf("UID:%s\r\n", t.UID))
	}
	if t.Summary != "" {
		b.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", t.Summary))
	}
	if t.Description != "" {
		b.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", t.Description))
	}
	if t.Status != "" {
		b.WriteString(fmt.Sprintf("STATUS:%s\r\n", t.Status))
	}
	if t.Priority > 0 {
		b.WriteString(fmt.Sprintf("PRIORITY:%d\r\n", t.Priority))
	}
	if t.StartTime != nil {
		b.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatTime(*t.StartTime)))
	}
	if t.Due != nil {
		b.WriteString(fmt.Sprintf("DUE:%s\r\n", formatTime(*t.Due)))
	}
	if t.Completed != nil {
		b.WriteString(fmt.Sprintf("COMPLETED:%s\r\n", formatTime(*t.Completed)))
	}
	if t.PercentComplete > 0 {
		b.WriteString(fmt.Sprintf("PERCENT-COMPLETE:%d\r\n", t.PercentComplete))
	}
	if t.RecurrenceRule != "" {
		b.WriteString(fmt.Sprintf("RRULE:%s\r\n", t.RecurrenceRule))
	}

	for _, p := range t.Properties {
		writeProperty(b, p)
	}

	for _, alarm := range t.Alarms {
		writeAlarm(b, alarm)
	}

	b.WriteString("END:VTODO\r\n")
}

// writeJournal writes a journal to the builder.
func writeJournal(b *strings.Builder, j *Journal) {
	b.WriteString("BEGIN:VJOURNAL\r\n")

	if j.UID != "" {
		b.WriteString(fmt.Sprintf("UID:%s\r\n", j.UID))
	}
	if j.Summary != "" {
		b.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", j.Summary))
	}
	if j.Description != "" {
		b.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", j.Description))
	}
	if j.Status != "" {
		b.WriteString(fmt.Sprintf("STATUS:%s\r\n", j.Status))
	}

	for _, p := range j.Properties {
		writeProperty(b, p)
	}

	b.WriteString("END:VJOURNAL\r\n")
}

// writeAttendee writes an attendee property.
func writeAttendee(b *strings.Builder, a *Attendee) {
	b.WriteString("ATTENDEE")
	if a.CN != "" {
		b.WriteString(fmt.Sprintf(";CN=%s", a.CN))
	}
	if a.Role != "" {
		b.WriteString(fmt.Sprintf(";ROLE=%s", a.Role))
	}
	if a.PartStat != "" {
		b.WriteString(fmt.Sprintf(";PARTSTAT=%s", a.PartStat))
	}
	if a.Rsvp {
		b.WriteString(";RSVP=TRUE")
	}
	b.WriteString(fmt.Sprintf(":mailto:%s\r\n", a.Email))
}

// writeAlarm writes an alarm sub-component.
func writeAlarm(b *strings.Builder, a *Alarm) {
	b.WriteString("BEGIN:VALARM\r\n")
	if a.Action != "" {
		b.WriteString(fmt.Sprintf("ACTION:%s\r\n", a.Action))
	}
	if a.Trigger != "" {
		b.WriteString(fmt.Sprintf("TRIGGER:%s\r\n", a.Trigger))
	}
	if a.Description != "" {
		b.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", a.Description))
	}
	if a.Repeat > 0 {
		b.WriteString(fmt.Sprintf("REPEAT:%d\r\n", a.Repeat))
	}
	b.WriteString("END:VALARM\r\n")
}

// writeTimeZone writes a timezone component.
func writeTimeZone(b *strings.Builder, tz *TimeZone) {
	b.WriteString("BEGIN:VTIMEZONE\r\n")
	if tz.ID != "" {
		b.WriteString(fmt.Sprintf("TZID:%s\r\n", tz.ID))
	}
	if tz.Standard != nil {
		writeObservance(b, "STANDARD", tz.Standard)
	}
	if tz.Daylight != nil {
		writeObservance(b, "DAYLIGHT", tz.Daylight)
	}
	b.WriteString("END:VTIMEZONE\r\n")
}

// writeObservance writes a STANDARD or DAYLIGHT sub-component.
func writeObservance(b *strings.Builder, name string, o *TimeZoneObservance) {
	b.WriteString(fmt.Sprintf("BEGIN:%s\r\n", name))
	if o.OffsetFrom != "" {
		b.WriteString(fmt.Sprintf("TZOFFSETFROM:%s\r\n", o.OffsetFrom))
	}
	if o.OffsetTo != "" {
		b.WriteString(fmt.Sprintf("TZOFFSETTO:%s\r\n", o.OffsetTo))
	}
	if o.StartTime != nil {
		b.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatTime(*o.StartTime)))
	}
	if o.RecurrenceRule != "" {
		b.WriteString(fmt.Sprintf("RRULE:%s\r\n", o.RecurrenceRule))
	}
	if o.Abbreviation != "" {
		b.WriteString(fmt.Sprintf("TZNAME:%s\r\n", o.Abbreviation))
	}
	b.WriteString(fmt.Sprintf("END:%s\r\n", name))
}

// formatTime formats a time.Time to iCalendar format.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("20060102T150405Z")
}
