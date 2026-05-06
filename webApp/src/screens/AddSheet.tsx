import React, { useEffect, useRef, useState } from 'react';
import { GlassSurface, Icon } from '../primitives';
import type { Theme } from '../tokens';
import { CADENCE_SECONDS } from '../api';

type CadenceKey = keyof typeof CADENCE_SECONDS;
const CADENCES: CadenceKey[] = ['5m', '15m', '1h', '6h', '1d'];
const CADENCE_LABEL: Record<CadenceKey, string> = {
  '5m': 'every 5 minutes', '15m': 'every 15 minutes',
  '1h': 'every hour', '6h': 'every 6 hours', '1d': 'once a day',
};

const SUGGESTIONS = ['blockchain meetups in Lisbon', 'EU AI Act updates', 'Phoebe Bridgers tour'];

export function AddSheet({
  theme, open, onClose, onCreate, onAI,
}: {
  theme: Theme;
  open: boolean;
  onClose: () => void;
  onCreate: (q: string, type: 'event' | 'news', cadenceSeconds: number) => Promise<unknown>;
  onAI: () => void;
}) {
  const [query, setQuery] = useState('');
  const [type, setType] = useState<'event' | 'news'>('event');
  const [cadence, setCadence] = useState<CadenceKey>('1h');
  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      setQuery(''); setType('event'); setCadence('1h');
      setSubmitting(false); setSuccess(false);
      setTimeout(() => inputRef.current?.focus(), 280);
    }
  }, [open]);

  const valid = query.trim().length >= 3 && !submitting;

  const submit = async () => {
    if (!valid) return;
    setSubmitting(true);
    try {
      await onCreate(query.trim(), type, CADENCE_SECONDS[cadence]);
      setSuccess(true);
      setTimeout(() => { setSuccess(false); onClose(); }, 600);
    } catch {
      setSubmitting(false);
    }
  };

  const hint = !query
    ? 'Tell me what to keep an eye on. The more specific, the better.'
    : `Checks ${CADENCE_LABEL[cadence]} for new ${type === 'event' ? 'events' : 'news'} matching this query.`;

  return (
    <>
      <div onClick={onClose} style={{
        position: 'absolute', inset: 0, zIndex: 90,
        background: open ? 'rgba(0,0,0,0.4)' : 'transparent',
        backdropFilter: open ? 'blur(2px)' : 'none',
        WebkitBackdropFilter: open ? 'blur(2px)' : 'none',
        opacity: open ? 1 : 0, pointerEvents: open ? 'auto' : 'none',
        transition: 'opacity 0.32s cubic-bezier(0.32,0.72,0,1)',
      }}/>
      <div style={{
        position: 'absolute', left: 0, right: 0, bottom: 0, zIndex: 95,
        transform: open ? 'translateY(0)' : 'translateY(100%)',
        transition: 'transform 0.42s cubic-bezier(0.32,0.72,0,1)',
      }}>
        <GlassSurface theme={theme} radius={28} intense style={{
          borderTopLeftRadius: 28, borderTopRightRadius: 28,
          borderBottomLeftRadius: 0, borderBottomRightRadius: 0,
          minHeight: 480, paddingBottom: 28,
          background: theme.bgElev1,
        }}>
          <div style={{
            width: 36, height: 5, borderRadius: 999,
            background: theme.label4, margin: '8px auto 0',
          }}/>
          <div style={{ padding: '24px 22px 0' }}>
            <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', marginBottom: 22 }}>
              <h2 style={{
                fontSize: 26, fontWeight: 700, letterSpacing: -0.6,
                color: theme.label, margin: 0,
              }}>New watcher</h2>
              <div onClick={onClose} style={{
                width: 30, height: 30, borderRadius: 999,
                background: theme.bgElev3,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                cursor: 'pointer',
              }}>
                <Icon name="x" size={14} color={theme.label2} weight="medium"/>
              </div>
            </div>

            <div style={{
              borderBottom: `0.5px solid ${theme.separator}`,
              paddingBottom: 18, marginBottom: 22,
            }}>
              <input
                ref={inputRef}
                value={query}
                onChange={e => setQuery(e.target.value)}
                onKeyDown={e => { if (e.key === 'Enter') submit(); }}
                placeholder="What should I watch?"
                style={{
                  width: '100%', border: 'none', outline: 'none',
                  background: 'transparent',
                  fontFamily: 'inherit',
                  fontSize: 22, fontWeight: 500, letterSpacing: -0.4,
                  color: theme.label,
                  caretColor: theme.accent,
                  padding: 0,
                }}
              />
              <div style={{ marginTop: 10, display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                {SUGGESTIONS.map(s => (
                  <div key={s} onClick={() => setQuery(s)} style={{
                    fontSize: 12, color: theme.label3,
                    padding: '5px 10px', borderRadius: 999,
                    background: theme.chipBg, cursor: 'pointer',
                  }}>{s}</div>
                ))}
              </div>
            </div>

            <SectionLabel theme={theme}>Type</SectionLabel>
            <div style={{
              display: 'flex', padding: 3, marginBottom: 20,
              background: theme.bgElev3, borderRadius: 10,
              position: 'relative',
            }}>
              <div style={{
                position: 'absolute', top: 3, bottom: 3,
                left: type === 'event' ? 3 : '50%',
                width: 'calc(50% - 3px)',
                background: theme.bgElev2, borderRadius: 8,
                boxShadow: '0 1px 2px rgba(0,0,0,0.1)',
                transition: 'left 0.28s cubic-bezier(0.32,0.72,0,1)',
              }}/>
              {([
                { id: 'event', icon: 'calendar' as const, label: 'Event' },
                { id: 'news',  icon: 'newspaper' as const, label: 'News' },
              ] as const).map(opt => (
                <div key={opt.id} onClick={() => setType(opt.id)} style={{
                  flex: 1, padding: '8px 0',
                  display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6,
                  fontSize: 14, fontWeight: 600,
                  color: type === opt.id ? theme.label : theme.label3,
                  position: 'relative', zIndex: 1, cursor: 'pointer',
                  transition: 'color 0.2s',
                }}>
                  <Icon name={opt.icon} size={14} color={type === opt.id ? theme.label : theme.label3}/>
                  {opt.label}
                </div>
              ))}
            </div>

            <SectionLabel theme={theme}>Cadence</SectionLabel>
            <div style={{ display: 'flex', gap: 6, marginBottom: 22 }}>
              {CADENCES.map(c => {
                const sel = cadence === c;
                return (
                  <div key={c} onClick={() => setCadence(c)} style={{
                    flex: 1, padding: '10px 0',
                    borderRadius: 999,
                    background: sel ? theme.accent : theme.bgElev3,
                    color: sel ? (theme.mode === 'dark' ? '#000' : '#fff') : theme.label2,
                    fontSize: 14, fontWeight: 600,
                    textAlign: 'center', cursor: 'pointer',
                    transition: 'background 0.2s, transform 0.15s',
                    transform: sel ? 'scale(1.0)' : 'scale(0.98)',
                    letterSpacing: 0.2,
                  }}>{c}</div>
                );
              })}
            </div>

            <div style={{
              fontSize: 13, color: theme.label3, lineHeight: 1.4,
              padding: '0 2px', marginBottom: 28, minHeight: 38,
            }}>{hint}</div>

            <button
              onClick={onAI}
              style={{
                width: '100%', padding: '10px 12px', marginBottom: 10,
                borderRadius: 12, border: 'none',
                background: 'transparent',
                color: theme.label2,
                fontSize: 13,
                cursor: 'pointer',
                display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
              }}
            >
              <Icon name="sparkle" size={13} weight="bold" color={theme.accent}/>
              <span>Need suggestions? <span style={{ color: theme.accent, fontWeight: 600 }}>Try AI</span></span>
            </button>

            <button onClick={submit} disabled={!valid} style={{
              width: '100%', padding: '16px',
              borderRadius: 14, border: 'none',
              background: valid ? theme.accent : theme.bgElev3,
              color: valid ? (theme.mode === 'dark' ? '#000' : '#fff') : theme.label3,
              fontSize: 16, fontWeight: 600, letterSpacing: 0.2,
              cursor: valid ? 'pointer' : 'default',
              display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
              transition: 'all 0.2s',
            }}>
              {success ? (
                <>
                  <Icon name="check" size={18} weight="bold"
                    color={theme.mode === 'dark' ? '#000' : '#fff'}/>
                  Watching
                </>
              ) : (
                <>
                  <Icon name="eye" size={16}
                    color={valid ? (theme.mode === 'dark' ? '#000' : '#fff') : theme.label3}/>
                  {submitting ? 'Creating…' : 'Watch'}
                </>
              )}
            </button>
          </div>
        </GlassSurface>
      </div>
    </>
  );
}

function SectionLabel({ theme, children }: { theme: Theme; children: React.ReactNode }) {
  return (
    <div style={{
      fontSize: 11, fontWeight: 600, color: theme.label3,
      letterSpacing: 0.6, textTransform: 'uppercase', marginBottom: 8,
    }}>{children}</div>
  );
}
