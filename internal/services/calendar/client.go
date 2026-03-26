package calendar

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

// Client fetches and parses iCal feeds.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new calendar client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{httpClient: httpClient}
}

// FetchEvents fetches one or more iCal URLs and returns upcoming events.
func (c *Client) FetchEvents(urls []string, days int) (*CalendarData, error) {
	if days <= 0 {
		days = 14
	}

	now := time.Now()
	from := now.Add(-24 * time.Hour) // include today's past events
	to := now.Add(time.Duration(days) * 24 * time.Hour)

	var allEvents []CalendarEvent

	for _, url := range urls {
		events, err := c.fetchURL(url, from, to)
		if err != nil {
			// Log error but continue with other URLs
			continue
		}
		allEvents = append(allEvents, events...)
	}

	// Sort by start time
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Start < allEvents[j].Start
	})

	return &CalendarData{Events: allEvents}, nil
}

func (c *Client) fetchURL(url string, from, to time.Time) ([]CalendarEvent, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body from %s: %w", url, err)
	}

	parsed, err := ParseICal(string(body), from, to)
	if err != nil {
		return nil, fmt.Errorf("parsing ical from %s: %w", url, err)
	}

	events := make([]CalendarEvent, 0, len(parsed))
	for _, p := range parsed {
		ev := CalendarEvent{
			UID:         p.uid,
			Summary:     p.summary,
			Description: p.description,
			Location:    p.location,
			AllDay:      p.allDay,
			CalendarURL: url,
		}
		if p.allDay {
			ev.Start = p.dtStart.Format("2006-01-02")
			ev.End = p.dtEnd.Format("2006-01-02")
		} else {
			ev.Start = p.dtStart.Format(time.RFC3339)
			ev.End = p.dtEnd.Format(time.RFC3339)
		}
		events = append(events, ev)
	}

	return events, nil
}
