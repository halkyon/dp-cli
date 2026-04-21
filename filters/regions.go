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

func NewRegions(client *api.Client, cacheDuration time.Duration, cacheDir string) (*Regions, error) {
	c, err := cache.New[[]string]("regions", cacheDuration, cacheDir)
	if err != nil {
		return nil, err
	}
	return &Regions{
		cache:  c,
		client: client,
	}, nil
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

	if err := r.cache.Set(regions, 0); err != nil {
		return nil, err
	}
	return regions, nil
}

func (r *Regions) Clear() error {
	return r.cache.Clear()
}
