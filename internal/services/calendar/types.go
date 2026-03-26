package calendar

import "time"

// CalendarEvent represents a single calendar event.
type CalendarEvent struct {
	UID         string `json:"uid"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
	Start       string `json:"start"`
	End         string `json:"end"`
	AllDay      bool   `json:"all_day"`
	Color       string `json:"color,omitempty"`
	CalendarURL string `json:"calendar_url,omitempty"`
}

// CalendarData holds the list of upcoming events.
type CalendarData struct {
	Events []CalendarEvent `json:"events"`
}

// parsedEvent is an intermediate struct used during iCal parsing.
type parsedEvent struct {
	uid         string
	summary     string
	description string
	location    string
	dtStart     time.Time
	dtEnd       time.Time
	allDay      bool
}
