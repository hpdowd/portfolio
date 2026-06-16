# syntax=docker/dockerfile:1
#
# Three-stage build producing one small, static image:
#   1. node  — build the Astro site to web/dist
#   2. golang — build the Go binary, embedding that web/dist
#   3. distroless — copy just the binary into a minimal, non-root runtime
#
# Pin the base image tags to exact digests/patches before relying on this in
# production (the homelab convention is to verify versions, never trust a
# floating tag).

# ---- stage 1: build the static site ----
FROM node:22-alpine AS web
WORKDIR /src/web
# Copy manifests first so `npm ci` is cached unless dependencies change. The
# lock file is optional on the very first build (ci needs it; install creates it).
COPY web/package.json web/package-lock.json* ./
RUN npm ci || npm install
COPY web/ ./
RUN npm run build            # -> /src/web/dist

# ---- stage 2: build the Go binary (embeds the site) ----
FROM golang:1.23-alpine AS app
WORKDIR /src
COPY . .
# Overlay the freshly built site so `//go:embed all:web/dist` has real content
# (the local/committed web/dist is excluded via .dockerignore).
COPY --from=web /src/web/dist ./web/dist
ARG VERSION=dev
ARG COMMIT=none
RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
      -o /out/portfolio .

# ---- stage 3: minimal runtime ----
# distroless/static: no shell, no package manager, runs as an unprivileged user.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=app /out/portfolio /portfolio
EXPOSE 8080 9090
USER nonroot:nonroot
ENTRYPOINT ["/portfolio"]
