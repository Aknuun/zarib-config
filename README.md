# zarib-config

Measures the real traffic ratio of a V2Ray or WireGuard subscription. It downloads a known amount through one config and compares it with the usage reported by the panel.

Live: <https://zarib.app11.ir/>

## Install on a personal server

On a fresh Ubuntu or Debian server with a domain pointing to it:

```bash
curl -fsSL https://raw.githubusercontent.com/Aknuun/zarib-config/main/web/deploy/install.sh | sudo DOMAIN=zarib.app11.ir bash
```

The installer sets up nginx, a systemd service, and a Let's Encrypt certificate. It installs to `/opt/v2ray-ratio`.

## Notes

- The subscription must expose `subscription-userinfo`.
- Results are estimates. Panels can report usage late or count it differently.
- Subscription links, configs, and test results are not stored.
- Backups: `stats.json` in `/opt/v2ray-ratio/.cache/v2ray-ratio-web/` and `/etc/zarib-config/telegram.env`.
