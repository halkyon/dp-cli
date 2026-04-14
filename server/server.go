package server

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/halkyon/dp/api"
)

type Server struct {
	Name            string   `json:"name"`
	Alias           string   `json:"alias,omitempty"`
	IP              string   `json:"ip,omitempty"`
	IPType          string   `json:"ipType,omitempty"`
	AdditionalIPs   []string `json:"additionalIPs,omitempty"`
	Hostname        string   `json:"hostname,omitempty"`
	Location        string   `json:"location,omitempty"`
	Uptime          string   `json:"uptime,omitempty"`
	Price           float64  `json:"price,omitempty"`
	Currency        string   `json:"currency,omitempty"`
	BillingPeriod   string   `json:"billingPeriod,omitempty"`
	Status          string   `json:"status"`
	PowerStatus     string   `json:"powerStatus,omitempty"`
	OperatingSystem string   `json:"operatingSystem,omitempty"`
	CPU             string   `json:"cpu,omitempty"`
	Memory          string   `json:"memory,omitempty"`
	Storage         string   `json:"storage,omitempty"`
	Uplink          string   `json:"uplink,omitempty"`
	RAID            string   `json:"raid,omitempty"`
	TrafficPlan     string   `json:"trafficPlan,omitempty"`
	IPMIURL         string   `json:"ipmiUrl,omitempty"`
	IPMIUser        string   `json:"ipmiUser,omitempty"`
	BGP             bool     `json:"bgp,omitempty"`
	DDOS            string   `json:"ddos,omitempty"`
	LinkAggregation bool     `json:"linkAggregation,omitempty"`
	Tags            []string `json:"tags,omitempty"`
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
	Name        string          `json:"name"`
	Alias       string          `json:"alias"`
	Hostname    string          `json:"hostname"`
	Uptime      int             `json:"uptime"`
	StatusV2    string          `json:"statusV2"`
	PowerStatus string          `json:"powerStatus"`
	Location    locationInfo    `json:"location"`
	Network     networkInfo     `json:"network"`
	Billing     billingInfo     `json:"billing"`
	TrafficPlan trafficPlanInfo `json:"trafficPlan"`
	System      systemInfo      `json:"system"`
	Hardware    hardwareInfo    `json:"hardware"`
	Tags        []tagInfo       `json:"tags"`
}

type trafficPlanInfo struct {
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Bandwidth float64 `json:"bandwidth"`
}

type locationInfo struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

type tagInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type networkInfo struct {
	IPAddresses        []ipAddress `json:"ipAddresses"`
	UplinkCapacity     int         `json:"uplinkCapacity"`
	HasBGP             bool        `json:"hasBgp"`
	HasLinkAggregation bool        `json:"hasLinkAggregation"`
	DDOSShieldLevel    string      `json:"ddosShieldLevel"`
	IPMI               ipmiInfo    `json:"ipmi"`
}

type ipmiInfo struct {
	IP       string `json:"ip"`
	Username string `json:"username"`
}

type ipAddress struct {
	IP          string `json:"ip"`
	IsPrimary   bool   `json:"isPrimary"`
	Type        string `json:"type"`
	IsBgpPrefix bool   `json:"isBgpPrefix"`
}

type billingInfo struct {
	SubscriptionItem subscriptionItem `json:"subscriptionItem"`
}

type subscriptionItem struct {
	Price                  float64                 `json:"price"`
	Currency               string                  `json:"currency"`
	SubscriptionItemDetail *subscriptionItemDetail `json:"subscriptionItemDetail"`
}

type subscriptionItemDetail struct {
	Server *struct {
		Name string `json:"name"`
	} `json:"server"`
}

type systemInfo struct {
	OperatingSystem operatingSystemInfo `json:"operatingSystem"`
	Raid            string              `json:"raid"`
}

type hardwareInfo struct {
	CPUs    []cpuInfo     `json:"cpus"`
	Storage []storageInfo `json:"storage"`
	RAMs    []ramInfo     `json:"rams"`
}

type cpuInfo struct {
	Name  string `json:"name"`
	Cores int    `json:"cores"`
}

type storageInfo struct {
	Size int    `json:"size"`
	Type string `json:"type"`
}

type ramInfo struct {
	Size int `json:"size"`
}

