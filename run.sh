#!/usr/bin/env bash
set -euo pipefail

REPO="${ZARIB_REPO:-Aknuun/zarib-config}"
REF="${ZARIB_REF:-main}"
CACHE="${XDG_CACHE_HOME:-${HOME:-/tmp}/.cache}/zarib-config"
GO_VERSION="1.22.12"
GO_ROOT="$CACHE/go"
GO_BIN="$GO_ROOT/bin/go"
TMP_DIR=""

log() { printf '%s\n' "$*" >&2; }
die() { log "error: $*"; exit 1; }

cleanup() {
    if [ -n "${TMP_DIR:-}" ]; then
        rm -rf "$TMP_DIR"
    fi
}
trap cleanup EXIT

need_cmd() {
    command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

download() {
    local url="$1" out="$2"
    if [ -t 2 ]; then
        curl -fL --progress-bar --retry 3 --connect-timeout 15 "$url" -o "$out"
    else
        curl -fsSL --retry 3 --connect-timeout 15 "$url" -o "$out"
    fi
}

verify_sha() {
    local file="$1" sha="$2"
    if command -v sha256sum >/dev/null 2>&1; then
        (cd "$(dirname "$file")" && printf '%s  %s\n' "$sha" "$(basename "$file")" | sha256sum -c - >/dev/null)
    elif command -v shasum >/dev/null 2>&1; then
        (cd "$(dirname "$file")" && printf '%s  %s\n' "$sha" "$(basename "$file")" | shasum -a 256 -c - >/dev/null)
    else
        log "warning: no sha256 tool found; skipping checksum verification"
        return 0
    fi
}

download_go() {
    local file="$1" out="$2" sha="$3" url
    local urls=(
        "https://go.dev/dl/${file}"
        "https://mirrors.aliyun.com/golang/${file}"
        "https://mirrors.nju.edu.cn/golang/${file}"
    )
    for url in "${urls[@]}"; do
        if ! curl -fsIL --max-time 15 "$url" >/dev/null 2>&1; then
            log "source unavailable, trying another mirror..."
            continue
        fi
        log "Downloading from ${url}"
        if download "$url" "$out" && verify_sha "$out" "$sha"; then
            return 0
        fi
        rm -f "$out"
        log "download or checksum failed, trying another mirror..."
    done
    return 1
}

go_ok() {
    local bin="$1" version major minor rest
    [ -x "$bin" ] || return 1
    version="$("$bin" version 2>/dev/null | awk '{print $3}')"
    version="${version#go}"
    major="${version%%.*}"
    rest="${version#*.}"
    minor="${rest%%.*}"
    case "$major" in ''|*[!0-9]*) return 1 ;; esac
    case "$minor" in ''|*[!0-9]*) return 1 ;; esac
    [ "$major" -gt 1 ] || { [ "$major" -eq 1 ] && [ "$minor" -ge 22 ]; }
}

install_go() {
    local os arch file sha
    case "$(uname -s)" in
        Linux) os=linux ;;
        Darwin) os=darwin ;;
        *) die "unsupported OS: $(uname -s). Install Go 1.22+ manually." ;;
    esac
    case "$(uname -m)" in
        x86_64|amd64) arch=amd64 ;;
        arm64|aarch64) arch=arm64 ;;
        *) die "unsupported architecture: $(uname -m). Install Go 1.22+ manually." ;;
    esac
    case "$os/$arch" in
        linux/amd64) sha=4fa4f869b0f7fc6bb1eb2660e74657fbf04cdd290b5aef905585c86051b34d43 ;;
        linux/arm64) sha=fd017e647ec28525e86ae8203236e0653242722a7436929b1f775744e26278e7 ;;
        darwin/amd64) sha=e7bbe07e96f0bd3df04225090fe1e7852ed33af37c43a23e16edbbb3b90a5b7c ;;
        darwin/arm64) sha=416c35218edb9d20990b5d8fc87be655d8b39926f15524ea35c66ee70273050d ;;
        *) die "unsupported platform: $os/$arch. Install Go 1.22+ manually." ;;
    esac

    need_cmd curl
    need_cmd tar
    file="go${GO_VERSION}.${os}-${arch}.tar.gz"
    url="https://go.dev/dl/${file}"
    log "Go 1.22+ not found; installing Go ${GO_VERSION} for ${os}/${arch}..."
    log "Downloading Go (about 70 MB). This can take a minute; progress is shown below."
    TMP_DIR="$(mktemp -d)"
    download_go "$file" "$TMP_DIR/$file" "$sha" \
        || die "could not download Go; install Go 1.22+ manually"
    rm -rf "$GO_ROOT"
    mkdir -p "$GO_ROOT"
    tar -xzf "$TMP_DIR/$file" -C "$GO_ROOT" --strip-components=1
    rm -rf "$TMP_DIR"
    TMP_DIR=""
    [ -x "$GO_BIN" ] || die "Go installation failed"
}

pick_go() {
    if go_ok "$GO_BIN"; then
        export GOROOT="$GO_ROOT"
        export PATH="$GO_ROOT/bin:$PATH"
        return
    fi
    if command -v go >/dev/null 2>&1 && go_ok "$(command -v go)"; then
        GO_BIN="$(command -v go)"
        return
    fi
    install_go
    export GOROOT="$GO_ROOT"
    export PATH="$GO_ROOT/bin:$PATH"
}

fetch_source() {
    local url top
    case "$REF" in
        v[0-9]*) url="https://codeload.github.com/${REPO}/tar.gz/refs/tags/${REF}" ;;
        *) url="https://codeload.github.com/${REPO}/tar.gz/refs/heads/${REF}" ;;
    esac
    need_cmd curl
    need_cmd tar
    log "Downloading ${REPO}@${REF}..."
    TMP_DIR="$(mktemp -d)"
    download "$url" "$TMP_DIR/src.tar.gz"
    tar -xzf "$TMP_DIR/src.tar.gz" -C "$TMP_DIR"
    top="$(find "$TMP_DIR" -mindepth 1 -maxdepth 1 -type d -print -quit)"
    [ -n "$top" ] || die "source download failed"
    rm -rf "$SRC"
    mkdir -p "$(dirname "$SRC")"
    mv "$top" "$SRC"
    touch "$SRC/.downloaded"
    rm -rf "$TMP_DIR"
    TMP_DIR=""
}

main() {
    need_cmd curl
    need_cmd tar
    pick_go

    local here src stamp
    here="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" 2>/dev/null && pwd || pwd)"
    if [ -f "$here/go.mod" ] && grep -q '^module github.com/Aknuun/zarib-config' "$here/go.mod" 2>/dev/null; then
        src="$here"
    else
        src="$CACHE/src-${REF//\//_}"
        stamp="$src/.downloaded"
        if [ "$REF" = "main" ] || [ ! -d "$src" ] || [ -z "$(find "$stamp" -mtime -1 2>/dev/null)" ] || [ -n "${ZARIB_UPDATE:-}" ]; then
            SRC="$src"
            fetch_source
        fi
    fi

    cd "$src"
    mkdir -p "$CACHE/bin"
    log "Building zarib-config..."
    CGO_ENABLED=0 "$GO_BIN" build -trimpath -ldflags="-s -w" -o "$CACHE/bin/zarib-config" ./cmd/zarib-config
    exec "$CACHE/bin/zarib-config" "$@"
}

main "$@"
