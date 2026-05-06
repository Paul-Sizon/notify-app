import React, { useRef, useState } from 'react';
import { Chip, Icon, WaveGlyph } from '../primitives';
import type { Theme } from '../tokens';
import type { EnrichedSub } from '../state';
import { cadenceLabel } from '../api';

export function HomeScreen({
  theme, subs, onOpen, onAdd, onAI, onDelete, refreshing,
}: {
  theme: Theme;
  subs: EnrichedSub[];
  onOpen: (s: EnrichedSub) => void;
  onAdd: () => void;
  onAI: () => void;
  onDelete: (id: string) => void;
  refreshing: boolean;
}) {
  if (subs.length === 0) {
    return (
      <div style={{
        height: '100%', display: 'flex', flexDirection: 'column',
        alignItems: 'center', justifyContent: 'center', padding: 32, gap: 24,
      }}>
        <div style={{ position: 'relative' }}>
          <div style={{
            width: 80, height: 80, borderRadius: 999,
            border: `1px solid ${theme.hairline}`,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>
            <Icon name="antenna" size={36} color={theme.label3}/>
          </div>
          <div style={{
            position: 'absolute', inset: -8, borderRadius: 999,
            border: `1px solid ${theme.hairline}`, opacity: 0.5, pointerEvents: 'none',
          }}/>
        </div>
        <div style={{ textAlign: 'center', maxWidth: 280 }}>
          <div style={{
            fontSize: 22, fontWeight: 600, color: theme.label,
            letterSpacing: -0.4, marginBottom: 8,
          }}>Nothing watched yet.</div>
          <div style={{ fontSize: 15, color: theme.label3, lineHeight: 1.4 }}>
            Add something specific. The narrower the query, the better the signals.
          </div>
        </div>
        <button onClick={onAdd} style={{
          padding: '12px 22px', borderRadius: 999,
          background: theme.accent, color: theme.mode === 'dark' ? '#000' : '#fff',
          border: 'none', fontSize: 15, fontWeight: 600,
          cursor: 'pointer',
          display: 'flex', alignItems: 'center', gap: 6,
        }}>
          <Icon name="plus" size={16} weight="bold" color={theme.mode === 'dark' ? '#000' : '#fff'}/>
          Add a watcher
        </button>
        <button onClick={onAI} style={{
          padding: '10px 18px', borderRadius: 999,
          background: 'transparent',
          color: theme.accent,
          border: `0.5px solid ${theme.accent}`,
          fontSize: 14, fontWeight: 600,
          cursor: 'pointer',
          display: 'flex', alignItems: 'center', gap: 6,
        }}>
          <Icon name="sparkle" size={14} weight="bold" color={theme.accent}/>
          Try AI suggestions
        </button>
      </div>
    );
  }

  const newCount = subs.filter(s => s.unread).length;

  return (
    <div style={{ height: '100%', overflow: 'auto', paddingBottom: 200 }}>
      <div style={{ padding: '60px 20px 20px' }}>
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 10 }}>
          <h1 style={{
            fontSize: 34, fontWeight: 700, letterSpacing: -1.0,
            color: theme.label, margin: 0, lineHeight: 1.05,
          }}>Watching</h1>
          <span style={{
            fontSize: 17, fontWeight: 500, color: theme.label3,
            fontVariantNumeric: 'tabular-nums',
          }}>{subs.length}</span>
          <div style={{ flex: 1 }}/>
          <button
            onClick={onAI}
            style={{
              display: 'flex', alignItems: 'center', gap: 6,
              padding: '7px 12px', borderRadius: 999,
              background: `${theme.accent}26`,
              border: `0.5px solid ${theme.accent}40`,
              color: theme.accent,
              fontSize: 13, fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            <Icon name="sparkle" size={12} weight="bold"/>
            AI
          </button>
        </div>
        <div style={{ fontSize: 13, color: theme.label3, marginTop: 6, letterSpacing: 0.1 }}>
          {newCount > 0 ? `${newCount} with new signals` : 'All caught up'}
        </div>
      </div>

      {refreshing && (
        <div style={{ padding: '0 20px 12px' }}>
          <div style={{
            padding: '10px 14px', display: 'flex', alignItems: 'center', gap: 10,
            background: theme.bgElev2, borderRadius: 14,
            border: `0.5px solid ${theme.hairline}`,
          }}>
            <WaveGlyph color={theme.accent} size={12} animated/>
            <div style={{ fontSize: 13, color: theme.label2, flex: 1 }}>
              Checking watchers…
            </div>
            <div style={{ fontSize: 11, color: theme.label3 }}>live</div>
          </div>
        </div>
      )}

      <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: 8 }}>
        {subs.map(s => (
          <SubscriptionCard key={s.id} sub={s} theme={theme} onOpen={onOpen} onDelete={onDelete}/>
        ))}
      </div>

      <div style={{
        textAlign: 'center', marginTop: 16, padding: '0 20px 20px',
        fontSize: 11, color: theme.label4, letterSpacing: 0.3,
        display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6,
      }}>
        <Icon name="shield" size={11} color={theme.label4}/>
        Tap a watcher to see signals
      </div>
    </div>
  );
}

