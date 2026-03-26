package calendar

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

// ParseICal parses an iCal/ICS string and returns events within the given time range.
// It handles VEVENT blocks with DTSTART, DTEND, SUMMARY, DESCRIPTION, LOCATION.
// Recurring events (RRULE) are not supported in this version.
func ParseICal(data string, from, to time.Time) ([]parsedEvent, error) {
	scanner := bufio.NewScanner(strings.NewReader(data))
	var events []parsedEvent
	var current *parsedEvent
	inEvent := false

	// Handle line unfolding (RFC 5545 §3.1): lines starting with space/tab are continuations
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') && len(lines) > 0 {
			lines[len(lines)-1] += line[1:]
		} else {
			lines = append(lines, line)
		}
	}

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")

		if line == "BEGIN:VEVENT" {
			inEvent = true
			current = &parsedEvent{}
			continue
		}

		if line == "END:VEVENT" && inEvent {
			inEvent = false
			if current.summary != "" && !current.dtStart.IsZero() {
				// Filter by time range
				if current.dtEnd.IsZero() {
					if current.allDay {
						current.dtEnd = current.dtStart.Add(24 * time.Hour)
					} else {
						current.dtEnd = current.dtStart.Add(time.Hour)
					}
				}
				// Include event if it overlaps with our time range
				if current.dtEnd.After(from) && current.dtStart.Before(to) {
					events = append(events, *current)
				}
			}
			current = nil
			continue
		}

		if !inEvent || current == nil {
			continue
		}

		// Parse property
		key, value := splitICalLine(line)
		switch {
		case key == "SUMMARY":
			current.summary = unescapeICal(value)
		case key == "DESCRIPTION":
			current.description = unescapeICal(value)
		case key == "LOCATION":
			current.location = unescapeICal(value)
		case key == "UID":
			current.uid = value
		case strings.HasPrefix(key, "DTSTART"):
			t, allDay, err := parseICalDate(key, value)
			if err == nil {
				current.dtStart = t
				current.allDay = allDay
			}
		case strings.HasPrefix(key, "DTEND"):
			t, _, err := parseICalDate(key, value)
			if err == nil {
				current.dtEnd = t
			}
		}
	}

	return events, nil
}

// splitICalLine splits "KEY;PARAM=val:value" into (key with params, value).
func splitICalLine(line string) (string, string) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return line, ""
	}
	return line[:idx], line[idx+1:]
}

// parseICalDate parses iCal date/datetime values.
// Supports formats: 20240101T120000Z, 20240101T120000, 20240101, with optional TZID.
func parseICalDate(key, value string) (time.Time, bool, error) {
	value = strings.TrimSpace(value)

	// Check for VALUE=DATE (all-day event)
	isDate := strings.Contains(key, "VALUE=DATE") && !strings.Contains(key, "VALUE=DATE-TIME")

	// Extract TZID if present
	var loc *time.Location
	if strings.Contains(key, "TZID=") {
		for _, param := range strings.Split(key, ";") {
			if strings.HasPrefix(param, "TZID=") {
				tzName := strings.TrimPrefix(param, "TZID=")
				if l, err := time.LoadLocation(tzName); err == nil {
					loc = l
				}
			}
		}
	}

	if isDate || len(value) == 8 {
		// Date only: 20240101
		t, err := time.Parse("20060102", value)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("parsing date %q: %w", value, err)
		}
		return t, true, nil
	}

	// DateTime formats
	if strings.HasSuffix(value, "Z") {
		// UTC: 20240101T120000Z
		t, err := time.Parse("20060102T150405Z", value)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("parsing datetime %q: %w", value, err)
		}
		return t, false, nil
	}

	// Local or with TZID: 20240101T120000
	t, err := time.Parse("20060102T150405", value)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("parsing datetime %q: %w", value, err)
	}

	if loc != nil {
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
	}

	return t, false, nil
}

// unescapeICal unescapes iCal text values.
func unescapeICal(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\N", "\n")
	s = strings.ReplaceAll(s, "\\,", ",")
	s = strings.ReplaceAll(s, "\\;", ";")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}
