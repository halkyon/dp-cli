package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/completion"
	"github.com/halkyon/dp/config"
	"github.com/halkyon/dp/server"
	"github.com/halkyon/dp/ssh"
)

var version = ""

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	switch os.Args[1] {
	case "show":
		return runShow(ctx)
	case "ssh":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: dp ssh <alias> [ssh flags...]")
		}
		return runSSH(ctx, os.Args[2:])
	case "completion":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: dp completion <bash|zsh|fish>")
		}
		return completion.Generate(completion.Shell(os.Args[2]))
	case "aliases":
		return completion.ListAliases(ctx)
	case "version":
		fmt.Println(getVersion())
		return nil
	case "-v", "--version":
		fmt.Println(getVersion())
		return nil
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [options]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  show [regex]       List servers (optional regex filter)\n")
	fmt.Fprintf(os.Stderr, "  ssh <alias>        SSH to server (alias or user@alias) [ssh flags...]\n")
	fmt.Fprintf(os.Stderr, "  completion <shell> Generate completion script (bash|zsh|fish)\n")
	fmt.Fprintf(os.Stderr, "  aliases            List all server aliases (for completion)\n")
	fmt.Fprintf(os.Stderr, "  version            Print version\n")
}

func getVersion() string {
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

func runShow(ctx context.Context) error {
	var opts server.Options

	for i := 2; i < len(os.Args); i++ {
		if !strings.HasPrefix(os.Args[i], "-") {
			opts.Filter = os.Args[i]
		}
	}

	apiKey, err := config.GetAPIKey()
	if err != nil {
		return err
	}
	if apiKey == "" {
		return api.ErrMissingAPIKey
	}

	client, err := api.NewClient(apiKey)
	if err != nil {
		return err
	}

	servers, err := server.FetchAll(ctx, client)
	if err != nil {
		return err
	}

	servers = server.Filter(servers, opts)

	output, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	fmt.Println(string(output))

	return nil
}

func runSSH(ctx context.Context, args []string) error {
	servers, err := server.FetchAll(ctx, getClient())
	if err != nil {
		return err
	}

	return ssh.Run(ctx, servers, args)
}

func getClient() *api.Client {
	apiKey, _ := config.GetAPIKey()
	client, _ := api.NewClient(apiKey)
	return client
}
