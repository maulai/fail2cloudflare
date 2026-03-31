package action

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"

	sqlc "github.com/maulai/fail2cloudflare/internal/db/generated"
)

type BanConfig struct {
	IP      string
	Jail    string
	Comment string
}

func ParseBanArgs(args []string) (BanConfig, error) {
	if len(args) < 2 {
		return BanConfig{}, fmt.Errorf("IP and Jail must be specified")
	}
	if ip := net.ParseIP(strings.TrimSpace(args[0])); ip == nil {
		return BanConfig{}, fmt.Errorf("IP not valid")
	}
	if jail := strings.TrimSpace(args[1]); jail == "" {
		return BanConfig{}, fmt.Errorf("Jail must be specified")
	}
	var comment string
	if len(args) > 3 {
		comment = strings.Join(args[2:], " ")
	}
	return BanConfig{
		IP:      args[0],
		Jail:    strings.TrimSpace(args[1]),
		Comment: comment,
	}, nil
}

func RunBanIP(ctx context.Context, cfg BanConfig, q *sqlc.Queries) error {
	rows, err := q.InsertBanEntry(ctx, sqlc.InsertBanEntryParams{
		Ip:   cfg.IP,
		Jail: cfg.Jail,
		Comment: sql.NullString{
			String: cfg.Comment,
			Valid:  cfg.Comment != "",
		},
	})
	if err != nil {
		return err
	}
	if rows == 1 {
		err = q.UpsertBannedIPAfterNewBanEntry(ctx, cfg.IP)
		if err != nil {
			return err
		}
	}
	return nil
}
