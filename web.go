package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// distFS holds the built Astro site, baked into the binary at compile time.
//
// The `all:` prefix makes go:embed include files whose names start with `.` or
// `_` — Astro emits fingerprinted assets under `_astro/`, so this matters.
//
// `web/dist` is a BUILD ARTIFACT, not source: it is produced by `npm run build`
// (see web/). The repo commits only `web/dist/.gitkeep` so this directive
// compiles on a fresh checkout; the Docker image copies the real build in
// before `go build`. For a local run, build the site first:
//
//	cd web && npm install && npm run build
//
//go:embed all:web/dist
var distFS embed.FS

// staticHandler serves the embedded site. Astro's fingerprinted assets (their
// hash changes whenever their content does) get long-lived immutable caching;
// everything else is marked no-cache so HTML is always revalidated.
func staticHandler() http.Handler {
	sub, err := fs.Sub(distFS, "web/dist")
	if err != nil {
		// Unreachable: the path is fixed and guaranteed by the embed directive.
		panic(err)
	}
	files := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/_astro/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	})
}
