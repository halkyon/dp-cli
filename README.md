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
2. `~/.config/dp/credentials` config file

Example config file:

```ini
api_key = your-api-key
```

## Commands

### show - Show server info

```bash
# List all servers (sorted by price)
dp show

# Filter by regex (matches name or alias)
dp show "dp-prod-edge-mia-.+"
```

### ssh - SSH to server

```bash
# SSH using ubuntu user (default)
dp ssh "dp-prod-edge-mia1-11"

# SSH as different user
dp ssh "root@dp-prod-edge-mia1-11"
```

### aliases - List all server aliases

```bash
dp aliases
```

## Shell completion setup

After sourcing the completion script (see below), tab completion will work for:
- Commands (show, ssh, completion, aliases)
- Alias names when using `show` or `ssh`

### zsh

```bash
dp completion zsh > ~/.zsh/completions/_dp
```

### bash

```bash
dp completion bash > /etc/bash_completion.d/dp
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

    apiKey, err := config.GetAPIKey()
    if err != nil {
        log.Fatal(err)
    }
    if apiKey == "" {
        log.Fatal(api.ErrMissingAPIKey)
    }

    client, err := api.NewClient(apiKey)
    if err != nil {
        log.Fatal(err)
    }

    servers, err := server.FetchAll(ctx, client)
    if err != nil {
        log.Fatal(err)
    }

    for _, s := range servers {
        fmt.Printf("%s (%s) - %s\n", s.Name, s.Alias, s.IP)
    }
}
```

### Filter servers

```go
import "github.com/halkyon/dp/server"

servers := server.Filter(servers, server.Options{Filter: "prod-.*"})
```

### Cached aliases

```go
import (
    "github.com/halkyon/dp/aliases"
    "github.com/halkyon/dp/api"
)

cache := aliases.New(client)
list, err := cache.Get(ctx)
```
