# Portfolio — design research: the AI aesthetic to avoid, and what professionals do

*Research doc, 2026-06-17. Written after the "Topology" home view (Concept C from
`portfolio-ui-overhaul.md`) was built, shipped, and **rejected** by Henry — "I
don't like it." Goal of this doc: ground the next direction in evidence, not
vibes. It's the external-research companion to `portfolio-ui-overhaul.md` (which
holds the concept wireframes) and feeds whatever direction we build next.*

**The brief, restated:** a portfolio/CV for a **platform / DevOps / SRE** engineer
that looks deliberately *designed* — not like an AI default, not like a SaaS
template. Distinct, intuitive, professional, low-motion, honest about the system
it runs on. Avoid the AI look **completely**.

---

## Part 1 — The "AI aesthetic": the blocklist (and *why* it happens)

The single most useful finding from the research: AI-generated sites look the same
for a *mechanical* reason, not a taste reason. LLMs were trained on billions of
tokens of web code (≈2019–2024) heavily weighted toward **Tailwind CSS**, whose
default palette leads with `indigo-500 / violet-600 / purple-700`. So the model
"learned" that *default web design = blue-purple*. The same averaging produces a
predictable layout, font, and component vocabulary. To avoid the look you have to
**actively delete the defaults**, not lightly reskin them.
(Sources: dev.to/alanwest *Indigo-500*; prg.sh *Purple Gradient*; Medium *The
Purple Problem*; Shuffle.dev.)

### Colour — the #1 tell
- ❌ **Indigo/violet/purple gradients** — the "Purple Problem." Blue→purple hero
  gradient on a white background is the signature.
- ❌ Renaming Tailwind `indigo` to `primary` — it's still indigo. **Remove the
  default scale**, don't alias it.
- ❌ Timid, evenly-distributed palettes with low contrast; "Google-style"
  red/green/blue/purple callouts on a dashboard.
- ✅ **One dominant colour + a sharp, reserved accent.** Pick hues *outside*
  Tailwind's default range (the article's example: a warm `#e8775a` ember, not a
  cool blue). Reserve saturation for **meaning** (state: ok/warn/err).
- ✅ Anchor to a specific real-world palette reference (blueprint cyan-on-vellum;
  warm paper + ink; phosphor terminal) so the colour reads as a *choice*.

### Typography — the #2 tell
- ❌ **Inter as the entire identity.** Also Roboto, Arial, `system-ui`, and now
  **Space Grotesk** (models have started converging on it too). One weight, no
  hierarchy beyond font-size, default line-height/tracking = "technically readable
  and entirely forgettable."
- ✅ **A pairing with real contrast.** Proven combos from the research:
  - *Editorial:* a display **serif** (Fraunces, Playfair Display, Newsreader) +
    a mono or humanist sans for body.
  - *Engineering-first:* a **mono** (JetBrains Mono, IBM Plex Mono, Commit Mono)
    for metadata/data/labels + a grotesque/sans for prose. Mono "reinforces the
    exposed-structure philosophy — makes the interface feel engineered."
  - Other display options that don't read AI: Clash Display, Bricolage Grotesque.
- ✅ **Extreme weight contrast (100 vs 900) and big size jumps (3×+).** Tabular
  figures for all numbers. Self-host the font (no Google-Fonts-Inter default).

### Components & surfaces — the #3 tell
- ❌ `rounded-2xl shadow-lg p-6` on *everything*; soft shadows at ~0.1 opacity;
  **glassmorphism** (frosted translucent cards); pill badges for skills.
- ✅ **One radius vocabulary sitewide** — often **0** (hard edges) for a
  Swiss/brutalist/technical feel. Use **borders + colour contrast** for hierarchy
  instead of drop shadows (e.g. `border` + a faint tinted fill, not `shadow`).
- ✅ A small set of *fixed* card variants (default / inset / flat), not
  per-component drift.

### Layout — the #4 tell
- ❌ The template spine: **hero → 3 feature cards w/ line-icons → testimonials →
  pricing → CTA.** Centered, symmetric, uniform grids, identical whitespace
  rhythm, vertically-stacked full-width sections.
- ✅ **Asymmetry on purpose.** Off-center compositions, an explicit/visible grid,
  content that *breaks* the grid, art that bleeds off one edge. The article's
  example grid: `grid-template-columns: minmax(2rem,1fr) minmax(0,38rem)
  minmax(0,1fr)` (content offset left, art bleeds right). Density where there's
  something to say.

