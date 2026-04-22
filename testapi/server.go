package testapi

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

const defaultServerShutdownTimeout = 10 * time.Second

type Server struct {
	ln  net.Listener
	srv *http.Server
}

func NewServer() (*Server, error) {
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleGraphQL)

	srv := &http.Server{
		Addr:              ln.Addr().String(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Server{ln: ln, srv: srv}, nil
}

func (s *Server) Addr() string {
	return s.ln.Addr().String()
}

func (s *Server) Run(ctx context.Context) error {
	var g errgroup.Group
	g.Go(func() error {
		return s.srv.Serve(s.ln)
	})
	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(ctx, defaultServerShutdownTimeout)
		defer cancel()
		return s.srv.Shutdown(shutdownCtx)
	})
	return g.Wait()
}

func handleGraphQL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]any{"errors": []map[string]string{{"message": err.Error()}}}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	var req struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]any{"errors": []map[string]string{{"message": err.Error()}}}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := processQuery(req.Query, req.Variables)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func processQuery(query string, variables map[string]any) map[string]any {
	switch {
	case contains(query, "servers"):
		return map[string]any{"data": map[string]any{"servers": queryServers(variables)}}
	case contains(query, "locations"):
		return map[string]any{"data": map[string]any{"locations": queryLocations()}}
	case contains(query, "regions"):
		return map[string]any{"data": map[string]any{"locations": queryLocations()}}
	default:
		return map[string]any{"errors": []map[string]string{{"message": "unknown query"}}}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func queryServers(variables map[string]any) map[string]any {
	servers, _, err := loadTestData()
	if err != nil {
		return map[string]any{"errors": []map[string]string{{"message": err.Error()}}}
	}

	if input, ok := variables["input"].(map[string]any); ok {
		if filter, ok := input["filter"].(map[string]any); ok {
			servers = filterServers(servers, filter)
		}
	}

	entries := make([]map[string]any, len(servers))
	for i, s := range servers {
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
				"cpus":    s.Hardware.CPUs,
				"storage": s.Hardware.Storage,
				"rams":    s.Hardware.Rams,
			},
			"network": map[string]any{
				"ipAddresses":        s.Network.IPAddresses,
				"uplinkCapacity":     s.Network.UplinkCapacity,
				"hasBgp":             s.Network.HasBGP,
				"hasLinkAggregation": s.Network.HasLinkAggregation,
				"ddosShieldLevel":    s.Network.DDOSShieldLevel,
				"ipmi":               s.Network.IPMI,
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
			"tags": s.Tags,
		}
	}

	return map[string]any{
		"entriesTotalCount": len(servers),
		"pageCount":         1,
		"isLastPage":        true,
		"entries":           entries,
	}
}

func filterServers(servers []entry, filter map[string]any) []entry {
	var filtered []entry

	for _, s := range servers {
		match := true

		if nameIn, ok := filter["name_in"].([]any); ok {
			match = match && containsAny(nameIn, s.Name)
		}
		if aliasIn, ok := filter["alias_in"].([]any); ok {
			match = match && containsAny(aliasIn, s.Alias)
		}
		if locationIn, ok := filter["location_in"].([]any); ok {
			match = match && containsAny(locationIn, s.Location.Name)
		}
		if regionIn, ok := filter["region_in"].([]any); ok {
			match = match && containsAny(regionIn, s.Location.Region)
		}
		if statusIn, ok := filter["serverStatusV2_in"].([]any); ok {
			match = match && containsAny(statusIn, s.StatusV2)
		}
		if powerIn, ok := filter["powerStatus_in"].([]any); ok {
			match = match && containsAny(powerIn, s.PowerStatus)
		}

		if match {
			filtered = append(filtered, s)
		}
	}

	return filtered
}

func containsAny(list []any, value string) bool {
	for _, v := range list {
		if s, ok := v.(string); ok && s == value {
			return true
		}
	}
	return false
}

func queryLocations() []map[string]any {
	_, locations, err := loadTestData()
	if err != nil {
		return nil
	}

	result := make([]map[string]any, len(locations))
	for i, loc := range locations {
		result[i] = map[string]any{"name": loc.Name, "region": loc.Region}
	}
	return result
}
