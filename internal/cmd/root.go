// Package cmd implements the CLI commands for CalForge.
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EdgarOrtegaRamirez/calforge/internal/convert"
	"github.com/EdgarOrtegaRamirez/calforge/internal/ical"
	"github.com/EdgarOrtegaRamirez/calforge/internal/vcard"
)

const version = "1.0.0"

// Run executes the CLI with the given arguments.
func Run(args []string) error {
	if len(args) < 2 {
		printUsage()
		return nil
	}

	command := args[1]
	switch command {
	case "version", "--version", "-v":
		fmt.Printf("calforge %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	case "parse":
		return cmdParse(args[2:])
	case "validate":
		return cmdValidate(args[2:])
	case "stats":
		return cmdStats(args[2:])
	case "filter":
		return cmdFilter(args[2:])
	case "search":
		return cmdSearch(args[2:])
	case "convert":
		return cmdConvert(args[2:])
	case "merge":
		return cmdMerge(args[2:])
	case "create":
		return cmdCreate(args[2:])
	case "vcard":
		return cmdVCard(args[2:])
	default:
		return fmt.Errorf("unknown command: %s\nRun 'calforge help' for usage", command)
	}
	return nil
}

func printUsage() {
	fmt.Print(`CalForge - iCalendar & vCard Processing Toolkit

Usage:
  calforge <command> [options]

Commands:
  parse       Parse and display iCalendar or vCard files
  validate    Validate iCalendar or vCard files
  stats       Show calendar statistics
  filter      Filter events by criteria
  search      Search events by text
  convert     Convert between formats (ics↔json, ics↔csv, vcf↔json, vcf↔csv)
  merge       Merge multiple calendar files
  create      Create sample iCalendar or vCard files
  vcard       vCard-specific operations (parse, search, format)
  version     Show version
  help        Show this help message

Examples:
  calforge parse events.ics
  calforge validate calendar.ics
  calforge stats calendar.ics
  calforge filter --start 2026-01-01 --end 2026-12-31 calendar.ics
  calforge search --query "meeting" calendar.ics
  calforge convert --from ics --to json calendar.ics
  calforge convert --from vcf --to csv contacts.vcf
  calforge merge cal1.ics cal2.ics -o merged.ics
  calforge create --type ics --events 5
  calforge create --type vcf --contacts 3
  calforge vcard parse contacts.vcf
  calforge vcard search --query "john" contacts.vcf
  calforge vcard format contacts.vcf
`)
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}
	return string(data), nil
}

func detectType(path string) string {
	ext := strings.ToLower(path)
	if strings.HasSuffix(ext, ".ics") || strings.HasSuffix(ext, ".ical") || strings.HasSuffix(ext, ".ifb") {
		return "ical"
	}
	if strings.HasSuffix(ext, ".vcf") || strings.HasSuffix(ext, ".vcard") {
		return "vcard"
	}
	return "unknown"
}

func cmdParse(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: calforge parse <file.ics|file.vcf> [--json]")
	}

	path := args[0]
	outputJSON := false
	for i := 1; i < len(args); i++ {
		if args[i] == "--json" {
			outputJSON = true
		}
	}

	data, err := readFile(path)
	if err != nil {
		return err
	}

	fileType := detectType(path)
	if fileType == "unknown" {
		// Try to detect from content
		if strings.Contains(data, "BEGIN:VCALENDAR") {
			fileType = "ical"
		} else if strings.Contains(data, "BEGIN:VCARD") {
			fileType = "vcard"
		}
	}

	switch fileType {
	case "ical":
		cal, err := ical.Parse(data)
		if err != nil {
			return fmt.Errorf("parse iCalendar: %w", err)
		}
		if outputJSON {
			jsonStr, err := convert.EventsToJSON(cal)
			if err != nil {
				return err
			}
			fmt.Println(jsonStr)
		} else {
			fmt.Printf("Calendar: %s\n", cal.CalName)
			fmt.Printf("Events: %d\n", len(cal.Events))
			for i, e := range cal.Events {
				fmt.Printf("\n[%d] %s\n", i+1, e.Summary)
				if e.StartTime != nil {
					fmt.Printf("    Start: %s\n", e.StartTime.Format("2006-01-02 15:04"))
				}
				if e.EndTime != nil {
					fmt.Printf("    End:   %s\n", e.EndTime.Format("2006-01-02 15:04"))
				}
				if e.Location != "" {
					fmt.Printf("    Location: %s\n", e.Location)
				}
				if e.Description != "" {
					desc := e.Description
					if len(desc) > 80 {
						desc = desc[:77] + "..."
					}
					fmt.Printf("    Description: %s\n", desc)
				}
				if e.Status != "" {
					fmt.Printf("    Status: %s\n", e.Status)
				}
				if len(e.Attendees) > 0 {
					fmt.Printf("    Attendees: %d\n", len(e.Attendees))
				}
			}
		}
	case "vcard":
		contacts, err := vcard.Parse(data)
		if err != nil {
			return fmt.Errorf("parse vCard: %w", err)
		}
		if outputJSON {
			jsonStr, err := convert.ContactsToJSON(contacts)
			if err != nil {
				return err
			}
			fmt.Println(jsonStr)
		} else {
			for i, c := range contacts {
				fmt.Printf("\n[%d] %s\n", i+1, c.DisplayName())
				if c.Organization != "" {
					fmt.Printf("    Organization: %s\n", c.Organization)
				}
				if c.Title != "" {
					fmt.Printf("    Title: %s\n", c.Title)
				}
				for _, e := range c.Emails {
					fmt.Printf("    Email: %s (%s)\n", e.Address, e.Type)
				}
				for _, p := range c.Phones {
					fmt.Printf("    Phone: %s (%s)\n", p.Number, p.Type)
				}
			}
		}
	}
	return nil
}

