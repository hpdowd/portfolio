package main

import "net/http"

// contentSecurityPolicy is intentionally strict: the site loads only same-origin
// assets and talks only to its own /api. All CSS and JS are external same-origin
// files — Astro is configured with inlineStylesheets:'never' and the live wiring
// is served as ES modules (/home.js, /status.js, and their shared /lib.js) — so
// 'self' needs no 'unsafe-inline'. Same-origin module imports are allowed under
// script-src 'self', so lib.js needs no directive of its own. If the
// front-end ever loads a cross-origin asset (a web font, an embed), widen the
// matching directive here rather than reaching for 'unsafe-inline'.
const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self'; " +
	"img-src 'self' data:; " +
	"connect-src 'self'; " +
	"font-src 'self'; " +
	"base-uri 'none'; " +
	"form-action 'none'; " +
	"frame-ancestors 'none'; " +
	"object-src 'none'"

// securityHeaders sets conservative security headers on every response from the
// PUBLIC listener. It is applied only to :8080 — never the private :9090 metrics
// port, which is not internet-reachable and is scraped by VictoriaMetrics.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Frame-Options", "DENY") // legacy companion to frame-ancestors 'none'
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), interest-cohort=()")
		// Served behind Traefik TLS on an HTTPS-only domain; tell browsers to
		// remember that. (Covers henrydowd.dev + its HTTPS subdomains.)
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}
