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
	lines := unfoldICalLines(data)
	var events []parsedEvent
	var current *parsedEvent
	inEvent := false

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")

		if line == "BEGIN:VEVENT" {
			inEvent = true
			current = &parsedEvent{}
			continue
		}
		if line == "END:VEVENT" && inEvent {
			inEvent = false
			if ev := finalizeEvent(current, from, to); ev != nil {
				events = append(events, *ev)
			}
			current = nil
			continue
		}
		if !inEvent || current == nil {
			continue
		}
		parseICalProperty(current, line)
	}

	return events, nil
}

// unfoldICalLines handles RFC 5545 §3.1 line unfolding (continuation lines).
func unfoldICalLines(data string) []string {
	scanner := bufio.NewScanner(strings.NewReader(data))
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') && len(lines) > 0 {
			lines[len(lines)-1] += line[1:]
		} else {
			lines = append(lines, line)
		}
	}
	return lines
}

// finalizeEvent validates and filters a parsed event against the time range.
func finalizeEvent(ev *parsedEvent, from, to time.Time) *parsedEvent {
	if ev == nil || ev.summary == "" || ev.dtStart.IsZero() {
		return nil
	}
	if ev.dtEnd.IsZero() {
		if ev.allDay {
			ev.dtEnd = ev.dtStart.Add(24 * time.Hour)
		} else {
			ev.dtEnd = ev.dtStart.Add(time.Hour)
		}
	}
	if ev.dtEnd.After(from) && ev.dtStart.Before(to) {
		return ev
	}
	return nil
}

// parseICalProperty parses a single iCal property line into the event.
func parseICalProperty(ev *parsedEvent, line string) {
	key, value := splitICalLine(line)
	switch {
	case key == "SUMMARY":
		ev.summary = unescapeICal(value)
	case key == "DESCRIPTION":
		ev.description = unescapeICal(value)
	case key == "LOCATION":
		ev.location = unescapeICal(value)
	case key == "UID":
		ev.uid = value
	case strings.HasPrefix(key, "DTSTART"):
		if t, allDay, err := parseICalDate(key, value); err == nil {
			ev.dtStart = t
			ev.allDay = allDay
		}
	case strings.HasPrefix(key, "DTEND"):
		if t, _, err := parseICalDate(key, value); err == nil {
			ev.dtEnd = t
		}
	}
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

	loc := extractTZID(key)

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

// extractTZID extracts a TZID time.Location from an iCal property key, if present.
func extractTZID(key string) *time.Location {
	if !strings.Contains(key, "TZID=") {
		return nil
	}
	for _, param := range strings.Split(key, ";") {
		if !strings.HasPrefix(param, "TZID=") {
			continue
		}
		tzName := strings.TrimPrefix(param, "TZID=")
		if loc, err := time.LoadLocation(tzName); err == nil {
			return loc
		}
	}
	return nil
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