func cmdValidate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: calforge validate <file.ics|file.vcf>")
	}

	path := args[0]
	data, err := readFile(path)
	if err != nil {
		return err
	}

	fileType := detectType(path)
	if fileType == "unknown" {
		if strings.Contains(data, "BEGIN:VCALENDAR") {
			fileType = "ical"
		} else if strings.Contains(data, "BEGIN:VCARD") {
			fileType = "vcard"
		}
	}

	switch fileType {
	case "ical":
		cal, err := ical.Parse(data)
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}
		issues := ical.Validate(cal)
		if len(issues) == 0 {
			fmt.Println("✓ Valid iCalendar file")
		} else {
			fmt.Printf("Found %d issue(s):\n", len(issues))
			for _, issue := range issues {
				fmt.Printf("  ⚠ %s\n", issue)
			}
		}
	case "vcard":
		contacts, err := vcard.Parse(data)
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}
		totalIssues := 0
		for i, c := range contacts {
			issues := c.Validate()
			if len(issues) > 0 {
				fmt.Printf("Contact[%d] %s:\n", i+1, c.DisplayName())
				for _, issue := range issues {
					fmt.Printf("  ⚠ %s\n", issue)
				}
				totalIssues += len(issues)
			}
		}
		if totalIssues == 0 {
			fmt.Println("✓ All vCard contacts are valid")
		}
	}
	return nil
}

func cmdStats(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: calforge stats <file.ics>")
	}

	path := args[0]
	data, err := readFile(path)
	if err != nil {
		return err
	}

	cal, err := ical.Parse(data)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	stats := cal.Stats()
	fmt.Print(stats.StatsText())
	return nil
}

func cmdFilter(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: calforge filter --start <date> --end <date> <file.ics>")
	}

	var startDate, endDate, status, category, organizer string
	var filePath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--start":
			if i+1 < len(args) {
				startDate = args[i+1]
				i++
			}
		case "--end":
			if i+1 < len(args) {
				endDate = args[i+1]
				i++
			}
		case "--status":
			if i+1 < len(args) {
				status = args[i+1]
				i++
			}
		case "--category":
			if i+1 < len(args) {
				category = args[i+1]
				i++
			}
		case "--organizer":
			if i+1 < len(args) {
				organizer = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "-") {
				filePath = args[i]
			}
		}
	}

	if filePath == "" {
		return fmt.Errorf("no input file specified")
	}

	data, err := readFile(filePath)
	if err != nil {
		return err
	}

	cal, err := ical.Parse(data)
	if err != nil {
		return err
	}

	var events []*ical.Event

	if startDate != "" && endDate != "" {
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return fmt.Errorf("invalid start date: %w", err)
		}
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return fmt.Errorf("invalid end date: %w", err)
		}
		events = cal.FilterEventsByDateRange(start, end)
	} else if status != "" {
		events = cal.FilterEventsByStatus(status)
	} else if category != "" {
		events = cal.FilterEventsByCategory(category)
	} else if organizer != "" {
		events = cal.FilterEventsByOrganizer(organizer)
	}

	fmt.Printf("Found %d event(s):\n", len(events))
	for i, e := range events {
		startStr := ""
		if e.StartTime != nil {
			startStr = e.StartTime.Format("2006-01-02 15:04")
		}
		fmt.Printf("[%d] %s | %s\n", i+1, e.Summary, startStr)
	}
	return nil
}

