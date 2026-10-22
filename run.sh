#!/usr/bin/env bash
set -euo pipefail

RAW="${RATIO_TEST_RAW:-https://raw.githubusercontent.com/Aknuun/v2ray-ratio-test/main/ratio_test.py}"
CACHE="${RATIO_TEST_CACHE:-$HOME/.cache/ratio-test}"
SCRIPT="$CACHE/ratio_test.py"

if ! command -v python3 >/dev/null 2>&1; then
    echo "python3 نصب نیست." >&2
    exit 1
fi

mkdir -p "$CACHE"

fetch_script() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$RAW" -o "$SCRIPT.tmp"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$SCRIPT.tmp" "$RAW"
    else
        return 1
    fi
}

if fetch_script; then
    mv "$SCRIPT.tmp" "$SCRIPT"
elif [ ! -f "$SCRIPT" ]; then
    echo "دانلود اسکریپت ناموفق بود و نسخه ذخیره‌شده هم وجود ندارد." >&2
    exit 1
fi

chmod +x "$SCRIPT"
exec python3 "$SCRIPT" "$@"
