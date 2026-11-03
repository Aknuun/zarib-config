#!/usr/bin/env bash
# Archive zarib-config state and admin access, then send the archive to Telegram.
# Installed as /usr/local/bin/zarib-backup; scheduled by zarib-backup.timer.
#
# Optional overrides: APP_DIR, ETC_DIR, BACKUP_DIR, KEEP_DAYS.
set -euo pipefail
umask 077

APP_DIR="${APP_DIR:-/opt/v2ray-ratio}"
ETC_DIR="${ETC_DIR:-/etc/zarib-config}"
BACKUP_DIR="${BACKUP_DIR:-$APP_DIR/backups}"
CACHE_DIR="$APP_DIR/.cache/v2ray-ratio-web"
KEEP_DAYS="${KEEP_DAYS:-7}"

log() { printf '%s\n' "$*" >&2; }
die() { log "error: $*"; exit 1; }

# Read KEY=VALUE files without executing them.
load_env() {
  local file="$1" line key value
  [ -f "$file" ] || return 0
  while IFS= read -r line || [ -n "$line" ]; do
    line="${line%$'\r'}"
    case "$line" in ''|'#'*) continue ;; esac
    case "$line" in *=*) ;; *) continue ;; esac
    key="${line%%=*}"
    value="${line#*=}"
    case "$key" in ''|*[!A-Za-z0-9_]*) continue ;; esac
    case "$value" in
      \"*\") value="${value#\"}"; value="${value%\"}" ;;
      \'*\') value="${value#\'}"; value="${value%\'}" ;;
    esac
    export "$key=$value"
  done < "$file"
}

load_env "$ETC_DIR/telegram.env"
load_env "$ETC_DIR/proxy.env"

TOKEN="${ZARIB_TELEGRAM_TOKEN:-${TELEGRAM_TOKEN:-}}"
CHAT_ID="${ZARIB_TELEGRAM_CHAT_ID:-${TELEGRAM_CHAT_ID:-}}"
PROXY="${ZARIB_SUB_PROXY:-}"

mkdir -p "$BACKUP_DIR"
chmod 0700 "$BACKUP_DIR"

files=()
for path in "$CACHE_DIR/stats.json" "$CACHE_DIR/reports.jsonl" "$ETC_DIR/admin.env"; do
  [ -f "$path" ] && files+=("$path")
done
[ "${#files[@]}" -gt 0 ] || die "nothing to back up in $CACHE_DIR or $ETC_DIR"

stamp="$(date -u +%Y%m%d-%H%M%S)"
archive="$BACKUP_DIR/zarib-backup-$stamp.tar.gz"
tar_args=()
for path in "${files[@]}"; do
  tar_args+=(-C "$(dirname "$path")" "$(basename "$path")")
done
tar -czf "$archive" "${tar_args[@]}"
chmod 0600 "$archive"

find "$BACKUP_DIR" -maxdepth 1 -type f -name 'zarib-backup-*.tar.gz' -mtime +"$KEEP_DAYS" -delete

[ -n "$TOKEN" ] || die "TELEGRAM_TOKEN is not set in $ETC_DIR/telegram.env"
[ -n "$CHAT_ID" ] || die "TELEGRAM_CHAT_ID is not set in $ETC_DIR/telegram.env"

curl_args=(-fsS --max-time 120
  -F "chat_id=$CHAT_ID"
  -F "document=@$archive"
  -F "caption=zarib-config backup $stamp")
if [ -n "$PROXY" ]; then
  curl_args+=(-x "$PROXY")
fi

response="$(curl "${curl_args[@]}" "https://api.telegram.org/bot$TOKEN/sendDocument")" ||
  die "Telegram sendDocument failed"
case "$response" in
  *'"ok":true'*) ;;
  *) die "Telegram rejected the backup" ;;
esac

log "backup sent: $archive"
