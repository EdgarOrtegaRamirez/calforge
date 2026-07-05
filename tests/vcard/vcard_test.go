package vcard_test

import (
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/calforge/internal/vcard"
)

const testVCF = `BEGIN:VCARD
VERSION:4.0
FN:John Doe
N:Doe;John;;;
EMAIL;TYPE=WORK:john@example.com
EMAIL;TYPE=HOME:john.doe@personal.com
TEL;TYPE=CELL:+1-555-0101
TEL;TYPE=WORK:+1-555-0102
ADR;TYPE=WORK:;;123 Main St;Springfield;IL;62701;USA
ORG:Acme Corp
TITLE:Senior Engineer
BDAY:19900515
CATEGORIES:Friend,Colleague
NOTE:Met at conference
URL:https://johndoe.com
END:VCARD
BEGIN:VCARD
VERSION:4.0
FN:Jane Smith
N:Smith;Jane;;;
EMAIL;TYPE=WORK:jane@example.com
TEL;TYPE=HOME:+1-555-0201
ADR;TYPE=HOME:;;456 Oak Ave;Portland;OR;97201;USA
ORG:Tech Startup
TITLE:CEO
GENDER:F
END:VCARD`

func TestParse(t *testing.T) {
	contacts, err := vcard.Parse(testVCF)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(contacts) != 2 {
		t.Fatalf("Contacts = %d, want 2", len(contacts))
	}

	// Check first contact
	c := contacts[0]
	if c.FullName != "John Doe" {
		t.Errorf("FullName = %q, want %q", c.FullName, "John Doe")
	}
	if c.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", c.FirstName, "John")
	}
	if c.LastName != "Doe" {
		t.Errorf("LastName = %q, want %q", c.LastName, "Doe")
	}
	if c.Organization != "Acme Corp" {
		t.Errorf("Organization = %q, want %q", c.Organization, "Acme Corp")
	}
	if c.Title != "Senior Engineer" {
		t.Errorf("Title = %q, want %q", c.Title, "Senior Engineer")
	}
	if c.Birthday != "19900515" {
		t.Errorf("Birthday = %q, want %q", c.Birthday, "19900515")
	}
	if c.Notes != "Met at conference" {
		t.Errorf("Notes = %q, want %q", c.Notes, "Met at conference")
	}

	// Check emails
	if len(c.Emails) != 2 {
		t.Errorf("Emails = %d, want 2", len(c.Emails))
	} else {
		if c.Emails[0].Address != "john@example.com" {
			t.Errorf("Email[0] = %q, want %q", c.Emails[0].Address, "john@example.com")
		}
		if c.Emails[0].Type != "WORK" {
			t.Errorf("Email[0].Type = %q, want %q", c.Emails[0].Type, "WORK")
		}
	}

	// Check phones
	if len(c.Phones) != 2 {
		t.Errorf("Phones = %d, want 2", len(c.Phones))
	}

	// Check address
	if len(c.Addresses) != 1 {
		t.Errorf("Addresses = %d, want 1", len(c.Addresses))
	} else {
		if c.Addresses[0].Street != "123 Main St" {
			t.Errorf("Address.Street = %q, want %q", c.Addresses[0].Street, "123 Main St")
		}
		if c.Addresses[0].City != "Springfield" {
			t.Errorf("Address.City = %q, want %q", c.Addresses[0].City, "Springfield")
		}
		if c.Addresses[0].Country != "USA" {
			t.Errorf("Address.Country = %q, want %q", c.Addresses[0].Country, "USA")
		}
	}

	// Check categories
	if len(c.Categories) != 2 {
		t.Errorf("Categories = %d, want 2", len(c.Categories))
	}

	// Check second contact
	c2 := contacts[1]
	if c2.FullName != "Jane Smith" {
		t.Errorf("Contact[1].FullName = %q, want %q", c2.FullName, "Jane Smith")
	}
	if c2.Gender != "F" {
		t.Errorf("Contact[1].Gender = %q, want %q", c2.Gender, "F")
	}
}

