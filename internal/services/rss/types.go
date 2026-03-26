package rss

// FeedItem represents a single article from an RSS/Atom feed.
type FeedItem struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description,omitempty"`
	Published   string `json:"published,omitempty"`
	Author      string `json:"author,omitempty"`
	Source      string `json:"source,omitempty"`
}

// FeedData holds the list of feed items.
type FeedData struct {
	Items []FeedItem `json:"items"`
}
