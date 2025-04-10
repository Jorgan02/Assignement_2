package handler

import (
	"sync"
	"time"
)

type Features struct {
	Temperature      bool     `json:"temperature"`
	Precipitation    bool     `json:"precipitation"`
	Capital          bool     `json:"capital"`
	Coordinates      bool     `json:"coordinates"`
	Population       bool     `json:"population"`
	Area             bool     `json:"area"`
	TargetCurrencies []string `json:"targetCurrencies"`
}

type DashboardConfig struct {
	ID         string   `json:"id"`
	Country    string   `json:"country"`  // Full country name.
	ISOCode    string   `json:"isoCode"`  // Two-letter country code.
	Currency   string   `json:"currency"` // Three-letter currency code.
	Features   Features `json:"features"`
	LastChange string   `json:"lastChange"`
}

type Webhook struct {
	ID      string `firebase:"id" json:"id"`
	URL     string `firebase:"url" json:"url"`
	Country string `firebase:"country" json:"country"`
	Event   string `firebase:"event" json:"event"`
}

type Cache struct {
	Configs  map[string]DashboardConfig `json:"configs"`
	Webhooks map[string]Webhook         `json:"webhooks"`
	sync.RWMutex
}

var appCache = Cache{
	Configs:  make(map[string]DashboardConfig),
	Webhooks: make(map[string]Webhook),
}

var startTime = time.Now()

const cacheFile = "stored-data/cache.json"

type FeaturesUpdate struct {
	Temperature      *bool     `json:"temperature,omitempty"`
	Precipitation    *bool     `json:"precipitation,omitempty"`
	Capital          *bool     `json:"capital,omitempty"`
	Coordinates      *bool     `json:"coordinates,omitempty"`
	Population       *bool     `json:"population,omitempty"`
	Area             *bool     `json:"area,omitempty"`
	TargetCurrencies *[]string `json:"targetCurrencies,omitempty"`
}

type DashboardConfigUpdate struct {
	Country  *string         `json:"country,omitempty"`
	ISOCode  *string         `json:"isoCode,omitempty"`
	Currency *string         `json:"currency,omitempty"`
	Features *FeaturesUpdate `json:"features,omitempty"`
}
