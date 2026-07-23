/**
 * CV extractor: primary_cv.tex -> src/data/cv.ts.
 *
 * The .tex is the single source of truth for both the downloadable PDF
 * (pdflatex) and the /cv page. This script parses the LaTeX body with
 * unified-latex and emits structured, typed data the Astro page renders, so
 * the two can never drift. Run via `npm run cv:data` (and `cv:build`, which
 * also rebuilds the PDF). The generated cv.ts is committed, so the production
 * `astro build` imports it directly and never needs unified-latex.
 *
 * Scope: the document body from the first \section onward. The masthead
 * (name + contact) is decorative and stays hand-authored in cv.astro.
 *
 * Convention: a line beginning `%%WEB ` is a LaTeX comment (ignored by
 * pdflatex) that the extractor promotes to live content, so the web page can
 * carry detail the one-page PDF omits.
 */
import { readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { unified } from 'unified';
import { unifiedLatexFromString } from '@unified-latex/unified-latex-util-parse';
import { unifiedLatexToHast } from '@unified-latex/unified-latex-to-hast';
import rehypeStringify from 'rehype-stringify';

const SRC = fileURLToPath(new URL('./primary_cv.tex', import.meta.url));
const OUT = fileURLToPath(new URL('../src/data/cv.ts', import.meta.url));

// `\&` makes unified-latex's HTML conversion swallow the following space
// ("Infrastructure \& Home" -> "Infrastructure &Home"). Route it through a
// sentinel that is plain text to the converter, then restore it afterwards.
const AMP = '@@AMP@@';

// The semantic macros this file defines; unified-latex needs their arities to
// attach arguments rather than leaving trailing `{...}` as loose groups.
const macros = {
  section: { signature: 'm' },
  cventry: { signature: 'm m m' },
  cvsubtitle: { signature: 'm' },
  skillsline: { signature: 'm m' },
};

const parser = unified().use(unifiedLatexFromString, { macros });
const toHast = unified().use(unifiedLatexToHast).use(rehypeStringify);

function preprocess(tex) {
  // From \begin{document} (skipping the preamble, where \section also appears
  // inside \titleformat) to \end{document}. The masthead between \begin{document}
  // and the first real \section is ignored by the walk (no section open yet).
  const open = '\\begin{document}';
  const body = tex.slice(
    tex.indexOf(open) + open.length,
    tex.indexOf('\\end{document}'),
  );
  return body
    .replace(/^[ \t]*%%WEB ?/gm, '') // promote web-only lines to live content
    .replace(/\\&/g, AMP)
    .replace(/\\enspace\\textperiodcentered\\enspace/g, ' · ')
    .replace(/\\textperiodcentered/g, '·')
    .replace(/\\enspace/g, ' ')
    .replace(/\\noindent/g, '')
    .replace(/\\vspace\*?\{[^}]*\}/g, '')
    .replace(/\\LaTeX(\{\})?/g, 'LaTeX');
}

// Render a node array to an HTML fragment, normalising unified-latex's output
// to this site's markup: <code> for \texttt, external anchors for \href.
function renderRaw(nodes) {
  const tree = toHast.runSync({ type: 'root', content: nodes });
  return toHast.stringify(tree).trim();
}
function toHtml(nodes) {
  let h = renderRaw(nodes);
  if (/^<p>[\s\S]*<\/p>$/.test(h)) h = h.replace(/^<p>/, '').replace(/<\/p>$/, '');
  return h
    .replace(/<span class="texttt">/g, '<code>')
    .replace(/<\/span>/g, '</code>')
    .replace(/<b class="textbf">/g, '<strong>')
    .replace(/<\/b>/g, '</strong>')
    .replace(/<a class="href" href="([^"]*)">/g,
      (_, url) => `<a href="${url}" target="_blank" rel="noopener">`)
    .replace(new RegExp(AMP, 'g'), '&amp;')
    .replace(/\s+/g, ' ')
    .trim();
}
function toText(nodes) {
  return toHtml(nodes).replace(/<[^>]+>/g, '').replace(/&amp;/g, '&').trim();
}

const isStruct = (n) =>
  (n.type === 'macro' && ['section', 'skillsline', 'cventry', 'cvsubtitle'].includes(n.content)) ||
  (n.type === 'environment' && n.env === 'cvitems');

const lastEntry = (sec) => [...sec.blocks].reverse().find((b) => b.kind === 'entry');

function splitItems(nodes) {
  const groups = [];
  let cur = null;
  for (const n of nodes) {
    if (n.type === 'macro' && n.content === 'item') { cur = []; groups.push(cur); continue; }
    if (cur) cur.push(n);
  }
  return groups;
}

function extract(tex) {
  const ast = parser.parse(preprocess(tex));
  const content = ast.content;
  const sections = [];
  let sec = null;

  for (let i = 0; i < content.length; i++) {
    const n = content[i];

    if (n.type === 'macro' && n.content === 'section') {
      sec = { heading: toText(n.args[0].content), blocks: [] };
      sections.push(sec);
      continue;
    }
    if (!sec) continue;

    if (n.type === 'macro' && n.content === 'skillsline') {
      let block = sec.blocks.at(-1);
      if (!block || block.kind !== 'skills') { block = { kind: 'skills', rows: [] }; sec.blocks.push(block); }
      block.rows.push({ label: toText(n.args[0].content), items: toHtml(n.args[1].content) });
      continue;
    }
    if (n.type === 'macro' && n.content === 'cventry') {
      sec.blocks.push({
        kind: 'entry',
        title: toHtml(n.args[0].content),
        org: toText(n.args[1].content),
        dates: toText(n.args[2].content),
        bullets: [],
      });
      continue;
    }
    if (n.type === 'macro' && n.content === 'cvsubtitle') {
      const e = lastEntry(sec);
      if (e) e.subtitle = toText(n.args[0].content);
      continue;
    }
    if (n.type === 'environment' && n.env === 'cvitems') {
      const e = lastEntry(sec);
      const bullets = splitItems(n.content).map(toHtml).filter(Boolean);
      if (e) e.bullets.push(...bullets);
      continue;
    }

    // Loose prose: gather a run up to the next structural node.
    const run = [];
    let j = i;
    while (j < content.length && !isStruct(content[j])) run.push(content[j++]);
    i = j - 1;
    const html = toHtml(run);
    if (html) sec.blocks.push({ kind: 'prose', html });
  }
  return sections;
}

const sections = extract(readFileSync(SRC, 'utf8'));

const banner = `// AUTO-GENERATED from web/cv/primary_cv.tex by \`npm run cv:data\`.
// Do not edit by hand — edit the .tex and regenerate.
`;
const types = `export interface SkillRow { label: string; items: string; }
export interface CvEntry { kind: 'entry'; title: string; org: string; dates: string; subtitle?: string; bullets: string[]; }
export interface CvSkills { kind: 'skills'; rows: SkillRow[]; }
export interface CvProse { kind: 'prose'; html: string; }
export type CvBlock = CvEntry | CvSkills | CvProse;
export interface CvSection { heading: string; blocks: CvBlock[]; }
`;
const body = `${banner}\n${types}\nexport const sections: CvSection[] = ${JSON.stringify(sections, null, 2)};\n`;
writeFileSync(OUT, body);
console.log(`cv.ts written: ${sections.length} sections, ${sections.reduce((a, s) => a + s.blocks.length, 0)} blocks`);
