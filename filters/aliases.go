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

func NewAliases(client *api.Client, cacheDuration time.Duration) *Aliases {
	return &Aliases{
		cache:  cache.New[[]string]("aliases", cacheDuration),
		client: client,
	}
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

	a.cache.Set(aliases, 0)

	return aliases, nil
}
