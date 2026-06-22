# portfolio

A self-hosted portfolio / CV site that runs as just another app on my homelab
Kubernetes cluster — and reports on that cluster while it does.

It is a **single, dependency-free Go binary** that embeds a static
[Astro](https://astro.build) front-end and serves a small live API. At runtime
it reads cluster metrics from VictoriaMetrics and recent commits from Gitea over
the cluster-internal network, and it exposes its own Prometheus metrics, so it
appears as a monitored service in the very stack it surfaces.

- **Source repo (this):** what to build — the site and the service.
- **Config repo (`henry/homelab`):** how it runs — the Kubernetes manifests
  ArgoCD reconciles. CI here writes the new image tag into that repo.

## Why it's shaped this way

| Decision | Reason |
|---|---|
| One Go binary, site embedded via `go:embed` | One image, one Deployment, one thing to operate — least to build and maintain |
| Standard library only, no dependencies | Reproducible builds with no module proxy, a tiny static image, no `go.sum` to manage |
| Two HTTP listeners (`:8080` public, `:9090` private) | A single process can serve the site + API publicly while keeping `/metrics` and `/healthz` off the internet (the Ingress only routes `:8080`) |
| No credentials in the pod | VictoriaMetrics needs no auth; the Gitea feed reads a **public** repo anonymously. The public-facing pod holds no secrets to leak |

## Architecture

```
                     Internet (henrydowd.dev)
                              │
                   Cloudflare ─ tunnel ─ Traefik Ingress
                              │  (routes :8080 only)
        ┌─────────────────────▼──────────────────────────┐
        │  portfolio (one Go binary, one pod)             │
        │                                                 │
        │   :8080  ── /            embedded Astro site     │  ◀ public
        │          └─ /api/status  curated JSON  ┐        │
        │          └─ /api/git     curated JSON  │        │
        │                                        │        │
        │   :9090  ── /metrics  (Prometheus)     │        │  ◀ private
        │          └─ /healthz  (k8s probes)     │        │     (scrape + probes)
        └───────────────────────────┬───────────┼────────┘
                                     │           │
          cluster-internal, no auth  │           │  cluster-internal, anonymous
                                     ▼           ▼
                         VictoriaMetrics      Gitea (public henry/homelab)
                         count(up==1), …      recent commits
```

The browser only ever receives curated JSON; it never touches VictoriaMetrics or
Gitea directly.

## Repository layout

```
.
├── main.go                 entrypoint: the two listeners + graceful shutdown
├── web.go                  //go:embed of the built site + static handler
├── security.go             security headers + strict CSP on the public listener
├── go.mod                  module (no dependencies)
├── internal/
│   ├── config/             environment-driven configuration
│   ├── cache/              stale-while-revalidate TTL cache (one value)
│   ├── metrics/            hand-rolled Prometheus exposition + middleware
│   ├── vm/                 VictoriaMetrics client (fixed instant queries)
│   ├── gitea/              Gitea client (anonymous public-repo reads)
│   └── api/                /api/status and /api/git handlers
├── web/                    Astro front-end (static; built into web/dist)
│   ├── src/                Base layout, Home.astro, cv/about pages, global styles
│   ├── public/             favicon, home.js (live wiring), resume.pdf, robots.txt, sitemap.xml, og-image.png
│   └── dist/               build output — embedded into the binary (git-ignored)
├── docs/design.md          design rationale: anti-AI-aesthetic research + decisions
├── Dockerfile              3-stage: build site → build binary → distroless
└── .github/workflows/       build.yml — CI: test → build on GitHub → push GHCR → deploy
```

## Local development

The front-end and back-end build independently; the binary embeds whatever is in
`web/dist` at compile time.

```bash
# 1. Build the static site (also generates web/dist for the embed).
cd web
npm install        # first time only; commit the resulting package-lock.json
npm run build
cd ..

# 2. Build and run the binary.
go run .
```

Then:

- site + API → <http://localhost:8080>
- metrics + health → <http://localhost:9090/metrics>, `/healthz`

With the defaults, the Gitea feed works against the real public instance, while
VictoriaMetrics (cluster-only) is unreachable locally — so `/api/status` shows
the service's own uptime and degrades the cluster tiles gracefully. That is the
intended fail-soft behaviour.

For rapid front-end iteration use Astro's dev server (`cd web && npm run dev`);
note the `/api/*` calls only resolve when the Go binary is serving them.

## Configuration

All via environment variables; every value has a working default.

| Variable | Default | Purpose |
|---|---|---|
| `PORTFOLIO_ADDR` | `:8080` | Public listener (site + `/api`) |
| `PORTFOLIO_METRICS_ADDR` | `:9090` | Private listener (`/metrics` + `/healthz`) |
| `VM_URL` | `http://localhost:8428` | VictoriaMetrics base URL (cluster-internal) |
| `GITEA_URL` | `https://git.henrydowd.dev` | Gitea base URL |
| `GITEA_REPO` | `henry/homelab` | `owner/name` of the public repo for the activity feed |
| `PORTFOLIO_CACHE_TTL` | `30s` | How long upstream data is reused before a background refresh |
| `PORTFOLIO_REQUEST_TIMEOUT` | `5s` | Per-call timeout for upstream requests |

In the cluster, `VM_URL` and `GITEA_URL` are set to the in-cluster Service
addresses by the Deployment in the homelab repo.

## HTTP surface

| Port | Path | Notes |
|---|---|---|
| 8080 | `/` | Embedded static site (public) |
| 8080 | `/api/status` | Service uptime + cluster snapshot (curated JSON) |
| 8080 | `/api/git` | Recent commits of the public homelab repo (curated JSON) |
| 9090 | `/metrics` | Prometheus exposition — **not** Ingress-routed |
| 9090 | `/healthz` | Liveness/readiness — **not** Ingress-routed |

## Metrics

Exposed at `:9090/metrics` (hand-written exposition; see `internal/metrics`):

- `portfolio_build_info{version,commit}`
- `portfolio_http_requests_total{route,method,code}`
- `portfolio_http_request_duration_seconds` (histogram, by `route`)
- `portfolio_http_requests_in_flight`
- `portfolio_upstream_up{upstream}` — `vm` / `gitea` reachability
- `go_goroutines`

Routes are folded into a small fixed set of labels to keep cardinality bounded.

## Security headers & tests

Every response on the public `:8080` listener carries a strict
`Content-Security-Policy` — same-origin `script-src` / `style-src` with no
`unsafe-inline`, which is why all CSS and JS are external same-origin files (Astro
builds with `inlineStylesheets: 'never'` and the live script is served from
`/home.js`) — plus `X-Content-Type-Options`, `Referrer-Policy`, `X-Frame-Options`,
`Permissions-Policy`, and HSTS. The private `:9090` listener is deliberately left
bare. See `security.go`.

The pure logic is unit-tested with `go test ./...`: the VictoriaMetrics scalar
parser, the Gitea commit curation, the metrics exposition format, the cache's
stale-while-revalidate behaviour, and the security headers. CI runs the tests as
a gate before the image is built.

The front-end's visual direction (a Swiss/editorial, Gruvbox-themed layout that
deliberately avoids the generic "AI-generated" aesthetic) is documented in
[`docs/design.md`](docs/design.md).

