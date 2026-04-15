package filters

import (
	"context"
	"time"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/cache"
)

type Locations struct {
	cache  *cache.Cache[[]string]
	client *api.Client
}

func NewLocations(client *api.Client, cacheDuration time.Duration) *Locations {
	return &Locations{
		cache:  cache.New[[]string]("locations", cacheDuration),
		client: client,
	}
}

func (l *Locations) Get(ctx context.Context) ([]string, error) {
	var locations []string
	if l.cache.Get(&locations) {
		return locations, nil
	}

	var data locationsData
	if err := l.client.Query(ctx, locationsQuery, nil, &data); err != nil {
		return nil, err
	}

	locations = make([]string, len(data.Locations))
	for i, loc := range data.Locations {
		locations[i] = loc.Name
	}

	l.cache.Set(locations, 0)
	return locations, nil
}

func (l *Locations) Clear() error {
	return l.cache.Clear()
}

type locationNode struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

type locationsData struct {
	Locations []locationNode `json:"locations"`
}

const locationsQuery = `query {
	locations {
		name
		region
	}
}`
