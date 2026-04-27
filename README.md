# dp

CLI tool for interacting with DataPacket's GraphQL API to list and SSH into servers.

## Requirements

- Go 1.25+ (if building from source)
- DataPacket readonly API key (generate this at https://app.datapacket.com/settings/security)

## Installation

Download a pre-built binary from [releases](https://github.com/halkyon/dp-cli/releases) or build from source:

```bash
make build
```

## Configuration

The API key is loaded in the following order of priority:

1. `DATAPACKET_API_KEY` environment variable
2. `api_key` in config file

The output format is loaded in the following order of priority:

1. `-o` / `--output` flag
2. `output` in config file
3. default `json`

Config is loaded from two files:
- `~/.config/dp/config` - output, api_url, cache durations
- `~/.config/dp/credentials` - api_key

Example `~/.config/dp/config`:

```ini
output = table
api_url = https://api.example.com/graphql
aliases_cache = 24h
locations_cache = 168h
regions_cache = 168h
```

Example `~/.config/dp/credentials`:

```ini
api_key = your-api-key
```

## Commands

### servers

List servers with optional filters.

```bash
# List all servers
dp servers

# Output formats: json, table, csv, raw (default json)
dp servers -o table
dp servers -o csv
dp servers -o raw -q Name

# Shorthand flags
dp servers -o raw -q Name           # same as --output raw --query Name
dp servers -ow                       # same as --output-wide

# Combine with filters (flags after command)
dp servers -o csv --alias=dp-select-storage-slo-4
dp servers -o table --location=London

# Show more fields in table/csv
dp servers -o table -ow
dp servers -o csv --output-wide

# Filter by name (repeatable)
dp servers -n DP-12345 -n DP-67890
dp servers --name DP-12345 --name DP-67890

# Filter by alias (repeatable)
dp servers -a my-server -a prod-server
dp servers --alias my-server --alias prod-server

# Filter by location (repeatable)
dp servers -l Amsterdam -l Berlin
dp servers --location Amsterdam --location Berlin

# Filter by region (repeatable)
dp servers -r EU -r NA
dp servers --region EU --region NA

# Filter by status (repeatable)
dp servers -s ACTIVE -s MAINTENANCE
dp servers --status ACTIVE --status MAINTENANCE

# Filter by power status (repeatable)
dp servers -p ON
dp servers --power ON

# Filter by tag (repeatable)
dp servers -t env=prod -t team=backend
dp servers --tag env=prod --tag team=backend

# Output specific field(s) (repeatable)
dp servers -q Name -q Alias -q IP
dp servers --query Name --query Alias --query IP

# Raw output for scripting/completions (newline-separated values)
dp servers -o raw -q Name           # outputs: DP-12345\nDP-67890
dp servers -o raw -q Name -q Alias # outputs: DP-12345 my-server\nDP-67890 prod-server

# Fuzzy filter using jq
dp servers | jq '[.[] | select(.alias) | select(.alias | test("dp-prod-edge-fra01-[0-9]"))]'
```

### ssh

SSH to server by alias.

```bash
# SSH using default user (root on linux, admin on windows)
dp ssh my-server

# SSH as different user
dp ssh root@my-server

# SSH with specific user
dp ssh --user=admin my-server
```

## Shell completion setup

After sourcing the completion script (see below), tab completion will work for:
- Commands (`servers`, `ssh`, `completion`)
- Alias names when using `-a` / `--alias`
- Location names when using `-l` / `--location`
- Region names when using `-r` / `--region`
- Power statuses when using `-p` / `--power`
- Server statuses when using `-s` / `--status`
- Server names when using `-n` / `--name`
- Tags when using `-t` / `--tag`
- Queryable fields when using `-q` / `--query`

### bash

```bash
# User-level (no root required)
mkdir -p ~/.local/share/bash-completion/completions
dp completion bash > ~/.local/share/bash-completion/completions/dp

# Or system-wide (requires root)
dp completion bash > /etc/bash_completion.d/dp
```

### zsh

```bash
# Add to ~/.zshrc first if not already present:
# fpath+=(~/.zsh/completions)
mkdir -p ~/.zsh/completions
dp completion zsh > ~/.zsh/completions/_dp
```

### fish

```bash
dp completion fish > ~/.config/fish/completions/dp.fish
```

## SDK usage

The dp packages can be imported for programmatic access to the DataPacket API.

### API client

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/halkyon/dp/api"
    "github.com/halkyon/dp/config"
    "github.com/halkyon/dp/server"
)

func main() {
    ctx := context.Background()

    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }
    if cfg.APIKey == "" {
        log.Fatal(api.ErrMissingAPIKey)
    }

    client, err := api.NewClient(cfg.APIKey)
    if err != nil {
        log.Fatal(err)
    }

    servers, err := server.List(ctx, client)
    if err != nil {
        log.Fatal(err)
    }

    for _, s := range servers {
        fmt.Printf("%s (%s) - %s\n", s.Name, s.Alias, s.IP)
    }
}
```

### List with filters

```go
servers, err := server.List(ctx, client,
    server.WithName("DP-12345", "DP-67890"),
    server.WithAlias("my-server", "prod-server"),
    server.WithLocation("Amsterdam", "Berlin"),
    server.WithRegion("EU", "NA"),
    server.WithStatus("ACTIVE", "MAINTENANCE"),
    server.WithPower("ON"),
    server.WithTag("env=prod", "team=backend"),
)
```
