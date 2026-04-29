package cli

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/cache"
	"github.com/halkyon/dp/completion"
	"github.com/halkyon/dp/config"
	"github.com/halkyon/dp/filters"
	"github.com/halkyon/dp/internal/output"
	"github.com/halkyon/dp/server"
	"github.com/halkyon/dp/ssh"
)

var version = ""

func GetVersion() string {
	revision := "unknown"

	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				revision = setting.Value
				break
			}
		}
	}

	if version == "" {
		return revision
	}

	if revision == version {
		return version
	}

	return fmt.Sprintf("%s (%s)", version, revision)
}

type CLI struct {
	cfg    *config.Config
	client api.Querier
}

func New(cfg *config.Config, client api.Querier) *CLI {
	return &CLI{cfg: cfg, client: client}
}

func (c *CLI) ShowServers(ctx context.Context, opts server.Options, outputFormat string, wide bool) error {
	servers, err := server.List(ctx, c.client, opts.ToOpts()...)
	if err != nil {
		return err
	}

	switch outputFormat {
	case "json":
		encoded, err := output.PrintJSON(servers, opts.Fields)
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
	case "table":
		fmt.Println(output.PrintTable(servers, wide, opts.Fields))
	case "csv":
		w := csv.NewWriter(os.Stdout)
		if err := output.PrintCSV(w, servers, wide, opts.Fields); err != nil {
			return fmt.Errorf("writing CSV: %w", err)
		}
	case "raw":
		fmt.Print(output.PrintRaw(servers, opts.Fields))
	default:
		return fmt.Errorf("unknown output format: %s (use json, table, or csv)", outputFormat)
	}

	return nil
}

func (c *CLI) SSH(ctx context.Context, opts server.Options, sshUser string, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: dp ssh <alias> [ssh flags...]")
	}

	alias := args[0]
	if strings.Contains(alias, "@") {
		alias = strings.SplitN(alias, "@", 2)[1]
	}
	if alias != "" {
		opts.Alias = append(opts.Alias, alias)
	}

	opts.Fields = []string{"Name", "Alias", "IP", "OperatingSystem"}

	servers, err := server.List(ctx, c.client, opts.ToOpts()...)
	if err != nil {
		return err
	}

	return ssh.Run(ctx, servers, sshUser, args)
}

func GenerateCompletion(shell string) error {
	return completion.Generate(completion.Shell(shell))
}

func (c *CLI) Filter(ctx context.Context, filterType string) error {
	var filter interface {
		Get(context.Context) ([]string, error)
	}

	switch filterType {
	case "aliases":
		filter = filters.NewAliases(c.client, c.cfg.AliasesCache)
	case "locations":
		filter = filters.NewLocations(c.client, c.cfg.LocationsCache)
	case "regions":
		filter = filters.NewRegions(c.client, c.cfg.RegionsCache)
	case "power":
		filter = filters.NewPower()
	case "status":
		filter = filters.NewStatus()
	default:
		return fmt.Errorf("unknown filter type: %s", filterType)
	}

	var list []string
	var err error

	type cacheable interface {
		CacheDuration() time.Duration
		CacheKey() string
	}

	if ca, ok := filter.(cacheable); ok && ca.CacheDuration() > 0 {
		c, err := cache.New[[]string](ca.CacheKey(), ca.CacheDuration(), "")
		if err != nil {
			return err
		}
		if !c.Get(&list) {
			list, err = filter.Get(ctx)
			if err != nil {
				return err
			}
			if err = c.Set(list, 0); err != nil {
				return err
			}
		}
	} else {
		list, err = filter.Get(ctx)
		if err != nil {
			return err
		}
	}

	for _, item := range list {
		fmt.Println(item)
	}
	return nil
}

func Fields() {
	for _, f := range output.QueryableFields {
		fmt.Println(f)
	}
}
