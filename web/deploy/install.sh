#!/usr/bin/env bash
set -euo pipefail

DOMAIN="${DOMAIN:-zarib-config.videobazi.com}"
OLD_DOMAIN="${OLD_DOMAIN:-v2ray-ratio.videobazi.com}"
APP_DIR="${APP_DIR:-/opt/v2ray-ratio}"
HERE="$(cd "$(dirname "$0")" && pwd)"
WEB_DIR="$(cd "$HERE/.." && pwd)"
REPO_DIR="$(cd "$WEB_DIR/.." && pwd)"

if [ "$(id -u)" -ne 0 ]; then
  echo "Run as root." >&2
  exit 1
fi

apt-get update
apt-get install -y --no-install-recommends nginx certbot curl ca-certificates
if ! command -v go >/dev/null 2>&1; then
  apt-get install -y --no-install-recommends golang-go
fi
if ! command -v go >/dev/null 2>&1; then
  echo "Go 1.22 or newer is required to build zarib-config." >&2
  exit 1
fi

if ! id -u ratio >/dev/null 2>&1; then
  useradd --system --home "$APP_DIR" --shell /usr/sbin/nologin ratio
fi

mkdir -p "$APP_DIR/bin" "$APP_DIR/.cache/ratio-test"
(
  cd "$REPO_DIR"
  CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$APP_DIR/bin/zarib-config" ./cmd/zarib-config
)
rm -rf "$APP_DIR/app" "$APP_DIR/venv"
mkdir -p "$APP_DIR/app"
cp -r "$WEB_DIR/static" "$APP_DIR/app/static"
find "$APP_DIR/app" -type f -exec chmod 0644 {} +
find "$APP_DIR/app" -type d -exec chmod 0755 {} +

if [ -x /root/.cache/ratio-test/xray ]; then
  install -m 0755 /root/.cache/ratio-test/xray "$APP_DIR/.cache/ratio-test/xray"
fi

chown -R ratio:ratio "$APP_DIR/.cache"
chmod -R u+rwX "$APP_DIR/.cache"
chown -R root:root "$APP_DIR/app" "$APP_DIR/bin"

install -m 0644 "$HERE/v2ray-ratio.service" /etc/systemd/system/v2ray-ratio.service
systemctl daemon-reload
systemctl enable v2ray-ratio.service
systemctl restart v2ray-ratio.service

install -m 0644 "$HERE/nginx-v2ray-ratio.conf" /etc/nginx/sites-available/v2ray-ratio
ln -sf /etc/nginx/sites-available/v2ray-ratio /etc/nginx/sites-enabled/v2ray-ratio
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx

if [ "${SKIP_CERTBOT:-0}" != "1" ]; then
  CERT_DOMAINS=(-d "$DOMAIN")
  if [ -n "${OLD_DOMAIN:-}" ]; then
    CERT_DOMAINS+=(-d "$OLD_DOMAIN")
  fi
  certbot --nginx "${CERT_DOMAINS[@]}" --non-interactive --agree-tos \
    --register-unsafely-without-email --redirect
fi

echo "Done: https://$DOMAIN"