func cmdSearch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: calforge search --query <text> <file.ics>")
	}

	var query string
	var filePath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--query", "-q":
			if i+1 < len(args) {
				query = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "-") {
				filePath = args[i]
			}
		}
	}

	if filePath == "" {
		return fmt.Errorf("no input file specified")
	}
	if query == "" {
		return fmt.Errorf("no search query specified")
	}

	data, err := readFile(filePath)
	if err != nil {
		return err
	}

	cal, err := ical.Parse(data)
	if err != nil {
		return err
	}

	events := cal.SearchEvents(query)
	fmt.Printf("Found %d event(s) matching %q:\n", len(events), query)
	for i, e := range events {
		startStr := ""
		if e.StartTime != nil {
			startStr = e.StartTime.Format("2006-01-02 15:04")
		}
		fmt.Printf("[%d] %s | %s\n", i+1, e.Summary, startStr)
	}
	return nil
}

func cmdConvert(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: calforge convert --from <ics|vcf> --to <json|csv> <input> [-o output]")
	}

	var fromFormat, toFormat, inputPath, outputPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from":
			if i+1 < len(args) {
				fromFormat = args[i+1]
				i++
			}
		case "--to":
			if i+1 < len(args) {
				toFormat = args[i+1]
				i++
			}
		case "-o", "--output":
			if i+1 < len(args) {
				outputPath = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "-") && inputPath == "" {
				inputPath = args[i]
			}
		}
	}

	if inputPath == "" {
		return fmt.Errorf("no input file specified")
	}

	// Auto-detect format if not specified
	if fromFormat == "" {
		fromFormat = detectType(inputPath)
	}

	data, err := readFile(inputPath)
	if err != nil {
		return err
	}

	var output string

	switch fromFormat {
	case "ics", "ical":
		cal, errInner := ical.Parse(data)
		if errInner != nil {
			return errInner
		}
		switch toFormat {
		case "json":
			var err error
			output, err = convert.EventsToJSON(cal)
			if err != nil {
				return err
			}
		case "csv":
			var err error
			output, err = convert.EventsToCSV(cal)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported output format: %s", toFormat)
		}
	case "vcf", "vcard":
		contacts, errInner := vcard.Parse(data)
		if errInner != nil {
			return errInner
		}
		switch toFormat {
		case "json":
			var err error
			output, err = convert.ContactsToJSON(contacts)
			if err != nil {
				return err
			}
		case "csv":
			var err error
			output, err = convert.ContactsToCSV(contacts)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported output format: %s", toFormat)
		}
	default:
		return fmt.Errorf("unsupported input format: %s", fromFormat)
	}

	if err != nil {
		return err
	}

	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		fmt.Printf("Written to %s\n", outputPath)
	} else {
		fmt.Print(output)
	}
	return nil
}

func cmdMerge(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: calforge merge <file1.ics> <file2.ics> [...] [-o output.ics]")
	}

	var files []string
	var outputPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o", "--output":
			if i+1 < len(args) {
				outputPath = args[i+1]
				i++
			}
		default:
			files = append(files, args[i])
		}
	}

	if len(files) < 2 {
		return fmt.Errorf("need at least 2 files to merge")
	}

	merged := &ical.Calendar{
		ProdID:  "-//CalForge//calforge//EN",
		Version: "2.0",
		Events:  make([]*ical.Event, 0),
	}

	for _, path := range files {
		data, err := readFile(path)
		if err != nil {
			return err
		}
		cal, err := ical.Parse(data)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		merged.Events = append(merged.Events, cal.Events...)
	}

	output := merged.Serialize()

	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		fmt.Printf("Merged %d events to %s\n", len(merged.Events), outputPath)
	} else {
		fmt.Print(output)
	}
	return nil
}

