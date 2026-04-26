package filters

import (
	"context"
	"time"

	"github.com/halkyon/dp/api"
)

type Locations struct {
	client        api.Querier
	cacheDuration time.Duration
}

func NewLocations(client api.Querier, cacheDuration time.Duration) *Locations {
	return &Locations{
		client:        client,
		cacheDuration: cacheDuration,
	}
}

func (l *Locations) CacheDuration() time.Duration { return l.cacheDuration }
func (l *Locations) CacheKey() string             { return "locations" }

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

func (l *Locations) Get(ctx context.Context) ([]string, error) {
	var data locationsData
	if err := l.client.Query(ctx, locationsQuery, nil, &data); err != nil {
		return nil, err
	}

	locations := make([]string, len(data.Locations))
	for i, loc := range data.Locations {
		locations[i] = loc.Name
	}

	return locations, nil
}
