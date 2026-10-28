#!/usr/bin/env bash
# Fresh-server installer and disaster-recovery script for zarib-config.
#
# From a checkout:
#   sudo DOMAIN=zarib.example.com bash web/deploy/install.sh
#
# From a blank server (no checkout needed):
#   curl -fsSL https://raw.githubusercontent.com/Aknuun/zarib-config/main/web/deploy/install.sh \
#     | sudo DOMAIN=zarib.example.com bash
#
# Point the domain's A/AAAA record at this server before running it, otherwise
# only HTTP will be configured. Re-running is safe: it rebuilds the binary and
# static files, keeps $APP_DIR/.cache (stats), and restarts the service.
#
# Everything needed to recreate the site is in this repo. The only things that
# are not are the stats file and the Telegram bot credentials; see README.md.
set -euo pipefail

DOMAIN="${DOMAIN:-zarib-config.videobazi.com}"
OLD_DOMAIN="${OLD_DOMAIN:-}"
APP_DIR="${APP_DIR:-/opt/v2ray-ratio}"
SKIP_CERTBOT="${SKIP_CERTBOT:-0}"
LETSENCRYPT_EMAIL="${LETSENCRYPT_EMAIL:-}"
ZARIB_REPO="${ZARIB_REPO:-Aknuun/zarib-config}"
ZARIB_REF="${ZARIB_REF:-main}"
SRC_DIR="${SRC_DIR:-/opt/zarib-config-src}"
GO_VERSION="1.22.12"
GO_ROOT="/opt/zarib-config-go"

log() { printf '%s\n' "$*" >&2; }
die() { log "error: $*"; exit 1; }

[ "$(id -u)" -eq 0 ] || die "Run as root."
[ "$(uname -s)" = "Linux" ] || die "This installer supports Linux only."
case "$DOMAIN" in
  ''|*[!A-Za-z0-9.-]*) die "DOMAIN must be a hostname such as zarib.example.com." ;;
esac
if [ -n "$OLD_DOMAIN" ]; then
  case "$OLD_DOMAIN" in
    *[!A-Za-z0-9.-]*) die "OLD_DOMAIN must be a hostname such as old.example.com." ;;
  esac
fi

if { [ -n "${TELEGRAM_TOKEN:-}" ] && [ -z "${TELEGRAM_CHAT_ID:-}" ]; } ||
   { [ -z "${TELEGRAM_TOKEN:-}" ] && [ -n "${TELEGRAM_CHAT_ID:-}" ]; }; then
  die "Set both TELEGRAM_TOKEN and TELEGRAM_CHAT_ID, or neither."
fi

TMP="$(mktemp -d)"
cleanup() { [ -n "${TMP:-}" ] && rm -rf "$TMP"; }
trap cleanup EXIT

log "Installing system packages ..."
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y --no-install-recommends \
  ca-certificates certbot curl nginx python3-certbot-nginx tar

# Use the checkout this script lives in, an existing download, or fetch the repo.
SRC=""
SCRIPT="${BASH_SOURCE[0]:-}"
if [ -n "$SCRIPT" ] && [ -f "$SCRIPT" ]; then
  CANDIDATE="$(cd "$(dirname "$SCRIPT")/../.." && pwd)"
  if [ -f "$CANDIDATE/go.mod" ] && grep -q '^module github.com/Aknuun/zarib-config' "$CANDIDATE/go.mod"; then
    SRC="$CANDIDATE"
  fi
fi
# Reuse a real checkout at SRC_DIR, but refresh a copy this script downloaded
# earlier so a re-run picks up the latest source.
if [ -z "$SRC" ] && [ -f "$SRC_DIR/go.mod" ] && [ ! -f "$SRC_DIR/.downloaded" ] &&
   grep -q '^module github.com/Aknuun/zarib-config' "$SRC_DIR/go.mod"; then
  SRC="$SRC_DIR"
fi
if [ -z "$SRC" ]; then
  case "$ZARIB_REF" in
    v[0-9]*) URL="https://codeload.github.com/$ZARIB_REPO/tar.gz/refs/tags/$ZARIB_REF" ;;
    *) URL="https://codeload.github.com/$ZARIB_REPO/tar.gz/refs/heads/$ZARIB_REF" ;;
  esac
  log "Downloading $ZARIB_REPO@$ZARIB_REF ..."
  curl -fsSL --retry 3 "$URL" -o "$TMP/src.tar.gz"
  tar -xzf "$TMP/src.tar.gz" -C "$TMP"
  TOP="$(find "$TMP" -mindepth 1 -maxdepth 1 -type d -print -quit)"
  [ -n "$TOP" ] || die "Could not unpack the source archive."
  rm -rf "$SRC_DIR"
  mkdir -p "$(dirname "$SRC_DIR")"
  mv "$TOP" "$SRC_DIR"
  touch "$SRC_DIR/.downloaded"
  SRC="$SRC_DIR"
fi

go_ok() {
  local version major rest minor
  version="$("$1" version 2>/dev/null | awk '{print $3}')" || return 1
  version="${version#go}"
  major="${version%%.*}"
  rest="${version#*.}"
  minor="${rest%%.*}"
  case "$major" in ''|*[!0-9]*) return 1 ;; esac
  case "$minor" in ''|*[!0-9]*) return 1 ;; esac
  [ "$major" -gt 1 ] || { [ "$major" -eq 1 ] && [ "$minor" -ge 22 ]; }
}

GO_BIN=""
if command -v go >/dev/null 2>&1 && go_ok "$(command -v go)"; then
  GO_BIN="$(command -v go)"
