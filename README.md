# zarib-config

Measures the real traffic ratio of a V2Ray or WireGuard subscription. It downloads a known amount through one config and compares the panel usage before and after.

A result of `1.38x` means that for every 1 MB downloaded, 1.38 MB is deducted from the subscription.

Live instance: https://zarib.app11.ir/

The number is an estimate. Panels can report usage late or count traffic differently, so use `--sync` when the panel has a fixed update interval.

## Requirements

- Linux or macOS
- `curl` and `tar`
- A subscription that returns `subscription-userinfo`

Go 1.22+ and Xray are downloaded automatically by `run.sh`.

## Quick start

```bash
bash -c "$(curl -fsSL https://github.com/Aknuun/zarib-config/raw/refs/heads/main/run.sh)"
```

The script opens a terminal menu. You can test one or several configs, choose the test size, pick the Xray core, and enable background measurement.

To update:

```bash
ZARIB_UPDATE=1 bash -c "$(curl -fsSL https://github.com/Aknuun/zarib-config/raw/refs/heads/main/run.sh)"
```

Pin a release with `ZARIB_REF=v1.2.15`.

## Build

```bash
go build -o zarib-config ./cmd/zarib-config
```

## CLI

```bash
./zarib-config --sub "https://panel.example.com/sub/..." --list
./zarib-config --sub "https://panel.example.com/sub/..." --pick 5
./zarib-config --sub "https://panel.example.com/sub/..." --pick 5 --sync --size 10
./zarib-config --config "vless://..." --json-out result.json
```

Supported configs:

- `vless://`
- `vmess://`
- `trojan://`
- `ss://`
- `wireguard://`
- A JSON client config (a single object or an array of configs)

Run `./zarib-config --help` for the rest of the flags.

## Web server

```bash
./zarib-config serve --addr 127.0.0.1:8000 --static-dir web/static
```

Or run it through the quick-start script:

```bash
bash -c "$(curl -fsSL https://github.com/Aknuun/zarib-config/raw/refs/heads/main/run.sh)" _ serve --addr 0.0.0.0:8000
```

Put the web server behind a reverse proxy and add authentication if it is public.

## How it works

1. Read the subscription usage.
2. Optionally wait one update interval and measure background usage.
3. Start Xray and download data through the selected config.
4. Read the usage again and calculate:

```text
coefficient = (usage delta - background) / downloaded bytes
```

## Web UI

The web UI can inspect subscriptions, show configs and their locations, run ping and tunnel checks, test one or several configs, show live counters, and send error reports to Telegram when configured.

## Privacy

The service does not store the subscription URL, configs, or test results. Web jobs are kept in memory for one hour. Temporary files are deleted, and the local SOCKS listener only binds to loopback.

## Deployment

On a fresh Ubuntu server:

```bash
curl -fsSL https://raw.githubusercontent.com/Aknuun/zarib-config/main/web/deploy/install.sh | sudo DOMAIN=zarib.example.com bash
```

Or from a checkout:

```bash
sudo DOMAIN=zarib.example.com bash web/deploy/install.sh
```

The installer sets up nginx, certbot, the Go binary, a systemd unit, and the static UI. It is safe to run again.

| Variable | Purpose |
| --- | --- |
| `DOMAIN` | Domain served by nginx |
| `OLD_DOMAIN` | Optional second domain |
| `APP_DIR` | Install directory (default `/opt/v2ray-ratio`) |
| `SKIP_CERTBOT` | Skip certificate issuance |
| `LETSENCRYPT_EMAIL` | Email for Let's Encrypt |
| `TELEGRAM_TOKEN` / `TELEGRAM_CHAT_ID` | Optional error reports |
| `ZARIB_REF` | Git ref to install (default `main`) |
| `SRC_DIR` | Source cache (default `/opt/zarib-config-src`) |

To keep the web counters and Telegram settings after a server rebuild, back up:

- `$APP_DIR/.cache/v2ray-ratio-web/stats.json`
- `/etc/zarib-config/telegram.env`

## Tests

```bash
go test ./...
```

## Notes

- A test really uses traffic from the subscription.
- The CLI downloads in 20 MB chunks. The web UI defaults to 25 MB and accepts 5–100 MB.
- 10–50 MB is usually enough. Very small tests are noisy.
- The tunnel check sends the config host to a public geo API, not the full config.
- Panels that update usage slowly can need a longer wait; `--sync` handles the common case.
