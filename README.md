# V2Ray Config Coefficient Tester

This script downloads a known amount of data through a V2Ray config and compares it with the volume deducted from your subscription, so you can measure the **real traffic coefficient** of the config.

Example: you download 10 MB and the subscription drops by 15 MB. The coefficient is `1.5x`, meaning 50% more than the real traffic is deducted.

A hosted bilingual (Persian/English) version is available at **https://zarib-config.videobazi.com**. The web UI defaults to the browser language and a light theme, includes a dark-mode toggle, and can share a result image with the site address through the native share sheet.

## Requirements

- Python 3.9 or newer
- `curl`
- Internet access
- Xray: if it is not installed, the correct build is downloaded automatically from GitHub.
- The subscription must expose the `subscription-userinfo` header (most Marzban-based panels do).

## Run on a server

Run this directly on the server:

```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/Aknuun/zarib-config/main/run.sh)"
```

Or use the more robust form that keeps interactive input working:

```bash
curl -fsSL https://raw.githubusercontent.com/Aknuun/zarib-config/main/run.sh -o /tmp/ratio.sh && bash /tmp/ratio.sh
```

`run.sh` uses the local `ratio_test.py` when present; otherwise it downloads/updates the tester in `~/.cache/ratio-test`. Xray is also downloaded automatically when needed.

## Usage

Interactive:

```bash
python3 ratio_test.py
```

It asks for the subscription URL. If no config is passed, it extracts the configs from the subscription, lists all of them, and asks which number to test. Then it asks for the test size. The minimum and default size is **10 MB**. After the download it waits 20 seconds for the panel to record usage and keeps polling if needed.

List configs from a subscription without consuming traffic:

```bash
python3 ratio_test.py --sub "https://example.com/sub/xxxx" --list
```

Pick config number 5 without a prompt:

```bash
python3 ratio_test.py --sub "https://example.com/sub/xxxx" --pick 5
```

Non-interactive:

```bash
python3 ratio_test.py \
  --sub "https://example.com/sub/xxxx" \
  --config "vless://uuid@host:443?type=ws&security=tls#%C3%971.5"
```

Use a full JSON client config:

```bash
python3 ratio_test.py --sub "https://example.com/sub/xxxx" --config ./config.json
```

Save the result as JSON:

```bash
python3 ratio_test.py --sub "..." --config "..." --json-out result.json
```

### Options

| Option | Description |
| --- | --- |
| `--sub` | Subscription URL |
| `--config` | Config link or JSON file path |
| `--size` | Test size in MB (minimum 10, default 10) |
| `--wait` | Seconds to wait before reading usage (default 20) |
| `--poll-timeout` | Maximum seconds to wait for usage to be recorded (default 60) |
| `--drift-window` | Seconds to sample background usage before the test; 0 disables it (default 10) |
| `--list` | Only list configs from the subscription and exit |
| `--pick` | Config number inside the subscription for automatic selection |
| `--chunk` | Download request size in MB (default 5) |
| `--timeout` | Maximum seconds per curl request (default 180) |
| `--test-url` | Test download URL; `{bytes}` is replaced with the size |
| `--probe-url` | Probe URL used to check connectivity |
| `--probe-timeout` | Seconds per probe URL (default 20) |
| `--probe-only` | Only check that the config connects; do not download test data |
| `--probe-retries` | Extra connection attempts after a failed probe (default 0) |
| `--probe-wait` | Seconds to wait between probe attempts (default 10) |
| `--core` | Xray core: `auto`, `latest`, `legacy` (v1.8.24), or `system` (default `auto`) |
| `--claimed` | Claimed coefficient, if not present in the config name |
| `--json-out` | Save the result to a JSON file |
| `--keep-temp` | Keep temporary files and the Xray log for debugging |

The subscription can be passed with `--sub` or entered interactively. Without `--config`, configs are read from the subscription and shown for selection. Both URI links (`vless://`, `vmess://`, `trojan://`, `ss://`) and full Xray JSON configs are supported. If a subscription contains a single top-level JSON client config (or a JSON array of them), it is accepted directly; you no longer need to paste it manually. You can still pass a config manually as a link or a JSON file path.

