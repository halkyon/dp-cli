package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Cache[T any] struct {
	Path     string
	MaxAge   time.Duration
	Duration time.Duration
}

type cacheFile[T any] struct {
	Expiry   time.Time     `json:"expiry"`
	Duration time.Duration `json:"duration"`
	Data     T             `json:"data"`
}

func New[T any](name string, maxAge time.Duration) *Cache[T] {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "dp")
	_ = os.MkdirAll(cacheDir, 0755)

	return &Cache[T]{
		Path:   filepath.Join(cacheDir, name+".json"),
		MaxAge: maxAge,
	}
}

func (c *Cache[T]) Get(data *T) bool {
	contents, err := os.ReadFile(c.Path)
	if err != nil {
		return false
	}

	var cached cacheFile[T]
	if err := json.Unmarshal(contents, &cached); err != nil {
		return false
	}

	if time.Now().After(cached.Expiry) {
		return false
	}

	*data = cached.Data
	return true
}

func (c *Cache[T]) Set(data T, duration time.Duration) {
	cached := cacheFile[T]{
		Expiry:   time.Now().Add(c.MaxAge),
		Duration: duration,
		Data:     data,
	}

	contents, err := json.Marshal(cached)
	if err != nil {
		return
	}

	_ = os.WriteFile(c.Path, contents, 0644)
}
