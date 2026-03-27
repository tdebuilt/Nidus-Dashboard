package rss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

// Client fetches and parses RSS/Atom feeds.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new RSS client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{httpClient: httpClient}
}

// FetchFeeds fetches one or more RSS/Atom feed URLs and returns combined items.
func (c *Client) FetchFeeds(ctx context.Context, urls []string, maxItems int) (*FeedData, error) {
	if maxItems <= 0 {
		maxItems = 20
	}

	var allItems []FeedItem

	for _, url := range urls {
		items, err := c.fetchURL(ctx, url)
		if err != nil {
			// Log error but continue with other URLs
			continue
		}
		allItems = append(allItems, items...)
	}

	// Sort by published date (newest first)
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].Published > allItems[j].Published
	})

	// Limit to maxItems
	if len(allItems) > maxItems {
		allItems = allItems[:maxItems]
	}

	return &FeedData{Items: allItems}, nil
}

func (c *Client) fetchURL(ctx context.Context, url string) ([]FeedItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for %s: %w", url, err)
	}
	req.Header.Set("User-Agent", "Nidus-Dashboard/1.0")
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml")

	resp, err := c.httpClient.Do(req)
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

	return Parse(body, url)
}
