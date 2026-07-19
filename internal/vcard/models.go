// Package vcard implements vCard (RFC 6350) parsing and generation.
package vcard

import (
	"fmt"
	"strings"
)

// Contact represents a vCard contact.
type Contact struct {
	Version      string
	FirstName    string
	LastName     string
	MiddleName   string
	Prefix       string
	Suffix       string
	FullName     string
	Nickname     []string
	Birthday     string
	Anniversary  string
	Gender       string
	Organization string
	Department   string
	Title        string
	Role         string
	Emails       []Email
	Phones       []Phone
	Addresses    []Address
	URLs         []string
	Photos       []Photo
	Notes        string
	Categories   []string
	TimeZone     string
	Geo          *Geo
	IMs          []IM
	Relations    []Relation
	Properties   []Property
}

// Email represents a vCard email address.
type Email struct {
	Address string
	Type    string // HOME, WORK, etc.
	Primary bool
}

// Phone represents a vCard phone number.
type Phone struct {
	Number  string
	Type    string // HOME, WORK, CELL, FAX, etc.
	Primary bool
}

// Address represents a vCard address.
type Address struct {
	Type       string // HOME, WORK
	Street     string
	City       string
	Region     string
	PostalCode string
	Country    string
	POBox      string
	Extended   string
	Primary    bool
}

// Photo represents a vCard photo.
type Photo struct {
	URI      string
	Type     string // JPEG, PNG
	Encoding string // URL, B64
}

// Geo represents geographic coordinates.
type Geo struct {
	Latitude  float64
	Longitude float64
}

// IM represents an instant messaging handle.
type IM struct {
	Handle string
	Type   string // SIP, XMPP, etc.
}

// Relation represents a relationship.
type Relation struct {
	Name string
	Type string // spouse, child, parent, etc.
}

// Property represents a custom vCard property.
type Property struct {
	Name       string
	Value      string
	Parameters map[string][]string
}

// Parse parses a vCard string into one or more Contact objects.
func Parse(input string) ([]*Contact, error) {
	var contacts []*Contact
	var current *Contact

	// Normalize line endings
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.ReplaceAll(input, "\r", "\n")

	lines := strings.Split(input, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle folded lines (RFC 6350 §3.1)
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if current != nil && len(current.Properties) > 0 {
				last := &current.Properties[len(current.Properties)-1]
				last.Value += line[1:]
			}
			continue
		}

		if strings.ToUpper(line) == "BEGIN:VCARD" {
			current = &Contact{
				Emails:     make([]Email, 0),
				Phones:     make([]Phone, 0),
				Addresses:  make([]Address, 0),
				Properties: make([]Property, 0),
			}
			continue
		}

		if strings.ToUpper(line) == "END:VCARD" {
			if current != nil {
				contacts = append(contacts, current)
			}
			current = nil
			continue
		}

		if current != nil {
			prop := parseProperty(line)
			applyProperty(current, prop)
		}
	}

	if len(contacts) == 0 {
		return nil, fmt.Errorf("no vCard found in input")
	}

	return contacts, nil
}

// parsedProp represents a parsed vCard property.
type parsedProp struct {
	name   string
	params map[string][]string
	value  string
}

// parseProperty parses a single vCard property line.
func parseProperty(line string) parsedProp {
	p := parsedProp{
		params: make(map[string][]string),
	}

	// Find the property name (before first ';')
	semiIdx := strings.Index(line, ";")
	colonIdx := strings.Index(line, ":")

	if colonIdx < 0 {
		p.name = strings.ToUpper(line)
		return p
	}

	if semiIdx >= 0 && semiIdx < colonIdx {
		p.name = strings.ToUpper(line[:semiIdx])
		// Parse parameters
		paramStr := line[semiIdx+1 : colonIdx]
		segments := strings.Split(paramStr, ";")
		for _, seg := range segments {
			eqIdx := strings.Index(seg, "=")
			if eqIdx < 0 {
				continue
			}
			key := strings.TrimSpace(seg[:eqIdx])
			val := strings.TrimSpace(seg[eqIdx+1:])
			p.params[key] = append(p.params[key], val)
		}
	} else {
		p.name = strings.ToUpper(line[:colonIdx])
	}

	p.value = line[colonIdx+1:]
	return p
}

