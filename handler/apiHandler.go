package handler

import (
	"assignment_02/api"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// --------------------------
// Allowed API Helper Functions
// --------------------------

func fetchCountryDetails(query string) (name string, capital string, lat float64, lon float64, iso string, currencyCode string, population float64, area float64, err error) {
	trimmed := strings.TrimSpace(query)
	var reqURL string
	if len(trimmed) == 2 {
		reqURL = fmt.Sprintf("%s%s?fullText=true", api.CountriesAPIIso, strings.ToLower(trimmed))
	} else {
		reqURL = fmt.Sprintf("%s%s?fullText=true", api.CountriesApi, strings.ToLower(trimmed))
	}

	resp, err := http.Get(reqURL)
	if err != nil {
		err = fmt.Errorf("error calling countries API: %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("countries API returned status %d", resp.StatusCode)
		return
	}

	var results []struct {
		Name struct {
			Common string `json:"common"`
		} `json:"name"`
		Capital    []string `json:"capital"`
		Cca2       string   `json:"cca2"`
		Currencies map[string]struct {
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
		} `json:"currencies"`
		Latlng     []float64 `json:"latlng"`
		Population float64   `json:"population"`
		Area       float64   `json:"area"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&results); err != nil {
		err = fmt.Errorf("error decoding countries API response: %w", err)
		return
	}

	if len(results) < 1 {
		err = fmt.Errorf("no results found in countries API")
		return
	}

	res := results[0]
	name = res.Name.Common
	iso = res.Cca2
	if len(res.Capital) > 0 {
		capital = res.Capital[0]
	}
	if len(res.Latlng) >= 2 {
		lat = res.Latlng[0]
		lon = res.Latlng[1]
	}
	for k := range res.Currencies {
		currencyCode = k
		break
	}
	population = res.Population
	area = res.Area
	return
}

func fetchCurrencyRates(currency string) (map[string]float64, error) {
	url := api.CurrencyApi + currency
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error calling currency API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("currency API returned status %d", resp.StatusCode)
	}
	var data struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding currency API response: %w", err)
	}
	return data.Rates, nil
}

type GeoResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Name      string  `json:"name"`
	} `json:"results"`
}

func getWeather(city string) (float64, float64, error) {
	geoURL := api.WeatherCoordinates + url.QueryEscape(city) + api.CountShow
	geoResp, err := http.Get(geoURL)
	if err != nil {
		return 0, 0, fmt.Errorf("error calling geocoding API: %w", err)
	}
	defer geoResp.Body.Close()
	if geoResp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("geocoding API returned status %d", geoResp.StatusCode)
	}
	var geoData GeoResponse
	if err := json.NewDecoder(geoResp.Body).Decode(&geoData); err != nil {
		return 0, 0, fmt.Errorf("error decoding geocoding API response: %w", err)
	}
	if len(geoData.Results) < 1 {
		return 0, 0, fmt.Errorf("no geocoding results found")
	}
	lat := geoData.Results[0].Latitude
	lon := geoData.Results[0].Longitude

	weatherURL := fmt.Sprintf("%slatitude=%f&longitude=%f%s", api.WeatherConditions, lat, lon, api.WeatherShow)
	weatherResp, err := http.Get(weatherURL)
	if err != nil {
		return 0, 0, fmt.Errorf("error calling weather API: %w", err)
	}
	defer weatherResp.Body.Close()
	if weatherResp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("weather API returned status %d", weatherResp.StatusCode)
	}
	var weatherData struct {
		Hourly struct {
			Temperature2m []float64 `json:"temperature_2m"`
			Precipitation []float64 `json:"precipitation"`
		} `json:"hourly"`
	}
	if err := json.NewDecoder(weatherResp.Body).Decode(&weatherData); err != nil {
		return 0, 0, fmt.Errorf("error decoding weather API response: %w", err)
	}
	temp := -99.0
	precip := -99.0
	if len(weatherData.Hourly.Temperature2m) > 0 {
		temp = weatherData.Hourly.Temperature2m[0]
	}
	if len(weatherData.Hourly.Precipitation) > 0 {
		precip = weatherData.Hourly.Precipitation[0]
	}
	return temp, precip, nil
}
