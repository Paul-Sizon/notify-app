// Relative-time + day-bucket helpers. Mirror SignalsView grouping in design.

export function relTime(iso: string | null | undefined, now = Date.now()): string {
  if (!iso) return '—';
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return '—';
  const diff = Math.max(0, now - t);
  const m = Math.round(diff / 60_000);
  if (m < 1)    return 'just now';
  if (m < 60)   return `${m}m ago`;
  const h = Math.round(m / 60);
  if (h < 24)   return `${h}h ago`;
  const d = Math.round(h / 24);
  if (d === 1)  return 'Yesterday';
  if (d < 7)    return `${d}d ago`;
  if (d < 30)   return `${Math.round(d / 7)}w ago`;
  return new Date(t).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

export type DayBucket = 'Today' | 'Yesterday' | 'This week' | 'Earlier';

export function dayBucket(iso: string | null | undefined, now = new Date()): DayBucket {
  if (!iso) return 'Earlier';
  const t = new Date(iso);
  if (Number.isNaN(t.getTime())) return 'Earlier';
  const startOfDay = (d: Date) => {
    const x = new Date(d); x.setHours(0, 0, 0, 0); return x.getTime();
  };
  const today = startOfDay(now);
  const tDay  = startOfDay(t);
  const dayMs = 86_400_000;
  if (tDay === today)            return 'Today';
  if (tDay === today - dayMs)    return 'Yesterday';
  if (tDay >= today - 6 * dayMs) return 'This week';
  return 'Earlier';
}

export const BUCKET_ORDER: DayBucket[] = ['Today', 'Yesterday', 'This week', 'Earlier'];

export function host(url: string | null | undefined, fallback = '') {
  if (!url) return fallback;
  try { return new URL(url).host.replace(/^www\./, ''); }
  catch { return fallback; }
}
