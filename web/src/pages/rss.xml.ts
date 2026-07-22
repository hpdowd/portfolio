/**
 * RSS 2.0 feed for the homelab notes. Hand-rolled so the front-end keeps `astro` as its
 * only dependency (no @astrojs/rss). Prerendered at build time like every other
 * route. Drafts are excluded.
 */
import type { APIRoute } from 'astro';
import { getCollection } from 'astro:content';

const SITE = 'https://henrydowd.dev';

const esc = (s: string) =>
  s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

export const GET: APIRoute = async () => {
  const entries = (await getCollection('writing', ({ data }) => !data.draft)).sort(
    (a, b) => b.data.date.valueOf() - a.data.date.valueOf(),
  );

  const items = entries
    .map(({ id, data }) => `
    <item>
      <title>${esc(data.title)}</title>
      <link>${SITE}/homelab/notes/${id}/</link>
      <guid>${SITE}/homelab/notes/${id}/</guid>
      <description>${esc(data.description)}</description>
      <pubDate>${data.date.toUTCString()}</pubDate>
    </item>`)
    .join('');

  const xml = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Henry Dowd — Homelab notes</title>
    <link>${SITE}/homelab/</link>
    <description>Write-ups and post-mortems from running a self-hosted Kubernetes homelab.</description>
    <language>en-ie</language>${items}
  </channel>
</rss>
`;

  return new Response(xml, { headers: { 'Content-Type': 'application/xml; charset=utf-8' } });
};
