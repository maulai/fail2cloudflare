package worker

import (
	"context"
	"time"

	"github.com/maulai/fail2cloudflare/internal/cloudflare"
	sqlc "github.com/maulai/fail2cloudflare/internal/db/generated"
	"github.com/maulai/fail2cloudflare/internal/logging"
)

type Config struct {
	BatchSize         int
	Debounce          time.Duration
	ReconcileInterval time.Duration
	RetryDelay        time.Duration
	PollInterval      time.Duration
}

func (c Config) withDefaults() Config {
	if c.BatchSize <= 0 {
		c.BatchSize = 100
	}
	if c.Debounce <= 0 {
		c.Debounce = 500 * time.Millisecond
	}
	if c.ReconcileInterval <= 0 {
		c.ReconcileInterval = 2 * time.Minute
	}
	if c.RetryDelay <= 0 {
		c.RetryDelay = 10 * time.Second
	}
	if c.PollInterval <= 0 {
		c.PollInterval = 1 * time.Second
	}
	return c
}

type Worker struct {
	q      *sqlc.Queries
	cf     *cloudflare.Client
	cfg    Config
	logger logging.Logger
}

func New(q *sqlc.Queries, cf *cloudflare.Client, logger logging.Logger, cfg Config) *Worker {
	cfg = cfg.withDefaults()

	return &Worker{
		q:      q,
		cf:     cf,
		cfg:    cfg,
		logger: logger,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	pollTicker := time.NewTicker(w.cfg.PollInterval)
	defer pollTicker.Stop()

	reconcileTicker := time.NewTicker(w.cfg.ReconcileInterval)
	defer reconcileTicker.Stop()

	for {
		if err := w.processPendingAdds(ctx); err != nil {
			if err := w.waitRetry(ctx, err); err != nil {
				return err
			}
			continue
		}

		if err := w.processPendingDeletes(ctx); err != nil {
			if err := w.waitRetry(ctx, err); err != nil {
				return err
			}
			continue
		}

		select {
		case <-ctx.Done():
			return nil
		case <-pollTicker.C:
		case <-reconcileTicker.C:
			if err := w.reconcile(ctx); err != nil {
				if err := w.waitRetry(ctx, err); err != nil {
					return err
				}
			}
		}
	}
}

func (w *Worker) waitRetry(ctx context.Context, cause error) error {
	w.logger.Error("cloudflare error, retrying...", cause)

	timer := time.NewTimer(w.cfg.RetryDelay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
