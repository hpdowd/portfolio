// lib.js — helpers shared by the live front-end modules (home.js, status.js).
//
// A same-origin ES module the others import, so the CSP stays script-src 'self'.
// Only the genuinely common primitives live here; page-specific rendering stays
// in each page's own module.

/** querySelector shorthand, scoped to a root (defaults to document). */
export const q = (sel, root = document) => root.querySelector(sel);

/** Fetch same-origin JSON, returning null on any network or HTTP error. */
export async function getJSON(path) {
  try {
    const res = await fetch(path, { headers: { Accept: 'application/json' } });
    return res.ok ? await res.json() : null;
  } catch {
    return null;
  }
}

/** Write text only on change, so aria-live regions don't re-announce. */
export function setText(el, next) {
  if (el && el.textContent !== next) el.textContent = next;
}

/**
 * Set a status class (`${prefix}-${state}`) on an element, clearing the other
 * states first. Used for the pulsing health dots. `states` lists every state
 * class the prefix can take so stale ones are removed.
 */
export function applyDotState(el, prefix, state, states) {
  if (!el) return;
  for (const s of states) el.classList.remove(`${prefix}-${s}`);
  el.classList.add(`${prefix}-${state}`);
}
