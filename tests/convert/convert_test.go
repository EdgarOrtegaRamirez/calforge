package convert_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/calforge/internal/convert"
	"github.com/EdgarOrtegaRamirez/calforge/internal/ical"
	"github.com/EdgarOrtegaRamirez/calforge/internal/vcard"
)

const testICS = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:test-1@test.com
SUMMARY:Team Meeting
DESCRIPTION:Weekly sync
DTSTART:20260710T100000Z
DTEND:20260710T110000Z
LOCATION:Conference Room
STATUS:CONFIRMED
CATEGORIES:Work
ATTENDEE;CN=Alice:mailto:alice@example.com
END:VEVENT
BEGIN:VEVENT
UID:test-2@test.com
SUMMARY:Lunch
DTSTART:20260710T120000Z
DTEND:20260710T130000Z
STATUS:TENTATIVE
END:VEVENT
END:VCALENDAR`

const testVCF = `BEGIN:VCARD
VERSION:4.0
FN:John Doe
N:Doe;John;;;
EMAIL;TYPE=WORK:john@example.com
TEL;TYPE=CELL:+1-555-0101
ADR;TYPE=HOME:;;123 Main St;Springfield;IL;62701;USA
ORG:Acme Corp
TITLE:Engineer
NOTE:Test contact
END:VCARD`

func TestEventsToJSON(t *testing.T) {
	cal, err := ical.Parse(testICS)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsonStr, err := convert.EventsToJSON(cal)
	if err != nil {
		t.Fatalf("EventsToJSON failed: %v", err)
	}

	var events []convert.EventJSON
	if err := json.Unmarshal([]byte(jsonStr), &events); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("Events count = %d, want 2", len(events))
	}
	if events[0].Summary != "Team Meeting" {
		t.Errorf("Event[0].Summary = %q, want %q", events[0].Summary, "Team Meeting")
	}
	if events[0].Location != "Conference Room" {
		t.Errorf("Event[0].Location = %q, want %q", events[0].Location, "Conference Room")
	}
	if len(events[0].Attendees) != 1 {
		t.Errorf("Event[0].Attendees = %d, want 1", len(events[0].Attendees))
	}
}

func TestEventsToCSV(t *testing.T) {
	cal, err := ical.Parse(testICS)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	csvStr, err := convert.EventsToCSV(cal)
	if err != nil {
		t.Fatalf("EventsToCSV failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(csvStr), "\n")
	if len(lines) != 3 { // header + 2 events
		t.Errorf("CSV lines = %d, want 3", len(lines))
	}
	if !strings.Contains(lines[0], "uid") {
		t.Error("CSV header missing 'uid' column")
	}
	if !strings.Contains(lines[1], "Team Meeting") {
		t.Error("CSV row 1 missing 'Team Meeting'")
	}
}

func TestContactsToJSON(t *testing.T) {
	contacts, err := vcard.Parse(testVCF)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsonStr, err := convert.ContactsToJSON(contacts)
	if err != nil {
		t.Fatalf("ContactsToJSON failed: %v", err)
	}

	var result []convert.ContactJSON
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Contacts count = %d, want 1", len(result))
	}
	if result[0].FullName != "John Doe" {
		t.Errorf("FullName = %q, want %q", result[0].FullName, "John Doe")
	}
	if len(result[0].Emails) != 1 {
		t.Errorf("Emails = %d, want 1", len(result[0].Emails))
	}
}

func TestContactsToCSV(t *testing.T) {
	contacts, err := vcard.Parse(testVCF)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	csvStr, err := convert.ContactsToCSV(contacts)
	if err != nil {
		t.Fatalf("ContactsToCSV failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(csvStr), "\n")
	if len(lines) != 2 { // header + 1 contact
		t.Errorf("CSV lines = %d, want 2", len(lines))
	}
	if !strings.Contains(lines[0], "full_name") {
		t.Error("CSV header missing 'full_name' column")
	}
	if !strings.Contains(lines[1], "John Doe") {
		t.Error("CSV row missing 'John Doe'")
	}
}
