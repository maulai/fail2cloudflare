package worker

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/maulai/fail2cloudflare/internal/cloudflare"
	sqlc "github.com/maulai/fail2cloudflare/internal/db/generated"
)

func (w *Worker) processPendingAdds(ctx context.Context) error {
	rows, err := w.q.ListPendingAdds(ctx, int64(w.cfg.BatchSize))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	createItems := make([]cloudflare.CreateItemRequest, 0, len(rows))
	for _, row := range rows {
		if row.Ip == "" {
			continue
		}
		createItems = append(createItems, cloudflare.CreateItemRequest{
			IP:      row.Ip,
			Comment: row.Comment.String,
		})
	}

	if len(createItems) == 0 {
		return nil
	}

	w.logger.Info(fmt.Sprintf("Adding %d ips...", len(createItems)))

	if err := w.cf.AddIPsAndWait(ctx, createItems, 2*time.Second); err != nil {
		return err
	}

	items, err := w.cf.ListAllItems(ctx)
	if err != nil {
		return err
	}

	itemIDByIP := make(map[string]string, len(items))
	for _, item := range items {
		if item.IP == "" || item.ID == "" {
			continue
		}
		itemIDByIP[item.IP] = item.ID
	}

	for _, createItem := range createItems {
		itemID := itemIDByIP[createItem.IP]
		if itemID == "" {
			return fmt.Errorf("cloudflare item id for ip %s not found after add", createItem.IP)
		}

		err := w.q.MarkSynced(ctx, sqlc.MarkSyncedParams{
			Ip:       createItem.IP,
			CfItemID: sql.NullString{String: itemID, Valid: true},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) processPendingDeletes(ctx context.Context) error {
	rows, err := w.q.ListPendingDeletes(ctx, int64(w.cfg.BatchSize))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	itemIDs := make([]string, 0, len(rows))
	ipByItemID := make(map[string]string, len(rows))

	for _, row := range rows {
		if !row.CfItemID.Valid || row.CfItemID.String == "" || row.Ip == "" {
			continue
		}

		itemID := row.CfItemID.String
		itemIDs = append(itemIDs, itemID)
		ipByItemID[itemID] = row.Ip
	}

	if len(itemIDs) == 0 {
		return nil
	}

	w.logger.Info(fmt.Sprintf("Deleting %d ips...", len(itemIDs)))

	if err := w.cf.DeleteItemsAndWait(ctx, itemIDs, 2*time.Second); err != nil {
		return err
	}

	for _, itemID := range itemIDs {
		ip := ipByItemID[itemID]
		if ip == "" {
			continue
		}

		if err := w.q.DeleteAfterRemoteDelete(ctx, ip); err != nil {
			return err
		}
	}

	return nil
}
