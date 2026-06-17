// topology.js — live wiring for the home "Topology" view.
//
// Served as a same-origin static ES module (not an Astro-hoisted inline script)
// so the Content-Security-Policy can stay `script-src 'self'` with no inline
// allowance. It mirrors the typed helpers in src/lib/api.ts; this file is plain
// JS because public/ assets are shipped as-is, with no bundling step.
//
// It only toggles classes and writes textContent on elements Astro rendered
// (which carry the component's scoped-style attribute), so the scoped CSS in
// Topology.astro styles the live states.

const INTERVAL_MS = 5000;

async function getJSON(path) {
  try {
    const res = await fetch(path, { headers: { Accept: 'application/json' } });
    return res.ok ? await res.json() : null;
  } catch {
    return null;
  }
}
const getStatus = () => getJSON('/api/status');
const getGit = () => getJSON('/api/git');

// Honest, specific copy for each node — no marketing, just what it does.
const DETAIL = {
  henry: {
    role: 'the operator',
    body: 'I build and run this platform end to end. Every node here is something I deploy, monitor, and keep alive.',
    href: '/cv/', hrefLabel: 'See the CV →',
  },
  gitea: {
    role: 'self-hosted Git · git.henrydowd.dev',
    body: 'Source of truth for both the app and the cluster-config repos. It push-mirrors to GitHub so image builds can run on hosted runners.',
  },
  actions: {
    role: 'CI on GitHub-hosted runners',
    body: 'The in-cluster runner has no Docker daemon, so the container image is built here and pushed to the registry. The mirror skips this job on Gitea.',
  },
  ghcr: {
    role: 'GitHub Container Registry',
    body: 'Holds the public image ghcr.io/hpdowd/portfolio, pulled anonymously by the cluster — no pull secret, no credential on the pod.',
  },
  argocd: {
    role: 'GitOps controller',
    body: 'Watches the homelab config repo and continuously reconciles the cluster to match it. Self-healing is on, so drift is corrected automatically.',
  },
  portfolio: {
    role: 'this site',
    body: 'One static Go binary embedding this front-end, serving the page plus a small live /api. It holds no secrets and is scraped back into VictoriaMetrics — so it monitors the platform that runs it.',
    href: 'https://git.henrydowd.dev/henry/portfolio', hrefLabel: 'Source →',
  },
  vm: {
    role: 'VictoriaMetrics · observability',
    body: 'Scrapes every target in the cluster — including this pod — and answers the queries behind the live panel you are reading. The loop back to portfolio is the recursion.',
  },
};

function setText(el, next) {
  // Only write when it changed, so aria-live regions don't re-announce noise.
  if (el && el.textContent !== next) el.textContent = next;
}

function paint(id, state, live = false) {
  for (const dot of document.querySelectorAll(`[data-dot="${id}"]`)) {
    dot.classList.remove('is-ok', 'is-warn', 'is-err', 'is-idle');
    dot.classList.add(`is-${state}`);
    dot.classList.toggle('is-live', live && state === 'ok');
  }
}

async function tick() {
  const [status, git] = await Promise.all([getStatus(), getGit()]);

  // Cluster aggregate → header + the colour of the structural infra nodes.
  const c = status?.cluster;
  const healthy = !!(status?.live && c && c.alerts_firing === 0 && c.targets_up === c.targets_total);
  if (status?.live && c) {
    setText(
      document.querySelector('[data-cluster]'),
      `lab ${healthy ? 'up' : 'degraded'} · ${c.targets_up}/${c.targets_total} targets · ${c.alerts_firing} alert${c.alerts_firing === 1 ? '' : 's'}`,
    );
  } else {
    setText(document.querySelector('[data-cluster]'), status ? 'backend up · cluster data pending' : 'unavailable');
  }
  paint('cluster', !status ? 'err' : healthy ? 'ok' : 'warn');

  // Directly-probed nodes.
  paint('portfolio', status ? 'ok' : 'err', true);
  paint('vm', status?.live ? 'ok' : 'warn', !!status?.live);
  paint('gitea', git?.live ? 'ok' : 'warn', !!git?.live);

  // The portfolio node wears its own request rate — the recursion, in a number.
  const rate = c ? `${c.request_rate.toFixed(2)} req/s` : '— req/s';
  for (const el of document.querySelectorAll('[data-rate], [data-status="portfolio"]')) setText(el, rate);

  // Gitea's outline status carries the latest commit when we have it.
  if (git?.commits?.length) {
    setText(document.querySelector('[data-status="gitea"]'), `${git.commits[0].sha} · latest`);
  }

  // Structural infra: green only when the whole cluster reports healthy.
  const infra = !status ? 'err' : healthy ? 'ok' : 'warn';
  for (const id of ['actions', 'ghcr', 'argocd']) paint(id, infra);
}

// ---- Detail dialog ----------------------------------------------------------
const dialog = document.querySelector('dialog.topo__panel');

function openNode(id) {
  const d = DETAIL[id];
  if (!d || !dialog) return;
  setText(dialog.querySelector('[data-panel-title]'), id);
  setText(dialog.querySelector('[data-panel-role]'), d.role);
  setText(dialog.querySelector('[data-panel-body]'), d.body);
  const link = dialog.querySelector('[data-panel-link]');
  if (link) {
    if (d.href) {
      link.href = d.href;
      link.textContent = d.hrefLabel ?? 'Open →';
      link.hidden = false;
      link.target = d.href.startsWith('http') ? '_blank' : '_self';
    } else {
      link.hidden = true;
    }
  }
  dialog.showModal();
}

for (const n of document.querySelectorAll('[data-node]')) {
  n.addEventListener('click', () => openNode(n.dataset.node));
  n.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      openNode(n.dataset.node);
    }
  });
}
dialog?.querySelector('[data-panel-close]')?.addEventListener('click', () => dialog.close());
// Click on the backdrop (outside the article) closes too.
dialog?.addEventListener('click', (e) => {
  if (e.target === dialog) dialog.close();
});

void tick();
setInterval(() => void tick(), INTERVAL_MS);
