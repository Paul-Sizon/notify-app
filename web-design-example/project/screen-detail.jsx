// SignalDetailView (subscription history)

function SignalCell({ signal, theme, onTap }) {
  const [pressed, setPressed] = React.useState(false);
  return (
    <div
      onClick={() => onTap?.(signal)}
      onPointerDown={() => setPressed(true)}
      onPointerUp={() => setPressed(false)}
      onPointerLeave={() => setPressed(false)}
      style={{
        padding: '14px 16px',
        background: theme.bgElev2,
        borderRadius: 14,
        cursor: 'pointer',
        position: 'relative',
        transform: pressed ? 'scale(0.985)' : 'scale(1)',
        transition: 'transform 0.18s cubic-bezier(0.32,0.72,0,1)',
      }}
    >
      {signal.new && (
        <div style={{
          position: 'absolute', left: -4, top: 18,
          width: 6, height: 6, borderRadius: 999,
          background: theme.accent,
          boxShadow: `0 0 10px ${theme.accentGlow}`,
        }}/>
      )}
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10 }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{
            fontSize: 16, fontWeight: 600, color: theme.label,
            letterSpacing: -0.3, lineHeight: 1.3, marginBottom: 6,
            textWrap: 'pretty',
          }}>{signal.title}</div>
          <div style={{
            fontSize: 14, color: theme.label2, lineHeight: 1.4,
            marginBottom: 10, textWrap: 'pretty',
          }}>{signal.summary}</div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Chip theme={theme} small>
              <Icon name="globe" size={9} color={theme.label2}/> {signal.source}
            </Chip>
            <span style={{ fontSize: 11, color: theme.label3 }}>{signal.time}</span>
          </div>
        </div>
        <Icon name="arrow-up-right" size={14} color={theme.label4} style={{ marginTop: 4 }}/>
      </div>
    </div>
  );
}

function DetailScreen({ theme, sub, onBack, onOpenSignal, onTriggerAgent }) {
  if (!sub) return null;
  const groups = {};
  sub.signals.forEach(s => {
    groups[s.day] = groups[s.day] || [];
    groups[s.day].push(s);
  });
  const order = ['Today', 'Yesterday', 'This week', 'Earlier'].filter(g => groups[g]);

  return (
    <div style={{ height: '100%', overflow: 'auto', paddingBottom: 80 }}>
      {/* Toolbar */}
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
          <div onClick={onTriggerAgent} style={{
            width: 36, height: 36, borderRadius: 999,
            background: theme.bgElev2,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            cursor: 'pointer',
          }}>
            <Icon name="arrow-clockwise" size={16} color={theme.accent}/>
          </div>
          <div style={{
            width: 36, height: 36, borderRadius: 999,
            background: theme.bgElev2,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            cursor: 'pointer',
          }}>
            <Icon name="ellipsis" size={16} color={theme.accent}/>
          </div>
        </div>
      </div>

      {/* Hero */}
      <div style={{ padding: '12px 22px 24px' }}>
        <h1 style={{
          fontSize: 30, fontWeight: 700, letterSpacing: -0.8,
          color: theme.label, margin: 0, lineHeight: 1.1,
          textWrap: 'pretty',
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
          <span>every {sub.cadence}</span>
          <span style={{ width: 2, height: 2, borderRadius: 999, background: theme.label4 }}/>
          <span>{sub.signals.length} signals</span>
        </div>
        <div style={{
          marginTop: 14, height: 1, background: theme.hairline,
          width: '100%',
        }}/>
        <div style={{
          marginTop: 10, fontSize: 11, color: theme.label3,
          letterSpacing: 0.4, display: 'flex', alignItems: 'center', gap: 6,
        }}>
          <WaveGlyph color={theme.label3} size={9}/>
          Last checked {sub.lastRun}
        </div>
      </div>

      {/* Signals grouped */}
      {order.map(g => (
        <div key={g} style={{ marginBottom: 24 }}>
          <div style={{
            padding: '0 22px 8px', fontSize: 11, fontWeight: 600,
            color: theme.label3, letterSpacing: 0.6, textTransform: 'uppercase',
          }}>{g}</div>
          <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: 6 }}>
            {groups[g].map(s => (
              <SignalCell key={s.id} signal={s} theme={theme} onTap={onOpenSignal}/>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

window.DetailScreen = DetailScreen;
