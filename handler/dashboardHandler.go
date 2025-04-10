package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) > 4 && pathParts[4] != "" {
		switch r.Method {
		case http.MethodGet:
			handleGetRegistration(w, r)
		case http.MethodPut:
			handleUpdateRegistration(w, r)
		case http.MethodDelete:
			handleDeleteRegistration(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	switch r.Method {
	case http.MethodPost:
		handleCreateRegistration(w, r)
	case http.MethodGet:
		handleListRegistrations(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleCreateRegistration(w http.ResponseWriter, r *http.Request) {
	var config DashboardConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(config.Country) == "" && strings.TrimSpace(config.ISOCode) != "" {
		name, _, _, _, iso, currency, _, _, err := fetchCountryDetails(config.ISOCode)
		if err == nil && name != "" {
			config.Country = name
			config.ISOCode = strings.ToUpper(iso)
		}
		if err == nil && currency != "" {
			config.Currency = strings.ToUpper(currency)
		}
	}
	if strings.TrimSpace(config.ISOCode) == "" && strings.TrimSpace(config.Country) != "" {
		name, _, _, _, iso, currency, _, _, err := fetchCountryDetails(config.Country)
		if err == nil && name != "" {
			config.Country = name
			config.ISOCode = strings.ToUpper(iso)
		}
		if err == nil && currency != "" {
			config.Currency = strings.ToUpper(currency)
		}
	}
	if strings.TrimSpace(config.ISOCode) != "" && strings.TrimSpace(config.Country) != "" {
		name, _, _, _, iso, currency, _, _, err := fetchCountryDetails(config.Country)
		if err == nil && name != "" {
			config.Country = name
			config.ISOCode = strings.ToUpper(iso)
		}
		if err == nil && currency != "" {
			config.Currency = strings.ToUpper(currency)
		}
	}
	config.ID = generateID()
	config.LastChange = time.Now().Format("20060102 15:04")
	appCache.Lock()
	appCache.Configs[config.ID] = config
	appCache.Unlock()
	if err := saveCache(); err != nil {
		log.Println("Error saving cache:", err)
	}

	// Trigger REGISTER webhook notifications.
	sendWebhookNotification("REGISTER", config.ISOCode)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(config)
}

func handleListRegistrations(w http.ResponseWriter, r *http.Request) {
	appCache.RLock()
	defer appCache.RUnlock()
	var configs []DashboardConfig
	for _, cfg := range appCache.Configs {
		configs = append(configs, cfg)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configs)
}

func handleGetRegistration(w http.ResponseWriter, r *http.Request) {
	// This just returns the raw stored config as is.
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[4] == "" {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}
	id := parts[4]
	appCache.RLock()
	config, exists := appCache.Configs[id]
	appCache.RUnlock()
	if !exists {
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func handleUpdateRegistration(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[4] == "" {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}
	id := parts[4]
	var updateData DashboardConfigUpdate
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	log.Printf("Update payload received for ID %s: %+v\n", id, updateData)
	appCache.Lock()
	existing, exists := appCache.Configs[id]
	if !exists {
		appCache.Unlock()
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}
	if updateData.Country != nil {
		if strings.TrimSpace(*updateData.Country) != "" {
			name, _, _, _, iso, currency, _, _, err := fetchCountryDetails(*updateData.Country)
			if err == nil {
				existing.Country = name
				existing.ISOCode = strings.ToUpper(iso)
				existing.Currency = strings.ToUpper(currency)
				log.Printf("Country updated via lookup: %s, ISO: %s, Currency: %s", name, iso, currency)
			} else {
				log.Printf("Warning: failed country lookup: %v", err)
				existing.Country = *updateData.Country
			}
		} else {
			existing.Country = ""
		}
	} else if updateData.ISOCode != nil && strings.TrimSpace(*updateData.ISOCode) != "" && updateData.Country == nil {
		name, _, _, _, iso, currency, _, _, err := fetchCountryDetails(*updateData.ISOCode)
		if err == nil {
			existing.Country = name
			existing.ISOCode = strings.ToUpper(iso)
			existing.Currency = strings.ToUpper(currency)
			log.Printf("ISO updated via lookup: %s, Country: %s, Currency: %s", iso, name, currency)
		} else {
			log.Printf("Warning: failed ISO lookup: %v", err)
			existing.ISOCode = *updateData.ISOCode
		}
	}
	if updateData.Currency != nil && strings.TrimSpace(*updateData.Currency) != "" {
		existing.Currency = strings.ToUpper(*updateData.Currency)
	}
	if updateData.Features != nil {
		if updateData.Features.Temperature != nil {
			existing.Features.Temperature = *updateData.Features.Temperature
		}
		if updateData.Features.Precipitation != nil {
			existing.Features.Precipitation = *updateData.Features.Precipitation
		}
		if updateData.Features.Capital != nil {
			existing.Features.Capital = *updateData.Features.Capital
		}
		if updateData.Features.Coordinates != nil {
			existing.Features.Coordinates = *updateData.Features.Coordinates
		}
		if updateData.Features.Population != nil {
			existing.Features.Population = *updateData.Features.Population
		}
		if updateData.Features.Area != nil {
			existing.Features.Area = *updateData.Features.Area
		}
		if updateData.Features.TargetCurrencies != nil {
			existing.Features.TargetCurrencies = *updateData.Features.TargetCurrencies
		}
	}
	existing.LastChange = time.Now().Format("20060102 15:04")
	appCache.Configs[id] = existing
	appCache.Unlock()
	log.Printf("Updated config for ID %s: %+v\n", id, existing)
	if err := saveCache(); err != nil {
		log.Println("Error saving cache:", err)
	}

	// Trigger CHANGE webhook notifications.
	sendWebhookNotification("CHANGE", existing.ISOCode)

	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteRegistration(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[4] == "" {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}
	id := parts[4]
	appCache.Lock()
	if _, exists := appCache.Configs[id]; !exists {
		appCache.Unlock()
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}
	config := appCache.Configs[id]
	delete(appCache.Configs, id)
	appCache.Unlock()
	if err := saveCache(); err != nil {
		log.Println("Error saving cache:", err)
	}

	// Trigger DELETE webhook notifications.
	sendWebhookNotification("DELETE", config.ISOCode)

	w.WriteHeader(http.StatusNoContent)
}

func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[4] == "" {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}
	id := parts[4]
	appCache.RLock()
	config, exists := appCache.Configs[id]
	appCache.RUnlock()
	if !exists {
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}
	lookupKey := config.Country
	if strings.TrimSpace(lookupKey) == "" {
		lookupKey = config.ISOCode
	}
	_, cap, lat, lon, _, _, population, area, err := fetchCountryDetails(lookupKey)
	if err != nil {
		log.Printf("Error fetching country details: %v", err)
		cap = "Unknown"
		lat, lon = 0, 0
	}
	rates, err := fetchCurrencyRates(config.Currency)
	if err != nil {
		log.Printf("Error fetching currency rates for base %s: %v", config.Currency, err)
		rates = map[string]float64{}
	}
	targetCurrencies := make(map[string]interface{})
	for _, cur := range config.Features.TargetCurrencies {
		if rate, ok := rates[cur]; ok {
			targetCurrencies[cur] = rate
		} else {
			targetCurrencies[cur] = 1.0
		}
	}

	var temperature, precipitation float64

	city := config.Country
	if city == "" {
		city = lookupKey
	}

	if config.Features.Temperature || config.Features.Precipitation {
		temp, precip, wErr := getWeather(city)
		if wErr != nil {
			log.Printf("Error fetching weather for %s: %v", city, wErr)
		} else {
			if config.Features.Temperature {
				temperature = temp
			}
			if config.Features.Precipitation {
				precipitation = precip
			}
		}
	}

	// Build the "populated" JSON with numeric placeholders from the real config + lookups.
	populated := map[string]interface{}{
		"country": config.Country,
		"isoCode": config.ISOCode,
		"features": map[string]interface{}{
			"temperature":      temperature,
			"precipitation":    precipitation,
			"capital":          cap, // A real city name if found
			"coordinates":      map[string]float64{"latitude": lat, "longitude": lon},
			"population":       population,
			"area":             area,
			"targetCurrencies": targetCurrencies,
		},
		"lastRetrieval": time.Now().Format("20060102 15:04"),
	}

	featuresMap := populated["features"].(map[string]interface{})

	if !config.Features.Capital {
		delete(featuresMap, "capital")
	}

	if !config.Features.Coordinates {
		delete(featuresMap, "coordinates")
	}

	if !config.Features.Population {
		delete(featuresMap, "population")
	}

	if !config.Features.Area {
		delete(featuresMap, "area")
	}

	if len(config.Features.TargetCurrencies) == 0 {
		delete(featuresMap, "targetCurrencies")
	}

	if !config.Features.Temperature {
		delete(featuresMap, "temperature")
	}

	if !config.Features.Precipitation {
		delete(featuresMap, "precipitation")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(populated)

	// Trigger INVOKE webhook notifications (done asynchronously so it does not block the response :3)
	go sendWebhookNotification("INVOKE", config.ISOCode)
}
