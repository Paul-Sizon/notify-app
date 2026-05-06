import { useEffect, useState } from 'react';
import { GlassSurface, Icon } from '../primitives';
import type { Theme } from '../tokens';

const PHASES = [
  { id: 'searching', name: 'Searching',       sub: 'querying sources',           ms: 1400 },
  { id: 'reading',   name: 'Reading',         sub: 'fetching pages',             ms: 1800 },
  { id: 'verifying', name: 'Verifying',       sub: 'cross-referencing sources',  ms: 1500 },
  { id: 'filtering', name: 'Filtering noise', sub: 'discarding known items',     ms: 1100 },
  { id: 'done',      name: 'Done',            sub: 'wrapping up',                ms: 0 },
] as const;

export function LiveAgent({
  theme, open, onClose,
}: { theme: Theme; open: boolean; onClose: () => void }) {
  const phaseIndex = useAgentPhase(open);
  const done = phaseIndex >= PHASES.length - 1;

  return (
    <>
      <div onClick={done ? onClose : undefined} style={{
        position: 'absolute', inset: 0, zIndex: 90,
        background: open ? 'rgba(0,0,0,0.5)' : 'transparent',
        backdropFilter: open ? 'blur(8px)' : 'none',
        WebkitBackdropFilter: open ? 'blur(8px)' : 'none',
        opacity: open ? 1 : 0, pointerEvents: open ? 'auto' : 'none',
        transition: 'opacity 0.4s cubic-bezier(0.32,0.72,0,1)',
      }}/>
      <div style={{
        position: 'absolute', left: 0, right: 0, bottom: 0, zIndex: 95,
        transform: open ? 'translateY(0)' : 'translateY(100%)',
        transition: 'transform 0.5s cubic-bezier(0.32,0.72,0,1)',
      }}>
        <GlassSurface theme={theme} radius={28} intense style={{
          borderTopLeftRadius: 28, borderTopRightRadius: 28,
          borderBottomLeftRadius: 0, borderBottomRightRadius: 0,
          paddingBottom: 22,
          background: theme.mode === 'dark' ? 'rgba(20,20,22,0.92)' : 'rgba(250,250,252,0.92)',
        }}>
          <div style={{
            width: 36, height: 5, borderRadius: 999,
            background: theme.label4, margin: '8px auto 4px',
          }}/>
          <div style={{
            padding: '14px 22px 4px',
            display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Icon name="sparkle" size={16} color={theme.accent}/>
              <div style={{
                fontSize: 11, fontWeight: 700, color: theme.label2,
                letterSpacing: 1.2, textTransform: 'uppercase',
              }}>Agent · live</div>
            </div>
            <div onClick={onClose} style={{
              width: 26, height: 26, borderRadius: 999,
              background: theme.bgElev3,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              cursor: 'pointer',
            }}>
              <Icon name="x" size={11} color={theme.label2} weight="medium"/>
            </div>
          </div>

          <div style={{ padding: '8px 22px 22px' }}>
            {PHASES.map((p, i) => {
              const status = i < phaseIndex ? 'done' : i === phaseIndex ? 'active' : 'pending';
              return (
                <div key={p.id} style={{
                  display: 'flex', alignItems: 'flex-start', gap: 14,
                  padding: '14px 0',
                  borderBottom: i < PHASES.length - 1 ? `0.5px solid ${theme.hairline}` : 'none',
                  opacity: status === 'done' ? 0.55 : 1,
                  transition: 'opacity 0.4s',
                }}>
                  <div style={{
                    width: 26, height: 26, borderRadius: 999,
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    flexShrink: 0, marginTop: 2,
                  }}>
                    {status === 'pending' && (
                      <div style={{
                        width: 8, height: 8, borderRadius: 999, background: theme.label4,
                      }}/>
                    )}
                    {status === 'active' && (
                      <div style={{ position: 'relative', width: 26, height: 26 }}>
                        <div style={{
                          position: 'absolute', inset: 0, borderRadius: 999,
                          border: `2px solid ${theme.accent}`,
                          animation: 'pulse-ring 1.6s ease-out infinite',
                        }}/>
                        <div style={{
                          position: 'absolute', inset: 8, borderRadius: 999,
                          background: theme.accent,
                          boxShadow: `0 0 16px ${theme.accentGlow}`,
                          animation: 'pulse-dot 1.2s ease-in-out infinite',
                        }}/>
                      </div>
                    )}
                    {status === 'done' && (
                      <div style={{
                        width: 22, height: 22, borderRadius: 999,
                        background: theme.accent,
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                      }}>
                        <Icon name="check" size={12} weight="bold"
                          color={theme.mode === 'dark' ? '#000' : '#fff'}/>
                      </div>
                    )}
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{
                      fontSize: 16, fontWeight: 600, color: theme.label,
                      letterSpacing: -0.2,
                    }}>{p.name}</div>
                    <div style={{ fontSize: 12, color: theme.label3, marginTop: 2 }}>{p.sub}</div>
                  </div>
                </div>
              );
            })}
          </div>

          <div style={{ padding: '0 22px' }}>
            {done ? (
              <button onClick={onClose} style={{
                width: '100%', padding: '15px',
                borderRadius: 14, border: 'none',
                background: theme.accent,
                color: theme.mode === 'dark' ? '#000' : '#fff',
                fontSize: 16, fontWeight: 600, fontFamily: 'inherit',
                cursor: 'pointer', letterSpacing: 0.2,
                display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
              }}>
                <Icon name="check" size={16} weight="bold"
                  color={theme.mode === 'dark' ? '#000' : '#fff'}/>
                Done
              </button>
            ) : (
              <div style={{
                textAlign: 'center', fontSize: 11, color: theme.label3,
                letterSpacing: 0.4, padding: '8px 0',
              }}>Working… this closes itself when done.</div>
            )}
          </div>
        </GlassSurface>
      </div>
    </>
  );
}

function useAgentPhase(active: boolean): number {
  const [phaseIndex, setPhaseIndex] = useState(0);
  useEffect(() => {
    if (!active) { setPhaseIndex(0); return; }
    setPhaseIndex(0);
    const timers: number[] = [];
    let t = 0;
    PHASES.forEach((_, i) => {
      t += i === 0 ? 200 : PHASES[i - 1].ms;
      timers.push(window.setTimeout(() => setPhaseIndex(i), t));
    });
    return () => timers.forEach(clearTimeout);
  }, [active]);
  return phaseIndex;
}
