package finance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	chartBaseURL  = "https://query1.finance.yahoo.com/v8/finance/chart"
	searchBaseURL = "https://query1.finance.yahoo.com/v1/finance/search"
	userAgent     = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// Client communicates with the Yahoo Finance API.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a Yahoo Finance API client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{httpClient: httpClient}
}

// GetQuotes fetches real-time quotes for the given symbols using the v8 chart API.
// Requests are made in parallel (one per symbol).
func (c *Client) GetQuotes(ctx context.Context, symbols []string) (*QuotesResponse, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols provided")
	}

	type result struct {
		quote Quote
		err   error
	}

	results := make([]result, len(symbols))
	var wg sync.WaitGroup

	for i, symbol := range symbols {
		wg.Add(1)
		go func(idx int, sym string) {
			defer wg.Done()
			q, err := c.fetchChart(ctx, sym)
			results[idx] = result{quote: q, err: err}
		}(i, symbol)
	}
	wg.Wait()

	quotes := make([]Quote, 0, len(symbols))
	for _, r := range results {
		if r.err == nil {
			quotes = append(quotes, r.quote)
		}
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("failed to fetch any quotes")
	}

	return &QuotesResponse{
		Quotes:    quotes,
		FetchedAt: time.Now().Unix(),
	}, nil
}

func (c *Client) fetchChart(ctx context.Context, symbol string) (Quote, error) {
	apiURL := fmt.Sprintf("%s/%s?range=1d&interval=1d", chartBaseURL, url.PathEscape(symbol))

	var resp yahooChartResponse
	if err := c.doRequest(ctx, apiURL, &resp); err != nil {
		return Quote{}, fmt.Errorf("fetching chart for %s: %w", symbol, err)
	}
	if len(resp.Chart.Result) == 0 {
		return Quote{}, fmt.Errorf("no chart data for %s", symbol)
	}

	return buildQuoteFromChart(resp.Chart.Result[0]), nil
}

// buildQuoteFromChart converts a Yahoo chart result into a Quote.
func buildQuoteFromChart(result yahooChartResult) Quote {
	meta := result.Meta
	price := meta.RegularMarketPrice
	prevClose := meta.ChartPreviousClose
	change := price - prevClose
	changePct := 0.0
	if prevClose != 0 {
		changePct = (change / prevClose) * 100
	}

	openPrice := 0.0
	if len(result.Indicators.Quote) > 0 && len(result.Indicators.Quote[0].Open) > 0 {
		openPrice = result.Indicators.Quote[0].Open[0]
	}

	name := meta.ShortName
	if name == "" {
		name = meta.LongName
	}

	return Quote{
		Symbol:        meta.Symbol,
		ShortName:     name,
		QuoteType:     meta.InstrumentType,
		Currency:      meta.Currency,
		Price:         math.Round(price*100) / 100,
		Change:        math.Round(change*100) / 100,
		ChangePercent: math.Round(changePct*100) / 100,
		Volume:        meta.RegularMarketVolume,
		Open:          openPrice,
		DayHigh:       meta.RegularMarketDayHigh,
		DayLow:        meta.RegularMarketDayLow,
		MarketCap:     0,
		MarketState:   "REGULAR",
	}
}

// Search finds symbols matching the query (for autocomplete).
func (c *Client) Search(ctx context.Context, query string) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("empty search query")
	}

	apiURL := fmt.Sprintf("%s?q=%s&quotesCount=10&newsCount=0&enableFuzzyQuery=false",
		searchBaseURL, url.QueryEscape(query))

	var resp yahooSearchResponse
	if err := c.doRequest(ctx, apiURL, &resp); err != nil {
		return nil, fmt.Errorf("searching symbols: %w", err)
	}

	results := make([]SearchResult, 0, len(resp.Quotes))
	for _, q := range resp.Quotes {
		results = append(results, SearchResult{
			Symbol:   q.Symbol,
			Name:     q.ShortName,
			Type:     q.QuoteType,
			Exchange: q.Exchange,
		})
	}

	return results, nil
}

func (c *Client) doRequest(ctx context.Context, apiURL string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}
