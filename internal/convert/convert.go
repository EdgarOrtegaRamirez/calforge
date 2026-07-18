// Package convert provides format conversion between iCal, vCard, JSON, and CSV.
package convert

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/EdgarOrtegaRamirez/calforge/internal/ical"
	"github.com/EdgarOrtegaRamirez/calforge/internal/vcard"
)

// EventJSON is the JSON representation of an event.
type EventJSON struct {
	UID         string   `json:"uid"`
	Summary     string   `json:"summary"`
	Description string   `json:"description,omitempty"`
	Location    string   `json:"location,omitempty"`
	URL         string   `json:"url,omitempty"`
	Status      string   `json:"status,omitempty"`
	StartTime   string   `json:"start_time,omitempty"`
	EndTime     string   `json:"end_time,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	Attendees   []string `json:"attendees,omitempty"`
}

// ContactJSON is the JSON representation of a contact.
type ContactJSON struct {
	FullName     string           `json:"full_name"`
	FirstName    string           `json:"first_name,omitempty"`
	LastName     string           `json:"last_name,omitempty"`
	Organization string           `json:"organization,omitempty"`
	Title        string           `json:"title,omitempty"`
	Emails       []ContactEmail   `json:"emails,omitempty"`
	Phones       []ContactPhone   `json:"phones,omitempty"`
	Addresses    []ContactAddress `json:"addresses,omitempty"`
	Notes        string           `json:"notes,omitempty"`
	Categories   []string         `json:"categories,omitempty"`
	Birthday     string           `json:"birthday,omitempty"`
}

// ContactEmail represents an email in JSON.
type ContactEmail struct {
	Address string `json:"address"`
	Type    string `json:"type,omitempty"`
}

// ContactPhone represents a phone in JSON.
type ContactPhone struct {
	Number string `json:"number"`
	Type   string `json:"type,omitempty"`
}

// ContactAddress represents an address in JSON.
type ContactAddress struct {
	Street     string `json:"street,omitempty"`
	City       string `json:"city,omitempty"`
	Region     string `json:"region,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
	Type       string `json:"type,omitempty"`
}

// EventsToJSON converts iCalendar events to JSON.
func EventsToJSON(cal *ical.Calendar) (string, error) {
	events := make([]EventJSON, 0)
	for _, e := range cal.Events {
		ej := EventJSON{
			UID:         e.UID,
			Summary:     e.Summary,
			Description: e.Description,
			Location:    e.Location,
			URL:         e.URL,
			Status:      e.Status,
			Categories:  e.Categories,
		}
		if e.StartTime != nil {
			ej.StartTime = e.StartTime.Format(time.RFC3339)
		}
		if e.EndTime != nil {
			ej.EndTime = e.EndTime.Format(time.RFC3339)
		}
		for _, a := range e.Attendees {
			ej.Attendees = append(ej.Attendees, a.Email)
		}
		events = append(events, ej)
	}

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal JSON: %w", err)
	}
	return string(data), nil
}

// EventsToCSV converts iCalendar events to CSV.
func EventsToCSV(cal *ical.Calendar) (string, error) {
	var b strings.Builder
	w := csv.NewWriter(&b)

	// Header
	header := []string{"uid", "summary", "description", "location", "status", "start_time", "end_time", "categories"}
	if err := w.Write(header); err != nil {
		return "", fmt.Errorf("write CSV header: %w", err)
	}

	for _, e := range cal.Events {
		startStr := ""
		if e.StartTime != nil {
			startStr = e.StartTime.Format(time.RFC3339)
		}
		endStr := ""
		if e.EndTime != nil {
			endStr = e.EndTime.Format(time.RFC3339)
		}

		record := []string{
			e.UID,
			e.Summary,
			e.Description,
			e.Location,
			e.Status,
			startStr,
			endStr,
			strings.Join(e.Categories, ";"),
		}
		if err := w.Write(record); err != nil {
			return "", fmt.Errorf("write CSV record: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("flush CSV: %w", err)
	}

	return b.String(), nil
}

// ContactsToJSON converts vCard contacts to JSON.
func ContactsToJSON(contacts []*vcard.Contact) (string, error) {
	result := make([]ContactJSON, 0)
	for _, c := range contacts {
		cj := ContactJSON{
			FullName:     c.FullName,
			FirstName:    c.FirstName,
			LastName:     c.LastName,
			Organization: c.Organization,
			Title:        c.Title,
			Notes:        c.Notes,
			Categories:   c.Categories,
			Birthday:     c.Birthday,
		}
		for _, e := range c.Emails {
			cj.Emails = append(cj.Emails, ContactEmail{Address: e.Address, Type: e.Type})
		}
		for _, p := range c.Phones {
			cj.Phones = append(cj.Phones, ContactPhone{Number: p.Number, Type: p.Type})
		}
		for _, a := range c.Addresses {
			cj.Addresses = append(cj.Addresses, ContactAddress{
				Street:     a.Street,
				City:       a.City,
				Region:     a.Region,
				PostalCode: a.PostalCode,
				Country:    a.Country,
				Type:       a.Type,
			})
		}
		result = append(result, cj)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal JSON: %w", err)
	}
	return string(data), nil
}

// ContactsToCSV converts vCard contacts to CSV.
func ContactsToCSV(contacts []*vcard.Contact) (string, error) {
	var b strings.Builder
	w := csv.NewWriter(&b)

	header := []string{"full_name", "first_name", "last_name", "organization", "title", "email", "phone", "city", "country", "notes"}
	if err := w.Write(header); err != nil {
		return "", fmt.Errorf("write CSV header: %w", err)
	}

	for _, c := range contacts {
		email := ""
		if len(c.Emails) > 0 {
			email = c.Emails[0].Address
		}
		phone := ""
		if len(c.Phones) > 0 {
			phone = c.Phones[0].Number
		}
		city := ""
		country := ""
		if len(c.Addresses) > 0 {
			city = c.Addresses[0].City
			country = c.Addresses[0].Country
		}

		record := []string{
			c.FullName,
			c.FirstName,
			c.LastName,
			c.Organization,
			c.Title,
			email,
			phone,
			city,
			country,
			c.Notes,
		}
		if err := w.Write(record); err != nil {
			return "", fmt.Errorf("write CSV record: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("flush CSV: %w", err)
	}

	return b.String(), nil
}
