package ssh

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"dp/internal/api"
	"dp/internal/server"
)

func Run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dp ssh <alias> [ssh flags...]")
	}

	arg := args[0]
	user := "ubuntu"
	alias := arg

	if strings.Contains(arg, "@") {
		parts := strings.SplitN(arg, "@", 2)
		user = parts[0]
		alias = parts[1]
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

	var target *server.Server
	for i := range servers {
		if strings.EqualFold(servers[i].Alias, alias) {
			target = &servers[i]
			break
		}
	}

	if target == nil {
		return fmt.Errorf("no server found with alias %q", alias)
	}

	if target.IP == "" {
		return fmt.Errorf("server %q has no IP address", alias)
	}

	sshArgs := append([]string{fmt.Sprintf("%s@%s", user, target.IP)}, args[1:]...)
	sshCmd := exec.Command("ssh", sshArgs...)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	return sshCmd.Run()
}
