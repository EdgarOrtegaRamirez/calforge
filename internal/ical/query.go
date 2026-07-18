package ical

import (
	"fmt"
	"strings"
	"time"
)

// Validate checks a Calendar for common issues and returns a list of problems.
func Validate(cal *Calendar) []string {
	var issues []string

	if cal.ProdID == "" {
		issues = append(issues, "missing PRODID")
	}
	if cal.Version == "" {
		issues = append(issues, "missing VERSION")
	}

	// Check events
	uids := make(map[string]bool)
	for i, e := range cal.Events {
		if e.UID == "" {
			issues = append(issues, fmt.Sprintf("event[%d]: missing UID", i))
		} else {
			if uids[e.UID] {
				issues = append(issues, fmt.Sprintf("event[%d]: duplicate UID %q", i, e.UID))
			}
			uids[e.UID] = true
		}
		if e.Summary == "" {
			issues = append(issues, fmt.Sprintf("event[%d] (UID=%s): missing SUMMARY", i, e.UID))
		}
		if e.StartTime != nil && e.EndTime != nil {
			if e.EndTime.Before(*e.StartTime) {
				issues = append(issues, fmt.Sprintf("event[%d] (UID=%s): DTEND before DTSTART", i, e.UID))
			}
		}
		// Check attendees
		for j, a := range e.Attendees {
			if a.Email == "" {
				issues = append(issues, fmt.Sprintf("event[%d].attendee[%d]: missing email", i, j))
			}
		}
	}

	// Check todos
	for i, t := range cal.Todos {
		if t.UID == "" {
			issues = append(issues, fmt.Sprintf("todo[%d]: missing UID", i))
		}
		if t.PercentComplete < 0 || t.PercentComplete > 100 {
			issues = append(issues, fmt.Sprintf("todo[%d]: PERCENT-COMPLETE out of range (%d)", i, t.PercentComplete))
		}
		if t.Priority < 0 || t.Priority > 9 {
			issues = append(issues, fmt.Sprintf("todo[%d]: PRIORITY out of range (%d)", i, t.Priority))
		}
	}

	return issues
}

// FilterEventsByDateRange returns events that overlap with the given date range.
func (c *Calendar) FilterEventsByDateRange(start, end time.Time) []*Event {
	var result []*Event
	for _, e := range c.Events {
		if e.StartTime == nil {
			continue
		}
		eventEnd := end
		if e.EndTime != nil {
			eventEnd = *e.EndTime
		}
		if e.StartTime.Before(end) && eventEnd.After(start) {
			result = append(result, e)
		}
	}
	return result
}

// FilterEventsByStatus returns events matching the given status.
func (c *Calendar) FilterEventsByStatus(status string) []*Event {
	status = strings.ToUpper(status)
	var result []*Event
	for _, e := range c.Events {
		if strings.ToUpper(e.Status) == status {
			result = append(result, e)
		}
	}
	return result
}

// FilterEventsByOrganizer returns events organized by the given email.
func (c *Calendar) FilterEventsByOrganizer(email string) []*Event {
	email = strings.ToLower(email)
	var result []*Event
	for _, e := range c.Events {
		if e.Organizer != nil && strings.ToLower(e.Organizer.Email) == email {
			result = append(result, e)
		}
	}
	return result
}

// FilterEventsByCategory returns events with the given category.
func (c *Calendar) FilterEventsByCategory(category string) []*Event {
	category = strings.ToLower(category)
	var result []*Event
	for _, e := range c.Events {
		for _, cat := range e.Categories {
			if strings.ToLower(cat) == category {
				result = append(result, e)
				break
			}
		}
	}
	return result
}

// SearchEvents searches events by text in summary, description, and location.
func (c *Calendar) SearchEvents(query string) []*Event {
	query = strings.ToLower(query)
	var result []*Event
	for _, e := range c.Events {
		if strings.Contains(strings.ToLower(e.Summary), query) ||
			strings.Contains(strings.ToLower(e.Description), query) ||
			strings.Contains(strings.ToLower(e.Location), query) {
			result = append(result, e)
		}
	}
	return result
}

