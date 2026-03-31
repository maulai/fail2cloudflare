package action

import (
	"context"
	"fmt"
	"net"
	"strings"

	sqlc "github.com/maulai/fail2cloudflare/internal/db/generated"
)

type UnbanConfig struct {
	IP   string
	Jail string
}

func ParseUnbanArgs(args []string) (UnbanConfig, error) {
	if len(args) < 2 {
		return UnbanConfig{}, fmt.Errorf("IP and Jail must be specified")
	}
	if ip := net.ParseIP(strings.TrimSpace(args[0])); ip == nil {
		return UnbanConfig{}, fmt.Errorf("IP not valid")
	}
	if jail := strings.TrimSpace(args[1]); jail == "" {
		return UnbanConfig{}, fmt.Errorf("Jail must be specified")
	}
	return UnbanConfig{
		IP:   args[0],
		Jail: strings.TrimSpace(args[1]),
	}, nil
}

func RunUnbanIP(ctx context.Context, cfg UnbanConfig, q *sqlc.Queries) error {
	rows, err := q.DeleteBanEntry(ctx, sqlc.DeleteBanEntryParams{
		Ip:   cfg.IP,
		Jail: cfg.Jail,
	})
	if err != nil {
		return err
	}

	if rows == 1 {
		err = q.DecrementBannedIPAfterBanEntryDelete(ctx, cfg.IP)
		if err != nil {
			return err
		}
	}
	return nil
}
