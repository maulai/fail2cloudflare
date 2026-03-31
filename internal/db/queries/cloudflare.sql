-- name: ListPendingAdds :many
SELECT
  banned_ips.*,
  (
    SELECT be.comment
    FROM ban_entries be
    WHERE be.ip = banned_ips.ip
    ORDER BY be.created_at DESC
    LIMIT 1
  ) AS comment
FROM banned_ips
WHERE sync_state = 'pending_add'
  AND ban_count > 0
ORDER BY updated_at ASC
LIMIT sqlc.arg(limit_rows);

-- name: ListPendingDeletes :many
SELECT *
FROM banned_ips
WHERE sync_state = 'pending_delete'
  AND ban_count = 0
  AND cf_item_id IS NOT NULL
  AND cf_item_id != ''
ORDER BY updated_at ASC
LIMIT sqlc.arg(limit_rows);

-- name: MarkSynced :exec
UPDATE banned_ips
SET
  sync_state = 'synced',
  cf_item_id = sqlc.arg(cf_item_id),
  updated_at = CURRENT_TIMESTAMP
WHERE ip = sqlc.arg(ip)
  AND ban_count > 0;

-- name: DeleteAfterRemoteDelete :exec
DELETE FROM banned_ips
WHERE ip = sqlc.arg(ip)
  AND ban_count = 0
  AND sync_state = 'pending_delete';