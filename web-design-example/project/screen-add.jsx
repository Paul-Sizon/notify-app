// AddSubscriptionView (sheet)

function AddSheet({ theme, open, onClose, onCreate }) {
  const [query, setQuery] = React.useState('');
  const [type, setType] = React.useState('event');
  const [cadence, setCadence] = React.useState('1h');
  const [success, setSuccess] = React.useState(false);
  const inputRef = React.useRef(null);

  React.useEffect(() => {
    if (open) {
      setQuery(''); setType('event'); setCadence('1h'); setSuccess(false);
      setTimeout(() => inputRef.current?.focus(), 280);
    }
  }, [open]);

  const valid = query.trim().length >= 3;

  const submit = () => {
    if (!valid) return;
    setSuccess(true);
    setTimeout(() => {
      onCreate?.({ query: query.trim(), type, cadence });
      setSuccess(false);
    }, 500);
  };

  // Hint copy
  const hint = (() => {
    if (!query) return 'Tell me what to keep an eye on. The more specific, the better.';
    const verb = type === 'event' ? 'events' : 'news';
    const cad = { '5m': 'every 5 minutes', '15m': 'every 15 minutes', '1h': 'every hour', '6h': 'every 6 hours', '1d': 'once a day' }[cadence];
    return `Checks ${cad} for new ${verb} matching this query.`;
  })();

  return (
    <>
      {/* Backdrop */}
      <div onClick={onClose} style={{
        position: 'absolute', inset: 0, zIndex: 90,
        background: open ? 'rgba(0,0,0,0.4)' : 'transparent',
        backdropFilter: open ? 'blur(2px)' : 'none',
        WebkitBackdropFilter: open ? 'blur(2px)' : 'none',
        opacity: open ? 1 : 0, pointerEvents: open ? 'auto' : 'none',
        transition: 'opacity 0.32s cubic-bezier(0.32,0.72,0,1)',
      }}/>
      {/* Sheet */}
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
          {/* grabber */}
          <div style={{
            width: 36, height: 5, borderRadius: 999,
            background: theme.label4,
            margin: '8px auto 0',
          }}/>
          <div style={{ padding: '24px 22px 0' }}>
            <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', marginBottom: 22 }}>
              <h2 style={{
                fontSize: 26, fontWeight: 700, letterSpacing: -0.6,
                color: theme.label, margin: 0,
              }}>New watcher</h2>
              <div onClick={onClose} style={{
                width: 30, height: 30, borderRadius: 999,
                background: theme.bgElev3, display: 'flex',
                alignItems: 'center', justifyContent: 'center',
                cursor: 'pointer',
              }}>
                <Icon name="x" size={14} color={theme.label2} weight="medium"/>
              </div>
            </div>

            {/* Input */}
            <div style={{
              borderBottom: `0.5px solid ${theme.separator}`,
              paddingBottom: 18, marginBottom: 22,
            }}>
              <input
                ref={inputRef}
                value={query}
                onChange={(e) => setQuery(e.target.value)}
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
                {['blockchain meetups in Lisbon', 'EU AI Act updates', 'Phoebe Bridgers tour'].map(s => (
                  <div key={s} onClick={() => setQuery(s)} style={{
                    fontSize: 12, color: theme.label3,
                    padding: '5px 10px', borderRadius: 999,
                    background: theme.chipBg, cursor: 'pointer',
                  }}>{s}</div>
                ))}
              </div>
            </div>

            {/* Type */}
            <div style={{ marginBottom: 20 }}>
              <div style={{
                fontSize: 11, fontWeight: 600, color: theme.label3,
                letterSpacing: 0.6, textTransform: 'uppercase', marginBottom: 8,
              }}>Type</div>
              <div style={{
                display: 'flex', padding: 3,
                background: theme.bgElev3, borderRadius: 10,
                position: 'relative',
              }}>
                <div style={{
                  position: 'absolute', top: 3, bottom: 3,
                  left: type === 'event' ? 3 : '50%',
                  width: 'calc(50% - 3px)',
                  background: theme.bgElev2,
                  borderRadius: 8,
                  boxShadow: '0 1px 2px rgba(0,0,0,0.1)',
                  transition: 'left 0.28s cubic-bezier(0.32,0.72,0,1)',
                }}/>
                {[
                  { id: 'event', icon: 'calendar', label: 'Event' },
                  { id: 'news', icon: 'newspaper', label: 'News' },
                ].map(opt => (
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
            </div>

            {/* Cadence */}
            <div style={{ marginBottom: 22 }}>
              <div style={{
                fontSize: 11, fontWeight: 600, color: theme.label3,
                letterSpacing: 0.6, textTransform: 'uppercase', marginBottom: 8,
              }}>Cadence</div>
              <div style={{ display: 'flex', gap: 6 }}>
                {['5m', '15m', '1h', '6h', '1d'].map(c => {
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
            </div>

            {/* Hint */}
            <div style={{
              fontSize: 13, color: theme.label3, lineHeight: 1.4,
              padding: '0 2px', marginBottom: 28, minHeight: 38,
              textWrap: 'pretty',
            }}>{hint}</div>

            {/* Watch button */}
            <button onClick={submit} disabled={!valid} style={{
              width: '100%', padding: '16px',
              borderRadius: 14, border: 'none',
              background: valid ? theme.accent : theme.bgElev3,
              color: valid
                ? (theme.mode === 'dark' ? '#000' : '#fff')
                : theme.label3,
              fontSize: 16, fontWeight: 600, letterSpacing: 0.2,
              fontFamily: 'inherit',
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
                  Watch
                </>
              )}
            </button>
          </div>
        </GlassSurface>
      </div>
    </>
  );
}

window.AddSheet = AddSheet;
