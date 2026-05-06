// LiveAgentView — three treatment variations
// 1. 'phases'   — phase rows with status indicators (the brief)
// 2. 'wave'     — animated waveform/oscilloscope as the focus
// 3. 'terminal' — console stream of what the agent is reading

const AGENT_PHASES = [
  { id: 'searching',  name: 'Searching',       sub: 'brave.com → results',                         duration: 1400 },
  { id: 'reading',    name: 'Reading',         sub: '12 pages',                                    duration: 1800 },
  { id: 'verifying',  name: 'Verifying',       sub: 'cross-referencing 4 sources',                 duration: 1500 },
  { id: 'filtering',  name: 'Filtering noise', sub: 'discarded 23 known items',                    duration: 1100 },
  { id: 'done',       name: 'Done',            sub: '2 new signals',                               duration: 0 },
];

function useAgentSimulation(active) {
  const [phaseIndex, setPhaseIndex] = React.useState(0);
  const [logs, setLogs] = React.useState([]);

  React.useEffect(() => {
    if (!active) { setPhaseIndex(0); setLogs([]); return; }
    setPhaseIndex(0); setLogs([]);
    let t = 0;
    const timers = [];
    const logLines = [
      { t: 200,  text: 'GET https://search.brave.com/?q=blockchain+meetups+curitiba', tag: 'http' },
      { t: 600,  text: '→ 18 results returned', tag: 'ok' },
      { t: 1100, text: 'fetch lu.ma/eth-curitiba-jun', tag: 'http' },
      { t: 1500, text: 'fetch web3brasil.org/regions/pr', tag: 'http' },
      { t: 1900, text: 'extract: 4 events, 2 articles', tag: 'ok' },
      { t: 2400, text: 'compare against 47 known items', tag: 'check' },
      { t: 2900, text: 'similarity 0.94 — flag duplicate', tag: 'warn' },
      { t: 3400, text: 'verify date, location, status', tag: 'check' },
      { t: 3900, text: 'cross-ref tezos.foundation', tag: 'http' },
      { t: 4400, text: 'discard: 23 cached items', tag: 'discard' },
      { t: 5100, text: '+ ETH Curitiba — June Builders Night', tag: 'signal' },
      { t: 5500, text: '+ Web3 Brasil South Chapter', tag: 'signal' },
      { t: 5900, text: 'done. 2 new signals.', tag: 'done' },
    ];
    logLines.forEach(l => {
      timers.push(setTimeout(() => {
        setLogs(prev => [...prev, { ...l, id: Math.random() }]);
      }, l.t));
    });

    AGENT_PHASES.forEach((p, i) => {
      t += i === 0 ? 200 : AGENT_PHASES[i-1].duration;
      timers.push(setTimeout(() => setPhaseIndex(i), t));
    });

    return () => timers.forEach(clearTimeout);
  }, [active]);

  return { phaseIndex, logs, total: AGENT_PHASES.length };
}

