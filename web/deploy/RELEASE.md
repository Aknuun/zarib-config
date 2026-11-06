# Zarib release checklist

1. Run `CGO_ENABLED=0 go test ./...`.
2. Run `npm ci && npm run build` in `web/react`.
3. Back up `/opt/v2ray-ratio` and confirm the latest Telegram backup.
4. Build with `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /tmp/zarib-config ./cmd/zarib-config`.
5. Install the binary, replace static assets atomically, restart `v2ray-ratio`.
6. Check `/healthz`, `/api/ui`, `/admin`, and one known subscription before announcing.
7. Roll back by restoring the previous binary/assets and restarting; server-only env stays outside the repo.

Never commit `admin.env`, proxy credentials, Telegram tokens, or subscription links.
