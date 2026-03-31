-- name: ListSyncedIPs :many
SELECT *
FROM banned_ips
WHERE sync_state = 'synced'
  AND ban_count > 0
ORDER BY updated_at ASC
LIMIT sqlc.arg(limit_rows);

-- name: SetPendingAdd :exec
UPDATE banned_ips
SET
  sync_state = 'pending_add',
  cf_item_id = NULL,
  updated_at = CURRENT_TIMESTAMP
WHERE ip = sqlc.arg(ip)
  AND sync_state = 'synced'
  AND ban_count > 0;