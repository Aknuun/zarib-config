# Zarib web (React + Vite)

New frontend for Zarib. The Go backend, API contracts, and `web/static` are untouched.

## Stack

- React 19 + Vite + TypeScript
- `@cloudflare/kumo` (Base UI + Phosphor icons), Tailwind CSS v4
- Local Estedad woff2 files in `src/assets/fonts` (no CDN)
- Build base is `/static/`, so the Go server serves `/static/assets/*` and the app at `/`

## Build

Requires Node 22 and npm 9.

```bash
cd web/react
npm install
npm run build
```

Output: `web/react/dist/`.

## Preview

### With the Go backend (recommended)

```bash
cd /root/zarib-config
go run ./cmd/zarib-config serve --addr 127.0.0.1:8000 --static-dir web/react/dist
```

Open <http://127.0.0.1:8000/>. This uses the real API and does not touch `web/static`.

### With Vite only

```bash
cd web/react
npm run dev
```

Vite serves the app at <http://127.0.0.1:5173/static/>. API calls need the backend; there is no dev proxy configured.

## Deploy

The owner can deploy by copying the build over the live static directory:

```bash
cp -r web/react/dist/assets web/static/
cp web/react/dist/index.html web/static/index.html
cp web/react/dist/theme-boot.js web/static/
cp web/react/dist/favicon.svg web/static/
```

`dist/assets` contains the JS, CSS, and Estedad fonts. `web/static/fonts`, `app.js`, and the old CSS can stay in place until they are removed; the new `index.html` does not reference them.

## Behavior notes

- API paths and payloads are unchanged (`/api/inspect`, `/api/jobs`, `/api/jobs/{id}`, cancel, presence, geo, ping, report, `/healthz`).
- Active jobs are stored in `sessionStorage` under `ratio-active-job-react` and resumed after reload.
- The share image is a 1080 px wide PNG canvas. Web Share is used when available, with download + caption copy as fallback.
- Theme and language live in `localStorage` (`ratio-theme`, `ratio-lang`); `public/theme-boot.js` applies them before React mounts to avoid a flash.
