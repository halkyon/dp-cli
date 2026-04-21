package filters

import (
	"context"
	"time"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/cache"
	"github.com/halkyon/dp/server"
)

type Aliases struct {
	cache  *cache.Cache[[]string]
	client *api.Client
}

func NewAliases(client *api.Client, cacheDuration time.Duration, cacheDir string) (*Aliases, error) {
	c, err := cache.New[[]string]("aliases", cacheDuration, cacheDir)
	if err != nil {
		return nil, err
	}
	return &Aliases{
		cache:  c,
		client: client,
	}, nil
}

func (a *Aliases) Get(ctx context.Context) ([]string, error) {
	var aliases []string
	if a.cache.Get(&aliases) {
		return aliases, nil
	}

	servers, err := server.List(ctx, a.client)
	if err != nil {
		return nil, err
	}

	aliases = make([]string, 0, len(servers))
	for _, srv := range servers {
		if srv.Alias != "" {
			aliases = append(aliases, srv.Alias)
		}
	}

	if err := a.cache.Set(aliases, 0); err != nil {
		return nil, err
	}

	return aliases, nil
}
