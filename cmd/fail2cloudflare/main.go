package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/maulai/fail2cloudflare/internal/action"
	"github.com/maulai/fail2cloudflare/internal/cli"
	"github.com/maulai/fail2cloudflare/internal/cloudflare"
	"github.com/maulai/fail2cloudflare/internal/config"
	"github.com/maulai/fail2cloudflare/internal/db"
	sqlc "github.com/maulai/fail2cloudflare/internal/db/generated"
	"github.com/maulai/fail2cloudflare/internal/logging"
	"github.com/maulai/fail2cloudflare/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	isDev, args := extractDevFlag(os.Args)
	if isDev {
		if err := godotenv.Load(".env"); err != nil {
			panic(err)
		}
	}

	cfg, err := config.NewAppConfig()
	if err != nil {
		panic(err)
	}

	logger := logging.New(logging.Config{
		FilePath: cfg.LogPath,
		Format:   logging.FormatJSON,
	})

	parsed, err := cli.Parse(args)
	if err != nil {
		cli.PrintUsage()
		logger.Error("", err)
		return
	}

	sqlitedb, err := db.Init(cfg.DBPath)
	if err != nil {
		logger.Error("could not initialize database", err)
		return
	}
	defer sqlitedb.Close()

	q := sqlc.New(sqlitedb)

	switch parsed.Mode {
	case cli.ModeBanIP:
		cfg, err := action.ParseBanArgs(parsed.Args)
		if err != nil {
			logger.Error("could not parse ban args", err)
			return
		}
		logger.Info(fmt.Sprintf("Banning IP: %s, Jail: %s...", cfg.IP, cfg.Jail))
		if err := action.RunBanIP(ctx, cfg, q); err != nil {
			logger.Error("could not run ban action", err)
			return
		}

	case cli.ModeUnbanIP:
		cfg, err := action.ParseUnbanArgs(parsed.Args)
		if err != nil {
			logger.Error("could not parse unban args", err)
			return
		}
		logger.Info(fmt.Sprintf("Unbanning IP: %s, Jail: %s...", cfg.IP, cfg.Jail))
		if err := action.RunUnbanIP(ctx, cfg, q); err != nil {
			logger.Error("could not run unban action", err)
			return
		}

	case cli.ModeMigrate:
		if err := db.Migrate(sqlitedb); err != nil {
			logger.Error("could not migrate database", err)
			return
		}

	case cli.ModeWorker:
		logger.Info("Started worker")
		cf := cloudflare.New(cloudflare.Config{
			APIToken:  cfg.ApiToken,
			AccountID: cfg.AccountID,
			ListID:    cfg.ListID,
		})
		w := worker.New(q, cf, logger, worker.Config{})
		if err := w.Run(ctx); err != nil {
			logger.Error("worker error", err)
			return
		}
	}
}

func extractDevFlag(args []string) (bool, []string) {
	out := make([]string, 0, len(args))
	isDev := false

	for _, arg := range args {
		if arg == "--dev" {
			isDev = true
			continue
		}
		out = append(out, arg)
	}

	return isDev, out
}
