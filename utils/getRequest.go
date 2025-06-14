package utils

import (
	"encoding/json"
	"io"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	bodyCache  = map[*gin.Context]map[string]interface{}{}
	cacheMutex sync.RWMutex
)

// request mirip Laravel's request('key') untuk ambil data dari JSON body
func Request(c *gin.Context, key string) interface{} {
	cacheMutex.RLock()
	data, found := bodyCache[c]
	cacheMutex.RUnlock()

	if !found {
		// Baca body dan cache
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Println("Failed to read body:", err)
			return nil
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(body, &parsed); err != nil {
			log.Println("Failed to parse JSON:", err)
			return nil
		}

		cacheMutex.Lock()
		bodyCache[c] = parsed
		cacheMutex.Unlock()

		data = parsed

		// Set ulang body agar bisa dibaca ulang kalau dibutuhkan
		c.Request.Body = io.NopCloser(io.LimitReader(io.NopCloser(io.MultiReader()), int64(len(body))))
	}

	return data[key]
}