// GetUpcomingEvents returns events sorted by start time within the next N days.
func (c *Calendar) GetUpcomingEvents(days int) []*Event {
	now := time.Now()
	future := now.AddDate(0, 0, days)
	events := c.FilterEventsByDateRange(now, future)

	// Sort by start time
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].StartTime != nil && events[j].StartTime != nil {
				if events[j].StartTime.Before(*events[i].StartTime) {
					events[i], events[j] = events[j], events[i]
				}
			}
		}
	}
	return events
}

// Stats returns calendar statistics.
func (c *Calendar) Stats() CalendarStats {
	stats := CalendarStats{
		TotalEvents:    len(c.Events),
		TotalTodos:     len(c.Todos),
		TotalJournals:  len(c.Journals),
		TotalFreeBusy:  len(c.FreeBusy),
		TotalTimeZones: len(c.TimeZones),
	}

	statusCounts := make(map[string]int)
	categoryCounts := make(map[string]int)
	attendeeEmails := make(map[string]bool)

	for _, e := range c.Events {
		if e.Status != "" {
			statusCounts[strings.ToUpper(e.Status)]++
		}
		for _, cat := range e.Categories {
			categoryCounts[cat]++
		}
		for _, a := range e.Attendees {
			attendeeEmails[strings.ToLower(a.Email)] = true
		}
		if e.StartTime != nil && (stats.EarliestEvent.IsZero() || e.StartTime.Before(stats.EarliestEvent)) {
			stats.EarliestEvent = *e.StartTime
		}
		if e.EndTime != nil && (stats.LatestEvent.IsZero() || e.EndTime.After(stats.LatestEvent)) {
			stats.LatestEvent = *e.EndTime
		}
	}

	stats.StatusCounts = statusCounts
	stats.CategoryCounts = categoryCounts
	stats.UniqueAttendees = len(attendeeEmails)

	return stats
}

// CalendarStats holds calendar statistics.
type CalendarStats struct {
	TotalEvents     int
	TotalTodos      int
	TotalJournals   int
	TotalFreeBusy   int
	TotalTimeZones  int
	UniqueAttendees int
	EarliestEvent   time.Time
	LatestEvent     time.Time
	StatusCounts    map[string]int
	CategoryCounts  map[string]int
}

// StatsText returns a human-readable stats string.
func (s CalendarStats) StatsText() string {
	var b strings.Builder
	b.WriteString("Calendar Statistics\n")
	b.WriteString("==================\n")
	b.WriteString(fmt.Sprintf("Events:        %d\n", s.TotalEvents))
	b.WriteString(fmt.Sprintf("Todos:         %d\n", s.TotalTodos))
	b.WriteString(fmt.Sprintf("Journals:      %d\n", s.TotalJournals))
	b.WriteString(fmt.Sprintf("Free/Busy:     %d\n", s.TotalFreeBusy))
	b.WriteString(fmt.Sprintf("Time Zones:    %d\n", s.TotalTimeZones))
	b.WriteString(fmt.Sprintf("Attendees:     %d\n", s.UniqueAttendees))

	if !s.EarliestEvent.IsZero() {
		b.WriteString(fmt.Sprintf("Earliest:      %s\n", s.EarliestEvent.Format("2006-01-02 15:04")))
	}
	if !s.LatestEvent.IsZero() {
		b.WriteString(fmt.Sprintf("Latest:        %s\n", s.LatestEvent.Format("2006-01-02 15:04")))
	}

	if len(s.StatusCounts) > 0 {
		b.WriteString("\nBy Status:\n")
		for status, count := range s.StatusCounts {
			b.WriteString(fmt.Sprintf("  %s: %d\n", status, count))
		}
	}
	if len(s.CategoryCounts) > 0 {
		b.WriteString("\nBy Category:\n")
		for cat, count := range s.CategoryCounts {
			b.WriteString(fmt.Sprintf("  %s: %d\n", cat, count))
		}
	}

	return b.String()
}
