package rss

import (
	"testing"
)

func TestParseRSS(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Blog</title>
    <item>
      <title>First Article</title>
      <link>https://example.com/1</link>
      <description>&lt;p&gt;Hello &amp; world&lt;/p&gt;</description>
      <pubDate>Mon, 01 Mar 2026 10:00:00 +0000</pubDate>
      <author>Alice</author>
    </item>
    <item>
      <title>Second Article</title>
      <link>https://example.com/2</link>
      <description>Plain text description</description>
      <pubDate>Sun, 28 Feb 2026 08:00:00 +0000</pubDate>
    </item>
  </channel>
</rss>`)

	items, err := Parse(data, "https://example.com/feed")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if items[0].Title != "First Article" {
		t.Errorf("expected 'First Article', got %q", items[0].Title)
	}
	if items[0].Link != "https://example.com/1" {
		t.Errorf("expected link 'https://example.com/1', got %q", items[0].Link)
	}
	if items[0].Description != "Hello & world" {
		t.Errorf("expected 'Hello & world', got %q", items[0].Description)
	}
	if items[0].Author != "Alice" {
		t.Errorf("expected author 'Alice', got %q", items[0].Author)
	}
	if items[0].Source != "Test Blog" {
		t.Errorf("expected source 'Test Blog', got %q", items[0].Source)
	}
}

func TestParseAtom(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Atom Blog</title>
  <entry>
    <title>Atom Entry</title>
    <link href="https://example.com/atom/1" rel="alternate"/>
    <summary>A summary</summary>
    <published>2026-03-01T12:00:00Z</published>
    <author><name>Bob</name></author>
  </entry>
  <entry>
    <title>Second Entry</title>
    <link href="https://example.com/atom/2"/>
    <content>Full content here</content>
    <updated>2026-02-28T10:00:00Z</updated>
  </entry>
</feed>`)

	items, err := Parse(data, "https://example.com/atom")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if items[0].Title != "Atom Entry" {
		t.Errorf("expected 'Atom Entry', got %q", items[0].Title)
	}
	if items[0].Link != "https://example.com/atom/1" {
		t.Errorf("expected link, got %q", items[0].Link)
	}
	if items[0].Author != "Bob" {
		t.Errorf("expected author 'Bob', got %q", items[0].Author)
	}
	if items[0].Source != "Atom Blog" {
		t.Errorf("expected source 'Atom Blog', got %q", items[0].Source)
	}

	// Second entry uses content instead of summary, and updated instead of published
	if items[1].Description != "Full content here" {
		t.Errorf("expected 'Full content here', got %q", items[1].Description)
	}
	if items[1].Published == "" {
		t.Error("expected published date from updated field")
	}
}

func TestParseInvalidFormat(t *testing.T) {
	data := []byte(`<html><body>Not a feed</body></html>`)

	_, err := Parse(data, "https://example.com")
	if err == nil {
		t.Fatal("expected error for invalid feed format")
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<p>Hello</p>", "Hello"},
		{"<b>Bold</b> and <i>italic</i>", "Bold and italic"},
		{"No tags", "No tags"},
		{"&amp; entity", "& entity"},
	}

	for _, tc := range tests {
		result := stripHTML(tc.input)
		if result != tc.expected {
			t.Errorf("stripHTML(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestNormalizeDate(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"Mon, 01 Mar 2026 10:00:00 +0000", true},
		{"2026-03-01T12:00:00Z", true},
		{"2026-03-01T12:00:00+01:00", true},
		{"", false},
	}

	for _, tc := range tests {
		result := normalizeDate(tc.input)
		if tc.valid && result == "" {
			t.Errorf("normalizeDate(%q) returned empty", tc.input)
		}
		if !tc.valid && result != "" {
			t.Errorf("normalizeDate(%q) expected empty, got %q", tc.input, result)
		}
	}
}

func TestTruncateText(t *testing.T) {
	short := "Hello"
	if truncateText(short, 10) != "Hello" {
		t.Error("should not truncate short text")
	}

	long := "This is a very long text that should be truncated"
	result := truncateText(long, 20)
	if len([]rune(result)) != 21 { // 20 chars + ellipsis
		t.Errorf("expected 21 runes, got %d: %q", len([]rune(result)), result)
	}
}
