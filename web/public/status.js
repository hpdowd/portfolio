// status.js — live wiring for the /status page.
//
// Same-origin external module (not inline) so the CSP stays script-src 'self'.
// Reads /api/uptime and paints the 30-day availability strip, the 1/7/30-day
// rollups, and the overall state line. Shared fetch/DOM primitives come from
// ./lib.js.

import { q, getJSON } from './lib.js';

const POLL_MS = 30000;

// avail is a 0..1 fraction, or negative for "no data".
function pct(v) {
  return v < 0 ? '—' : `${(v * 100).toFixed(2)}%`;
}
function stateOf(v) {
  if (v < 0) return 'na';
  if (v >= 0.999) return 'ok';
  if (v >= 0.99) return 'warn';
  return 'err';
}
const STATE_LABEL = {
  ok: 'All systems operational',
  warn: 'Minor degradation',
  err: 'Degraded',
  na: 'Awaiting data',
};

function render(u) {
  q('[data-roll="1"]').textContent = pct(u.avg_1d);
  q('[data-roll="7"]').textContent = pct(u.avg_7d);
  q('[data-roll="30"]').textContent = pct(u.avg_30d);

  const strip = q('[data-strip]');
  strip.replaceChildren();
  for (const d of u.days) {
    const cell = document.createElement('span');
    cell.className = `cell cell--${stateOf(d.avail)}`;
    cell.title = `${d.date} · ${pct(d.avail)}`;
    strip.appendChild(cell);
  }

  // Overall state from the most recent day with data.
  const recent = [...u.days].reverse().find((d) => d.avail >= 0);
  const st = recent ? stateOf(recent.avail) : 'na';
  const dot = q('[data-dot-status]');
  if (dot) dot.className = `sdot sdot--${st}`;
  q('[data-state]').textContent = STATE_LABEL[st];
}

async function tick() {
  const r = await getJSON('/api/uptime');
  if (r && r.uptime && Array.isArray(r.uptime.days) && r.uptime.days.length) {
    render(r.uptime);
  } else {
    q('[data-state]').textContent = r ? 'Awaiting data' : 'Status feed unavailable';
  }
}

void tick();
setInterval(() => void tick(), POLL_MS);
