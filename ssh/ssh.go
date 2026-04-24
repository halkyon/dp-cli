package ssh

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/halkyon/dp/server"
)

func Run(ctx context.Context, servers []server.Server, username string, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: dp ssh <alias> [ssh flags...]")
	}

	alias := args[0]
	if strings.Contains(alias, "@") {
		parts := strings.SplitN(alias, "@", 2)
		username = parts[0]
		alias = parts[1]
	}

	var target *server.Server
	for i := range servers {
		if strings.EqualFold(servers[i].Alias, alias) || strings.EqualFold(servers[i].Name, alias) {
			target = &servers[i]
			break
		}
	}

	if target == nil {
		return fmt.Errorf("no server found with alias or name %q", alias)
	}

	if target.IP == "" {
		return fmt.Errorf("server %q has no IP address", alias)
	}

	if username == "" {
		username = defaultUsername(target.OperatingSystem)
	}

	sshArgs := append([]string{fmt.Sprintf("%s@%s", username, target.IP)}, args[1:]...)
	sshCmd := exec.CommandContext(ctx, "ssh", sshArgs...)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	return sshCmd.Run()
}

func defaultUsername(os string) string {
	os = strings.ToLower(os)
	if strings.Contains(os, "windows") {
		return "admin"
	}
	return "root"
}
