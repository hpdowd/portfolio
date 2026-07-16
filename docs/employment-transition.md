# Employment transition — flipping the site from "seeking" to "employed"

The site is currently written as a job-seeker's portfolio: it *targets* infrastructure
roles rather than stating a current one. When you start a relevant job, the positioning
flips from *"here's proof I can do this"* to *"this is what I do"*, and the homelab stops
being the whole argument and becomes the flourish.

This is a **content checklist**, not a code change — the source is kept clean (no
commented-out stubs). Each item gives the current text and the shape of the replacement.
Search each string to find it. `<COMPANY>`, `<TITLE>`, and `<MON YYYY>` are placeholders.

Positioning verified across the whole site on 2026-07-16 (see git history around the
`copy: reposition portfolio toward infrastructure / systems engineering` commit); this
doc is the forward-looking companion to that work.

---

## 1. The one high-value structural add: a real Experience entry (CV)

`web/src/pages/cv.astro` → the `Experience` section. Add this **above** the existing
"Summer Student — Human Resources" entry, so the job leads your professional story:

```html
<div class="cv__entry">
  <div class="cv__entryhead">
    <h3><TITLE> <span class="cv__org">· <COMPANY></span></h3>
    <span class="cv__dates"><MON YYYY> – Present</span>
  </div>
  <ul>
    <li>…</li>
  </ul>
</div>
```

Once this exists, consider moving the whole `Experience` section **above** `Projects`
in the page — employment becomes the stronger section, and the homelab under Projects
reframes from "my only real infra experience" to "and I do it at home too, from the
bare metal up".

## 2. Machine-readable: add `worksFor` (and a real `jobTitle`) to the JSON-LD

This is what an ATS or Google's knowledge graph reads.

- `web/src/pages/index.astro` — the `Person` in `homeLd` (`@id … #person`). It already
  has `jobTitle: 'Infrastructure engineer'`; refine it to your real title and add, next
  to `alumniOf`:
  ```js
  worksFor: { '@type': 'Organization', name: '<COMPANY>' },
  ```
- `web/src/pages/cv.astro` — the `personLd` object. It has **no** `jobTitle` or
  `worksFor` yet; add both alongside `alumniOf`:
  ```js
  jobTitle: '<TITLE>',
  worksFor: { '@type': 'Organization', name: '<COMPANY>' },
  ```

## 3. Kill the "seeking / targeting" language

| File | Current | Change to |
|---|---|---|
| `web/src/pages/about.astro` (body) | *"I'm looking for a graduate role where that's the actual job: infrastructure engineering, systems, and the low-level plumbing underneath, not just the platform on top."* | State the job: *"I work as \<TITLE\> at \<COMPANY\>; the homelab is where I still get to own the whole stack, from the bare metal up."* |
| `web/src/pages/cv.astro` (Profile) | *"Targeting graduate roles in infrastructure engineering, systems, and SRE."* | *"Now working as \<TITLE\> at \<COMPANY\>."* |
| `web/src/layouts/Base.astro` (default `description`) | *"…a Pure Mathematics & Computer Science graduate targeting infrastructure engineering / systems / SRE roles…"* | *"…a Pure Mathematics & Computer Science graduate working as \<TITLE\> at \<COMPANY\>…"* |
| `web/src/pages/cv.astro` (`description` prop) | *"…targeting graduate infrastructure engineering / systems / SRE roles."* | *"…\<TITLE\> at \<COMPANY\>."* |
| `web/src/pages/cv.astro` (`personLd.description`) | *"…targeting graduate infrastructure engineering / systems / SRE roles."* | *"…\<TITLE\> at \<COMPANY\>."* |
| `web/src/pages/about.astro` (`description` prop) | *"…graduate of Maynooth focused on infrastructure engineering…"* | optional: *"…now working as \<TITLE\>…"* |

## 4. Optional touches

- **Masthead eyebrow** (`web/src/components/Home.astro`) is `Infrastructure · Systems ·
  Linux`. It can stay as a role-agnostic keyword line, or become your literal title.
- **OG image** (`web/og-image.svg`, see §5) — the tagline *"I build and run infrastructure
  from the bare metal up."* is role-agnostic and can stay. Only touch it if you want the
  share card to name the role.
- **Tone** — once employed, the copy can lean slightly warmer and less "here's my pitch";
  it's a standing professional presence, not a job-hunt artifact.

## 5. Regenerating the OG share card (needed only if you change §4's SVG)

The social card `web/public/og-image.png` (1200×630) is **generated from
`web/og-image.svg`** — that SVG is the source of truth. Edit the SVG, then regenerate
with the documented command (also in the README's "SEO & social metadata" section):

```bash
rsvg-convert -w 1200 -h 630 web/og-image.svg -o web/public/og-image.png
```

The SVG's font stack (`'Charter','Georgia',…`) resolves to **Charter** on this machine
now that it's installed; librsvg renders text via fontconfig, so it matches the site.

## 6. Watch for a title/policy constraint

Depending on your employer's policy you may prefer *"an infrastructure role in \<sector\>"*
over naming the company. Decide that before filling `<COMPANY>` above.

## 7. Deploy

Same as always: commit + push to Gitea → GitHub CI builds the image and pins the SHA into
the `homelab` repo → ArgoCD rolls it out. See the README's "Build & deploy (CI/CD)".
