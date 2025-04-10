package handler

import (
	"assignment_02/api"
	"encoding/json"
	"net/http"
	"time"
)

var HttpClient = http.DefaultClient

// Handler for the status endpoint
// This function checks the status of various APIs and returns their status as a JSON response
func HandleStatus(w http.ResponseWriter, r *http.Request) {
	// Ping the API
	status := func(url string) int {
		resp, err := HttpClient.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			return 200
		}
		if resp != nil {
			defer resp.Body.Close()
		}
		return 500
	}

	countriesStatus := status(api.CountriesApiAll)
	meteoStatus := status(api.WeatherConditions)
	currencyStatus := status(api.CurrencyApiStatus)

	result := map[string]interface{}{
		"countries_api":   countriesStatus,
		"meteo_api":       meteoStatus,
		"currency_api":    currencyStatus,
		"notification_db": 200,
		"webhooks":        len(appCache.Webhooks),
		"version":         "v1",
		"uptime":          int(time.Since(startTime).Seconds()),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
