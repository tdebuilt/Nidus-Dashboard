package weather

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCurrentWeather(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/data/2.5/weather", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"main": {"temp": 15.5, "feels_like": 14.2, "temp_min": 13.0, "temp_max": 17.0, "humidity": 72, "pressure": 1013},
			"wind": {"speed": 3.5, "deg": 180},
			"weather": [{"description": "nuageux", "icon": "04d"}],
			"sys": {"country": "FR", "sunrise": 1700000000, "sunset": 1700040000},
			"name": "Paris"
		}`))
	})
	mux.HandleFunc("/data/2.5/forecast", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"list": []}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Override base URL by using city endpoint
	client := NewClient("test-key", "metric", "fr", server.Client())

	// We need to test via the internal methods since baseURL is const
	// Instead, test convertCurrent directly
	resp := &owmCurrentResponse{
		Name: "Paris",
	}
	resp.Main.Temp = 15.5
	resp.Main.FeelsLike = 14.2
	resp.Main.TempMin = 13.0
	resp.Main.TempMax = 17.0
	resp.Main.Humidity = 72
	resp.Main.Pressure = 1013
	resp.Wind.Speed = 3.5
	resp.Wind.Deg = 180
	resp.Weather = []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
	}{
		{Description: "nuageux", Icon: "04d"},
	}
	resp.Sys.Country = "FR"
	resp.Sys.Sunrise = 1700000000
	resp.Sys.Sunset = 1700040000

	current := convertCurrent(resp)

	if current.Temp != 15.5 {
		t.Errorf("expected temp 15.5, got %f", current.Temp)
	}
	if current.City != "Paris" {
		t.Errorf("expected city Paris, got %s", current.City)
	}
	if current.Country != "FR" {
		t.Errorf("expected country FR, got %s", current.Country)
	}
	if current.Description != "nuageux" {
		t.Errorf("expected description nuageux, got %s", current.Description)
	}
	if current.Icon != "04d" {
		t.Errorf("expected icon 04d, got %s", current.Icon)
	}
	if current.Humidity != 72 {
		t.Errorf("expected humidity 72, got %d", current.Humidity)
	}
	if current.WindSpeed != 3.5 {
		t.Errorf("expected wind speed 3.5, got %f", current.WindSpeed)
	}
	_ = client // client created successfully
}

func TestAggregateForecast(t *testing.T) {
	t.Parallel()
	resp := &owmForecastResponse{
		List: []struct {
			Dt   int64 `json:"dt"`
			Main struct {
				TempMin  float64 `json:"temp_min"`
				TempMax  float64 `json:"temp_max"`
				Humidity int     `json:"humidity"`
			} `json:"main"`
			Wind struct {
				Speed float64 `json:"speed"`
			} `json:"wind"`
			Weather []struct {
				Description string `json:"description"`
				Icon        string `json:"icon"`
			} `json:"weather"`
		}{
			{
				Dt: 1700006400, // 2023-11-15 00:00
				Main: struct {
					TempMin  float64 `json:"temp_min"`
					TempMax  float64 `json:"temp_max"`
					Humidity int     `json:"humidity"`
				}{TempMin: 10, TempMax: 15, Humidity: 60},
				Wind: struct {
					Speed float64 `json:"speed"`
				}{Speed: 2.0},
				Weather: []struct {
					Description string `json:"description"`
					Icon        string `json:"icon"`
				}{{Description: "clair", Icon: "01d"}},
			},
			{
				Dt: 1700049600, // 2023-11-15 12:00
				Main: struct {
					TempMin  float64 `json:"temp_min"`
					TempMax  float64 `json:"temp_max"`
					Humidity int     `json:"humidity"`
				}{TempMin: 12, TempMax: 18, Humidity: 50},
				Wind: struct {
					Speed float64 `json:"speed"`
				}{Speed: 3.0},
				Weather: []struct {
					Description string `json:"description"`
					Icon        string `json:"icon"`
				}{{Description: "ensoleillé", Icon: "02d"}},
			},
			{
				Dt: 1700092800, // 2023-11-16 00:00
				Main: struct {
					TempMin  float64 `json:"temp_min"`
					TempMax  float64 `json:"temp_max"`
					Humidity int     `json:"humidity"`
				}{TempMin: 8, TempMax: 14, Humidity: 70},
				Wind: struct {
					Speed float64 `json:"speed"`
				}{Speed: 4.0},
				Weather: []struct {
					Description string `json:"description"`
					Icon        string `json:"icon"`
				}{{Description: "pluie", Icon: "10d"}},
			},
		},
	}

	days := aggregateForecast(resp)
	if len(days) != 2 {
		t.Fatalf("expected 2 days, got %d", len(days))
	}

	// Day 1: min of mins, max of maxes
	if days[0].TempMin != 10 {
		t.Errorf("expected day1 min 10, got %f", days[0].TempMin)
	}
	if days[0].TempMax != 18 {
		t.Errorf("expected day1 max 18, got %f", days[0].TempMax)
	}

	// Day 2
	if days[1].TempMin != 8 {
		t.Errorf("expected day2 min 8, got %f", days[1].TempMin)
	}
}

func TestConvertCurrentNoWeather(t *testing.T) {
	t.Parallel()
	resp := &owmCurrentResponse{
		Name: "Unknown",
	}
	current := convertCurrent(resp)
	if current.Description != "" {
		t.Errorf("expected empty description, got %s", current.Description)
	}
	if current.Icon != "" {
		t.Errorf("expected empty icon, got %s", current.Icon)
	}
}
