// Typed fetch helpers for the same-origin /api/* endpoints served by the Go
// backend.
//
// Every call fails soft (resolves to null) so the page degrades gracefully when
// the backend — or one of its upstreams — is unavailable. The components treat
// null as "show the unavailable state" rather than throwing.

export interface Status {
  service: { uptime_seconds: number; version: string };
  cluster: {
    targets_up: number;
    targets_total: number;
    alerts_firing: number;
    request_rate: number;
  } | null;
  live: boolean;
}

export interface Commit {
  sha: string;
  summary: string;
  author: string;
  url: string;
  when: string;
}

export interface GitFeed {
  commits: Commit[];
  live: boolean;
}

async function getJSON<T>(path: string): Promise<T | null> {
  try {
    const res = await fetch(path, { headers: { Accept: 'application/json' } });
    if (!res.ok) return null;
    return (await res.json()) as T;
  } catch {
    return null;
  }
}

export const getStatus = () => getJSON<Status>('/api/status');
export const getGit = () => getJSON<GitFeed>('/api/git');
