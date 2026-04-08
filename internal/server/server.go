package server

import (
	"context"
	"regexp"
	"sort"

	"dp/internal/api"
)

type Server struct {
	Name        string  `json:"name"`
	Alias       string  `json:"alias,omitempty"`
	IP          string  `json:"ip,omitempty"`
	Price       float64 `json:"price,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	Status      string  `json:"status"`
	PowerStatus string  `json:"powerStatus,omitempty"`
}

type Options struct {
	Filter string
}

func FetchAll(ctx context.Context, client *api.Client) ([]Server, error) {
	var allServers []serverNode

	pageIndex := 0
	pageSize := 50

	for {
		var data serversData
		if err := client.Query(ctx, serversQuery, map[string]any{
			"input": map[string]any{
				"pageIndex": pageIndex,
				"pageSize":  pageSize,
			},
		}, &data); err != nil {
			return nil, err
		}

		allServers = append(allServers, data.Servers.Entries...)

		if data.Servers.IsLastPage {
			break
		}

		pageIndex++
	}

	servers := convertServers(allServers)

	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Price > servers[j].Price
	})

	return servers, nil
}

func Filter(servers []Server, opts Options) []Server {
	if opts.Filter != "" {
		servers = filterResults(servers, opts.Filter)
	}

	return servers
}

type serverNode struct {
	Name        string      `json:"name"`
	Alias       string      `json:"alias"`
	Uptime      int         `json:"uptime"`
	StatusV2    string      `json:"statusV2"`
	PowerStatus string      `json:"powerStatus"`
	Network     networkInfo `json:"network"`
	Billing     billingInfo `json:"billing"`
}

type networkInfo struct {
	IPAddresses []ipAddress `json:"ipAddresses"`
}

type ipAddress struct {
	IP        string `json:"ip"`
	IsPrimary bool   `json:"isPrimary"`
}

type billingInfo struct {
	SubscriptionItem subscriptionItem `json:"subscriptionItem"`
}

type subscriptionItem struct {
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
}

type serversData struct {
	Servers struct {
		EntriesTotalCount int          `json:"entriesTotalCount"`
		PageCount         int          `json:"pageCount"`
		IsLastPage        bool         `json:"isLastPage"`
		Entries           []serverNode `json:"entries"`
	} `json:"servers"`
}

const serversQuery = `query($input: PaginatedServersInput) {
	servers(input: $input) {
		entriesTotalCount
		pageCount
		isLastPage
		entries {
			name
			alias
			uptime
			statusV2
			powerStatus
			network {
				ipAddresses {
					ip
					isPrimary
				}
			}
			billing {
				subscriptionItem {
					price
					currency
				}
			}
		}
	}
}`

func filterResults(servers []Server, filterArg string) []Server {
	nameRegex := regexp.MustCompile(filterArg)
	aliasRegex := regexp.MustCompile(filterArg)

	filtered := make([]Server, 0, len(servers))

	for _, srv := range servers {
		if nameRegex.MatchString(srv.Name) || (srv.Alias != "" && aliasRegex.MatchString(srv.Alias)) {
			filtered = append(filtered, srv)
		}
	}

	return filtered
}

func convertServers(allServers []serverNode) []Server {
	servers := make([]Server, len(allServers))

	for i, srv := range allServers {
		ip := extractPrimaryIP(srv.Network.IPAddresses)

		servers[i] = Server{
			Name:        srv.Name,
			Alias:       srv.Alias,
			IP:          ip,
			Price:       srv.Billing.SubscriptionItem.Price,
			Currency:    srv.Billing.SubscriptionItem.Currency,
			Status:      srv.StatusV2,
			PowerStatus: srv.PowerStatus,
		}
	}

	return servers
}

func extractPrimaryIP(addresses []ipAddress) string {
	for _, addr := range addresses {
		if addr.IsPrimary {
			return addr.IP
		}
	}

	if len(addresses) > 0 {
		return addresses[0].IP
	}

	return ""
}
