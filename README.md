# zarib-config

Measure the real traffic ratio of a V2Ray subscription. Download a known amount of data, compare it with the subscription usage delta, and get a coefficient such as `1.5x`.

Hosted web UI: **https://zarib-config.videobazi.com**

## Quick start

```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/Aknuun/zarib-config/main/run.sh)"
```

Requirements: Python 3.9+, `curl`, and a subscription that exposes usage via `subscription-userinfo`. Xray is downloaded automatically.

## CLI

```bash
python3 ratio_test.py
python3 ratio_test.py --sub "https://example.com/sub/xxxx" --list
python3 ratio_test.py --sub "https://example.com/sub/xxxx" --pick 5
python3 ratio_test.py --sub "https://example.com/sub/xxxx" --config "vless://..." --json-out result.json
```

Configs can be a URI (`vless://`, `vmess://`, `trojan://`, `ss://`) or a JSON file. Run `python3 ratio_test.py --help` for all options.

## How it works

1. Read the subscription usage.
2. Start Xray locally and download the test data through the selected config.
3. Read usage again and calculate:

```
coefficient = (upload delta + download delta) / downloaded bytes
```

## Privacy

No subscription URL, config, or result is persisted. The URL is used only during the test and is deleted with the job's temporary files when the test finishes. Results stay in memory for up to one hour for the web UI and are then pruned.

## Notes

- The test consumes real traffic.
- The subscription must expose usage; otherwise the ratio cannot be measured.
- Panels that record usage with a delay may need a longer wait.
- For best accuracy, use a larger test size and disconnect other devices.
- SOCKS listens on `127.0.0.1` only; temporary files use `0600`.

## Tests

```bash
python3 -m unittest discover -s tests -v
```
