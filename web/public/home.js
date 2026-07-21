// home.js — live wiring for the Swiss/editorial home view.
//
// Same-origin external module (not an inline script) so the CSP stays
// script-src 'self'. Reads /api/status + /api/git and writes the live homelab
// figures into the masthead badge and the right-hand telemetry aside, as plain
// text/typography — never a separate dashboard. Also renders a small req/s
// sparkline from the rolling poll history and a "latest change" line.
//
// Shared fetch/DOM primitives come from ./lib.js.

import { q, getJSON, setText, applyDotState } from './lib.js';

const INTERVAL_MS = 5000;
const MAX_POINTS = 40; // ~3.3 min of req/s history at the poll interval
const DOT_STATES = ['ok', 'warn', 'err'];

function paintDot(state) {
  for (const dot of document.querySelectorAll('[data-dot="cluster"]')) {
    applyDotState(dot, 'is', state, DOT_STATES);
  }
}

// Per-service health for the infrastructure diagram: a node's dot is shown only
// when VictoriaMetrics actually reports that component (services[key] present),
// then coloured by its up/down state. Unmonitored nodes stay dotless.
function paintServices(services) {
  for (const chip of document.querySelectorAll('[data-target]')) {
    const dot = chip.querySelector('.chip__dot');
    if (!dot) continue;
    const key = chip.getAttribute('data-target');
    const known = services && Object.prototype.hasOwnProperty.call(services, key);
    if (known) {
      dot.hidden = false;
      applyDotState(dot, 'is', services[key] ? 'ok' : 'err', DOT_STATES);
    } else {
      dot.hidden = true;
      dot.classList.remove('is-ok', 'is-warn', 'is-err');
    }
  }
}

// Rolling req/s sparkline, drawn into the aside's <polyline>.
const rateHistory = [];
function renderSparkline(rate) {
  rateHistory.push(rate);
  if (rateHistory.length > MAX_POINTS) rateHistory.shift();

  const line = q('[data-spark] polyline');
  if (!line || rateHistory.length < 2) return;

  const w = 120;
  const h = 26;
  const pad = 2;
  const max = Math.max(...rateHistory);
  const min = Math.min(...rateHistory);
  const span = max - min || 1; // avoid divide-by-zero on a flat line
  const n = rateHistory.length;

  const points = rateHistory
    .map((v, i) => {
      const x = (i / (n - 1)) * w;
      const y = h - pad - ((v - min) / span) * (h - pad * 2);
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(' ');
  line.setAttribute('points', points);
}

function relativeTime(iso) {
  const s = (Date.now() - new Date(iso).getTime()) / 1000;
  if (s < 3600) return `${Math.max(1, Math.floor(s / 60))}m ago`;
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
  return `${Math.floor(s / 86400)}d ago`;
}

async function tick() {
  const [status, git] = await Promise.all([getJSON('/api/status'), getJSON('/api/git')]);
  const c = status?.cluster;
  const healthy = !!(status?.live && c && c.alerts_firing === 0 && c.targets_up === c.targets_total);
  const build = (status?.service?.version || '').slice(0, 7);

  // Masthead badge state.
  setText(q('[data-live-state]'), !status ? 'offline' : healthy ? 'live' : 'degraded');
  paintDot(!status ? 'err' : healthy ? 'ok' : 'warn');

  // Aside readout.
  setText(q('[data-targets]'), c ? `${c.targets_up}/${c.targets_total}` : '—');
  setText(q('[data-alerts]'), c ? String(c.alerts_firing) : '—');
  setText(q('[data-rate]'), c ? c.request_rate.toFixed(2) : '—');
  setText(q('[data-build]'), build || '—');
  if (c) renderSparkline(c.request_rate);
  paintServices(c?.services);

  // Latest homelab commit.
  if (git?.commits?.length) {
    const k = git.commits[0];
    setText(q('[data-latest]'), `${k.sha} · ${k.summary} · ${relativeTime(k.when)}`);
  } else {
    setText(q('[data-latest]'), git ? 'no recent commits' : 'feed unavailable');
  }
}

void tick();
setInterval(() => void tick(), INTERVAL_MS);

// Scroll-spy: mark the rail link whose section is currently in view, so the
// index doubles as a "you are here" indicator as content grows.
const railLinks = new Map();
for (const a of document.querySelectorAll('.index a')) {
  const id = a.getAttribute('href')?.slice(1);
  if (id) railLinks.set(id, a);
}
if (railLinks.size && 'IntersectionObserver' in window) {
  const observer = new IntersectionObserver(
    (entries) => {
      for (const e of entries) {
        if (!e.isIntersecting) continue;
        for (const a of railLinks.values()) a.classList.remove('is-active');
        railLinks.get(e.target.id)?.classList.add('is-active');
      }
    },
    { rootMargin: '-25% 0px -65% 0px' },
  );
  for (const id of railLinks.keys()) {
    const section = document.getElementById(id);
    if (section) observer.observe(section);
  }
}