elif [ -x "$GO_ROOT/bin/go" ] && go_ok "$GO_ROOT/bin/go"; then
  GO_BIN="$GO_ROOT/bin/go"
else
  case "$(uname -m)" in
    x86_64|amd64) ARCH=amd64; SHA=4fa4f869b0f7fc6bb1eb2660e74657fbf04cdd290b5aef905585c86051b34d43 ;;
    aarch64|arm64) ARCH=arm64; SHA=fd017e647ec28525e86ae8203236e0653242722a7436929b1f775744e26278e7 ;;
    *) die "Unsupported architecture: $(uname -m). Install Go 1.22+ manually." ;;
  esac
  FILE="go${GO_VERSION}.linux-${ARCH}.tar.gz"
  log "Go 1.22+ not found; installing Go ${GO_VERSION} ..."
  curl -fL --retry 3 --connect-timeout 15 "https://go.dev/dl/$FILE" -o "$TMP/$FILE"
  (cd "$TMP" && printf '%s  %s\n' "$SHA" "$FILE" | sha256sum -c - >/dev/null)
  rm -rf "$GO_ROOT"
  mkdir -p "$GO_ROOT"
  tar -xzf "$TMP/$FILE" -C "$GO_ROOT" --strip-components=1
  GO_BIN="$GO_ROOT/bin/go"
fi

log "Building zarib-config ..."
mkdir -p "$APP_DIR/bin"
(
  cd "$SRC"
  CGO_ENABLED=0 "$GO_BIN" build -trimpath -ldflags="-s -w" \
    -o "$APP_DIR/bin/zarib-config" ./cmd/zarib-config
)

if ! id -u ratio >/dev/null 2>&1; then
  useradd --system --home "$APP_DIR" --shell /usr/sbin/nologin ratio
fi

rm -rf "$APP_DIR/app"
mkdir -p "$APP_DIR/app"
cp -a "$SRC/web/static" "$APP_DIR/app/static"
find "$APP_DIR/app" -type f -exec chmod 0644 {} +
find "$APP_DIR/app" -type d -exec chmod 0755 {} +

mkdir -p "$APP_DIR/.cache/ratio-test"
if [ -x /root/.cache/ratio-test/xray ]; then
  install -m 0755 /root/.cache/ratio-test/xray "$APP_DIR/.cache/ratio-test/xray"
fi
chown -R ratio:ratio "$APP_DIR/.cache"
chmod -R u+rwX "$APP_DIR/.cache"
chown -R root:root "$APP_DIR/app" "$APP_DIR/bin"

if [ -n "${TELEGRAM_TOKEN:-}" ]; then
  mkdir -p /etc/zarib-config
  (
    umask 077
    printf 'ZARIB_TELEGRAM_TOKEN=%s\nZARIB_TELEGRAM_CHAT_ID=%s\n' \
      "$TELEGRAM_TOKEN" "$TELEGRAM_CHAT_ID" > /etc/zarib-config/telegram.env
  )
  chmod 0600 /etc/zarib-config/telegram.env
fi

log "Installing the systemd service ..."
sed "s|/opt/v2ray-ratio|$APP_DIR|g" "$SRC/web/deploy/v2ray-ratio.service" \
  > /etc/systemd/system/v2ray-ratio.service
chmod 0644 /etc/systemd/system/v2ray-ratio.service
systemctl daemon-reload
systemctl enable v2ray-ratio.service
systemctl restart v2ray-ratio.service

SERVER_NAMES="$DOMAIN"
if [ -n "$OLD_DOMAIN" ]; then
  SERVER_NAMES="$SERVER_NAMES $OLD_DOMAIN"
fi

log "Configuring nginx for $SERVER_NAMES ..."
SITE=/etc/nginx/sites-available/v2ray-ratio
install -m 0644 "$SRC/web/deploy/nginx-v2ray-ratio.conf" "$SITE"
sed -i "s|^[[:space:]]*server_name .*|    server_name $SERVER_NAMES;|" "$SITE"
ln -sf "$SITE" /etc/nginx/sites-enabled/v2ray-ratio
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx

if [ "$SKIP_CERTBOT" = "1" ]; then
  log "Skipping certbot (SKIP_CERTBOT=1). Site is on http://$DOMAIN"
else
  CERT_ARGS=(--nginx --non-interactive --agree-tos --redirect)
  if [ -n "$LETSENCRYPT_EMAIL" ]; then
    CERT_ARGS+=(-m "$LETSENCRYPT_EMAIL")
  else
    CERT_ARGS+=(--register-unsafely-without-email)
  fi
  CERT_ARGS+=(-d "$DOMAIN")
  if [ -n "$OLD_DOMAIN" ]; then
    CERT_ARGS+=(-d "$OLD_DOMAIN")
  fi
  if ! certbot "${CERT_ARGS[@]}"; then
    log "warning: HTTPS setup failed. Make sure $DOMAIN points to this server, then run:"
    log "  certbot --nginx -d $DOMAIN --redirect"
    log "The site is reachable over http://$DOMAIN until then."
  fi
fi

if [ -f /etc/zarib-config/telegram.env ]; then
  log "Telegram error reports: enabled (/etc/zarib-config/telegram.env)."
else
  log "Telegram error reports: disabled (set TELEGRAM_TOKEN and TELEGRAM_CHAT_ID to enable)."
fi

log "Done. Open https://$DOMAIN (or http://$DOMAIN if HTTPS was skipped)."
log "Logs: journalctl -u v2ray-ratio -f"
