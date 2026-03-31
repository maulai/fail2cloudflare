#!/usr/bin/env bash
set -euo pipefail

APP_NAME="fail2cloudflare"
INSTALL_BIN="/usr/local/bin/${APP_NAME}"
CONFIG_DIR="/etc/fail2cloudflare"
ENV_FILE="${CONFIG_DIR}/fail2cloudflare.env"
DATA_DIR="/var/lib/fail2cloudflare"
LOG_DIR="/var/log/fail2cloudflare"
SERVICE_FILE="/etc/systemd/system/${APP_NAME}.service"
REPO="maulai/fail2cloudflare"
ASSET="fail2cloudflare_linux_amd64"

echo "==> Installing ${APP_NAME}"

if [[ $EUID -ne 0 ]]; then
  echo "Please run this script with sudo or as root."
  exit 1
fi

curl -fsSL -o /tmp/fail2cloudflare \
  "https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "==> Installing binary to ${INSTALL_BIN}"
install -m 755 "/tmp/${APP_NAME}" "${INSTALL_BIN}"
rm -f "/tmp/${APP_NAME}"

echo "==> Creating directories"
mkdir -p "${CONFIG_DIR}"
mkdir -p "${DATA_DIR}"
mkdir -p "${LOG_DIR}"

chmod 755 "${DATA_DIR}" "${LOG_DIR}"
chmod 755 "${CONFIG_DIR}"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "==> Creating environment file at ${ENV_FILE}"
  cat > "${ENV_FILE}" <<'EOF'
CF_API_TOKEN=
CF_ACCOUNT_ID=
CF_LIST_ID=
EOF
  chmod 600 "${ENV_FILE}"
else
  echo "==> Environment file already exists: ${ENV_FILE}"
fi

echo "==> Creating systemd service at ${SERVICE_FILE}"
cat > "${SERVICE_FILE}" <<EOF
[Unit]
Description=fail2cloudflare worker
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root
EnvironmentFile=${ENV_FILE}
WorkingDirectory=${DATA_DIR}
ExecStart=${INSTALL_BIN} worker
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

echo "==> Reloading systemd"
systemctl daemon-reload

echo
echo "Installation finished."
echo
echo "Next steps:"
echo "1. Edit ${ENV_FILE} and fill in:"
echo "   - CF_API_TOKEN"
echo "   - CF_ACCOUNT_ID"
echo "   - CF_LIST_ID"
echo
echo "2. Run migrations:"
echo "   ${INSTALL_BIN} migrate"
echo
echo "3. Start and enable the worker:"
echo "   systemctl enable --now ${APP_NAME}.service"
echo
echo "4. Check logs:"
echo "   systemctl status ${APP_NAME}.service"
echo "   tail ${LOG_DIR}/fail2cloudflare.log -f"