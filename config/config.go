package config

import (
	"os"
	"time"

	"gopkg.in/ini.v1"
)

type Config struct {
	APIKey         string
	Output         string
	APIURL         string
	AliasesCache   time.Duration
	LocationsCache time.Duration
	RegionsCache   time.Duration
}

func getConfigPath() string {
	home := os.Getenv("HOME")
	if home == "" {
		return ""
	}
	return home + "/.config/dp/credentials"
}

func parseDuration(s string) time.Duration {
	s = s + "s"
	d, err := time.ParseDuration(s)
	if err == nil {
		return d
	}
	return time.Hour
}

func Load() (*Config, error) {
	var cfg Config

	path := getConfigPath()
	if path == "" {
		return &cfg, nil
	}

	data, err := ini.Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &cfg, nil
		}
		return nil, err
	}

	cfg.APIKey = data.Section("").Key("api_key").String()
	cfg.Output = data.Section("").Key("output").String()
	cfg.APIURL = data.Section("").Key("api_url").String()

	if aliases := data.Section("").Key("aliases_cache").String(); aliases != "" {
		cfg.AliasesCache = parseDuration(aliases)
	} else {
		cfg.AliasesCache = time.Hour
	}

	if locations := data.Section("").Key("locations_cache").String(); locations != "" {
		cfg.LocationsCache = parseDuration(locations)
	} else {
		cfg.LocationsCache = 7 * 24 * time.Hour
	}

	if regions := data.Section("").Key("regions_cache").String(); regions != "" {
		cfg.RegionsCache = parseDuration(regions)
	} else {
		cfg.RegionsCache = 7 * 24 * time.Hour
	}

	return &cfg, nil
}

func GetAPIKey() (string, error) {
	if key := os.Getenv("DATAPACKET_API_KEY"); key != "" {
		return key, nil
	}

	cfg, err := Load()
	if err != nil {
		return "", err
	}

	if cfg.APIKey == "" {
		return "", nil
	}

	return cfg.APIKey, nil
}

func GetOutput() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	return cfg.Output, nil
}

func GetAPIURL() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	return cfg.APIURL, nil
}

func GetAliasesCache() time.Duration {
	cfg, _ := Load()
	if cfg.AliasesCache > 0 {
		return cfg.AliasesCache
	}
	return time.Hour
}

func GetLocationsCache() time.Duration {
	cfg, _ := Load()
	if cfg.LocationsCache > 0 {
		return cfg.LocationsCache
	}
	return 7 * 24 * time.Hour
}

func GetRegionsCache() time.Duration {
	cfg, _ := Load()
	if cfg.RegionsCache > 0 {
		return cfg.RegionsCache
	}
	return 7 * 24 * time.Hour
}
