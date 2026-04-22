package testapi

import (
	"context"
	"encoding/json"
	"strings"
)

type MockQuerier struct{}

func (m *MockQuerier) Query(ctx context.Context, query string, variables map[string]any, result any) error {
	if strings.Contains(query, "locations") || strings.Contains(query, "regions") {
		_, locations, err := loadTestData()
		if err != nil {
			return err
		}
		locs := make([]map[string]any, len(locations))
		for i, loc := range locations {
			locs[i] = map[string]any{"name": loc.Name, "region": loc.Region}
		}
		resp := map[string]any{
			"data": map[string]any{
				"locations": locs,
			},
		}
		respBytes, err := json.Marshal(resp)
		if err != nil {
			return err
		}
		var gqlResp struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(respBytes, &gqlResp); err != nil {
			return err
		}
		return json.Unmarshal(gqlResp.Data, result)
	}

	servers, _, err := loadTestData()
	if err != nil {
		return err
	}

	if input, ok := variables["input"].(map[string]any); ok {
		if filter, ok := input["filter"].(map[string]any); ok {
			convertedFilter := make(map[string]any)
			for k, v := range filter {
				if strSlice, ok := v.([]string); ok {
					anySlice := make([]any, len(strSlice))
					for i, s := range strSlice {
						anySlice[i] = s
					}
					convertedFilter[k] = anySlice
				} else {
					convertedFilter[k] = v
				}
			}
			servers = filterServers(servers, convertedFilter)
		}
	}

	resp := map[string]any{
		"data": map[string]any{
			"servers": serversToGraphQLResponse(servers),
		},
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	var gqlResp struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(respBytes, &gqlResp); err != nil {
		return err
	}

	return json.Unmarshal(gqlResp.Data, result)
}

func serversToGraphQLResponse(servers []entry) map[string]any {
	entries := make([]map[string]any, len(servers))
	for i, s := range servers {
		var tags []map[string]any
		for _, t := range s.Tags {
			tags = append(tags, map[string]any{"key": t.Key, "value": t.Value})
		}

		var cpus []map[string]any
		for _, c := range s.Hardware.CPUs {
			cpus = append(cpus, map[string]any{"name": c.Name, "cores": c.Cores})
		}

		var storage []map[string]any
		for _, st := range s.Hardware.Storage {
			storage = append(storage, map[string]any{"size": st.Size, "type": st.Type})
		}

		var rams []map[string]any
		for _, r := range s.Hardware.Rams {
			rams = append(rams, map[string]any{"size": r.Size})
		}

		var ipAddresses []map[string]any
		for _, ip := range s.Network.IPAddresses {
			ipAddresses = append(ipAddresses, map[string]any{
				"ip":          ip.IP,
				"isPrimary":   ip.IsPrimary,
				"type":        ip.Type,
				"isBgpPrefix": ip.IsBgpPrefix,
			})
		}

		var ipmi map[string]any
		if s.Network.IPMI.IP != "" || s.Network.IPMI.Username != "" {
			ipmi = map[string]any{"ip": s.Network.IPMI.IP, "username": s.Network.IPMI.Username}
		}

		entries[i] = map[string]any{
			"name":        s.Name,
			"alias":       s.Alias,
			"hostname":    s.Hostname,
			"uptime":      s.Uptime,
			"statusV2":    s.StatusV2,
			"powerStatus": s.PowerStatus,
			"location": map[string]any{
				"name":   s.Location.Name,
				"region": s.Location.Region,
			},
			"system": map[string]any{
				"operatingSystem": map[string]any{"name": s.System.OperatingSystem.Name},
				"raid":            s.System.Raid,
			},
			"hardware": map[string]any{
				"cpus":    cpus,
				"storage": storage,
				"rams":    rams,
			},
			"network": map[string]any{
				"ipAddresses":        ipAddresses,
				"uplinkCapacity":     s.Network.UplinkCapacity,
				"hasBgp":             s.Network.HasBGP,
				"hasLinkAggregation": s.Network.HasLinkAggregation,
				"ddosShieldLevel":    s.Network.DDOSShieldLevel,
				"ipmi":               ipmi,
			},
			"trafficPlan": map[string]any{
				"name": s.TrafficPlan.Name, "type": s.TrafficPlan.Type, "bandwidth": s.TrafficPlan.Bandwidth,
			},
			"billing": map[string]any{
				"subscriptionItem": map[string]any{
					"price":    s.Billing.SubscriptionItem.Price,
					"currency": s.Billing.SubscriptionItem.Currency,
					"subscriptionItemDetail": map[string]any{
						"server": map[string]any{"name": s.Billing.SubscriptionItem.SubscriptionItemDetail.Server.Name},
					},
				},
			},
			"tags": tags,
		}
	}

	return map[string]any{
		"entriesTotalCount": len(servers),
		"pageCount":         1,
		"isLastPage":        true,
		"entries":           entries,
	}
}
