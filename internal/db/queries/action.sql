-- name: InsertBanEntry :execrows
INSERT OR IGNORE INTO ban_entries (
  ip,
  jail,
  comment,
  created_at
) VALUES (
  sqlc.arg(ip),
  sqlc.arg(jail),
  sqlc.narg(comment),
  CURRENT_TIMESTAMP
);

-- name: UpsertBannedIPAfterNewBanEntry :exec
INSERT INTO banned_ips (
  ip,
  ban_count,
  sync_state,
  updated_at,
  created_at
) VALUES (
  sqlc.arg(ip),
  1,
  'pending_add',
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
ON CONFLICT(ip) DO UPDATE SET
  ban_count = banned_ips.ban_count + 1,
  sync_state = 'pending_add',
  updated_at = CURRENT_TIMESTAMP;

-- name: DeleteBanEntry :execrows
DELETE FROM ban_entries
WHERE ip = sqlc.arg(ip)
  AND jail = sqlc.arg(jail);

-- name: DecrementBannedIPAfterBanEntryDelete :exec
UPDATE banned_ips
SET
  ban_count = CASE
    WHEN ban_count > 0 THEN ban_count - 1
    ELSE 0
  END,
  sync_state = CASE
    WHEN ban_count <= 1 THEN 'pending_delete'
    ELSE sync_state
  END,
  updated_at = CURRENT_TIMESTAMP
WHERE ip = sqlc.arg(ip);
