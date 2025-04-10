package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// --------------------------
// Cache Persistence Functions
// --------------------------

func LoadCache() error {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return err
	}
	appCache.Lock()
	defer appCache.Unlock()
	return json.Unmarshal(data, &appCache)
}

func saveCache() error {
	appCache.RLock()
	defer appCache.RUnlock()
	data, err := json.MarshalIndent(appCache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cacheFile, data, 0644)
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