## SEO & social metadata

Static, dependency-free, and hand-maintained (the site is only three routes):

- `web/public/robots.txt` — allows all crawlers and points at the sitemap.
- `web/public/sitemap.xml` — the three canonical URLs.
- **Open Graph / Twitter card** — `og:image` plus `twitter:card=summary_large_image`
  in the base layout reference `og-image.png` (1200×630), so a shared link renders
  a branded preview. The PNG is generated from `web/og-image.svg` (which mirrors the
  site palette and type); regenerate it after editing the SVG with:

  ```bash
  rsvg-convert -w 1200 -h 630 web/og-image.svg -o web/public/og-image.png
  ```
- **JSON-LD `Person`** on the CV page for rich-result eligibility. It is emitted as
  an inline `application/ld+json` data block, which the browser never executes, so
  the strict same-origin `script-src` CSP does not affect it.

## Build & deploy (CI/CD)

The in-cluster Gitea runner has no Docker daemon, so the image is **built on
GitHub's free hosted runners** — the same pattern NextKeep uses for its Android
build. The canonical repo on Gitea is push-mirrored to GitHub, and
`.github/workflows/build.yml` runs there on a push to `main`:

1. builds one image (Astro site embedded in the Go binary),
2. pushes it to **GHCR** as `ghcr.io/hpdowd/portfolio:<sha>` (and `:latest`),
3. if the `HOMELAB_TOKEN` secret is set, pins that `<sha>` into
   `k8s/apps/portfolio/deployment.yaml` in the homelab repo.

ArgoCD, watching the homelab repo, then rolls out the new image. The cluster
pulls from `ghcr.io` anonymously — exactly like it pulls Immich, no
`imagePullSecret`.

**Required once:**

- The Gitea → GitHub **push mirror** for this repo (so pushes reach GitHub).
- The GHCR package `portfolio` set to **public** after the first build
  (GitHub → your profile → Packages → portfolio → change visibility).
- For full GitOps, a Gitea PAT with write access to `henry/homelab` as the
  GitHub Actions secret **`HOMELAB_TOKEN`**. Without it the image still builds
  and `:latest` moves; the homelab tag is just bumped by hand.
- The deployment manifests in the homelab repo under `k8s/apps/portfolio/` plus
  the paired `k8s/apps/portfolio.yaml` Application.

GHCR pushes use the built-in `GITHUB_TOKEN` — no PAT needed for the registry.
Roll back by reverting the homelab commit.
