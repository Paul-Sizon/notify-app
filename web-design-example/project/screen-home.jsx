// Signal Monitor — SubscriptionsListView (home)

function SubscriptionCard({ sub, theme, density, onOpen, onDelete, onLongPress }) {
  const [pressed, setPressed] = React.useState(false);
  const [swipeX, setSwipeX] = React.useState(0);
  const [confirming, setConfirming] = React.useState(false);
  const startX = React.useRef(null);

  const dense = density === 'compact';
  const cozy = density === 'cozy';
  const padV = dense ? 12 : cozy ? 16 : 20;
  const padH = 16;

  const onPointerDown = (e) => { startX.current = e.clientX; };
  const onPointerMove = (e) => {
    if (startX.current == null) return;
    const dx = Math.min(0, e.clientX - startX.current);
    setSwipeX(Math.max(-100, dx));
  };
  const onPointerUp = () => {
    if (swipeX < -60) { setConfirming(true); setSwipeX(-80); }
    else { setSwipeX(0); }
    startX.current = null;
  };

  const typeIcon = sub.type === 'event' ? 'calendar' : 'newspaper';
  const resolved = sub.resolved;

  return (
    <div style={{ position: 'relative', userSelect: 'none' }}>
      {/* Delete background */}
      <div style={{
        position: 'absolute', inset: 0, display: 'flex',
        justifyContent: 'flex-end', alignItems: 'center',
        background: '#FF3B30', borderRadius: 16, paddingRight: 22,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: '#fff', fontSize: 13, fontWeight: 600 }}
             onClick={() => onDelete?.(sub.id)}>
          <Icon name="trash" size={18} color="#fff"/> Delete
        </div>
      </div>
      <div
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={onPointerUp}
        onPointerCancel={() => { setSwipeX(0); startX.current = null; }}
        onClick={() => swipeX === 0 && onOpen?.(sub)}
        onMouseDown={() => setPressed(true)}
        onMouseUp={() => setPressed(false)}
        onMouseLeave={() => setPressed(false)}
        onContextMenu={(e) => { e.preventDefault(); onLongPress?.(sub); }}
        style={{
          position: 'relative',
          background: resolved
            ? (theme.mode === 'dark' ? '#141416' : '#F4F4F7')
            : theme.bgElev2,
          borderRadius: 16,
          padding: `${padV}px ${padH}px`,
          transform: `translateX(${swipeX}px) scale(${pressed ? 0.985 : 1})`,
          transition: startX.current == null ? 'transform 0.25s cubic-bezier(0.32,0.72,0,1)' : 'none',
          boxShadow: theme.mode === 'dark'
            ? '0 0 0 0.5px rgba(255,255,255,0.04)'
            : '0 0 0 0.5px rgba(0,0,0,0.04)',
          cursor: 'pointer',
        }}
      >
        {/* unread dot */}
        {sub.unread && !resolved && (
          <div style={{
            position: 'absolute', left: -4, top: padV + 8,
            width: 6, height: 6, borderRadius: 999,
            background: theme.accent,
            boxShadow: `0 0 12px ${theme.accentGlow}`,
          }}/>
        )}
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{
              fontSize: dense ? 15 : 17,
              fontWeight: 600,
              color: resolved ? theme.label2 : theme.label,
              letterSpacing: -0.3,
              lineHeight: 1.25,
              marginBottom: dense ? 6 : 8,
              textWrap: 'pretty',
              textDecoration: resolved ? 'line-through' : 'none',
              textDecorationColor: resolved ? theme.label4 : 'transparent',
              textDecorationThickness: '1px',
            }}>{sub.query}</div>

            {resolved ? (
              <>
                <div style={{
                  display: 'flex', alignItems: 'center', gap: 6,
                  fontSize: 13, color: theme.label, fontWeight: 500,
                  marginBottom: 4, letterSpacing: -0.1,
                }}>
                  <Icon name="check" size={12} weight="bold" color={theme.accent}/>
                  {resolved.title}
                </div>
                <div style={{
                  fontSize: 12, color: theme.label3, lineHeight: 1.35,
                  textWrap: 'pretty',
                }}>{resolved.sub}</div>
              </>
            ) : (
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
                <Chip theme={theme} small>
                  <Icon name={typeIcon} size={10} color={theme.label2}/> {sub.type}
                </Chip>
                <Chip theme={theme} small>every {sub.cadence}</Chip>
                {sub.unread && <Chip theme={theme} small accent>
                  <WaveGlyph color={theme.accent} size={9} animated/> new
                </Chip>}
              </div>
            )}
          </div>
          <div style={{
            display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 4,
            flexShrink: 0, paddingTop: 2,
          }}>
            <div style={{
              fontSize: 12,
              color: resolved ? theme.accent : theme.label3,
              fontVariantNumeric: 'tabular-nums',
              fontWeight: resolved ? 600 : 400,
              letterSpacing: resolved ? 0.2 : 0,
            }}>
              {resolved ? resolved.date : sub.lastRun}
            </div>
            {!resolved && <Icon name="chevron-right" size={14} color={theme.label4}/>}
            {resolved && <div style={{
              fontSize: 10, color: theme.label4, letterSpacing: 0.5,
              textTransform: 'uppercase', fontWeight: 600,
            }}>resolved</div>}
          </div>
        </div>
      </div>
    </div>
  );
}

