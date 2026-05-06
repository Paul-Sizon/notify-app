import { Icon } from '../primitives';
import type { Theme } from '../tokens';
import type { EnrichedSub } from '../state';
import { BUCKET_ORDER, dayBucket, host, relTime, type DayBucket } from '../time';

interface AlertRow {
  id: string;
  query: string;
  title: string;
  source: string;
  time: string;
  bucket: DayBucket;
  unread: boolean;
}

export function AlertsScreen({ theme, subs }: { theme: Theme; subs: EnrichedSub[] }) {
  const all: AlertRow[] = [];
  for (const s of subs) {
    for (const sg of s.signals) {
      all.push({
        id: sg.id,
        query: s.query,
        title: sg.title,
        source: host(sg.url, sg.source_domains[0] ?? ''),
        time: relTime(sg.first_seen_at),
        bucket: dayBucket(sg.first_seen_at),
        unread: s.unread,
      });
    }
  }
  all.sort((a, b) => BUCKET_ORDER.indexOf(a.bucket) - BUCKET_ORDER.indexOf(b.bucket));
  const groups: Partial<Record<DayBucket, AlertRow[]>> = {};
  for (const a of all) (groups[a.bucket] ??= []).push(a);
  const order = BUCKET_ORDER.filter(b => groups[b]?.length);

  return (
    <div style={{ height: '100%', overflow: 'auto', paddingBottom: 140 }}>
      <div style={{ padding: '60px 20px 20px' }}>
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 10 }}>
          <h1 style={{
            fontSize: 34, fontWeight: 700, letterSpacing: -1.0,
            color: theme.label, margin: 0, lineHeight: 1.05,
          }}>Alerts</h1>
          <span style={{
            fontSize: 17, fontWeight: 500, color: theme.label3,
            fontVariantNumeric: 'tabular-nums',
          }}>{all.length}</span>
        </div>
        <div style={{ fontSize: 13, color: theme.label3, marginTop: 6 }}>
          Every signal the agent surfaced.
        </div>
      </div>

      {all.length === 0 && (
        <div style={{
          padding: '40px 20px', textAlign: 'center',
          color: theme.label3, fontSize: 14,
        }}>No alerts yet.</div>
      )}

      {order.map(g => (
        <div key={g} style={{ marginBottom: 24 }}>
          <div style={{
            padding: '0 22px 8px', fontSize: 11, fontWeight: 600,
            color: theme.label3, letterSpacing: 0.6, textTransform: 'uppercase',
          }}>{g}</div>
          <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: 6 }}>
            {groups[g]!.map(a => (
              <div key={a.id} style={{
                background: theme.bgElev2, borderRadius: 14,
                padding: '12px 14px', position: 'relative',
              }}>
                {a.unread && (
                  <div style={{
                    position: 'absolute', left: -4, top: 18,
                    width: 6, height: 6, borderRadius: 999,
                    background: theme.accent,
                    boxShadow: `0 0 10px ${theme.accentGlow}`,
                  }}/>
                )}
                <div style={{
                  display: 'flex', alignItems: 'baseline', justifyContent: 'space-between',
                  gap: 8, marginBottom: 4,
                }}>
                  <div style={{
                    fontSize: 11, fontWeight: 600, color: theme.accent,
                    letterSpacing: 0.3, textTransform: 'uppercase',
                    overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
                  }}>{a.query}</div>
                  <div style={{ fontSize: 11, color: theme.label3, flexShrink: 0 }}>{a.time}</div>
                </div>
                <div style={{
                  fontSize: 14, color: theme.label, fontWeight: 500,
                  letterSpacing: -0.2, lineHeight: 1.3, marginBottom: 4,
                }}>{a.title}</div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <Icon name="globe" size={10} color={theme.label3}/>
                  <span style={{ fontSize: 11, color: theme.label3 }}>{a.source || '—'}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
