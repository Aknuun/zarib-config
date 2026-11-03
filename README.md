# zarib-config

Measures the real traffic ratio of a V2Ray or WireGuard subscription. It downloads a known amount through one config and compares it with the usage reported by the panel.

Live: <https://zarib.app11.ir/>

## Run on a personal server

No root and no Go install needed. It downloads the prebuilt binary for Linux or macOS:

```bash
bash -c "$(curl -fsSL https://github.com/Aknuun/zarib-config/raw/refs/heads/main/run.sh)"
```

Set `ZARIB_UPDATE=1` to fetch the newest binary.

## Notes

- The subscription must expose `subscription-userinfo`.
- Results are estimates. Panels can report usage late or count it differently.
- Subscription links, configs, and test results are not stored.
- Error reports are stored without the subscription URL in `reports.jsonl` next to `stats.json`. The admin panel at `/admin` is enabled by setting `ADMIN_USER` and `ADMIN_PASSWORD` (for example in `/etc/zarib-config/admin.env`); without them it returns 503.
- Backups: `stats.json` and `reports.jsonl` in `/opt/v2ray-ratio/.cache/v2ray-ratio-web/`, plus `/etc/zarib-config/telegram.env` and `/etc/zarib-config/admin.env`.