func cmdCreate(args []string) error {
	fileType := "ics"
	numItems := 3

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--type":
			if i+1 < len(args) {
				fileType = args[i+1]
				i++
			}
		case "--events", "--contacts":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &numItems)
				i++
			}
		}
	}

	switch fileType {
	case "ics", "ical":
		cal := &ical.Calendar{
			ProdID:  "-//CalForge//calforge//EN",
			Version: "2.0",
			Events:  make([]*ical.Event, 0),
		}

		now := time.Now()
		for i := 0; i < numItems; i++ {
			start := now.AddDate(0, 0, i+1)
			end := start.Add(time.Hour)
			cal.Events = append(cal.Events, &ical.Event{
				UID:       fmt.Sprintf("event-%d@calforge", i+1),
				Summary:   fmt.Sprintf("Sample Event %d", i+1),
				StartTime: &start,
				EndTime:   &end,
				Status:    "CONFIRMED",
			})
		}

		fmt.Print(cal.Serialize())

	case "vcf", "vcard":
		contacts := make([]*vcard.Contact, 0)
		for i := 0; i < numItems; i++ {
			contacts = append(contacts, &vcard.Contact{
				FullName:  fmt.Sprintf("John Doe %d", i+1),
				FirstName: "John",
				LastName:  fmt.Sprintf("Doe %d", i+1),
				Emails: []vcard.Email{
					{Address: fmt.Sprintf("john%d@example.com", i+1), Type: "WORK"},
				},
				Phones: []vcard.Phone{
					{Number: fmt.Sprintf("+1-555-010%d", i+1), Type: "WORK"},
				},
			})
		}
		fmt.Print(vcard.SerializeMultiple(contacts))
	}
	return nil
}

func cmdVCard(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf(`usage:
  calforge vcard parse <file.vcf>
  calforge vcard search --query <text> <file.vcf>
  calforge vcard format <file.vcf>
  calforge vcard validate <file.vcf>`)
	}

	subCmd := args[0]
	switch subCmd {
	case "parse":
		if len(args) < 2 {
			return fmt.Errorf("usage: calforge vcard parse <file.vcf>")
		}
		data, err := readFile(args[1])
		if err != nil {
			return err
		}
		contacts, err := vcard.Parse(data)
		if err != nil {
			return err
		}
		for i, c := range contacts {
			fmt.Printf("\n[%d] %s\n", i+1, c.DisplayName())
			if c.Organization != "" {
				fmt.Printf("    Organization: %s\n", c.Organization)
			}
			for _, e := range c.Emails {
				fmt.Printf("    Email: %s\n", e.Address)
			}
			for _, p := range c.Phones {
				fmt.Printf("    Phone: %s\n", p.Number)
			}
			if len(c.Addresses) > 0 {
				a := c.Addresses[0]
				fmt.Printf("    Address: %s, %s, %s\n", a.Street, a.City, a.Country)
			}
		}
	case "search":
		var query, filePath string
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--query", "-q":
				if i+1 < len(args) {
					query = args[i+1]
					i++
				}
			default:
				if !strings.HasPrefix(args[i], "-") {
					filePath = args[i]
				}
			}
		}
		if filePath == "" || query == "" {
			return fmt.Errorf("usage: calforge vcard search --query <text> <file.vcf>")
		}
		data, err := readFile(filePath)
		if err != nil {
			return err
		}
		contacts, err := vcard.Parse(data)
		if err != nil {
			return err
		}
		queryLower := strings.ToLower(query)
		for _, c := range contacts {
			if strings.Contains(strings.ToLower(c.FullName), queryLower) ||
				strings.Contains(strings.ToLower(c.Organization), queryLower) ||
				strings.Contains(strings.ToLower(c.Emails[0].Address), queryLower) {
				fmt.Printf("%s - %s\n", c.DisplayName(), c.PrimaryEmail())
			}
		}
	case "format":
		if len(args) < 2 {
			return fmt.Errorf("usage: calforge vcard format <file.vcf>")
		}
		data, err := readFile(args[1])
		if err != nil {
			return err
		}
		contacts, err := vcard.Parse(data)
		if err != nil {
			return err
		}
		fmt.Print(vcard.SerializeMultiple(contacts))
	case "validate":
		if len(args) < 2 {
			return fmt.Errorf("usage: calforge vcard validate <file.vcf>")
		}
		data, err := readFile(args[1])
		if err != nil {
			return err
		}
		contacts, err := vcard.Parse(data)
		if err != nil {
			return err
		}
		totalIssues := 0
		for i, c := range contacts {
			issues := c.Validate()
			if len(issues) > 0 {
				fmt.Printf("[%d] %s:\n", i+1, c.DisplayName())
				for _, issue := range issues {
					fmt.Printf("  ⚠ %s\n", issue)
				}
				totalIssues += len(issues)
			}
		}
		if totalIssues == 0 {
			fmt.Println("✓ All contacts valid")
		}
	default:
		return fmt.Errorf("unknown vcard subcommand: %s", subCmd)
	}
	return nil
}
