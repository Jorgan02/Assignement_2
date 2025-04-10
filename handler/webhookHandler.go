package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

func sendWebhookNotification(event, country string) {
	appCache.RLock()
	var webhooks []Webhook
	for _, wh := range appCache.Webhooks {
		// Sjekker om webhooken er abonnert på denne hendelsen
		// og om webhookens land er enten tomt (global) eller matcher et land.
		if strings.EqualFold(wh.Event, event) && (strings.TrimSpace(wh.Country) == "" || strings.EqualFold(wh.Country, country)) {
			webhooks = append(webhooks, wh)
		}
	}
	appCache.RUnlock()

	for _, wh := range webhooks {
		// Forbereder payloaden for avsending
		payload := map[string]string{
			"id":      wh.ID,
			"country": country,
			"event":   event,
			"time":    time.Now().Format("20060102 15:04"),
		}
		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Println("Error marshaling webhook payload:", err)
			continue
		}
		// Sender POST-kallet asynkront
		go func(wh Webhook, data []byte) {
			resp, err := http.Post(wh.URL, "application/json", bytes.NewBuffer(data))
			if err != nil {
				log.Println("Error sending webhook to", wh.URL, ":", err)
				return
			}
			resp.Body.Close()
		}(wh, jsonData)
	}
}

func NotificationHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) > 4 && pathParts[4] != "" {
		switch r.Method {
		case http.MethodGet:
			handleGetWebhook(w, r)
			addDocument(w, r)
		case http.MethodDelete:
			handleDeleteWebhook(w, r)
			deleteDocument(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	switch r.Method {
	case http.MethodPost:
		handleCreateWebhook(w, r)
	case http.MethodGet:
		handleListWebhooks(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
	var webhook Webhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	webhook.ID = generateID()
	appCache.Lock()
	appCache.Webhooks[webhook.ID] = webhook
	appCache.Unlock()
	if err := saveCache(); err != nil {
		log.Println("Error saving cache:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(webhook)
}

func handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	appCache.RLock()
	defer appCache.RUnlock()
	var webhooks []Webhook
	for _, wh := range appCache.Webhooks {
		webhooks = append(webhooks, wh)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhooks)
}

func handleGetWebhook(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[4] == "" {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}
	id := parts[4]
	appCache.RLock()
	webhook, exists := appCache.Webhooks[id]
	appCache.RUnlock()
	if !exists {
		http.Error(w, "Webhook not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhook)
}

func handleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[4] == "" {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}
	id := parts[4]
	appCache.Lock()
	if _, exists := appCache.Webhooks[id]; !exists {
		appCache.Unlock()
		http.Error(w, "Webhook not found", http.StatusNotFound)
		return
	}
	delete(appCache.Webhooks, id)
	appCache.Unlock()
	if err := saveCache(); err != nil {
		log.Println("Error saving cache:", err)
	}
	w.WriteHeader(http.StatusNoContent)
}
