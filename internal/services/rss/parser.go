package rss

import (
	"encoding/xml"
	"fmt"
	"html"
	"strings"
	"time"
)

// RSS 2.0 structures
type rssDoc struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Author      string `xml:"author"`
	Creator     string `xml:"creator"`
}

// Atom structures
type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string     `xml:"title"`
	Links     []atomLink `xml:"link"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
	Author    atomAuthor `xml:"author"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

// Parse detects the feed format (RSS 2.0 or Atom) and parses it.
func Parse(data []byte, sourceURL string) ([]FeedItem, error) {
	// Try RSS first
	var rssFeed rssDoc
	if err := xml.Unmarshal(data, &rssFeed); err == nil && rssFeed.XMLName.Local == "rss" {
		return parseRSS(&rssFeed, sourceURL), nil
	}

	// Try Atom
	var atom atomFeed
	if err := xml.Unmarshal(data, &atom); err == nil && atom.XMLName.Local == "feed" {
		return parseAtom(&atom, sourceURL), nil
	}

	return nil, fmt.Errorf("unrecognized feed format")
}

func parseRSS(doc *rssDoc, sourceURL string) []FeedItem {
	source := doc.Channel.Title
	if source == "" {
		source = sourceURL
	}

	items := make([]FeedItem, 0, len(doc.Channel.Items))
	for _, item := range doc.Channel.Items {
		author := item.Author
		if author == "" {
			author = item.Creator
		}

		fi := FeedItem{
			Title:       cleanText(item.Title),
			Link:        item.Link,
			Description: truncateText(stripHTML(item.Description), 200),
			Published:   normalizeDate(item.PubDate),
			Author:      author,
			Source:      source,
		}
		items = append(items, fi)
	}
	return items
}

func parseAtom(feed *atomFeed, sourceURL string) []FeedItem {
	source := feed.Title
	if source == "" {
		source = sourceURL
	}

	items := make([]FeedItem, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		link := ""
		for _, l := range entry.Links {
			if l.Rel == "alternate" || l.Rel == "" {
				link = l.Href
				break
			}
		}
		if link == "" && len(entry.Links) > 0 {
			link = entry.Links[0].Href
		}

		desc := entry.Summary
		if desc == "" {
			desc = entry.Content
		}

		pubDate := entry.Published
		if pubDate == "" {
			pubDate = entry.Updated
		}

		fi := FeedItem{
			Title:       cleanText(entry.Title),
			Link:        link,
			Description: truncateText(stripHTML(desc), 200),
			Published:   normalizeDate(pubDate),
			Author:      entry.Author.Name,
			Source:      source,
		}
		items = append(items, fi)
	}
	return items
}

// normalizeDate tries to parse various date formats and returns RFC3339.
func normalizeDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"02 Jan 2006 15:04:05 -0700",
	}

	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.Format(time.RFC3339)
		}
	}
	return s // Return as-is if unparseable
}

// stripHTML removes HTML tags from text.
func stripHTML(s string) string {
	s = html.UnescapeString(s)
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

// cleanText trims and unescapes HTML entities.
func cleanText(s string) string {
	return strings.TrimSpace(html.UnescapeString(s))
}

// truncateText truncates text to maxLen characters with ellipsis.
func truncateText(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "…"
}
