package filters

import (
	"context"
	"sort"
	"time"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/cache"
)

type Regions struct {
	cache  *cache.Cache[[]string]
	client *api.Client
}

func NewRegions(client *api.Client) *Regions {
	return &Regions{
		cache:  cache.New[[]string]("regions", time.Hour),
		client: client,
	}
}

func (r *Regions) Get(ctx context.Context) ([]string, error) {
	var regions []string
	if r.cache.Get(&regions) {
		return regions, nil
	}

	var data locationsData
	if err := r.client.Query(ctx, locationsQuery, nil, &data); err != nil {
		return nil, err
	}

	regionSet := make(map[string]struct{})
	for _, loc := range data.Locations {
		regionSet[loc.Region] = struct{}{}
	}

	regions = make([]string, 0, len(regionSet))
	for region := range regionSet {
		regions = append(regions, region)
	}
	sort.Strings(regions)

	r.cache.Set(regions, 0)
	return regions, nil
}

func (r *Regions) Clear() error {
	return r.cache.Clear()
}
