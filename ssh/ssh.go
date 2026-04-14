package ssh

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/halkyon/dp/server"
)

func Run(ctx context.Context, servers []server.Server, username string, args []string, verbose bool) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dp ssh <alias> [ssh flags...]")
	}

	alias := args[0]
	if strings.Contains(alias, "@") {
		parts := strings.SplitN(alias, "@", 2)
		username = parts[0]
		alias = parts[1]
		if verbose {
			fmt.Fprintf(os.Stderr, "user from arg: %s\n", username)
		}
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

	if verbose {
		fmt.Fprintf(os.Stderr, "OS: %s\n", target.OperatingSystem)
	}

	if username == "" {
		username = defaultUsername(target.OperatingSystem)
		if verbose {
			fmt.Fprintf(os.Stderr, "default user from OS: %s\n", username)
		}
	}

	sshArgs := append([]string{fmt.Sprintf("%s@%s", username, target.IP)}, args[1:]...)
	if verbose {
		fmt.Fprintf(os.Stderr, "running: ssh %s\n", strings.Join(sshArgs, " "))
	}
	sshCmd := exec.Command("ssh", sshArgs...)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	return sshCmd.Run()
}

func defaultUsername(os string) string {
	os = strings.ToLower(os)
	if strings.Contains(os, "self") || os == "" {
		return "ubuntu"
	}
	return "root"
}
