package testapi

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed testdata/servers.json
var serversData []byte

//go:embed testdata/locations.json
var locationsData []byte

type entry struct {
	Name        string `json:"name"`
	Alias       string `json:"alias"`
	Hostname    string `json:"hostname"`
	Uptime      int    `json:"uptime"`
	StatusV2    string `json:"statusV2"`
	PowerStatus string `json:"powerStatus"`
	Location    struct {
		Name   string `json:"name"`
		Region string `json:"region"`
	} `json:"location"`
	System struct {
		OperatingSystem struct {
			Name string `json:"name"`
		} `json:"operatingSystem"`
		Raid string `json:"raid"`
	} `json:"system"`
	Hardware struct {
		CPUs    []CPUInfo     `json:"cpus"`
		Storage []StorageInfo `json:"storage"`
		Rams    []RAMInfo     `json:"rams"`
	} `json:"hardware"`
	Network struct {
		IPAddresses        []IPAddress `json:"ipAddresses"`
		UplinkCapacity     int         `json:"uplinkCapacity"`
		HasBGP             bool        `json:"hasBgp"`
		HasLinkAggregation bool        `json:"hasLinkAggregation"`
		DDOSShieldLevel    string      `json:"ddosShieldLevel"`
		IPMI               IPMIInfo    `json:"ipmi"`
	} `json:"network"`
	TrafficPlan struct {
		Name      string  `json:"name"`
		Type      string  `json:"type"`
		Bandwidth float64 `json:"bandwidth"`
	} `json:"trafficPlan"`
	Billing struct {
		SubscriptionItem struct {
			Price                  float64 `json:"price"`
			Currency               string  `json:"currency"`
			SubscriptionItemDetail struct {
				Server struct {
					Name string `json:"name"`
				} `json:"server"`
			} `json:"subscriptionItemDetail"`
		} `json:"subscriptionItem"`
	} `json:"billing"`
	Tags []TagInfo `json:"tags"`
}

type CPUInfo struct {
	Name  string `json:"name"`
	Cores int    `json:"cores"`
}

type StorageInfo struct {
	Size int    `json:"size"`
	Type string `json:"type"`
}

type RAMInfo struct {
	Size int `json:"size"`
}

type IPAddress struct {
	IP          string `json:"ip"`
	IsPrimary   bool   `json:"isPrimary"`
	Type        string `json:"type"`
	IsBgpPrefix bool   `json:"isBgpPrefix"`
}

type IPMIInfo struct {
	IP       string `json:"ip"`
	Username string `json:"username"`
}

type TagInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type locationEntry struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

type serversResponse struct {
	Entries []entry `json:"entries"`
}

type locationsResponse []locationEntry

func loadTestData() ([]entry, []locationEntry, error) {
	var servers serversResponse
	if err := json.Unmarshal(serversData, &servers); err != nil {
		return nil, nil, fmt.Errorf("unmarshaling servers: %w", err)
	}

	var locations locationsResponse
	if err := json.Unmarshal(locationsData, &locations); err != nil {
		return nil, nil, fmt.Errorf("unmarshaling locations: %w", err)
	}

	return servers.Entries, locations, nil
}
