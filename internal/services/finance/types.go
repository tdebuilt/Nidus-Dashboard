package finance

// Quote holds normalized quote data for a single symbol.
type Quote struct {
	Symbol        string  `json:"symbol"`
	ShortName     string  `json:"short_name"`
	QuoteType     string  `json:"quote_type"`
	Currency      string  `json:"currency"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Volume        int64   `json:"volume"`
	Open          float64 `json:"open"`
	DayHigh       float64 `json:"day_high"`
	DayLow        float64 `json:"day_low"`
	MarketCap     int64   `json:"market_cap"`
	MarketState   string  `json:"market_state"`
}

// QuotesResponse is returned to the frontend.
type QuotesResponse struct {
	Quotes    []Quote `json:"quotes"`
	FetchedAt int64   `json:"fetched_at"`
}

// SearchResult holds a single symbol search result.
type SearchResult struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Exchange string `json:"exchange"`
}

// yahooChartResponse maps the Yahoo Finance v8 chart API response.
type yahooChartResponse struct {
	Chart struct {
		Result []yahooChartResult `json:"result"`
		Error  interface{}        `json:"error"`
	} `json:"chart"`
}

type yahooChartResult struct {
	Meta struct {
		Currency             string  `json:"currency"`
		Symbol               string  `json:"symbol"`
		ShortName            string  `json:"shortName"`
		LongName             string  `json:"longName"`
		InstrumentType       string  `json:"instrumentType"`
		RegularMarketPrice   float64 `json:"regularMarketPrice"`
		RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
		RegularMarketDayLow  float64 `json:"regularMarketDayLow"`
		RegularMarketVolume  int64   `json:"regularMarketVolume"`
		ChartPreviousClose   float64 `json:"chartPreviousClose"`
		FiftyTwoWeekHigh     float64 `json:"fiftyTwoWeekHigh"`
		FiftyTwoWeekLow      float64 `json:"fiftyTwoWeekLow"`
	} `json:"meta"`
	Indicators struct {
		Quote []struct {
			Open   []float64 `json:"open"`
			High   []float64 `json:"high"`
			Low    []float64 `json:"low"`
			Close  []float64 `json:"close"`
			Volume []int64   `json:"volume"`
		} `json:"quote"`
	} `json:"indicators"`
}

// yahooSearchResponse maps the Yahoo Finance v1 search API response.
type yahooSearchResponse struct {
	Quotes []struct {
		Symbol    string `json:"symbol"`
		ShortName string `json:"shortname"`
		QuoteType string `json:"quoteType"`
		Exchange  string `json:"exchange"`
	} `json:"quotes"`
}
