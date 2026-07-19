package vcard

import (
	"fmt"
	"strings"
)

// Serialize converts a Contact to vCard format string.
func (c *Contact) Serialize() string {
	var b strings.Builder
	b.WriteString("BEGIN:VCARD\r\n")
	b.WriteString("VERSION:4.0\r\n")

	// Full name
	if c.FullName != "" {
		fmt.Fprintf(&b, "FN:%s\r\n", c.FullName)
	}

	// Structured name
	fmt.Fprintf(&b, "N:%s;%s;%s;%s;%s\r\n",
		c.LastName, c.FirstName, c.MiddleName, c.Prefix, c.Suffix)

	// Nickname
	if len(c.Nickname) > 0 {
		fmt.Fprintf(&b, "NICKNAME:%s\r\n", strings.Join(c.Nickname, ","))
	}

	// Birthday
	if c.Birthday != "" {
		b.WriteString(fmt.Sprintf("BDAY:%s\r\n", c.Birthday))
	}

	// Anniversary
	if c.Anniversary != "" {
		b.WriteString(fmt.Sprintf("ANNIVERSARY:%s\r\n", c.Anniversary))
	}

	// Gender
	if c.Gender != "" {
		b.WriteString(fmt.Sprintf("GENDER:%s\r\n", c.Gender))
	}

	// Organization
	if c.Organization != "" {
		b.WriteString(fmt.Sprintf("ORG:%s\r\n", c.Organization))
	}

	// Department
	if c.Department != "" {
		b.WriteString(fmt.Sprintf("DEPARTMENT:%s\r\n", c.Department))
	}

	// Title
	if c.Title != "" {
		b.WriteString(fmt.Sprintf("TITLE:%s\r\n", c.Title))
	}

	// Role
	if c.Role != "" {
		b.WriteString(fmt.Sprintf("ROLE:%s\r\n", c.Role))
	}

	// Emails
	for _, email := range c.Emails {
		b.WriteString("EMAIL")
		if email.Type != "" {
			b.WriteString(fmt.Sprintf(";TYPE=%s", email.Type))
		}
		if email.Primary {
			b.WriteString(";PREF=1")
		}
		b.WriteString(fmt.Sprintf(":%s\r\n", email.Address))
	}

	// Phones
	for _, phone := range c.Phones {
		b.WriteString("TEL")
		if phone.Type != "" {
			b.WriteString(fmt.Sprintf(";TYPE=%s", phone.Type))
		}
		if phone.Primary {
			b.WriteString(";PREF=1")
		}
		b.WriteString(fmt.Sprintf(":%s\r\n", phone.Number))
	}

	// Addresses
	for _, addr := range c.Addresses {
		b.WriteString("ADR")
		if addr.Type != "" {
			b.WriteString(fmt.Sprintf(";TYPE=%s", addr.Type))
		}
		if addr.Primary {
			b.WriteString(";PREF=1")
		}
		b.WriteString(fmt.Sprintf(":%s;%s;%s;%s;%s;%s;%s\r\n",
			addr.POBox, addr.Extended, addr.Street,
			addr.City, addr.Region, addr.PostalCode, addr.Country))
	}

	// URLs
	for _, url := range c.URLs {
		b.WriteString(fmt.Sprintf("URL:%s\r\n", url))
	}

	// Photos
	for _, photo := range c.Photos {
		b.WriteString("PHOTO")
		if photo.Type != "" {
			b.WriteString(fmt.Sprintf(";TYPE=%s", photo.Type))
		}
		if photo.Encoding != "" {
			b.WriteString(fmt.Sprintf(";ENCODING=%s", photo.Encoding))
		}
		b.WriteString(fmt.Sprintf(":%s\r\n", photo.URI))
	}

	// Notes
	if c.Notes != "" {
		b.WriteString(fmt.Sprintf("NOTE:%s\r\n", c.Notes))
	}

	// Categories
	if len(c.Categories) > 0 {
		b.WriteString(fmt.Sprintf("CATEGORIES:%s\r\n", strings.Join(c.Categories, ",")))
	}

	// Timezone
	if c.TimeZone != "" {
		b.WriteString(fmt.Sprintf("TZ:%s\r\n", c.TimeZone))
	}

	// Geo
	if c.Geo != nil {
		b.WriteString(fmt.Sprintf("GEO:%.6f;%.6f\r\n", c.Geo.Latitude, c.Geo.Longitude))
	}

	// IMs
	for _, im := range c.IMs {
		b.WriteString("IMPP")
		if im.Type != "" {
			b.WriteString(fmt.Sprintf(";TYPE=%s", im.Type))
		}
		b.WriteString(fmt.Sprintf(":%s\r\n", im.Handle))
	}

	// Relations
	for _, rel := range c.Relations {
		b.WriteString("RELATED")
		if rel.Type != "" {
			b.WriteString(fmt.Sprintf(";TYPE=%s", rel.Type))
		}
		b.WriteString(fmt.Sprintf(":%s\r\n", rel.Name))
	}

	// Custom properties
	for _, p := range c.Properties {
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

	b.WriteString("END:VCARD\r\n")
	return b.String()
}

// SerializeMultiple serializes multiple contacts into a single string.
func SerializeMultiple(contacts []*Contact) string {
	var b strings.Builder
	for _, c := range contacts {
		b.WriteString(c.Serialize())
	}
	return b.String()
}

// Validate checks a Contact for common issues.
func (c *Contact) Validate() []string {
	var issues []string

	if c.FullName == "" && c.FirstName == "" && c.LastName == "" {
		issues = append(issues, "missing name (FN or N)")
	}
	if len(c.Emails) == 0 {
		issues = append(issues, "no email addresses")
	}

	return issues
}

// DisplayName returns a formatted display name.
func (c *Contact) DisplayName() string {
	if c.FullName != "" {
		return c.FullName
	}
	parts := make([]string, 0)
	if c.Prefix != "" {
		parts = append(parts, c.Prefix)
	}
	if c.FirstName != "" {
		parts = append(parts, c.FirstName)
	}
	if c.MiddleName != "" {
		parts = append(parts, c.MiddleName)
	}
	if c.LastName != "" {
		parts = append(parts, c.LastName)
	}
	if c.Suffix != "" {
		parts = append(parts, c.Suffix)
	}
	return strings.Join(parts, " ")
}

// PrimaryEmail returns the primary email or the first email.
func (c *Contact) PrimaryEmail() string {
	for _, e := range c.Emails {
		if e.Primary {
			return e.Address
		}
	}
	if len(c.Emails) > 0 {
		return c.Emails[0].Address
	}
	return ""
}

// PrimaryPhone returns the primary phone or the first phone.
func (c *Contact) PrimaryPhone() string {
	for _, p := range c.Phones {
		if p.Primary {
			return p.Number
		}
	}
	if len(c.Phones) > 0 {
		return c.Phones[0].Number
	}
	return ""
}
