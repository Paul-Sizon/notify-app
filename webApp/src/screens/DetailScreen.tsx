import { useState } from 'react';
import { Chip, Icon, WaveGlyph } from '../primitives';
import type { Theme } from '../tokens';
import type { EnrichedSub } from '../state';
import { cadenceLabel } from '../api';
import { BUCKET_ORDER, dayBucket, host, relTime, type DayBucket } from '../time';
import type { SignalDTO } from '../api';

export function DetailScreen({
  theme, sub, onBack, onRun, refreshing,
}: {
  theme: Theme;
  sub: EnrichedSub;
  onBack: () => void;
  onRun: () => void;
  refreshing: boolean;
}) {
  const groups: Partial<Record<DayBucket, SignalDTO[]>> = {};
  for (const sg of sub.signals) {
    const b = dayBucket(sg.first_seen_at);
    (groups[b] ??= []).push(sg);
  }
  const order = BUCKET_ORDER.filter(b => groups[b]?.length);

  return (
    <div style={{ height: '100%', overflow: 'auto', paddingBottom: 80 }}>
      <div style={{
        padding: '54px 12px 0', display: 'flex',
        alignItems: 'center', justifyContent: 'space-between',
      }}>
        <div onClick={onBack} style={{
          display: 'flex', alignItems: 'center', gap: 2,
          color: theme.accent, fontSize: 17, padding: 8, cursor: 'pointer',
        }}>
          <Icon name="chevron-left" size={20} color={theme.accent} weight="medium"/>
          Watching
        </div>
        <div style={{ display: 'flex', gap: 4 }}>
          <div onClick={onRun} style={{
            width: 36, height: 36, borderRadius: 999,
            background: theme.bgElev2,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            cursor: 'pointer',
            opacity: refreshing ? 0.5 : 1,
          }}>
            <Icon name="arrow-clockwise" size={16} color={theme.accent}/>
          </div>
        </div>
      </div>

      <div style={{ padding: '12px 22px 24px' }}>
        <h1 style={{
          fontSize: 30, fontWeight: 700, letterSpacing: -0.8,
          color: theme.label, margin: 0, lineHeight: 1.1,
        }}>{sub.query}</h1>
        <div style={{
          marginTop: 10, display: 'flex', alignItems: 'center', gap: 10,
          fontSize: 13, color: theme.label3,
        }}>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
            <Icon name={sub.type === 'event' ? 'calendar' : 'newspaper'} size={11} color={theme.label3}/>
            {sub.type}
          </span>
          <span style={{ width: 2, height: 2, borderRadius: 999, background: theme.label4 }}/>
          <span>every {cadenceLabel(sub.cadence_seconds)}</span>
          <span style={{ width: 2, height: 2, borderRadius: 999, background: theme.label4 }}/>
          <span>{sub.signals.length} signals</span>
        </div>
        <div style={{ marginTop: 14, height: 1, background: theme.hairline }}/>
        <div style={{
          marginTop: 10, fontSize: 11, color: theme.label3,
          letterSpacing: 0.4, display: 'flex', alignItems: 'center', gap: 6,
        }}>
          <WaveGlyph color={theme.label3} size={9} animated={refreshing}/>
          {refreshing ? 'Checking now…' : `Last checked ${sub.lastRunRel}`}
        </div>
      </div>

      {sub.signals.length === 0 && !refreshing && (
        <div style={{
          padding: '20px 22px', fontSize: 14, color: theme.label3, textAlign: 'center',
        }}>No signals yet. Tap refresh to run the agent now.</div>
      )}

      {order.map(g => (
        <div key={g} style={{ marginBottom: 24 }}>
          <div style={{
            padding: '0 22px 8px', fontSize: 11, fontWeight: 600,
            color: theme.label3, letterSpacing: 0.6, textTransform: 'uppercase',
          }}>{g}</div>
          <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: 6 }}>
            {groups[g]!.map(sg => <SignalCell key={sg.id} signal={sg} theme={theme}/>)}
          </div>
        </div>
      ))}
    </div>
  );
}

function SignalCell({ signal, theme }: { signal: SignalDTO; theme: Theme }) {
  const [pressed, setPressed] = useState(false);
  const sourceLabel = host(signal.url, signal.source_domains[0] ?? '');
  const open = () => signal.url && window.open(signal.url, '_blank', 'noopener,noreferrer');
  return (
    <div
      onClick={open}
      onPointerDown={() => setPressed(true)}
      onPointerUp={() => setPressed(false)}
      onPointerLeave={() => setPressed(false)}
      style={{
        padding: '14px 16px',
        background: theme.bgElev2,
        borderRadius: 14,
        cursor: signal.url ? 'pointer' : 'default',
        position: 'relative',
        transform: pressed ? 'scale(0.985)' : 'scale(1)',
        transition: 'transform 0.18s cubic-bezier(0.32,0.72,0,1)',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10 }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{
            fontSize: 16, fontWeight: 600, color: theme.label,
            letterSpacing: -0.3, lineHeight: 1.3, marginBottom: 6,
          }}>{signal.title}</div>
          {signal.body && (
            <div style={{
              fontSize: 14, color: theme.label2, lineHeight: 1.4,
              marginBottom: 10,
            }}>{signal.body}</div>
          )}
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Chip theme={theme} small>
              <Icon name="globe" size={9} color={theme.label2}/> {sourceLabel || '—'}
            </Chip>
            <span style={{ fontSize: 11, color: theme.label3 }}>{relTime(signal.first_seen_at)}</span>
          </div>
        </div>
        {signal.url && <Icon name="arrow-up-right" size={14} color={theme.label4} style={{ marginTop: 4 }}/>}
      </div>
    </div>
  );
}
