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
    // stay `style-src 'self'` with no 'unsafe-inline'. The live script is
    // likewise a same-origin external file (public/topology.js).
    inlineStylesheets: 'never',
  },
});
