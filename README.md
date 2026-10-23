# zarib-config

Measure the real traffic ratio of a V2Ray subscription. Download a known amount of data, compare it with the subscription usage delta, and get a coefficient such as `1.5x`.

Hosted web UI: **https://zarib.appii.ir/**

## Quick start

Run this on your own Linux or macOS server:

```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/Aknuun/zarib-config/main/run.sh)"
```

Requirements: `curl` and `tar`. Go 1.22+ is installed into `~/.cache/zarib-config/go` automatically when needed (no root required). Xray is downloaded automatically. Your subscription must expose usage via `subscription-userinfo`.

To refresh the cached source after an update:

```bash
ZARIB_UPDATE=1 bash -c "$(curl -fsSL https://raw.githubusercontent.com/Aknuun/zarib-config/main/run.sh)"
```

You can also pin a release: `ZARIB_REF=v1.2.8`.

> Results are estimates. There is a possibility of error or inaccuracy because panels may record usage with a delay or report it differently.

## Build

```bash
go build -o zarib-config ./cmd/zarib-config
```

## CLI

```bash
./zarib-config
./zarib-config --sub "https://example.com/sub/xxxx" --list
./zarib-config --sub "https://example.com/sub/xxxx" --pick 5
./zarib-config --sub "https://example.com/sub/xxxx" --pick 5 --sync --size 10
./zarib-config --sub "https://example.com/sub/xxxx" --config "vless://..." --json-out result.json
```

Configs can be a URI (`vless://`, `vmess://`, `trojan://`, `ss://`) or a JSON file. Run `./zarib-config --help` for all options.

`--sync` waits for the panel's usage update cycle and measures background traffic across one full interval before testing. It is slower but avoids false coefficients on subscriptions that have other traffic. The web UI uses this mode by default. The default test size is 10 MB; larger tests (up to 50 MB) are more accurate.

## Web server

```bash
./zarib-config serve --addr 127.0.0.1:8000 --static-dir web/static
```

Or start the self-hosted web UI directly from the one-liner:

```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/Aknuun/zarib-config/main/run.sh)" _ serve --addr 0.0.0.0:8000
```

Binding to `0.0.0.0` exposes the UI to the network; put it behind a reverse proxy and authentication.

## How it works

1. Read the subscription usage.
2. In sync mode, observe one full panel update interval to measure background traffic.
3. Start Xray locally and download the test data through the selected config.
4. Read usage again, subtract the measured background, and calculate:

```
coefficient = (usage delta - background) / downloaded bytes
```

## Web extras

- **Tunnel column**: each config can be checked on demand. The server resolves the config host, looks up its country through `ipinfo.io` (falling back to `ipwho.is`), and treats servers located in Iran as tunnels.
- **Ping**: a TCP connect check with latency in milliseconds. ICMP is not used because many servers block it.
- **Live stats**: the header shows how many browsers are currently open on the site. Clicking it shows today's completed tests, how many had a coefficient, how many were normal (1x), invalid tests, tested volume, and the tunnel/non-tunnel split. Only aggregate counters are saved; no URL or config is stored.
- **Error reports**: the result card has a "Report a problem" button. It sends the description and an optional subscription link (at least 100 MB of volume) to the developer through a Telegram bot. Reporting is disabled unless the server has `ZARIB_TELEGRAM_TOKEN` and `ZARIB_TELEGRAM_CHAT_ID`; see the systemd unit in `web/deploy/` for the `EnvironmentFile` path.

## Privacy

No subscription URL, config, or result is persisted. The URL is used only during the test and is deleted with the job's temporary files when the test finishes. Results stay in memory for up to one hour for the web UI and are then pruned. Online presences use a random browser ID and expire automatically; they are never stored on disk.

The tunnel check sends only the config server's hostname/IP to the geolocation service; the result is cached in memory for 30 minutes. Aggregate daily counters are stored in the web cache directory.

## Notes

- The test consumes real traffic. The web UI asks you to confirm that no other device or person is using the subscription before each test.
- The subscription must expose usage; otherwise the ratio cannot be measured.
- Panels that record usage with a delay may need a longer wait; sync mode handles the common one-minute update cycle.
- The web UI waits 120 seconds for usage (every second for the first 10 seconds, then every 5 seconds). If usage is still missing, it offers a button to wait up to 5 more minutes (checked every 10 seconds) or cancel.
- The connectivity probe has a 20-second budget. A TCP ping can succeed while the node blocks the test destination or is too slow, so the probe may fail even when the server is reachable.
- Shared subscriptions can have background traffic. Sync mode measures and subtracts it, but a larger test size and disconnecting other devices still improve accuracy.
- For best accuracy, use a larger test size (10-50 MB) and disconnect other devices.
- SOCKS listens on `127.0.0.1` only; temporary files use `0600`.

## Tests

```bash
go test ./...
```

## Deployment

```bash
sudo bash web/deploy/install.sh
```