func TestSerialize(t *testing.T) {
	contacts, err := vcard.Parse(testVCF)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output := vcard.SerializeMultiple(contacts)

	// Re-parse
	contacts2, err := vcard.Parse(output)
	if err != nil {
		t.Fatalf("Re-parse failed: %v", err)
	}

	if len(contacts2) != len(contacts) {
		t.Errorf("Contacts count = %d, want %d", len(contacts2), len(contacts))
	}
	if contacts2[0].FullName != contacts[0].FullName {
		t.Errorf("FullName = %q, want %q", contacts2[0].FullName, contacts[0].FullName)
	}
}

func TestValidate(t *testing.T) {
	contacts, err := vcard.Parse(testVCF)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	issues := contacts[0].Validate()
	if len(issues) != 0 {
		t.Errorf("Expected no issues, got %d: %v", len(issues), issues)
	}

	// Test invalid contact
	bad := &vcard.Contact{
		FirstName: "Test",
	}
	issues = bad.Validate()
	if len(issues) == 0 {
		t.Error("Expected issues for contact without email")
	}
}

func TestDisplayName(t *testing.T) {
	c := &vcard.Contact{
		FullName: "John Doe",
	}
	if c.DisplayName() != "John Doe" {
		t.Errorf("DisplayName() = %q, want %q", c.DisplayName(), "John Doe")
	}

	c2 := &vcard.Contact{
		FirstName: "Jane",
		LastName:  "Smith",
	}
	if c2.DisplayName() != "Jane Smith" {
		t.Errorf("DisplayName() = %q, want %q", c2.DisplayName(), "Jane Smith")
	}

	c3 := &vcard.Contact{
		Prefix:    "Dr.",
		FirstName: "John",
		LastName:  "Doe",
		Suffix:    "Jr.",
	}
	if c3.DisplayName() != "Dr. John Doe Jr." {
		t.Errorf("DisplayName() = %q, want %q", c3.DisplayName(), "Dr. John Doe Jr.")
	}
}

func TestPrimaryEmail(t *testing.T) {
	c := &vcard.Contact{
		Emails: []vcard.Email{
			{Address: "work@example.com", Type: "WORK"},
			{Address: "home@example.com", Type: "HOME", Primary: true},
		},
	}
	if c.PrimaryEmail() != "home@example.com" {
		t.Errorf("PrimaryEmail() = %q, want %q", c.PrimaryEmail(), "home@example.com")
	}

	c2 := &vcard.Contact{
		Emails: []vcard.Email{
			{Address: "only@example.com"},
		},
	}
	if c2.PrimaryEmail() != "only@example.com" {
		t.Errorf("PrimaryEmail() = %q, want %q", c2.PrimaryEmail(), "only@example.com")
	}
}

func TestEmptyInput(t *testing.T) {
	_, err := vcard.Parse("")
	if err == nil {
		t.Error("Expected error for empty input")
	}
}

func TestSingleContact(t *testing.T) {
	vcf := `BEGIN:VCARD
VERSION:4.0
FN:Test User
N:User;Test;;;
EMAIL:test@example.com
END:VCARD`

	contacts, err := vcard.Parse(vcf)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(contacts) != 1 {
		t.Errorf("Contacts = %d, want 1", len(contacts))
	}
}

func TestFoldedLines(t *testing.T) {
	longNote := strings.Repeat("This is a very long note. ", 20)
	vcf := "BEGIN:VCARD\r\nVERSION:4.0\r\nFN:Test User\r\nN:User;Test;;;\r\nEMAIL:test@example.com\r\nNOTE:" + longNote + "\r\nEND:VCARD"

	contacts, err := vcard.Parse(vcf)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(contacts[0].Notes) == 0 {
		t.Error("Notes is empty")
	}
	if !strings.Contains(contacts[0].Notes, "This is a very long note.") {
		t.Errorf("Notes doesn't contain expected content: %q", contacts[0].Notes)
	}
}
