package worker

import (
	"context"
	"database/sql"

	"github.com/maulai/fail2cloudflare/internal/cloudflare"
	sqlc "github.com/maulai/fail2cloudflare/internal/db/generated"
)

func (w *Worker) reconcile(ctx context.Context) error {
	items, err := w.cf.ListAllItems(ctx)
	if err != nil {
		return err
	}

	remoteByIP := make(map[string]cloudflare.ListItem, len(items))
	for _, item := range items {
		if item.IP == "" {
			continue
		}
		remoteByIP[item.IP] = item
	}

	if err := w.reconcilePendingAdds(ctx, remoteByIP); err != nil {
		return err
	}

	if err := w.reconcilePendingDeletes(ctx, remoteByIP); err != nil {
		return err
	}

	if err := w.reconcileSyncedMissingFromRemote(ctx, remoteByIP); err != nil {
		return err
	}

	return nil
}

func (w *Worker) reconcilePendingAdds(ctx context.Context, remoteByIP map[string]cloudflare.ListItem) error {
	rows, err := w.q.ListPendingAdds(ctx, int64(w.cfg.BatchSize))
	if err != nil {
		return err
	}

	for _, row := range rows {
		if row.Ip == "" {
			continue
		}

		item, ok := remoteByIP[row.Ip]
		if !ok || item.ID == "" {
			continue
		}

		if err := w.q.MarkSynced(ctx, sqlc.MarkSyncedParams{
			Ip:       row.Ip,
			CfItemID: sql.NullString{String: item.ID, Valid: true},
		}); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) reconcilePendingDeletes(ctx context.Context, remoteByIP map[string]cloudflare.ListItem) error {
	rows, err := w.q.ListPendingDeletes(ctx, int64(w.cfg.BatchSize))
	if err != nil {
		return err
	}

	for _, row := range rows {
		if row.Ip == "" {
			continue
		}

		if _, ok := remoteByIP[row.Ip]; ok {
			continue
		}

		if err := w.q.DeleteAfterRemoteDelete(ctx, row.Ip); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) reconcileSyncedMissingFromRemote(ctx context.Context, remoteByIP map[string]cloudflare.ListItem) error {
	rows, err := w.q.ListSyncedIPs(ctx, int64(w.cfg.BatchSize))
	if err != nil {
		return err
	}

	for _, row := range rows {
		if row.Ip == "" {
			continue
		}

		if _, ok := remoteByIP[row.Ip]; ok {
			continue
		}

		if err := w.q.SetPendingAdd(ctx, row.Ip); err != nil {
			return err
		}
	}

	return nil
}