// ─── Variation 1: Phases (the brief) ───────────────────────────────
function PhasesView({ theme, phaseIndex }) {
  return (
    <div style={{ padding: '8px 22px 22px' }}>
      {AGENT_PHASES.map((p, i) => {
        const status = i < phaseIndex ? 'done' : i === phaseIndex ? 'active' : 'pending';
        return (
          <div key={p.id} style={{
            display: 'flex', alignItems: 'flex-start', gap: 14,
            padding: '14px 0',
            borderBottom: i < AGENT_PHASES.length - 1 ? `0.5px solid ${theme.hairline}` : 'none',
            opacity: status === 'done' ? 0.55 : 1,
            transform: status === 'pending' ? 'translateY(0)' : 'translateY(0)',
            transition: 'opacity 0.4s',
          }}>
            <div style={{
              width: 26, height: 26, borderRadius: 999,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              flexShrink: 0, marginTop: 2,
            }}>
              {status === 'pending' && (
                <div style={{
                  width: 8, height: 8, borderRadius: 999,
                  background: theme.label4,
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
                textDecoration: status === 'done' ? 'none' : 'none',
              }}>{p.name}</div>
              <div style={{
                fontSize: 12, color: theme.label3, marginTop: 2,
                fontVariantNumeric: 'tabular-nums',
              }}>{p.sub}</div>
            </div>
          </div>
        );
      })}
    </div>
  );
}

// ─── Variation 2: Wave (oscilloscope) ──────────────────────────────
function WaveView({ theme, phaseIndex, logs }) {
  const phase = AGENT_PHASES[phaseIndex];
  const done = phaseIndex >= AGENT_PHASES.length - 1;
  const recentLog = logs[logs.length - 1];

  // Generate oscilloscope path
  const points = 60;
  const [tick, setTick] = React.useState(0);
  React.useEffect(() => {
    const id = setInterval(() => setTick(t => t + 1), 50);
    return () => clearInterval(id);
  }, []);
  const path = React.useMemo(() => {
    const pts = [];
    for (let i = 0; i < points; i++) {
      const x = (i / (points - 1)) * 320;
      const t = tick * 0.1 + i * 0.4;
      // mix of frequencies — calm but alive
      const y = 60
        + Math.sin(t) * 14
        + Math.sin(t * 2.3) * 8
        + Math.sin(t * 0.7 + 1.2) * 6
        + (Math.random() - 0.5) * 1.5;
      pts.push(`${i === 0 ? 'M' : 'L'}${x},${y}`);
    }
    return pts.join(' ');
  }, [tick]);

  return (
    <div style={{ padding: '20px 22px 22px' }}>
      {/* Oscilloscope */}
      <div style={{
        position: 'relative', height: 130, marginBottom: 24,
        background: theme.mode === 'dark' ? '#000' : '#fafafa',
        borderRadius: 14, overflow: 'hidden',
        border: `0.5px solid ${theme.hairline}`,
      }}>
        {/* grid */}
        <svg width="100%" height="100%" viewBox="0 0 320 130" style={{ position: 'absolute', inset: 0 }}>
          {[0, 1, 2, 3, 4].map(i => (
            <line key={`h${i}`} x1="0" y1={i * 32.5} x2="320" y2={i * 32.5}
                  stroke={theme.hairline} strokeWidth="0.5"/>
          ))}
          {[0, 1, 2, 3, 4, 5, 6, 7, 8].map(i => (
            <line key={`v${i}`} x1={i * 40} y1="0" x2={i * 40} y2="130"
                  stroke={theme.hairline} strokeWidth="0.5"/>
          ))}
          {/* glow */}
          <path d={path} fill="none" stroke={theme.accent}
                strokeWidth="3" strokeLinecap="round"
                opacity="0.35" filter="blur(3px)"/>
          <path d={path} fill="none" stroke={theme.accent}
                strokeWidth="1.5" strokeLinecap="round"/>
        </svg>
        {/* readout corner */}
        <div style={{
          position: 'absolute', top: 8, left: 10,
          fontFamily: SM_TOKENS.fontMono, fontSize: 10,
          color: theme.accent, letterSpacing: 0.5,
          opacity: 0.85,
        }}>SIG · {String(phaseIndex + 1).padStart(2, '0')}/0{AGENT_PHASES.length}</div>
        <div style={{
          position: 'absolute', top: 8, right: 10,
          fontFamily: SM_TOKENS.fontMono, fontSize: 10,
          color: theme.label3, letterSpacing: 0.5,
        }}>{done ? 'IDLE' : 'LIVE'}</div>
      </div>

      <div style={{
        fontSize: 11, fontWeight: 600, color: theme.label3,
        letterSpacing: 0.8, textTransform: 'uppercase', marginBottom: 6,
      }}>Phase {phaseIndex + 1} of {AGENT_PHASES.length}</div>
      <div style={{
        fontSize: 28, fontWeight: 700, letterSpacing: -0.6,
        color: theme.label, lineHeight: 1.1, marginBottom: 8,
      }}>{phase?.name || 'Done'}</div>
      <div style={{
        fontSize: 14, color: theme.label2, marginBottom: 18,
        minHeight: 20, fontVariantNumeric: 'tabular-nums',
      }}>{phase?.sub || ''}</div>

      {/* progress bar */}
      <div style={{
        height: 4, background: theme.bgElev3, borderRadius: 999, overflow: 'hidden',
      }}>
        <div style={{
          width: `${((phaseIndex + (done ? 0 : 0.5)) / AGENT_PHASES.length) * 100}%`,
          height: '100%', background: theme.accent,
          boxShadow: `0 0 12px ${theme.accentGlow}`,
          transition: 'width 0.6s cubic-bezier(0.32,0.72,0,1)',
        }}/>
      </div>

      {recentLog && (
        <div style={{
          marginTop: 18, padding: '10px 12px',
          background: theme.bgElev2, borderRadius: 10,
          fontFamily: SM_TOKENS.fontMono, fontSize: 11,
          color: theme.label2, letterSpacing: 0.2,
        }}>
          <span style={{ color: theme.label4 }}>›</span> {recentLog.text}
        </div>
      )}
    </div>
  );
}

// ─── Variation 3: Terminal ─────────────────────────────────────────
const LOG_COLORS = (theme) => ({
  http: theme.label3,
  ok: theme.accent,
  check: theme.label2,
  warn: theme.mode === 'dark' ? '#FFB547' : '#B86E00',
  discard: theme.label4,
  signal: theme.accent,
  done: theme.accent,
});

function TerminalView({ theme, phaseIndex, logs }) {
  const scrollRef = React.useRef(null);
  React.useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs.length]);
  const colors = LOG_COLORS(theme);
  const phase = AGENT_PHASES[phaseIndex];
  const done = phaseIndex >= AGENT_PHASES.length - 1;

  return (
    <div style={{ padding: '8px 22px 22px' }}>
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        marginBottom: 12,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <div style={{
            width: 8, height: 8, borderRadius: 999,
            background: done ? theme.label3 : theme.accent,
            boxShadow: done ? 'none' : `0 0 10px ${theme.accentGlow}`,
            animation: done ? 'none' : 'pulse-dot 1.4s ease-in-out infinite',
          }}/>
          <div style={{ fontSize: 13, fontWeight: 600, color: theme.label }}>
            {phase?.name || 'Done'}
          </div>
          <div style={{
            fontFamily: SM_TOKENS.fontMono, fontSize: 11,
            color: theme.label3,
          }}>{phase?.sub || ''}</div>
        </div>
        <div style={{
          fontFamily: SM_TOKENS.fontMono, fontSize: 10,
          color: theme.label4, letterSpacing: 0.5,
        }}>{phaseIndex + 1}/{AGENT_PHASES.length}</div>
      </div>

      <div ref={scrollRef} style={{
        background: theme.mode === 'dark' ? '#0A0A0C' : '#FAFAFA',
        border: `0.5px solid ${theme.hairline}`,
        borderRadius: 12,
        padding: '12px 14px',
        height: 280,
        overflow: 'auto',
        fontFamily: SM_TOKENS.fontMono,
        fontSize: 11,
        lineHeight: 1.7,
        letterSpacing: 0.1,
      }}>
        {logs.map((l) => (
          <div key={l.id} style={{
            color: colors[l.tag] || theme.label2,
            display: 'flex', gap: 8,
            animation: 'log-in 0.3s ease-out',
          }}>
            <span style={{ color: theme.label4, flexShrink: 0 }}>
              {l.tag === 'signal' ? '+' : l.tag === 'discard' ? '−' : '›'}
            </span>
            <span style={{ wordBreak: 'break-word' }}>{l.text}</span>
          </div>
        ))}
        {!done && (
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 4 }}>
            <span style={{ color: theme.label4 }}>›</span>
            <span style={{
              width: 7, height: 12, background: theme.accent,
              animation: 'cursor-blink 1s steps(2) infinite',
              display: 'inline-block',
            }}/>
          </div>
        )}
      </div>
    </div>
  );
}

