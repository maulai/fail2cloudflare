CREATE TABLE goose_db_version (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version_id INTEGER NOT NULL,
		is_applied INTEGER NOT NULL,
		tstamp TIMESTAMP DEFAULT (datetime('now'))
	);
CREATE TABLE sqlite_sequence(name,seq);
CREATE TABLE banned_ips (
  ip TEXT PRIMARY KEY,
  ban_count INTEGER NOT NULL DEFAULT 1 CHECK (ban_count >= 0),
  sync_state TEXT NOT NULL CHECK (
    sync_state IN ('pending_add', 'synced', 'pending_delete', 'error')
  ),
  cf_item_id TEXT,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE ban_entries (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ip TEXT NOT NULL,
  jail TEXT NOT NULL,
  comment TEXT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (ip) REFERENCES banned_ips(ip) ON DELETE CASCADE,
  UNIQUE (ip, jail)
);