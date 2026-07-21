// @ts-check
import { defineConfig } from 'astro/config';

// Fully static build. The output (web/dist) is embedded into the Go binary,
// which serves it alongside the same-origin /api/* endpoints — so there is
// nothing to proxy or run server-side here.
export default defineConfig({
  output: 'static',
  site: 'https://henrydowd.dev',
  build: {
    // Keep fingerprinted assets under /_astro/ — the Go static handler applies
    // immutable caching to exactly that prefix (see web.go).
    assets: '_astro',
    // Externalise all CSS (no inline <style>) so the Content-Security-Policy can
    // stay `style-src 'self'` with no 'unsafe-inline'. The live scripts are
    // likewise same-origin external modules (public/home.js, status.js and their
    // shared lib.js), which is why script-src 'self' needs no 'unsafe-inline'.
    inlineStylesheets: 'never',
  },
});