// ─── Main LiveAgentView ────────────────────────────────────────────
function LiveAgent({ theme, open, variant = 'phases', onClose, onView }) {
  const { phaseIndex, logs } = useAgentSimulation(open);
  const done = phaseIndex >= AGENT_PHASES.length - 1;

  return (
    <>
      <div onClick={done ? onClose : null} style={{
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

          {/* Header */}
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

          {/* Variant body */}
          {variant === 'phases'   && <PhasesView   theme={theme} phaseIndex={phaseIndex}/>}
          {variant === 'wave'     && <WaveView     theme={theme} phaseIndex={phaseIndex} logs={logs}/>}
          {variant === 'terminal' && <TerminalView theme={theme} phaseIndex={phaseIndex} logs={logs}/>}

          {/* Footer */}
          <div style={{ padding: '0 22px' }}>
            {done ? (
              <button onClick={onView} style={{
                width: '100%', padding: '15px',
                borderRadius: 14, border: 'none',
                background: theme.accent,
                color: theme.mode === 'dark' ? '#000' : '#fff',
                fontSize: 16, fontWeight: 600, fontFamily: 'inherit',
                cursor: 'pointer', letterSpacing: 0.2,
                display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
              }}>
                <Icon name="arrow-up-right" size={16} weight="bold"
                  color={theme.mode === 'dark' ? '#000' : '#fff'}/>
                View 2 new signals
              </button>
            ) : (
              <div style={{
                textAlign: 'center', fontSize: 11, color: theme.label3,
                letterSpacing: 0.4, padding: '8px 0',
              }}>Tap anywhere outside to dismiss when done</div>
            )}
          </div>
        </GlassSurface>
      </div>
    </>
  );
}

window.LiveAgent = LiveAgent;
