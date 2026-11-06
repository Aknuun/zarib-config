#!/usr/bin/env bash
set -euo pipefail
version="${1:?usage: release.sh <version> (e.g. 1.3.1)}"
root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$root"

sed -i "s/Version = \"[0-9.]*\"/Version = \"$version\"/" internal/ratio/ratio.go
gofmt -w internal/ratio/ratio.go

CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /tmp/opencode/zarib-config ./cmd/zarib-config
(cd web/react && npm ci && npm run build)

echo "Built Zarib $version. Deploy the binary and web/react/dist, then commit and tag."
