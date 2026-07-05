package ical_test

import (
	"strings"
	"testing"
	"time"

	"github.com/EdgarOrtegaRamirez/calforge/internal/ical"
)

const testICS = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
X-WR-CALNAME:Test Calendar
BEGIN:VEVENT
UID:test-1@test.com
SUMMARY:Team Meeting
DESCRIPTION:Weekly team sync
DTSTART:20260710T100000Z
DTEND:20260710T110000Z
LOCATION:Conference Room A
STATUS:CONFIRMED
CATEGORIES:Meeting,Work
ATTENDEE;CN=Alice Smith;ROLE=REQ-PARTICIPANT;PARTSTAT=ACCEPTED:mailto:alice@example.com
ATTENDEE;CN=Bob Jones;ROLE=REQ-PARTICIPANT;PARTSTAT=TENTATIVE:mailto:bob@example.com
BEGIN:VALARM
ACTION:DISPLAY
DESCRIPTION:Meeting in 15 minutes
TRIGGER:-PT15M
END:VALARM
END:VEVENT
BEGIN:VEVENT
UID:test-2@test.com
SUMMARY:Lunch with Client
DTSTART:20260710T120000Z
DTEND:20260710T130000Z
LOCATION:Restaurant
STATUS:TENTATIVE
CATEGORIES:Work
END:VEVENT
BEGIN:VEVENT
UID:test-3@test.com
SUMMARY:Project Deadline
DTSTART:20260715T090000Z
DURATION:PT2H
STATUS:CONFIRMED
CATEGORIES:Deadline
END:VEVENT
BEGIN:VTODO
UID:todo-1@test.com
SUMMARY:Write documentation
STATUS:IN-PROCESS
PRIORITY:1
PERCENT-COMPLETE:50
END:VTODO
END:VCALENDAR`

func TestParse(t *testing.T) {
	cal, err := ical.Parse(testICS)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if cal.ProdID != "-//Test//Test//EN" {
		t.Errorf("ProdID = %q, want %q", cal.ProdID, "-//Test//Test//EN")
	}
	if cal.Version != "2.0" {
		t.Errorf("Version = %q, want %q", cal.Version, "2.0")
	}
	if cal.CalName != "Test Calendar" {
		t.Errorf("CalName = %q, want %q", cal.CalName, "Test Calendar")
	}

	// Events
	if len(cal.Events) != 3 {
		t.Fatalf("Events = %d, want 3", len(cal.Events))
	}

	e := cal.Events[0]
	if e.UID != "test-1@test.com" {
		t.Errorf("Event[0].UID = %q, want %q", e.UID, "test-1@test.com")
	}
	if e.Summary != "Team Meeting" {
		t.Errorf("Event[0].Summary = %q, want %q", e.Summary, "Team Meeting")
	}
	if e.StartTime == nil {
		t.Error("Event[0].StartTime is nil")
	} else {
		expected := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
		if !e.StartTime.Equal(expected) {
			t.Errorf("Event[0].StartTime = %v, want %v", e.StartTime, expected)
		}
	}
	if e.Location != "Conference Room A" {
		t.Errorf("Event[0].Location = %q, want %q", e.Location, "Conference Room A")
	}
	if e.Status != "CONFIRMED" {
		t.Errorf("Event[0].Status = %q, want %q", e.Status, "CONFIRMED")
	}
	if len(e.Categories) != 2 {
		t.Errorf("Event[0].Categories = %d, want 2", len(e.Categories))
	}
	if len(e.Attendees) != 2 {
		t.Errorf("Event[0].Attendees = %d, want 2", len(e.Attendees))
	}
	if len(e.Alarms) != 1 {
		t.Errorf("Event[0].Alarms = %d, want 1", len(e.Alarms))
	}

	// Check attendee details
	a := e.Attendees[0]
	if a.CN != "Alice Smith" {
		t.Errorf("Attendee[0].CN = %q, want %q", a.CN, "Alice Smith")
	}
	if a.Email != "alice@example.com" {
		t.Errorf("Attendee[0].Email = %q, want %q", a.Email, "alice@example.com")
	}
	if a.PartStat != "ACCEPTED" {
		t.Errorf("Attendee[0].PartStat = %q, want %q", a.PartStat, "ACCEPTED")
	}

	// Check duration event
	if cal.Events[2].Duration != "PT2H" {
		t.Errorf("Event[2].Duration = %q, want %q", cal.Events[2].Duration, "PT2H")
	}

	// Check todo
	if len(cal.Todos) != 1 {
		t.Fatalf("Todos = %d, want 1", len(cal.Todos))
	}
	todo := cal.Todos[0]
	if todo.Summary != "Write documentation" {
		t.Errorf("Todo.Summary = %q, want %q", todo.Summary, "Write documentation")
	}
	if todo.Priority != 1 {
		t.Errorf("Todo.Priority = %d, want 1", todo.Priority)
	}
	if todo.PercentComplete != 50 {
		t.Errorf("Todo.PercentComplete = %d, want 50", todo.PercentComplete)
	}
}

func TestSerialize(t *testing.T) {
	cal, err := ical.Parse(testICS)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output := cal.Serialize()

	// Re-parse to verify round-trip
	cal2, err := ical.Parse(output)
	if err != nil {
		t.Fatalf("Re-parse failed: %v", err)
	}

	if len(cal2.Events) != len(cal.Events) {
		t.Errorf("Events count = %d, want %d", len(cal2.Events), len(cal.Events))
	}
	if cal2.Events[0].Summary != cal.Events[0].Summary {
		t.Errorf("Event[0].Summary = %q, want %q", cal2.Events[0].Summary, cal.Events[0].Summary)
	}
}

func TestValidate(t *testing.T) {
	cal, err := ical.Parse(testICS)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	issues := ical.Validate(cal)
	if len(issues) != 0 {
		t.Errorf("Expected no issues, got %d: %v", len(issues), issues)
	}

	// Test validation with missing UID
	badCal := &ical.Calendar{
		ProdID:  "-//Test//Test//EN",
		Version: "2.0",
		Events: []*ical.Event{
			{Summary: "No UID Event"},
		},
	}
	issues = ical.Validate(badCal)
	if len(issues) == 0 {
		t.Error("Expected issues for missing UID")
	}
}

func TestFilterEvents(t *testing.T) {
	cal, err := ical.Parse(testICS)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Filter by date range
	start := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	events := cal.FilterEventsByDateRange(start, end)
	if len(events) != 2 {
		t.Errorf("Filter by date range: got %d events, want 2", len(events))
	}

	// Filter by status
	events = cal.FilterEventsByStatus("CONFIRMED")
	if len(events) != 2 {
		t.Errorf("Filter by status CONFIRMED: got %d events, want 2", len(events))
	}

	// Filter by category
	events = cal.FilterEventsByCategory("Meeting")
	if len(events) != 1 {
		t.Errorf("Filter by category Meeting: got %d events, want 1", len(events))
	}
}

func TestSearchEvents(t *testing.T) {
	cal, err := ical.Parse(testICS)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	events := cal.SearchEvents("meeting")
	if len(events) != 1 {
		t.Errorf("Search 'meeting': got %d events, want 1", len(events))
	}

	events = cal.SearchEvents("Client")
	if len(events) != 1 {
		t.Errorf("Search 'Client': got %d events, want 1", len(events))
	}
}

func TestStats(t *testing.T) {
	cal, err := ical.Parse(testICS)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	stats := cal.Stats()
	if stats.TotalEvents != 3 {
		t.Errorf("Stats.TotalEvents = %d, want 3", stats.TotalEvents)
	}
	if stats.TotalTodos != 1 {
		t.Errorf("Stats.TotalTodos = %d, want 1", stats.TotalTodos)
	}
	if stats.UniqueAttendees != 2 {
		t.Errorf("Stats.UniqueAttendees = %d, want 2", stats.UniqueAttendees)
	}
}

func TestEmptyInput(t *testing.T) {
	_, err := ical.Parse("")
	if err == nil {
		t.Error("Expected error for empty input")
	}
}

func TestMinimalICS(t *testing.T) {
	minimal := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:1@test.com
SUMMARY:Test
DTSTART:20260101T120000Z
END:VEVENT
END:VCALENDAR`

	cal, err := ical.Parse(minimal)
	if err != nil {
		t.Fatalf("Parse minimal ICS failed: %v", err)
	}
	if len(cal.Events) != 1 {
		t.Errorf("Events = %d, want 1", len(cal.Events))
	}
}

func TestLineUnfolding(t *testing.T) {
	longDesc := strings.Repeat("A", 200)
	ics := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:1@test.com
SUMMARY:Test
DESCRIPTION:` + longDesc + `
DTSTART:20260101T120000Z
END:VEVENT
END:VCALENDAR`

	cal, err := ical.Parse(ics)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if cal.Events[0].Description != longDesc {
		t.Errorf("Description length = %d, want %d", len(cal.Events[0].Description), len(longDesc))
	}
}