Placeholder entries that some panels put in subscriptions (for example a host like `ParsConnection` with no real domain) are skipped automatically. Full JSON configs are copied for the test and their custom routing rules are replaced with one rule that sends the test traffic to the selected proxy. This avoids missing `geoip.dat`/`geosite.dat` errors and port conflicts with the original config.

## How it works

1. The initial usage is read from the `subscription-userinfo` header.
2. Xray starts with the config and exposes a SOCKS proxy on `127.0.0.1` only.
3. The test data is downloaded through the proxy and the real received bytes are counted.
4. Before the test, background usage is sampled for 10 seconds. If the subscription is being used by another device, that rate is subtracted from the result.
5. After 20 seconds, the subscription is read again to capture the upload/download delta. If usage has not changed, the wait is repeated up to three times; in interactive mode the script asks before the second and third retries.
6. Formula:

```
coefficient = (upload delta + download delta) / actual downloaded bytes
```

If the config name contains something like `1.5x`, `x1.5`, `×1.2`, `ratio 1.5`, or a Persian coefficient label, the script treats it as the claimed coefficient and compares it with the measured value. Use `--claimed` to set it manually.

The connectivity probe tries Cloudflare first and then falls back to Google and Cloudflare's HTTP 204 endpoint. If all of them fail, the node is probably down, blocked, or too slow. Use `--probe-retries` and `--probe-wait` to retry a transient failure, or choose another config when prompted.

Some configs only work with a newer or older Xray core. `--core auto` tries the latest release, then the legacy v1.8.24 core, then any system binary. The JSON result includes the core that was actually used.

## Privacy

The tester does not persist subscription URLs, configs, or test results. During a test, the subscription URL is used only to fetch the config list; it is written to a private temporary file (`0600`) so the worker can read it, and that file and the job directory are deleted as soon as the job finishes or is cancelled. Temporary config files are also removed after the test. Results are kept in memory for up to one hour so the web UI can poll them, then they are pruned. No config or subscription data is written to a database, log, or long-term storage.

## Output

The result shows the real coefficient as a large number, followed by one plain sentence:

```
======================================================
              COEFFICIENT TEST RESULT
======================================================

  REAL COEFFICIENT

    ██████          ██████  ██████
        ██          ██  ██  ██  ██  ██  ██
    ██████          ██  ██  ██  ██    ██
    ██              ██  ██  ██  ██  ██  ██
    ██████    ██    ██████  ██████

  For every 1 MB downloaded, 2.00 MB is deducted from your subscription.

  Downloaded               50.00 MB
  Deducted                100.27 MB
  Claimed                   2.00x  matches the claim
  Remaining                91.17 GB  (~45.54 GB real)
======================================================
```

The wait for the panel is shown as a live countdown with a progress bar. Output is colored on a terminal; set `NO_COLOR=1` to disable colors.

## Limitations

- If the subscription has no `subscription-userinfo` header and no usage in the body, measurement is impossible.
- Some panels record usage with a delay or serve it from a cache; increase `--wait` and `--poll-timeout` in that case.
- If the subscription is used by another device at the same time, the script measures the background rate and subtracts it, but disconnecting other devices gives the most accurate result.
- The test consumes real traffic.
- Upload traffic (requests) is included in the usage delta, but it is negligible compared to the test size.
- Protocols not supported by Xray (such as Hysteria2 and TUIC) cannot be used as links; provide an Xray-compatible JSON config instead.
- For better accuracy, use a larger test size. A few percent of error is normal.

## Security

- The SOCKS proxy listens on `127.0.0.1` only.
- The temporary config file is created with `600` permissions.
- The Xray log is removed after the test (use `--keep-temp` to keep it).
- The script does not send any data to any third-party server.

## Tests

```bash
python3 -m unittest discover -s tests -v
```
