package calendar

import (
	"testing"
	"time"
)

func TestParseICalBasic(t *testing.T) {
	ical := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:event1@test
SUMMARY:Team Meeting
DESCRIPTION:Weekly sync
LOCATION:Room 42
DTSTART:20260301T100000Z
DTEND:20260301T110000Z
END:VEVENT
BEGIN:VEVENT
UID:event2@test
SUMMARY:Lunch
DTSTART:20260301T120000Z
DTEND:20260301T130000Z
END:VEVENT
END:VCALENDAR`

	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)

	events, err := ParseICal(ical, from, to)
	if err != nil {
		t.Fatalf("ParseICal error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].summary != "Team Meeting" {
		t.Errorf("expected 'Team Meeting', got %q", events[0].summary)
	}
	if events[0].description != "Weekly sync" {
		t.Errorf("expected 'Weekly sync', got %q", events[0].description)
	}
	if events[0].location != "Room 42" {
		t.Errorf("expected 'Room 42', got %q", events[0].location)
	}
	if events[1].summary != "Lunch" {
		t.Errorf("expected 'Lunch', got %q", events[1].summary)
	}
}

func TestParseICalAllDay(t *testing.T) {
	ical := `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:holiday@test
SUMMARY:Public Holiday
DTSTART;VALUE=DATE:20260301
DTEND;VALUE=DATE:20260302
END:VEVENT
END:VCALENDAR`

	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)

	events, err := ParseICal(ical, from, to)
	if err != nil {
		t.Fatalf("ParseICal error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if !events[0].allDay {
		t.Error("expected allDay=true")
	}
	if events[0].summary != "Public Holiday" {
		t.Errorf("expected 'Public Holiday', got %q", events[0].summary)
	}
}

func TestParseICalTimeRange(t *testing.T) {
	ical := `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:past@test
SUMMARY:Past Event
DTSTART:20260101T100000Z
DTEND:20260101T110000Z
END:VEVENT
BEGIN:VEVENT
UID:future@test
SUMMARY:Future Event
DTSTART:20260601T100000Z
DTEND:20260601T110000Z
END:VEVENT
END:VCALENDAR`

	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)

	events, err := ParseICal(ical, from, to)
	if err != nil {
		t.Fatalf("ParseICal error: %v", err)
	}

	if len(events) != 0 {
		t.Fatalf("expected 0 events in March range, got %d", len(events))
	}
}

func TestParseICalWithTZID(t *testing.T) {
	ical := `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:tz@test
SUMMARY:Paris Meeting
DTSTART;TZID=Europe/Paris:20260301T100000
DTEND;TZID=Europe/Paris:20260301T110000
END:VEVENT
END:VCALENDAR`

	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)

	events, err := ParseICal(ical, from, to)
	if err != nil {
		t.Fatalf("ParseICal error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	// 10:00 Europe/Paris = 09:00 UTC in March (CET = UTC+1)
	expectedHour := 9
	if events[0].dtStart.UTC().Hour() != expectedHour {
		t.Errorf("expected start hour %d UTC, got %d", expectedHour, events[0].dtStart.UTC().Hour())
	}
}

func TestParseICalLineUnfolding(t *testing.T) {
	ical := "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:fold@test\r\nSUMMARY:Very Long\r\n  Event Name\r\nDTSTART:20260301T100000Z\r\nDTEND:20260301T110000Z\r\nEND:VEVENT\r\nEND:VCALENDAR"

	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)

	events, err := ParseICal(ical, from, to)
	if err != nil {
		t.Fatalf("ParseICal error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].summary != "Very Long Event Name" {
		t.Errorf("expected 'Very Long Event Name', got %q", events[0].summary)
	}
}

func TestUnescapeICal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`Hello\, World`, "Hello, World"},
		{`Line 1\nLine 2`, "Line 1\nLine 2"},
		{`Semi\;colon`, "Semi;colon"},
		{`Back\\slash`, "Back\\slash"},
	}

	for _, tc := range tests {
		result := unescapeICal(tc.input)
		if result != tc.expected {
			t.Errorf("unescapeICal(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}
