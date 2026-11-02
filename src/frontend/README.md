# Timelapse frontend

The control panel served by the Go binary. People open it on a phone or laptop
joined to the Pi's own WiFi, usually outdoors and offline. It shows the latest
frame and when the next one will be taken, lets you download or delete photos,
take a framing preview, check temperatures and storage, adjust the schedule, and
read logs.

React 19, TypeScript, Vite. No UI framework: the whole app is about 80 KB of
gzipped JS plus two locally bundled font files, because it is embedded in the
Pi binary and served over a weak WiFi link. Nothing is loaded from the internet.

## Contract with the Go server

These must not change without changing `src/backend/main.go` and `src/backend/web/`:

- Output goes to `../backend/web/dist/` (Vite `outDir`), embedded by `src/backend/web` via `//go:embed dist`.
  `//go:embed` cannot reach outside its package, which is why the build is written into the backend.
- Assets are served under `/static/frontend/build/` (Vite `base`, `frontendPath` in `main.go`).
  This URL is independent of the on-disk location.
- `GET /` redirects to `/static/frontend/build/index.html`.
- API calls are same-origin. GET never changes state:
  `GET /photos`, `GET /file/:name`, `GET /file/last`, `GET /archive/{zip,tar}?f=…`, `GET /capture`,
  `GET|POST /configuration`, `GET /monitoring`, `GET /logs`, `GET /version`,
  `DELETE /file?f=…` (selected photos), `DELETE /photos` (all photos), `POST /admin/{restart,shutdown}`.
  The server rejects cross-site POST/DELETE requests (`http.CrossOriginProtection`), sends no CORS headers,
  and only answers to local host names (IPs, `localhost`, `*.local`, …; more via `-allowedHosts`) to block DNS rebinding.
  Payload types live in `src/api.ts` and mirror `src/backend/rest/model.go` and `src/backend/conf/settings.go`.
- Only intervals that divide an hour evenly (60 s to 1 h) are valid, with an offset below the interval.
  `src/schedule.ts` and the Go scheduler are both tested against `src/schedule-cases.json`.
- Navigation uses the URL hash (`#/settings`), so the server only ever serves `index.html`.

## Development

Requires Node 22.22+ or 24 LTS (see `.nvmrc`).

```bash
npm install
make run                 # in another terminal, from the repo root
npm start               # http://localhost:5173/static/frontend/build/
```

The dev server proxies API calls to `http://localhost:8080`. Point it at a real
Pi with `VITE_BACKEND=http://192.168.4.1:8080 npm start`.

| Command             | What it does                                          |
| ------------------- | ----------------------------------------------------- |
| `npm test`          | Runs the Vitest suite once (`npm run test:watch` to watch) |
| `npm run typecheck` | Type-checks without building                          |
| `npm run build`     | Type-checks, then builds into `../backend/web/dist/`                |
| `./build.sh`        | Same as build, with the git commit shown in Settings  |

## Layout

- `src/api.ts`: typed client and URL builders for every backend endpoint
- `src/schedule.ts`: a port of the Go capture scheduler, used for the countdown and dial
- `src/format.ts`: parsers for raw Pi output (`df -h`, `uptime`, temperatures, Go timestamps)
- `src/views/`: one component per section (Photos, Preview, Status, Settings, Logs)
- `src/components/`: the hour dial, photo viewer, dialogs and icons