function SubscriptionCard({
  sub, theme, onOpen, onDelete,
}: {
  sub: EnrichedSub;
  theme: Theme;
  onOpen: (s: EnrichedSub) => void;
  onDelete: (id: string) => void;
}) {
  const [pressed, setPressed] = useState(false);
  const [swipeX, setSwipeX] = useState(0);
  const startX = useRef<number | null>(null);

  const onPointerDown = (e: React.PointerEvent) => { startX.current = e.clientX; };
  const onPointerMove = (e: React.PointerEvent) => {
    if (startX.current == null) return;
    const dx = Math.min(0, e.clientX - startX.current);
    setSwipeX(Math.max(-100, dx));
  };
  const onPointerUp = () => {
    if (swipeX < -60) { onDelete(sub.id); setSwipeX(0); }
    else setSwipeX(0);
    startX.current = null;
  };

  const typeIcon = sub.type === 'event' ? 'calendar' : 'newspaper';

  return (
    <div style={{ position: 'relative', userSelect: 'none' }}>
      <div style={{
        position: 'absolute', inset: 0, display: 'flex',
        justifyContent: 'flex-end', alignItems: 'center',
        background: '#FF3B30', borderRadius: 16, paddingRight: 22,
      }}>
        <div style={{
          display: 'flex', alignItems: 'center', gap: 6,
          color: '#fff', fontSize: 13, fontWeight: 600,
        }}>
          <Icon name="trash" size={18} color="#fff"/> Delete
        </div>
      </div>
      <div
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={onPointerUp}
        onPointerCancel={() => { setSwipeX(0); startX.current = null; }}
        onClick={() => swipeX === 0 && onOpen(sub)}
        onMouseDown={() => setPressed(true)}
        onMouseUp={() => setPressed(false)}
        onMouseLeave={() => setPressed(false)}
        style={{
          position: 'relative',
          background: theme.bgElev2,
          borderRadius: 16,
          padding: '20px 16px',
          transform: `translateX(${swipeX}px) scale(${pressed ? 0.985 : 1})`,
          transition: startX.current == null ? 'transform 0.25s cubic-bezier(0.32,0.72,0,1)' : 'none',
          boxShadow: theme.mode === 'dark'
            ? '0 0 0 0.5px rgba(255,255,255,0.04)'
            : '0 0 0 0.5px rgba(0,0,0,0.04)',
          cursor: 'pointer',
        }}
      >
        {sub.unread && (
          <div style={{
            position: 'absolute', left: -4, top: 28,
            width: 6, height: 6, borderRadius: 999,
            background: theme.accent,
            boxShadow: `0 0 12px ${theme.accentGlow}`,
          }}/>
        )}
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{
              fontSize: 17, fontWeight: 600, color: theme.label,
              letterSpacing: -0.3, lineHeight: 1.25, marginBottom: 8,
            }}>{sub.query}</div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
              <Chip theme={theme} small>
                <Icon name={typeIcon} size={10} color={theme.label2}/> {sub.type}
              </Chip>
              <Chip theme={theme} small>every {cadenceLabel(sub.cadence_seconds)}</Chip>
              {sub.unread && (
                <Chip theme={theme} small accent>
                  <WaveGlyph color={theme.accent} size={9} animated/> {sub.newCount} new
                </Chip>
              )}
            </div>
          </div>
          <div style={{
            display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 4,
            flexShrink: 0, paddingTop: 2,
          }}>
            <div style={{
              fontSize: 12, color: theme.label3,
              fontVariantNumeric: 'tabular-nums',
            }}>{sub.lastRunRel}</div>
            <Icon name="chevron-right" size={14} color={theme.label4}/>
          </div>
        </div>
      </div>
    </div>
  );
}
