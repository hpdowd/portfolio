/**
 * Content collections. The `writing` collection backs /homelab/notes — prose
 * pieces authored as Markdown under src/content/writing. Uses Astro's built-in content
 * layer (glob loader), so no extra dependency: the whole front-end still ships
 * with `astro` as its only package.
 */
import { defineCollection, z } from 'astro:content';
import { glob } from 'astro/loaders';

const writing = defineCollection({
  loader: glob({ pattern: '**/*.md', base: './src/content/writing' }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    // Publication date; drives ordering and the RSS feed.
    date: z.coerce.date(),
    // Draft entries render (so they can be previewed by URL) but are kept out of
    // the index, the RSS feed and the sitemap until flipped to false.
    draft: z.boolean().default(false),
  }),
});

export const collections = { writing };
