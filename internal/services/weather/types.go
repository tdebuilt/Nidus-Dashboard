package weather

// CurrentWeather holds current weather data.
type CurrentWeather struct {
	Temp        float64 `json:"temp"`
	FeelsLike   float64 `json:"feels_like"`
	TempMin     float64 `json:"temp_min"`
	TempMax     float64 `json:"temp_max"`
	Humidity    int     `json:"humidity"`
	Pressure    int     `json:"pressure"`
	WindSpeed   float64 `json:"wind_speed"`
	WindDeg     int     `json:"wind_deg"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Sunrise     int64   `json:"sunrise"`
	Sunset      int64   `json:"sunset"`
}

// ForecastDay holds daily forecast data.
type ForecastDay struct {
	Date        string  `json:"date"`
	TempMin     float64 `json:"temp_min"`
	TempMax     float64 `json:"temp_max"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"wind_speed"`
}

// WeatherData combines current weather with forecast.
type WeatherData struct {
	Current  CurrentWeather `json:"current"`
	Forecast []ForecastDay  `json:"forecast"`
}

// OpenWeatherMap API response structures (internal)
type owmCurrentResponse struct {
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Humidity  int     `json:"humidity"`
		Pressure  int     `json:"pressure"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Sys struct {
		Country string `json:"country"`
		Sunrise int64  `json:"sunrise"`
		Sunset  int64  `json:"sunset"`
	} `json:"sys"`
	Name string `json:"name"`
}

type owmForecastResponse struct {
	List []struct {
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
	} `json:"list"`
}
