// Mirror of the Go version comparator. Compare the upstream version first,
// then an optional Fourgetu revision; preserve the official dev-channel rules.
function parseVersionParts(version: string): [number, number, number, number] | null {
  const match = version
    .trim()
    .replace(/^v/, '')
    .match(/^(\d+\.\d+\.\d+)(?:-fourgetu\.([1-9]\d*))?$/);
  if (!match) return null;
  const parts = match[1].split('.');
  if (parts.length !== 3) return null;
  const out: number[] = [];
  for (const part of parts) {
    if (!/^\d+$/.test(part)) return null;
    out.push(Number(part));
  }
  return [out[0], out[1], out[2], Number(match[2] ?? 0)];
}

// Format a panel version for display. Dev builds report a "dev+<commit>"
// identity (see config.GetPanelVersion); show those — and any other
// non-numeric label — verbatim. Semantic versions get a single normalized "v"
// prefix, so a raw "v3.4.0" tag and a bare "3.4.0" both render as "v3.4.0"
// instead of doubling up to "vv3.4.0".
export function formatPanelVersion(version: string | undefined | null): string {
  const v = (version || '').trim();
  if (!v) return '';
  const normalized = v.replace(/^v/i, '');
  return /^\d/.test(normalized) ? `v${normalized}` : v;
}

export function isPanelUpdateAvailable(latest: string, current: string): boolean {
  if (!latest || !current) return false;
  // A dev+<sha> label and a release tag sit on different channels and carry no
  // order, so a node moved to the other channel is not "behind" the master's latest.
  if (latest.trim().startsWith('dev+') !== current.trim().startsWith('dev+')) return false;
  const a = parseVersionParts(latest);
  const b = parseVersionParts(current);
  if (!a || !b) {
    return latest.trim().replace(/^v/, '') !== current.trim().replace(/^v/, '');
  }
  for (let i = 0; i < a.length; i++) {
    if (a[i] > b[i]) return true;
    if (a[i] < b[i]) return false;
  }
  return false;
}
