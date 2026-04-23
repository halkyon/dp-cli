package server

import (
	"context"
	"fmt"
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
	Status          string   `json:"status,omitempty"`
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
	DeviceType      string   `json:"deviceType,omitempty"`
}

type Options struct {
	Name     []string
	Alias    []string
	Location []string
	Region   []string
	Status   []string
	Power    []string
	Tag      []string
	Fields   []string
}

type Option func(*Options)

func WithName(names ...string) Option {
	return func(o *Options) {
		o.Name = append(o.Name, names...)
	}
}

func WithAlias(aliases ...string) Option {
	return func(o *Options) {
		o.Alias = append(o.Alias, aliases...)
	}
}

func WithLocation(locs ...string) Option {
	return func(o *Options) {
		o.Location = append(o.Location, locs...)
	}
}

func WithRegion(regions ...string) Option {
	return func(o *Options) {
		o.Region = append(o.Region, regions...)
	}
}

func WithStatus(statuses ...string) Option {
	return func(o *Options) {
		o.Status = append(o.Status, statuses...)
	}
}

func WithPower(powers ...string) Option {
	return func(o *Options) {
		o.Power = append(o.Power, powers...)
	}
}

func WithTag(tags ...string) Option {
	return func(o *Options) {
		o.Tag = append(o.Tag, tags...)
	}
}

func WithFields(fields ...string) Option {
	return func(o *Options) {
		o.Fields = append(o.Fields, fields...)
	}
}

func (o Options) ToOpts() []Option {
	var opts []Option
	if len(o.Name) > 0 {
		opts = append(opts, WithName(o.Name...))
	}
	if len(o.Alias) > 0 {
		opts = append(opts, WithAlias(o.Alias...))
	}
	if len(o.Location) > 0 {
		opts = append(opts, WithLocation(o.Location...))
	}
	if len(o.Region) > 0 {
		opts = append(opts, WithRegion(o.Region...))
	}
	if len(o.Status) > 0 {
		opts = append(opts, WithStatus(o.Status...))
	}
	if len(o.Power) > 0 {
		opts = append(opts, WithPower(o.Power...))
	}
	if len(o.Tag) > 0 {
		opts = append(opts, WithTag(o.Tag...))
	}
	if len(o.Fields) > 0 {
		opts = append(opts, WithFields(o.Fields...))
	}
	return opts
}

func List(ctx context.Context, client api.Querier, opts ...Option) ([]Server, error) {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	var result []node

	pageIndex := 0
	pageSize := 50

	input := map[string]any{
		"pageIndex": pageIndex,
		"pageSize":  pageSize,
	}

	if len(options.Name) > 0 || len(options.Alias) > 0 || len(options.Location) > 0 || len(options.Region) > 0 ||
		len(options.Status) > 0 || len(options.Power) > 0 || len(options.Tag) > 0 {
		filter := make(map[string]any)
		if len(options.Name) > 0 {
			filter["name_in"] = options.Name
		}
		if len(options.Alias) > 0 {
			filter["alias_in"] = options.Alias
		}
		if len(options.Location) > 0 {
			filter["location_in"] = options.Location
		}
		if len(options.Region) > 0 {
			filter["region_in"] = options.Region
		}
		if len(options.Status) > 0 {
			filter["serverStatusV2_in"] = options.Status
		}
		if len(options.Power) > 0 {
			filter["powerStatus_in"] = options.Power
		}
		if len(options.Tag) > 0 {
			var tags []map[string]string
			for _, t := range options.Tag {
				parts := strings.SplitN(t, "=", 2)
				if len(parts) != 2 {
					return nil, fmt.Errorf("invalid tag filter %q: expected key=value", t)
				}
				tags = append(tags, map[string]string{"key": parts[0], "value": parts[1]})
			}
			filter["tags_in"] = tags
		}
		input["filter"] = filter
	}

	query := serversQuery
	if len(options.Fields) > 0 {
		query = buildQuery(options.Fields)
	}

	for {
		var data serversData
		if err := client.Query(ctx, query, map[string]any{
			"input": input,
		}, &data); err != nil {
			return nil, err
		}

		result = append(result, data.Servers.Entries...)

		if data.Servers.IsLastPage || pageIndex >= data.Servers.PageCount-1 {
			break
		}

		pageIndex++
		input["pageIndex"] = pageIndex
	}

	servers := convertServers(result)

	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Price > servers[j].Price
	})

	return servers, nil
}

