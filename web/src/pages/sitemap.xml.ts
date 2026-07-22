/**
 * Sitemap, generated at build time. Replaces the hand-maintained
 * public/sitemap.xml: the static routes are listed once here, and the
 * /homelab/notes entries are derived from the content collection so new posts appear on their
 * own (drafts excluded). Kept as an endpoint rather than an @astrojs/sitemap
 * dependency to preserve the single-package front-end.
 */
import type { APIRoute } from 'astro';
import { getCollection } from 'astro:content';

const SITE = 'https://henrydowd.dev';

// Top-level routes. Add a page here when you add a route.
const staticRoutes = ['/', '/cv/', '/about/', '/homelab/', '/status/'];

export const GET: APIRoute = async () => {
  const posts = (await getCollection('writing', ({ data }) => !data.draft)).map(
    ({ id, data }) => ({ loc: `/homelab/notes/${id}/`, lastmod: data.date }),
  );

  const urls = [
    ...staticRoutes.map((loc) => ({ loc, lastmod: null as Date | null })),
    ...posts,
  ];

  const body = urls
    .map(({ loc, lastmod }) => {
      const last = lastmod ? `<lastmod>${lastmod.toISOString().slice(0, 10)}</lastmod>` : '';
      return `  <url><loc>${SITE}${loc}</loc>${last}</url>`;
    })
    .join('\n');

  const xml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${body}
</urlset>
`;

  return new Response(xml, { headers: { 'Content-Type': 'application/xml; charset=utf-8' } });
};