// applyProperty applies a parsed property to a Contact.
func applyProperty(c *Contact, p parsedProp) {
	switch p.name {
	case "VERSION":
		c.Version = p.value
	case "FN":
		c.FullName = p.value
	case "N":
		parts := strings.Split(p.value, ";")
		if len(parts) >= 5 {
			c.LastName = parts[0]
			c.FirstName = parts[1]
			c.MiddleName = parts[2]
			c.Prefix = parts[3]
			c.Suffix = parts[4]
		}
	case "NICKNAME":
		c.Nickname = strings.Split(p.value, ",")
	case "BDAY":
		c.Birthday = p.value
	case "ANNIVERSARY":
		c.Anniversary = p.value
	case "GENDER":
		c.Gender = p.value
	case "ORG":
		c.Organization = p.value
	case "DEPARTMENT":
		c.Department = p.value
	case "TITLE":
		c.Title = p.value
	case "ROLE":
		c.Role = p.value
	case "EMAIL":
		email := Email{Address: p.value}
		if types, ok := p.params["TYPE"]; ok && len(types) > 0 {
			email.Type = strings.ToUpper(types[0])
		}
		if prefs, ok := p.params["PREF"]; ok && len(prefs) > 0 {
			email.Primary = prefs[0] == "1"
		}
		c.Emails = append(c.Emails, email)
	case "TEL":
		phone := Phone{Number: p.value}
		if types, ok := p.params["TYPE"]; ok && len(types) > 0 {
			phone.Type = strings.ToUpper(types[0])
		}
		if prefs, ok := p.params["PREF"]; ok && len(prefs) > 0 {
			phone.Primary = prefs[0] == "1"
		}
		c.Phones = append(c.Phones, phone)
	case "ADR":
		addr := parseAddress(p.value)
		if types, ok := p.params["TYPE"]; ok && len(types) > 0 {
			addr.Type = strings.ToUpper(types[0])
		}
		if prefs, ok := p.params["PREF"]; ok && len(prefs) > 0 {
			addr.Primary = prefs[0] == "1"
		}
		c.Addresses = append(c.Addresses, addr)
	case "URL":
		c.URLs = append(c.URLs, p.value)
	case "PHOTO":
		photo := Photo{URI: p.value}
		if types, ok := p.params["TYPE"]; ok && len(types) > 0 {
			photo.Type = strings.ToUpper(types[0])
		}
		if enc, ok := p.params["ENCODING"]; ok && len(enc) > 0 {
			photo.Encoding = strings.ToUpper(enc[0])
		}
		c.Photos = append(c.Photos, photo)
	case "NOTE":
		c.Notes = p.value
	case "CATEGORIES":
		c.Categories = strings.Split(p.value, ",")
		for i := range c.Categories {
			c.Categories[i] = strings.TrimSpace(c.Categories[i])
		}
	case "TZ":
		c.TimeZone = p.value
	case "GEO":
		parts := strings.Split(p.value, ";")
		if len(parts) >= 2 {
			geo := &Geo{}
			n, err := fmt.Sscanf(parts[0], "%f", &geo.Latitude)
			if err != nil || n == 0 {
				geo.Latitude = 0
			}
			n, err = fmt.Sscanf(parts[1], "%f", &geo.Longitude)
			if err != nil || n == 0 {
				geo.Longitude = 0
			}
			c.Geo = geo
		}
	case "IMPP":
		im := IM{Handle: p.value}
		if types, ok := p.params["TYPE"]; ok && len(types) > 0 {
			im.Type = strings.ToUpper(types[0])
		}
		c.IMs = append(c.IMs, im)
	case "RELATED":
		rel := Relation{Name: p.value}
		if types, ok := p.params["TYPE"]; ok && len(types) > 0 {
			rel.Type = strings.ToUpper(types[0])
		}
		c.Relations = append(c.Relations, rel)
	default:
		c.Properties = append(c.Properties, Property{
			Name:       p.name,
			Value:      p.value,
			Parameters: p.params,
		})
	}
}

// parseAddress parses an ADR value string.
func parseAddress(value string) Address {
	parts := strings.Split(value, ";")
	addr := Address{}
	if len(parts) > 0 {
		addr.POBox = parts[0]
	}
	if len(parts) > 1 {
		addr.Extended = parts[1]
	}
	if len(parts) > 2 {
		addr.Street = parts[2]
	}
	if len(parts) > 3 {
		addr.City = parts[3]
	}
	if len(parts) > 4 {
		addr.Region = parts[4]
	}
	if len(parts) > 5 {
		addr.PostalCode = parts[5]
	}
	if len(parts) > 6 {
		addr.Country = parts[6]
	}
	return addr
}