type node struct {
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
	DeviceType  string          `json:"deviceType"`
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

type serverDetail struct {
	Name string `json:"name"`
}

type subscriptionItemDetail struct {
	Server *serverDetail `json:"server"`
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
		EntriesTotalCount int    `json:"entriesTotalCount"`
		PageCount         int    `json:"pageCount"`
		IsLastPage        bool   `json:"isLastPage"`
		Entries           []node `json:"entries"`
	} `json:"servers"`
}

// fieldBlocks maps each CLI output field to the top-level GraphQL block(s) it requires.
// Keys must match the case-insensitive names accepted by getFieldValue.
var fieldBlocks = map[string][]string{
	"name":            {"basic"},
	"alias":           {"basic"},
	"status":          {"basic"},
	"power":           {"basic"},
	"powerstatus":     {"basic"},
	"location":        {"location"},
	"ip":              {"network"},
	"iptype":          {"network"},
	"additionalips":   {"network"},
	"hostname":        {"basic"},
	"uptime":          {"basic"},
	"price":           {"billing"},
	"currency":        {"billing"},
	"billingperiod":   {"billing"},
	"os":              {"system"},
	"operatingsystem": {"system"},
	"cpu":             {"hardware"},
	"memory":          {"hardware"},
	"storage":         {"hardware"},
	"uplink":          {"network"},
	"raid":            {"system"},
	"trafficplan":     {"trafficPlan"},
	"ipmiurl":         {"network"},
	"ipmiuser":        {"network"},
	"bgp":             {"network"},
	"ddos":            {"network"},
	"linkaggregation": {"network"},
	"tags":            {"tags"},
	"devicetype":      {"basic"},
}

// allFieldBlocks lists every block in the order they should appear in the query.
var allFieldBlocks = []string{
	"basic", "location", "system", "hardware", "network", "trafficPlan", "billing", "tags",
}

var blockFragments = map[string]string{
	"basic":       "name alias hostname uptime statusV2 powerStatus deviceType",
	"location":    "location { name region }",
	"system":      "system { operatingSystem { name } raid }",
	"hardware":    "hardware { cpus { name cores } storage { size type } rams { size } }",
	"network":     "network { ipAddresses { ip isPrimary type } uplinkCapacity hasBgp hasLinkAggregation ddosShieldLevel ipmi { ip username } }",
	"trafficPlan": "trafficPlan { name type bandwidth }",
	"billing":     "billing { subscriptionItem { price currency subscriptionItemDetail { server { name } } } }",
	"tags":        "tags { key value }",
}

func buildQuery(fields []string) string {
	blockSet := make(map[string]struct{})
	for _, f := range fields {
		blocks, ok := fieldBlocks[strings.ToLower(f)]
		if !ok {
			continue
		}
		for _, b := range blocks {
			blockSet[b] = struct{}{}
		}
	}

	// If no recognized fields, include everything.
	includeAll := len(blockSet) == 0

	var fragments []string
	for _, name := range allFieldBlocks {
		if includeAll || func() bool { _, ok := blockSet[name]; return ok }() {
			fragments = append(fragments, blockFragments[name])
		}
	}

	return fmt.Sprintf(`query($input: PaginatedServersInput) {
	servers(input: $input) {
		entriesTotalCount
		pageCount
		isLastPage
		entries {
			%s
		}
	}
}`, strings.Join(fragments, "\n\t\t\t"))
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
		deviceType
	}
}`

func convertServers(items []node) []Server {
	servers := make([]Server, len(items))

	for i, srv := range items {
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

			var keys []storageKey
			for key := range counts {
				keys = append(keys, key)
			}
			sort.Slice(keys, func(i, j int) bool {
				if keys[i].typ != keys[j].typ {
					return keys[i].typ < keys[j].typ
				}
				return keys[i].size < keys[j].size
			})

			var parts []string
			for _, key := range keys {
				count := counts[key]
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
			DeviceType:      srv.DeviceType,
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
		additional = nil
		for _, addr := range addresses[1:] {
			additional = append(additional, addr.IP)
		}
	}

	return primary, ipType, additional
}
