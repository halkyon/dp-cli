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

func New[T any](name string, maxAge time.Duration, cacheDir string) (*Cache[T], error) {
	if cacheDir == "" {
		cacheDir = filepath.Join(os.Getenv("HOME"), ".cache", "dp")
	}
	if err := os.MkdirAll(cacheDir, 0750); err != nil {
		return nil, err
	}

	return &Cache[T]{
		Path:   filepath.Join(cacheDir, name+".json"),
		MaxAge: maxAge,
	}, nil
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

func (c *Cache[T]) Set(data T, duration time.Duration) error {
	cached := cacheFile[T]{
		Expiry:   time.Now().Add(c.MaxAge),
		Duration: duration,
		Data:     data,
	}

	contents, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	return os.WriteFile(c.Path, contents, 0600)
}

func (c *Cache[T]) Clear() error {
	return os.RemoveAll(c.Path)
}
