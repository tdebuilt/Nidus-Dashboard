package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api.openweathermap.org/data/2.5"

// Client communicates with the OpenWeatherMap API.
type Client struct {
	apiKey     string
	units      string
	lang       string
	httpClient *http.Client
}

// NewClient creates an OpenWeatherMap API client.
func NewClient(apiKey, units, lang string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	if units == "" {
		units = "metric"
	}
	if lang == "" {
		lang = "fr"
	}
	return &Client{
		apiKey:     apiKey,
		units:      units,
		lang:       lang,
		httpClient: httpClient,
	}
}

// GetWeather fetches current weather and 5-day forecast for a location.
// Location can be "city,country" or "lat,lon".
func (c *Client) GetWeather(ctx context.Context, lat, lon string) (*WeatherData, error) {
	current, err := c.getCurrent(ctx, lat, lon)
	if err != nil {
		return nil, err
	}

	forecast, err := c.getForecast(ctx, lat, lon)
	if err != nil {
		return nil, err
	}

	return &WeatherData{
		Current:  *current,
		Forecast: forecast,
	}, nil
}

// GetCurrentByCity fetches current weather by city name.
func (c *Client) GetCurrentByCity(ctx context.Context, city string) (*WeatherData, error) {
	url := fmt.Sprintf("%s/weather?q=%s&appid=%s&units=%s&lang=%s",
		baseURL, city, c.apiKey, c.units, c.lang)

	var resp owmCurrentResponse
	if err := c.get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("fetching current weather: %w", err)
	}

	current := convertCurrent(&resp)

	// Also fetch forecast by city
	forecastURL := fmt.Sprintf("%s/forecast?q=%s&appid=%s&units=%s&lang=%s",
		baseURL, city, c.apiKey, c.units, c.lang)

	var fResp owmForecastResponse
	if err := c.get(ctx, forecastURL, &fResp); err != nil {
		// Return current only if forecast fails
		return &WeatherData{Current: current}, nil
	}

	return &WeatherData{
		Current:  current,
		Forecast: aggregateForecast(&fResp),
	}, nil
}

func (c *Client) getCurrent(ctx context.Context, lat, lon string) (*CurrentWeather, error) {
	url := fmt.Sprintf("%s/weather?lat=%s&lon=%s&appid=%s&units=%s&lang=%s",
		baseURL, lat, lon, c.apiKey, c.units, c.lang)

	var resp owmCurrentResponse
	if err := c.get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("fetching current weather: %w", err)
	}

	current := convertCurrent(&resp)
	return &current, nil
}

func (c *Client) getForecast(ctx context.Context, lat, lon string) ([]ForecastDay, error) {
	url := fmt.Sprintf("%s/forecast?lat=%s&lon=%s&appid=%s&units=%s&lang=%s",
		baseURL, lat, lon, c.apiKey, c.units, c.lang)

	var resp owmForecastResponse
	if err := c.get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("fetching forecast: %w", err)
	}

	return aggregateForecast(&resp), nil
}

func convertCurrent(resp *owmCurrentResponse) CurrentWeather {
	cw := CurrentWeather{
		Temp:      resp.Main.Temp,
		FeelsLike: resp.Main.FeelsLike,
		TempMin:   resp.Main.TempMin,
		TempMax:   resp.Main.TempMax,
		Humidity:  resp.Main.Humidity,
		Pressure:  resp.Main.Pressure,
		WindSpeed: resp.Wind.Speed,
		WindDeg:   resp.Wind.Deg,
		City:      resp.Name,
		Country:   resp.Sys.Country,
		Sunrise:   resp.Sys.Sunrise,
		Sunset:    resp.Sys.Sunset,
	}
	if len(resp.Weather) > 0 {
		cw.Description = resp.Weather[0].Description
		cw.Icon = resp.Weather[0].Icon
	}
	return cw
}

// aggregateForecast groups 3-hour forecasts into daily summaries.
func aggregateForecast(resp *owmForecastResponse) []ForecastDay {
	type dayData struct {
		date    string
		tempMin float64
		tempMax float64
		icon    string
		desc    string
		hum     int
		wind    float64
		count   int
	}

	days := make(map[string]*dayData)
	var order []string

	for _, item := range resp.List {
		t := time.Unix(item.Dt, 0)
		date := t.Format("2006-01-02")

		d, exists := days[date]
		if !exists {
			d = &dayData{
				date:    date,
				tempMin: item.Main.TempMin,
				tempMax: item.Main.TempMax,
			}
			days[date] = d
			order = append(order, date)
		}

		if item.Main.TempMin < d.tempMin {
			d.tempMin = item.Main.TempMin
		}
		if item.Main.TempMax > d.tempMax {
			d.tempMax = item.Main.TempMax
		}

		d.hum += item.Main.Humidity
		d.wind += item.Wind.Speed
		d.count++

		// Use midday icon/description as representative
		hour := t.Hour()
		if hour >= 11 && hour <= 14 {
			if len(item.Weather) > 0 {
				d.icon = item.Weather[0].Icon
				d.desc = item.Weather[0].Description
			}
		}
	}

	result := make([]ForecastDay, 0, len(order))
	for _, date := range order {
		d := days[date]
		// Use first available icon if midday wasn't found
		if d.icon == "" {
			for _, item := range resp.List {
				t := time.Unix(item.Dt, 0)
				if t.Format("2006-01-02") == date && len(item.Weather) > 0 {
					d.icon = item.Weather[0].Icon
					d.desc = item.Weather[0].Description
					break
				}
			}
		}
		result = append(result, ForecastDay{
			Date:        d.date,
			TempMin:     d.tempMin,
			TempMax:     d.tempMax,
			Description: d.desc,
			Icon:        d.icon,
			Humidity:    d.hum / d.count,
			WindSpeed:   d.wind / float64(d.count),
		})
	}

	// Limit to 5 days
	if len(result) > 5 {
		result = result[:5]
	}

	return result
}

func (c *Client) get(ctx context.Context, url string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

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
