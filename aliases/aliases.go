package aliases

import (
	"context"
	"os"
	"time"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/cache"
	"github.com/halkyon/dp/server"
)

type AliasCache struct {
	cache  *cache.Cache[[]string]
	client *api.Client
}

func New(client *api.Client) *AliasCache {
	return &AliasCache{
		cache:  cache.New[[]string]("aliases", time.Hour),
		client: client,
	}
}

func (a *AliasCache) Get(ctx context.Context) ([]string, error) {
	var aliases []string
	if a.cache.Get(&aliases) {
		return aliases, nil
	}

	servers, err := server.FetchAll(ctx, a.client)
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

func (a *AliasCache) Clear() error {
	return os.Remove(cache.New[[]string]("aliases", time.Hour).Path)
}
