package filters

import (
	"context"
	"fmt"
	"time"

	"github.com/halkyon/dp/api"
)

type Aliases struct {
	client        api.Querier
	cacheDuration time.Duration
}

func NewAliases(client api.Querier, cacheDuration time.Duration) *Aliases {
	return &Aliases{
		client:        client,
		cacheDuration: cacheDuration,
	}
}

func (a *Aliases) CacheDuration() time.Duration { return a.cacheDuration }
func (a *Aliases) CacheKey() string             { return "aliases" }

type serverAliasNode struct {
	Alias string `json:"alias"`
}

type serverAliasesData struct {
	Servers struct {
		IsLastPage bool              `json:"isLastPage"`
		Entries    []serverAliasNode `json:"entries"`
	} `json:"servers"`
}

const serverAliasesQuery = `query($input: PaginatedServersInput) {
	servers(input: $input) {
		isLastPage
		entries {
			alias
		}
	}
}`

const maxAliasPages = 1000

func (a *Aliases) Get(ctx context.Context) ([]string, error) {
	var aliases []string

	pageIndex := 0
	pageSize := 50

	input := map[string]any{
		"pageIndex": pageIndex,
		"pageSize":  pageSize,
	}

	for {
		var data serverAliasesData
		if err := a.client.Query(ctx, serverAliasesQuery, map[string]any{
			"input": input,
		}, &data); err != nil {
			return nil, err
		}

		for _, srv := range data.Servers.Entries {
			if srv.Alias != "" {
				aliases = append(aliases, srv.Alias)
			}
		}

		if data.Servers.IsLastPage {
			break
		}

		pageIndex++
		if pageIndex > maxAliasPages {
			return nil, fmt.Errorf("alias pagination exceeded maximum of %d pages", maxAliasPages)
		}
		input["pageIndex"] = pageIndex
	}

	return aliases, nil
}
