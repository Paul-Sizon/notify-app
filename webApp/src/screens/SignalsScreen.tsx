import { Chip, Icon } from '../primitives';
import type { Theme } from '../tokens';
import type { EnrichedSub } from '../state';
import { dayBucket, host, relTime } from '../time';

export function SignalsScreen({ theme, subs }: { theme: Theme; subs: EnrichedSub[] }) {
  const all = subs.flatMap(s => s.signals.map(sg => ({
    ...sg,
    query: s.query,
    bucket: dayBucket(sg.first_seen_at),
    timeRel: relTime(sg.first_seen_at),
    sourceLabel: host(sg.url, sg.source_domains[0] ?? ''),
  })));
  all.sort((a, b) => Date.parse(b.first_seen_at) - Date.parse(a.first_seen_at));

  const stats = {
    today:   all.filter(a => a.bucket === 'Today').length,
    week:    all.filter(a => a.bucket === 'Today' || a.bucket === 'Yesterday' || a.bucket === 'This week').length,
    sources: new Set(all.map(a => a.sourceLabel).filter(Boolean)).size,
  };

  // 14-day spark — bucket signals into days from today backwards
  const buckets14 = new Array(14).fill(0);
  const dayMs = 86_400_000;
  const todayStart = new Date(); todayStart.setHours(0, 0, 0, 0);
  for (const a of all) {
    const t = Date.parse(a.first_seen_at);
    const d = Math.floor((todayStart.getTime() - t) / dayMs);
    if (d >= 0 && d < 14) buckets14[13 - d] += 1;
  }
  const max = Math.max(1, ...buckets14);

  return (
    <div style={{ height: '100%', overflow: 'auto', paddingBottom: 140 }}>
      <div style={{ padding: '60px 20px 16px' }}>
        <h1 style={{
          fontSize: 34, fontWeight: 700, letterSpacing: -1.0,
          color: theme.label, margin: 0, lineHeight: 1.05,
        }}>Signals</h1>
        <div style={{ fontSize: 13, color: theme.label3, marginTop: 6 }}>
          The firehose. Everything the agent surfaced.
        </div>
      </div>

      <div style={{
        margin: '0 16px 22px', padding: '14px 16px',
        background: theme.bgElev2, borderRadius: 14,
        display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10,
      }}>
        {[
          { label: 'today',     value: stats.today },
          { label: 'this week', value: stats.week },
          { label: 'sources',   value: stats.sources },
        ].map((s, i) => (
          <div key={s.label} style={{
            paddingLeft: i === 0 ? 0 : 12,
            borderLeft: i === 0 ? 'none' : `0.5px solid ${theme.separator}`,
          }}>
            <div style={{
              fontSize: 28, fontWeight: 700, color: theme.label,
              letterSpacing: -0.6, fontVariantNumeric: 'tabular-nums',
              lineHeight: 1, marginBottom: 4,
            }}>{s.value}</div>
            <div style={{
              fontSize: 11, color: theme.label3,
              letterSpacing: 0.4, textTransform: 'uppercase', fontWeight: 600,
            }}>{s.label}</div>
          </div>
        ))}
      </div>

      <div style={{
        margin: '0 16px 28px', padding: '14px 16px',
        background: theme.bgElev2, borderRadius: 14,
      }}>
        <div style={{
          fontSize: 11, fontWeight: 600, color: theme.label3,
          letterSpacing: 0.4, textTransform: 'uppercase', marginBottom: 12,
          display: 'flex', justifyContent: 'space-between',
        }}>
          <span>14-day activity</span>
          <span style={{ color: theme.label4 }}>signals/day</span>
        </div>
        <div style={{ display: 'flex', alignItems: 'flex-end', gap: 4, height: 48 }}>
          {buckets14.map((v, i) => {
            const h = Math.max(2, (v / max) * 48);
            const today = i === 13;
            return (
              <div key={i} style={{
                flex: 1, height: h, borderRadius: 2,
                background: today ? theme.accent : theme.bgElev3,
                boxShadow: today ? `0 0 10px ${theme.accentGlow}` : 'none',
              }}/>
            );
          })}
        </div>
      </div>

      <div style={{
        padding: '0 22px 8px', fontSize: 11, fontWeight: 600,
        color: theme.label3, letterSpacing: 0.6, textTransform: 'uppercase',
      }}>Latest</div>
      <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: 6 }}>
        {all.slice(0, 12).map(s => (
          <div key={s.id} style={{
            background: theme.bgElev2, borderRadius: 14,
            padding: '12px 14px',
          }}>
            <div style={{
              fontSize: 11, fontWeight: 600, color: theme.label3,
              letterSpacing: 0.3, textTransform: 'uppercase', marginBottom: 4,
              overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
            }}>{s.query}</div>
            <div style={{
              fontSize: 14, color: theme.label, fontWeight: 500,
              letterSpacing: -0.2, lineHeight: 1.3, marginBottom: 6,
            }}>{s.title}</div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Chip theme={theme} small>
                <Icon name="globe" size={9} color={theme.label2}/> {s.sourceLabel || '—'}
              </Chip>
              <span style={{ fontSize: 11, color: theme.label3 }}>{s.timeRel}</span>
            </div>
          </div>
        ))}
        {all.length === 0 && (
          <div style={{
            padding: '40px 0', textAlign: 'center',
            color: theme.label3, fontSize: 14,
          }}>No signals yet. Add a watcher.</div>
        )}
      </div>
    </div>
  );
}