function HomeScreen({ theme, subs, density, onOpen, onAdd, onDelete, onTriggerAgent, showAgentInline }) {
  const [refreshing, setRefreshing] = React.useState(false);

  const onRefresh = () => {
    setRefreshing(true);
    onTriggerAgent?.();
    setTimeout(() => setRefreshing(false), 600);
  };

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
            border: `1px solid ${theme.hairline}`,
            opacity: 0.5, pointerEvents: 'none',
          }}/>
        </div>
        <div style={{ textAlign: 'center', maxWidth: 280 }}>
          <div style={{ fontSize: 22, fontWeight: 600, color: theme.label, letterSpacing: -0.4, marginBottom: 8 }}>
            Nothing watched yet.
          </div>
          <div style={{ fontSize: 15, color: theme.label3, lineHeight: 1.4, textWrap: 'pretty' }}>
            Add something specific. The narrower the query, the better the signals.
          </div>
        </div>
        <button onClick={onAdd} style={{
          padding: '12px 22px', borderRadius: 999,
          background: theme.accent, color: theme.mode === 'dark' ? '#000' : '#fff',
          border: 'none', fontSize: 15, fontWeight: 600,
          fontFamily: 'inherit', cursor: 'pointer',
          display: 'flex', alignItems: 'center', gap: 6,
        }}>
          <Icon name="plus" size={16} weight="bold" color={theme.mode === 'dark' ? '#000' : '#fff'}/>
          Add a watcher
        </button>
      </div>
    );
  }

  // Group by status: pinned active → active → resolved
  const active = subs.filter(s => !s.resolved);
  const resolved = subs.filter(s => s.resolved);
  const pinned = active.filter(s => s.pinned);
  const rest = active.filter(s => !s.pinned);

  return (
    <div style={{
      height: '100%', overflow: 'auto',
      paddingBottom: 200,
    }}>
      {/* Large title */}
      <div style={{ padding: '60px 20px 20px' }}>
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 10 }}>
          <h1 style={{
            fontSize: 34, fontWeight: 700, letterSpacing: -1.0,
            color: theme.label, margin: 0, lineHeight: 1.05,
          }}>Watching</h1>
          <span style={{
            fontSize: 17, fontWeight: 500, color: theme.label3,
            fontVariantNumeric: 'tabular-nums',
          }}>{active.length}</span>
        </div>
        <div style={{ fontSize: 13, color: theme.label3, marginTop: 6, letterSpacing: 0.1 }}>
          {active.filter(s => s.unread).length} new since yesterday{resolved.length > 0 ? ` · ${resolved.length} resolved` : ''}
        </div>
      </div>

      {/* Inline agent indicator on refresh */}
      {(refreshing || showAgentInline) && (
        <div style={{ padding: '0 20px 12px' }}>
          <GlassSurface theme={theme} radius={14} style={{
            border: `0.5px solid ${theme.glassBorder}`,
          }}>
            <div style={{
              padding: '10px 14px', display: 'flex', alignItems: 'center', gap: 10,
            }}>
              <WaveGlyph color={theme.accent} size={12} animated/>
              <div style={{ fontSize: 13, color: theme.label2, flex: 1 }}>
                Agent is checking {active.length} watchers…
              </div>
              <div style={{ fontSize: 11, color: theme.label3, fontVariantNumeric: 'tabular-nums' }}>
                live
              </div>
            </div>
          </GlassSurface>
        </div>
      )}

      {/* Pinned section */}
      {pinned.length > 0 && (
        <>
          <div style={{
            padding: '0 22px 8px', fontSize: 11, fontWeight: 600,
            color: theme.label3, letterSpacing: 0.6, textTransform: 'uppercase',
            display: 'flex', alignItems: 'center', gap: 6,
          }}>
            <Icon name="pin" size={10} color={theme.label3}/> Pinned
          </div>
          <div style={{ padding: '0 16px 24px', display: 'flex', flexDirection: 'column', gap: 8 }}>
            {pinned.map(s => (
              <SubscriptionCard key={s.id} sub={s} theme={theme} density={density}
                onOpen={onOpen} onDelete={onDelete}/>
            ))}
          </div>
        </>
      )}

      {/* Rest */}
      {rest.length > 0 && (
        <>
          {pinned.length > 0 && (
            <div style={{
              padding: '0 22px 8px', fontSize: 11, fontWeight: 600,
              color: theme.label3, letterSpacing: 0.6, textTransform: 'uppercase',
            }}>
              All watchers
            </div>
          )}
          <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: 8 }}>
            {rest.map(s => (
              <SubscriptionCard key={s.id} sub={s} theme={theme} density={density}
                onOpen={onOpen} onDelete={onDelete}/>
            ))}
          </div>
        </>
      )}

      {/* Resolved section — confirmed events, no longer being watched */}
      {resolved.length > 0 && (
        <>
          <div style={{
            padding: '24px 22px 8px', fontSize: 11, fontWeight: 600,
            color: theme.label3, letterSpacing: 0.6, textTransform: 'uppercase',
            display: 'flex', alignItems: 'center', gap: 6,
          }}>
            <Icon name="check" size={11} color={theme.label3} weight="medium"/>
            Resolved
            <span style={{ color: theme.label4, fontWeight: 500, letterSpacing: 0.3, marginLeft: 2 }}>
              · paused
            </span>
          </div>
          <div style={{
            padding: '0 22px 12px', fontSize: 12, color: theme.label3, lineHeight: 1.4,
            textWrap: 'pretty',
          }}>
            These have a confirmed answer. Agent stopped checking. Swipe to archive.
          </div>
          <div style={{ padding: '0 16px 32px', display: 'flex', flexDirection: 'column', gap: 8 }}>
            {resolved.map(s => (
              <SubscriptionCard key={s.id} sub={s} theme={theme} density={density}
                onOpen={onOpen} onDelete={onDelete}/>
            ))}
          </div>
        </>
      )}

      {/* footer note */}
      <div style={{
        textAlign: 'center', marginTop: 12, padding: '0 20px 20px',
        fontSize: 11, color: theme.label4, letterSpacing: 0.3,
        display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6,
      }}>
        <Icon name="shield" size={11} color={theme.label4}/>
        Pull down to run all active
      </div>
    </div>
  );
}

window.HomeScreen = HomeScreen;
window.SubscriptionCard = SubscriptionCard;
