package cli

import (
	"fmt"
	"os"
)

type Mode string

const (
	ModeWorker  Mode = "worker"
	ModeBanIP   Mode = "banip"
	ModeUnbanIP Mode = "unbanip"
	ModeMigrate Mode = "migrate"
)

type Parsed struct {
	Mode Mode
	Args []string
}

func Parse(args []string) (Parsed, error) {
	if len(args) < 2 {
		return Parsed{}, fmt.Errorf("missing command")
	}

	switch args[1] {
	case "worker":
		return Parsed{
			Mode: ModeWorker,
			Args: args[2:],
		}, nil

	case "banip":
		return Parsed{
			Mode: ModeBanIP,
			Args: args[2:],
		}, nil

	case "unbanip":
		return Parsed{
			Mode: ModeUnbanIP,
			Args: args[2:],
		}, nil

	case "migrate":
		return Parsed{
			Mode: ModeMigrate,
			Args: args[2:],
		}, nil

	default:
		return Parsed{}, fmt.Errorf("unknown command: %s", args[1])
	}
}

func PrintUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  fail2cloudflare worker
  fail2cloudflare banip <ip> <jail> [comment]
  fail2cloudflare unbanip <ip> <jail>
  fail2cloudflare migrate`)
}