type operatingSystemInfo struct {
	Name string `json:"name"`
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
			hostname
			uptime
			statusV2
			powerStatus
			location {
				name
				region
			}
			system {
				operatingSystem {
					name
				}
				raid
			}
			hardware {
				cpus {
					name
					cores
				}
				storage {
					size
					type
				}
				rams {
					size
				}
			}
			network {
				ipAddresses {
					ip
					isPrimary
					type
					isBgpPrefix
				}
				uplinkCapacity
				hasBgp
				hasLinkAggregation
				ddosShieldLevel
				ipmi {
					ip
					username
				}
			}
			trafficPlan {
				name
				type
				bandwidth
			}
			billing {
				subscriptionItem {
					price
					currency
					subscriptionItemDetail {
						server {
							name
						}
					}
				}
			}
		}
	}
}`

func filterResults(servers []Server, filterArg string) []Server {
	nameRegex := regexp.MustCompile("(?i)" + filterArg)
	aliasRegex := regexp.MustCompile("(?i)" + filterArg)

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
		ip, ipType, additionalIPs := extractIPsWithType(srv.Network.IPAddresses)

		var cpu string
		if len(srv.Hardware.CPUs) > 0 {
			cpu = srv.Hardware.CPUs[0].Name
		}

		var memory string
		var totalRAM int
		for _, r := range srv.Hardware.RAMs {
			totalRAM += r.Size
		}
		if totalRAM > 0 {
			memory = fmt.Sprintf("%d GB", totalRAM)
		}

		var storage string
		if len(srv.Hardware.Storage) > 0 {
			type storageKey struct {
				size int
				typ  string
			}
			counts := make(map[storageKey]int)
			for _, s := range srv.Hardware.Storage {
				counts[storageKey{size: s.Size, typ: s.Type}]++
			}

			var parts []string
			for key, count := range counts {
				if count == 1 {
					parts = append(parts, fmt.Sprintf("%d GB %s", key.size, key.typ))
				} else {
					parts = append(parts, fmt.Sprintf("%dx %d GB %s", count, key.size, key.typ))
				}
			}
			storage = strings.Join(parts, ", ")
		}

		var uplink string
		if srv.Network.UplinkCapacity > 0 {
			uplink = fmt.Sprintf("%d Gbps", srv.Network.UplinkCapacity)
		}

		var tags []string
		for _, t := range srv.Tags {
			tags = append(tags, t.Key+"="+t.Value)
		}

		var uptime string
		if srv.Uptime > 0 {
			days := srv.Uptime / 86400
			hours := (srv.Uptime % 86400) / 3600
			if days > 0 {
				uptime = fmt.Sprintf("%d days", days)
			} else {
				uptime = fmt.Sprintf("%d hours", hours)
			}
		}

		var ipmiURL string
		if srv.Network.IPMI.IP != "" {
			ipmiURL = "https://" + srv.Network.IPMI.IP
		}

		var trafficPlan string
		if srv.TrafficPlan.Name != "" {
			trafficPlan = srv.TrafficPlan.Name
			if srv.TrafficPlan.Type != "" {
				trafficPlan = fmt.Sprintf("%s (%s)", trafficPlan, srv.TrafficPlan.Type)
			}
			if srv.TrafficPlan.Bandwidth > 0 {
				trafficPlan = fmt.Sprintf("%s, %d Gbps", trafficPlan, int(srv.TrafficPlan.Bandwidth))
			}
		}

		var billingPeriod string
		if srv.Billing.SubscriptionItem.SubscriptionItemDetail != nil &&
			srv.Billing.SubscriptionItem.SubscriptionItemDetail.Server != nil {
			billingPeriod = "MONTHLY"
		}

		servers[i] = Server{
			Name:            srv.Name,
			Alias:           srv.Alias,
			IP:              ip,
			IPType:          ipType,
			AdditionalIPs:   additionalIPs,
			Hostname:        srv.Hostname,
			Location:        srv.Location.Name,
			Uptime:          uptime,
			Price:           srv.Billing.SubscriptionItem.Price,
			Currency:        srv.Billing.SubscriptionItem.Currency,
			BillingPeriod:   billingPeriod,
			Status:          srv.StatusV2,
			PowerStatus:     srv.PowerStatus,
			OperatingSystem: srv.System.OperatingSystem.Name,
			CPU:             cpu,
			Memory:          memory,
			Storage:         storage,
			Uplink:          uplink,
			RAID:            srv.System.Raid,
			TrafficPlan:     trafficPlan,
			IPMIURL:         ipmiURL,
			IPMIUser:        srv.Network.IPMI.Username,
			BGP:             srv.Network.HasBGP,
			DDOS:            srv.Network.DDOSShieldLevel,
			LinkAggregation: srv.Network.HasLinkAggregation,
			Tags:            tags,
		}
	}

	return servers
}

func extractIPsWithType(addresses []ipAddress) (string, string, []string) {
	var primary string
	var ipType string
	var additional []string

	for _, addr := range addresses {
		if addr.IsPrimary {
			primary = addr.IP
			ipType = addr.Type
		} else {
			additional = append(additional, addr.IP)
		}
	}

	if primary == "" && len(addresses) > 0 {
		primary = addresses[0].IP
		ipType = addresses[0].Type
		for _, addr := range addresses[1:] {
			additional = append(additional, addr.IP)
		}
	}

	return primary, ipType, additional
}
