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

	// The pre-rendered branded 404 page (from src/pages/404.astro). If it's
	// somehow absent, we fall back to the FileServer's plain-text 404.
	notFound, _ := fs.ReadFile(sub, "404.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/_astro/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		if notFound != nil {
			w = &notFoundWriter{ResponseWriter: w, body: notFound}
		}
		files.ServeHTTP(w, r)
	})
}

// notFoundWriter swaps the FileServer's plain-text "404 page not found" for the
// site's branded 404 page, keeping a 404 status. It activates only when the
// wrapped handler writes a 404; every other response passes through untouched.
type notFoundWriter struct {
	http.ResponseWriter
	body      []byte
	intercept bool
}

func (w *notFoundWriter) WriteHeader(code int) {
	if code == http.StatusNotFound {
		w.intercept = true
		h := w.Header()
		h.Set("Content-Type", "text/html; charset=utf-8")
		h.Set("Cache-Control", "no-cache")
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		_, _ = w.ResponseWriter.Write(w.body)
		return
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *notFoundWriter) Write(b []byte) (int, error) {
	if w.intercept {
		return len(b), nil // swallow the FileServer's default 404 body
	}
	return w.ResponseWriter.Write(b)
}
