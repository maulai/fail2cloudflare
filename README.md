# fail2cloudflare

`fail2cloudflare` connects Fail2Ban with a Cloudflare Custom List.

## What problem does this solve?

When a website is behind the Cloudflare proxy, the origin server no longer sees the visitor’s IP address directly. Instead, proxied traffic reaches the origin from Cloudflare IP ranges. Cloudflare explicitly notes that all traffic to proxied DNS records passes through Cloudflare first, so the origin stops receiving traffic from individual visitor IPs and receives traffic from Cloudflare IPs instead.

That creates a practical problem for Fail2Ban. In a Cloudflare setup, Fail2Ban can only work correctly if your origin logs already contain the real client IP address instead of the proxy IP. Cloudflare documents that, by default, a Cloudflare IP is logged, and that the original visitor IP is provided in the `CF-Connecting-IP` header. You must therefore configure your web server or logging stack so that your access logs use the real visitor IP from `CF-Connecting-IP`. If you do not do that first, Fail2Ban may react to Cloudflare proxy IPs instead of the attacker’s real IP.

`fail2cloudflare` is meant for setups where that first step is already done correctly. Once your logs contain the real client IP, Fail2Ban can detect abusive visitors as usual. The next problem is enforcement: in a proxied setup, simply banning at the origin is often not enough, because traffic still arrives through Cloudflare. `fail2cloudflare` solves that by taking Fail2Ban ban and unban events and synchronizing them to a Cloudflare Custom List, so those decisions can be enforced where they matter in a Cloudflare-based architecture.

How it works:

- Fail2Ban calls the app whenever an IP is banned or unbanned
- the app stores that state locally in SQLite
- a background worker syncs the state to a Cloudflare Custom List
- the worker batches changes and retries on failure
- a reconcile step helps recover from crashes or temporary API issues

In short:

**Fail2Ban decides which IPs should be blocked, and `fail2cloudflare` makes sure that state is synchronized to Cloudflare in a robust way.**

---

## Usage

The application has four modes:

- `worker`
- `banip <ip> <jail> [comment]`
- `unbanip <ip> <jail>`
- `migrate`

### `banip`
Called by Fail2Ban when a new IP should be banned.
The IP is written to the local SQLite database.

### `unbanip`
Called by Fail2Ban when an IP should be removed.
The local database is updated accordingly.

### `worker`
Runs continuously in the background and syncs local state to Cloudflare.

### `migrate`
Applies the SQLite database migrations.

---

## Requirements

- Linux
- Fail2Ban
- Cloudflare Custom List
- Cloudflare WAF / custom rule that blocks requests from that list

If you want to build from source, you also need:

- Go

---

## Installation

### Option 1: Quick Install

You can install `fail2cloudflare` with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/maulai/fail2cloudflare/main/install.sh | sudo bash
```

This installer will:

- download and install the fail2cloudflare binary
- create the required directories
- create /etc/fail2cloudflare/fail2cloudflare.env
- install the systemd service for the worker

After the installation, edit the environment file and add your Cloudflare values:

```bash
sudo nano /etc/fail2cloudflare/fail2cloudflare.env
```

```env
CF_API_TOKEN=
CF_ACCOUNT_ID=
CF_LIST_ID=
```

Then run the database migrations and start the worker:

```bash
sudo /usr/local/bin/fail2cloudflare migrate
sudo systemctl daemon-reload
sudo systemctl enable --now fail2cloudflare.service
```

You can verify that the worker is running with:

```bash
sudo systemctl status fail2cloudflare.service
journalctl -u fail2cloudflare.service -f
```

### Option 2: Build from source

Build the binary:

```bash
go build -o fail2cloudflare ./cmd/fail2cloudflare
```

Install it:

```bash
sudo install -m 755 fail2cloudflare /usr/local/bin/fail2cloudflare
```

Create directories:

```bash
sudo mkdir -p /etc/fail2cloudflare
sudo mkdir -p /var/lib/fail2cloudflare
sudo mkdir -p /var/log/fail2cloudflare
```

Create the environment file:

```bash
sudo tee /etc/fail2cloudflare/fail2cloudflare.env > /dev/null <<'EOF_ENV'
CF_API_TOKEN=
CF_ACCOUNT_ID=
CF_LIST_ID=
EOF_ENV
```

Apply migrations:

```bash
sudo /usr/local/bin/fail2cloudflare migrate
```

A typical worker service looks like this:

```ini
[Unit]
Description=fail2cloudflare worker
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root
EnvironmentFile=/etc/fail2cloudflare/fail2cloudflare.env
WorkingDirectory=/var/lib/fail2cloudflare
ExecStart=/usr/local/bin/fail2cloudflare worker
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Save it as:

```bash
/etc/systemd/system/fail2cloudflare.service
```

Then enable it:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now fail2cloudflare.service
```

---

## Fail2Ban integration

In your Fail2Ban action file, you can use something like this:

```ini
actionban = /usr/local/bin/fail2cloudflare banip <ip> <name> "<matches>"
actionunban = /usr/local/bin/fail2cloudflare unbanip <ip> <name>
```

- `<ip>` is the banned IP address
- `<name>` is the Fail2Ban jail name
- `<matches>` can be used as a comment and later pushed to Cloudflare

---

## Cloudflare setup

Besides the Custom List itself, you still need a Cloudflare rule that blocks requests when the client IP is part of that list.

The idea is simply:

- if source IP is in the Custom List
- then block

---

## Development

For local development, the app expects a `.env` file in the **repository root directory**.

Example:

```env
CF_API_TOKEN=your_cloudflare_api_token
CF_ACCOUNT_ID=your_cloudflare_account_id
CF_LIST_ID=your_cloudflare_list_id
```

In development mode, you must append the `--dev` flag to your commands so the application loads the root `.env` file.

Examples:

```bash
go run ./cmd/fail2cloudflare migrate --dev
```

```bash
go run ./cmd/fail2cloudflare worker --dev
```

```bash
go run ./cmd/fail2cloudflare banip 1.2.3.4 sshd "test comment" --dev
```

```bash
go run ./cmd/fail2cloudflare unbanip 1.2.3.4 sshd --dev
```

---

## License

This project is licensed under the MIT License.
