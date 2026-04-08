package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"

	"dp/internal/api"
	"dp/internal/completion"
	"dp/internal/server"
	"dp/internal/ssh"
)

var version = ""

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	if err := run(ctx); err != nil {
		stop()
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	switch os.Args[1] {
	case "show":
		return runShow(ctx)
	case "ssh":
		return ssh.Run(ctx, os.Args[2:])
	case "completion":
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
	if version != "" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return "unknown"
}

func runShow(ctx context.Context) error {
	opts := server.Options{}

	for i := 2; i < len(os.Args); i++ {
		if !strings.HasPrefix(os.Args[i], "-") {
			opts.Filter = os.Args[i]
		}
	}

	apiKey := os.Getenv("DATAPACKET_API_KEY")
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
