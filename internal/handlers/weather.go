package handlers

import (
	"net/http"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/services/weather"
)

// WeatherHandler handles weather-related HTTP requests.
type WeatherHandler struct {
	Cache *cache.Cache
}

// GetWeather godoc
// @Summary Get current weather data
// @Description Returns weather data by city name or coordinates. API key and config come from query params.
// @Tags weather
// @Produce json
// @Param apikey query string true "OpenWeatherMap API key"
// @Param city query string false "City name"
// @Param lat query string false "Latitude"
// @Param lon query string false "Longitude"
// @Param units query string false "Units (metric, imperial, standard)"
// @Param lang query string false "Language code"
// @Success 200 {object} weather.WeatherData
// @Failure 400 {object} map[string]string
// @Failure 502 {object} map[string]string
// @Router /weather [get]
// @Security BearerAuth
func (h *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("apikey")
	if apiKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing apikey"})
		return
	}

	city := r.URL.Query().Get("city")
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")
	units := r.URL.Query().Get("units")
	lang := r.URL.Query().Get("lang")

	if city == "" && (lat == "" || lon == "") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing city or lat/lon"})
		return
	}

	// Cache key based on location
	cacheKey := "weather:" + city + lat + lon + units + lang
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client := weather.NewClient(apiKey, units, lang, nil)

	var data *weather.WeatherData
	var err error

	if city != "" {
		data, err = client.GetCurrentByCity(r.Context(), city)
	} else {
		data, err = client.GetWeather(r.Context(), lat, lon)
	}

	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to fetch weather: " + err.Error()})
		return
	}

	h.Cache.Set(cacheKey, data)
	writeJSON(w, http.StatusOK, data)
}