### Copy — a quieter but real tell
- ❌ "**Empower / Unlock / Transform**," two-noun feature titles ("Seamless
  Integration"), emoji section headers (🚀 ✨ 🔥), gradient text, "Let's build
  something amazing."
- ✅ At least one **specific claim with a number**; at least one sentence that
  "sounds like a real person wrote it." For this site, real facts (a request rate,
  a target count, a commit SHA) beat adjectives.

### Motion — the last tell
- ❌ Floating blobs, aurora backgrounds, on-scroll fade-up on every element,
  autoplaying carousels.
- ✅ Motion only when something *earned* it (a value that changed, a nav
  transition). Respect `prefers-reduced-motion`. The contrast of one live element
  against an otherwise-still page reads as confident, not busy.

### What AI consistently *forgets* (these are the "human-made" signals)
Visual hierarchy beyond size · intentional colour theory · type pairing · white
space *as* design · brand voice/personality · and **functional details**: input
validation, error states, **empty states**, focus styles, loading states.

---

## Part 2 — What professionals actually use (named styles)

These read as "designed" because they're a **committed, non-default choice**
followed everywhere. Distinctiveness = consistency of one strong idea.

| Style | What it is | Signals | Fit for an SRE/infra portfolio |
|---|---|---|---|
| **Swiss / International** | Strict grid, big type, neutral palette, hairline rules, asymmetric balance | discipline, taste, clarity | ★★★ evergreen, safe-classy, hard to get wrong |
| **Editorial** | Magazine layout, real display serif, long-form, marginalia | craft, writing, rigor | ★★ great if content-forward |
| **Brutalist / Neubrutalism** | Raw, high-contrast, hard edges, visible structure, mono, bold borders | transparency, no-BS, memorable | ★★ polarizing; strong but use with restraint |
| **Monospace / Terminal** ("Tactical Telemetry / CRT") | Dark, monospaced, high-density, tabular/HUD, system readouts | infra-native, engineering-first | ★★★ on-thesis for SRE; risk = looks like a toy if overdone |
| **Blueprint / "Swiss Industrial Print"** | Light, technical-drawing, callouts, thin strokes, schematic | architecture rigor | ★★★ on-thesis, calmer than terminal |

Sources converge on two infra-flavoured archetypes worth knowing by name:
**"Swiss Industrial Print"** (light, blueprint-driven, heavy sans) and **"Tactical
Telemetry / CRT Terminal"** (dark, monospaced, dense HUD). `JetBrains Mono` for
metadata/status is the canonical engineering-first move.
(Sources: Medium *Clear vs Bold: Swiss and Brutalism*; Superdesign style library;
neubrutalism.com.)

### Cross-cutting principles the good examples share
- **Treat it like a modern CV:** clear sections, generous space, low cognitive
  load, scannable. Standout engineer portfolios (Julian Ozen, Tom Weightman,
  Irene Alvarado, Emily Tagtow) win on *clarity + strong type + project stories
  with narrative context*, not effects. (Source: sitebuilderreport, 2026.)
- **Single page, smooth scroll** generally beats a multi-page site with heavy
  animation. Mobile-first, responsive, fast.
- **Semantic HTML, WCAG/a11y, SEO** (clean URLs, structured metadata) are table
  stakes — and for an SRE they're *part of the portfolio's argument*.
- **Make contact effortless** (in the header and the footer). Custom domain, one
  consistent identity. Keep it updated.

---

## Part 3 — Recommendation for *this* site

This is an SRE/platform portfolio served by one Go binary on a self-hosted
cluster, with live telemetry as the differentiator. The honest, on-thesis lane is
**"the system, observed"** — but the rejected Topology proved that leaning *all the
way* into a literal diagram is a step too far. The lesson: keep the **engineering-
honest typography + the live data**, drop the gimmick layout.

**Suggested direction (for discussion, not yet built):** a **Swiss / editorial
hybrid** — a strict typographic grid (the discipline + taste signal), set in a
**mono-led pairing** (mono for data/labels/metadata, a clean grotesque or text
serif for prose), **near-zero radius, borders-not-shadows**, one reserved signal
colour for live state. Content-forward (it's a CV), with the live homelab numbers
present as *typography* (tabular readouts inline), not as a separate "dashboard."
Calm, dense, distinctly engineered — without looking like Grafana *or* like an AI
template. This is essentially the doctrine of `portfolio-ui-overhaul.md` Concepts
B/D (Field Notebook / Changelog) and design-axis §6, now backed by the external
research.

**Decision still open** — Henry to weigh in on the lane (Swiss-editorial vs
mono-terminal vs blueprint) before another prototype. Build *one screen* and judge
it live, as before.

---

## Part 4 — Quick-reference do/don't (pin this when building)

**Never ship:** indigo/violet/purple (or any Tailwind default scale) · Inter or
Space Grotesk as the identity · `rounded-2xl` + soft shadow on everything ·
glassmorphism · centered hero → 3 icon cards → testimonials → CTA · emoji headers ·
gradient text · "Empower/Unlock/Transform" · floating-blob/aurora/fade-up-on-scroll
motion · pill skill badges.

**Always do:** delete default palettes (custom tokens only) · one committed type
pairing with weight/size contrast, self-hosted, tabular figures · one radius
vocabulary (lean to 0) · borders + contrast for hierarchy · asymmetric, visible
grid · one reserved signal colour for *meaning* · real numbers + human-voiced copy ·
semantic HTML, focus/empty/error states, `prefers-reduced-motion`, WCAG, SEO ·
motion only when data changed or on navigation.

---

## Part 5 — Using horizontal space on desktop (without hurting readability)

*Added 2026-06-17 after Henry noted the centered single column under-uses desktop width.*

The instinct is to widen the text column to fill the screen — but that **hurts**
readability: the optimal **measure** (line length) is ~45–75 characters, ~66
ideal, and a *larger* font wants a *shorter* line, not a longer one. So the
reading column should stay ~38–40rem regardless of screen width. (Bringhurst, via
UXPin / Baymard; Google Fonts Knowledge.)

The right way to use the extra width is **more columns, not wider text** — the
Swiss / editorial answer:
- **Multi-column grid.** Desktop editorial layouts lean on a 12-col grid with
  generous 24–32px gutters; here, a three-track layout (index rail · reading
  column · live aside) turns the side whitespace into *structured* columns.
- **Margin rails / sidenotes.** Secondary material — navigation, metadata, the
  live telemetry — lives in the margins beside the prose (the Tufte / Gwern
  sidenote pattern): at wide windows the rails occupy the space the text
  shouldn't.
- **Whitespace stays active.** Side whitespace is *not* wasted — it reduces eye
  strain and groups content. Use *some* of it with intent; don't fill every
  pixel (that's how the AI "feature grid" creeps back). Keep generous outer
  margins; no full-bleed.

**Applied here:** the home container widens to ~75rem and lays out as
`index rail | reading column (≤40rem) | live aside`. The reading measure is
unchanged (~66ch); the recovered width holds the section index (left) and a
persistent **live telemetry panel** (right) — req/s + a rolling sparkline,
targets, alerts, build, and the latest commit — which also gives the
"this page monitors itself" thesis a permanent home. Collapses to one column
(live panel as a band, index hidden) below 66rem.

## Part 6 — The AI writing voice (to avoid in the copy)

*Researched 2026-06-17 for populating the page — the prose equivalent of Parts
1–2. What makes writing read as machine-generated, so the site's words don't.*

**Vocabulary tells** (overused by LLMs, rare in real writing): delve, leverage,
robust, seamless, tapestry, realm, beacon, landscape, navigate (figuratively),
underscore, harness, illuminate, facilitate, bolster, testament, game-changer,
unlock, crucial, vital, pivotal. Three or four of these clustered in a short
passage is a near-certain tell.

**Phrase tells:** "In today's fast-paced world…", "It's not just X, it's Y"
(the contrast pattern, on repeat), "let's face it", "the bottom line", "Here's
the key takeaway", "at the end of the day".

**Structural tells:**
- The rigid **intro → three points → conclusion** shape; heavy signposting
  ("Firstly… Moreover… In conclusion…").
- **Tricolons everywhere** — ideas grouped in threes ("fast, reliable, and
  secure"). One is fine; a *pattern* of them is the tell.
- **Mechanical rhythm** — every sentence a similar length, or the "one short
  line. Then a break. Another short line." cadence on a loop.
- A relentlessly **even, polite, neutral tone**: no opinions, no friction.
- The **em dash is NOT a reliable tell** (a myth) — but in bulk it adds to the
  texture, so use in moderation.

**How to read human instead** (applied to all copy here):
- **Vary sentence length** — short for emphasis, long when the idea earns it; the
  occasional fragment, a real (non-rhetorical) question, a parenthetical aside.
- **Specifics over abstractions** — every "significant improvement" → *how much?*
  (0.79 vs 0.78 ROC AUC); every "many services" → *which?* Names, numbers, versions.
- **A point of view** — say what you actually think; a dry aside ("…with no
  patience left") reads human in a way no adjective can.
- **Read it aloud.** If you'd never say it, rewrite it.
- House voice for these pages: plain, concrete, lightly wry engineer-speak;
  concede limits honestly; let the real numbers (req/s, AUC, target counts) do
  the bragging.

## Part 7 — The infrastructure diagram (added 2026-07-14)

*Shipped to the home page as an `Infrastructure` section below the CV content.
This is the resolution of the tension Part 3 named: the **whole-page** "Topology"
was the rejected gimmick, but a **single disciplined diagram section** — built
from the site's own type system rather than as a flashy topology — is on-thesis
for a platform/SRE portfolio. The goal was a diagram that reads as an engineered
schematic, not as Grafana and not as an AI template.*

**Constraints it was built under.** No new dependencies; same tokens as the rest
of the page (mono/hairline/zero-radius, one reserved accent); CSP stays
`script-src 'self'` (no new inline script); must reflow to mobile without the
page scrolling horizontally.

**Formats prototyped and compared (all four built live, then judged).**

| Format | What it is | Why not chosen |
|---|---|---|
| A · Layered bands | Horizontal layers (ingress → routing → apps → platform → backup → metal), mono tag per layer, chips, connector arrows | Clear and scannable, but reads as a *list of tiers*, not as a system; the layering is the only relationship it shows |
| C · Left → right flow | Boxed stages (`ingress → edge → routing → apps`) with arrows, platform/metal as a base band | Good for a request *trace*, but wide (needs horizontal scroll on mobile) and it flattens the containment story |
| D · Monospace ASCII | One aligned box-drawing schematic in a `<pre>` | Most "engineered", but a fixed-width block scrolls on mobile and reads a touch toy/terminal; alignment is fragile to maintain |
| **B · Nested containment** ✅ | Boxes-within-boxes: `Proxmox host → k3s VMs → k3s cluster`, with LXCs and disks alongside, ingress annotated, backup leaving off-site | **Chosen.** It's the only format that shows *what runs inside what* — the actual value of the exhibit. Reflows cleanly (flex), and the three nesting levels read at a glance with a faint tint on the innermost box |

**Why B won.** The others render the topology as an *ordering* (tiers, or a
left-to-right pipeline); only nesting renders it as *containment*, which is the
honest shape of the system and the more interesting thing to show — the cluster
runs on VMs, which run on the host, with the LXCs as host-level siblings outside
k3s. It also carries the two facts the intro leans on: the Cloudflare-tunnel
ingress ("no open ports") annotated on the entry arrows, and `restic` as the one
path that leaves the premises.

**Live per-service health (added 2026-07-14, second pass).** The diagram started
with a single live dot on the `Portfolio` node; it now lights *every* monitored
node. Each node carries a `data-target`, and `home.js` shows+colours its dot only
when `/api/status`'s curated `services` map reports that component — so a dot
means "genuinely a scrape target," and the app nodes that aren't instrumented
stay honestly dotless. Building this surfaced a truth worth recording: the honest
version of "the diagram is alive" is that **most of the app tier isn't
monitored** — only the platform/observability plane is (portfolio, longhorn,
victoria-metrics, grafana, alertmanager, and — after adding scrapes in the
homelab repo — traefik and argocd). Lighting only the real targets reads as more
credible to an SRE than faking green on everything. Respects
`prefers-reduced-motion`.

**Companion: a public status page (`/status`).** The same self-monitoring thesis,
given its own room: 30-day platform availability (`avg` of up targets) as a
per-day strip plus 1/7/30-day rollups, read from `/api/uptime` (a VM range
query). It's the "here's my SLO, publicly" move that fits a platform/SRE
portfolio, kept in the same type system rather than looking like a Grafana embed.

**Lesson for next time.** A literal diagram is not off-limits; a literal diagram
*as the entire home layout* was. Scope the honest-but-showy idea to a section,
build it in the page's own visual language, and it stops reading as a gimmick.
And when a live element depends on real telemetry, let what's *actually* measured
drive it — don't invent green.

## Sources

- Shuffle.dev — *Why Do Most AI-Generated Websites Look the Same?* (2026-01) — https://shuffle.dev/blog/2026/01/why-do-most-ai-generated-websites-look-the-same/
- dev.to / Alan West — *Why Every AI-Built Website Looks the Same (Blame Tailwind's Indigo-500)* — https://dev.to/alanwest/why-every-ai-built-website-looks-the-same-blame-tailwinds-indigo-500-3h2p
- dev.to / Alan West — *How to fix the 'AI-generated' look in your frontend* — https://dev.to/alanwest/how-to-fix-the-ai-generated-look-in-your-frontend-1ahh
- prg.sh — *Why Your AI Keeps Building the Same Purple Gradient Website* — https://prg.sh/ramblings/Why-Your-AI-Keeps-Building-the-Same-Purple-Gradient-Website
- Medium / Kai Ni — *Why Do AI-Generated Websites Always Favour Blue-Purple Gradients?* — https://medium.com/@kai.ni/design-observation-why-do-ai-generated-websites-always-favour-blue-purple-gradients-ea91bf038d4c
- Medium — *The Purple Problem: Why AI Can't Stop Generating Purple Websites* — https://medium.com/@ai.in.motion.blog/the-purple-problem-why-ai-cant-stop-generating-purple-websites-4381fb066883
- freshlybrewed.co — *What's Wrong with AI-Generated Websites* — https://freshlybrewed.co/insights-news/ai-generated-websites/
- Medium / Manas Parashar — *Why Most AI-Generated Websites Feel The Same* — https://parashar--manas.medium.com/why-most-ai-generated-websites-feel-the-same-b62efaeb50fd
- Medium / Manoj Adhikari — *Clear vs Bold: Swiss and Brutalism for the Web* — https://medium.com/@manojadhikari57/clear-vs-bold-swiss-and-brutalism-for-the-web-2fbefedb5161
- neubrutalism.com — *The Definitive Guide to Neubrutalist Web Design* — https://neubrutalism.com/
- sitebuilderreport — *Software Engineer Portfolios: 15+ Well-Designed Examples (2026)* — https://www.sitebuilderreport.com/inspiration/software-engineer-portfolios
- zencoder.ai — *How to Create a Software Engineer Portfolio in 2026* — https://zencoder.ai/blog/how-to-create-software-engineer-portfolio
- UXPin — *Optimal Line Length for Readability: The 50–75 Character Rule* — https://www.uxpin.com/studio/blog/optimal-line-length-for-readability/
- Baymard — *Readability: The Optimal Line Length* — https://baymard.com/blog/line-length-readability
- Google Fonts Knowledge — *Understanding measure / line length* — https://fonts.google.com/knowledge/using_type/understanding_measure_line_length
- UXPin — *UI Grids: The Complete Guide to Grid Systems* — https://www.uxpin.com/studio/blog/ui-grids-how-to-guide/
- Gwern — *Sidenotes In Web Design* — https://gwern.net/sidenote
- Swiss Themes — *Swiss Design Principles Every Web Designer Should Know* — https://swissthemes.design/insights/swiss-design-for-web-designers
- Olivia Cal — *How to Spot AI Writing Tells: 17 Examples + AI Words Blacklist* — https://www.oliviacal.com/post/ai-writing-tells
- Wikipedia — *Signs of AI writing* — https://en.wikipedia.org/wiki/Wikipedia:Signs_of_AI_writing
- Grendesign — *AI Words and Phrases to Avoid in 2026* — https://grendesign.com.au/how-to-write-human-content-in-the-age-of-ai-a-comprehensive-list-of-ai-words-and-phrases-to-avoid-in-2026/
- George Kao — *How To Write Without Sounding Like AI* — https://www.georgekao.com/blog/118515-how-to-write-without-sounding-like-ai
- Plus AI — *The most overused ChatGPT words* — https://plusai.com/blog/the-most-overused-chatgpt-words
