// Alerts, Signals, Account screens

// ─── Alerts: log of past push notifications ─────────────────
function AlertsScreen({ theme, subs }) {
  // Build alert log from signals across all subs
  const alerts = [];
  subs.forEach(s => {
    s.signals.forEach(sg => {
      alerts.push({ ...sg, query: s.query, type: s.type, subId: s.id });
    });
  });
  alerts.sort((a, b) => {
    const order = { 'Today': 0, 'Yesterday': 1, 'This week': 2, 'Earlier': 3 };
    return (order[a.day] ?? 9) - (order[b.day] ?? 9);
  });

  const groups = {};
  alerts.forEach(a => { groups[a.day] = groups[a.day] || []; groups[a.day].push(a); });
  const order = ['Today', 'Yesterday', 'This week', 'Earlier'].filter(g => groups[g]);

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
          }}>{alerts.length}</span>
        </div>
        <div style={{ fontSize: 13, color: theme.label3, marginTop: 6 }}>
          Every signal that buzzed your phone.
        </div>
      </div>

      {order.map(g => (
        <div key={g} style={{ marginBottom: 24 }}>
          <div style={{
            padding: '0 22px 8px', fontSize: 11, fontWeight: 600,
            color: theme.label3, letterSpacing: 0.6, textTransform: 'uppercase',
          }}>{g}</div>
          <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: 6 }}>
            {groups[g].map(a => (
              <div key={a.id} style={{
                background: theme.bgElev2, borderRadius: 14,
                padding: '12px 14px', position: 'relative',
              }}>
                {a.new && (
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
                  <div style={{ fontSize: 11, color: theme.label3, flexShrink: 0 }}>
                    {a.time}
                  </div>
                </div>
                <div style={{
                  fontSize: 14, color: theme.label, fontWeight: 500,
                  letterSpacing: -0.2, lineHeight: 1.3, marginBottom: 4,
                  textWrap: 'pretty',
                }}>{a.title}</div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <Icon name="globe" size={10} color={theme.label3}/>
                  <span style={{ fontSize: 11, color: theme.label3 }}>{a.source}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

// ─── Signals: firehose of all signals, surfaced ─────────────
function SignalsScreen({ theme, subs }) {
  const all = [];
  subs.forEach(s => {
    s.signals.forEach(sg => {
      all.push({ ...sg, query: s.query, type: s.type });
    });
  });

  const stats = {
    today:  all.filter(a => a.day === 'Today').length,
    week:   all.filter(a => a.day === 'Today' || a.day === 'Yesterday' || a.day === 'This week').length,
    sources: new Set(all.map(a => a.source)).size,
  };

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

      {/* Stats strip */}
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
            position: 'relative',
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

      {/* Activity sparkline-ish row */}
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
        <div style={{
          display: 'flex', alignItems: 'flex-end', gap: 4, height: 48,
        }}>
          {[2, 0, 1, 3, 2, 0, 0, 4, 1, 2, 5, 1, 0, 3].map((v, i) => {
            const max = 5;
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

      {/* Recent list */}
      <div style={{
        padding: '0 22px 8px', fontSize: 11, fontWeight: 600,
        color: theme.label3, letterSpacing: 0.6, textTransform: 'uppercase',
      }}>Latest</div>
      <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: 6 }}>
        {all.slice(0, 8).map(s => (
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
              textWrap: 'pretty',
            }}>{s.title}</div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Chip theme={theme} small>
                <Icon name="globe" size={9} color={theme.label2}/> {s.source}
              </Chip>
              <span style={{ fontSize: 11, color: theme.label3 }}>{s.time}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Account ──────────────────────────────────────────────────
function AccountScreen({ theme, subs }) {
  const stats = {
    watching: subs.filter(s => !s.resolved).length,
    resolved: subs.filter(s => s.resolved).length,
    signals: subs.reduce((n, s) => n + s.signals.length, 0),
  };

  const Row = ({ icon, label, value, subtle, danger, top, bottom }) => (
    <div style={{
      display: 'flex', alignItems: 'center', gap: 12,
      padding: '14px 16px',
      background: theme.bgElev2,
      borderTopLeftRadius: top ? 14 : 0,
      borderTopRightRadius: top ? 14 : 0,
      borderBottomLeftRadius: bottom ? 14 : 0,
      borderBottomRightRadius: bottom ? 14 : 0,
      borderBottom: bottom ? 'none' : `0.5px solid ${theme.separator}`,
    }}>
      <div style={{
        width: 28, height: 28, borderRadius: 7,
        background: danger ? 'rgba(255,69,58,0.14)' : theme.chipBg,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        flexShrink: 0,
      }}>
        <Icon name={icon} size={15} color={danger ? '#FF453A' : theme.label2}/>
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{
          fontSize: 15, color: danger ? '#FF453A' : theme.label, fontWeight: 500,
          letterSpacing: -0.1,
        }}>{label}</div>
        {subtle && <div style={{ fontSize: 12, color: theme.label3, marginTop: 2 }}>{subtle}</div>}
      </div>
      {value && <div style={{ fontSize: 13, color: theme.label3 }}>{value}</div>}
      {!danger && <Icon name="chevron-right" size={14} color={theme.label4}/>}
    </div>
  );

  return (
    <div style={{ height: '100%', overflow: 'auto', paddingBottom: 140 }}>
      <div style={{ padding: '60px 20px 24px' }}>
        <h1 style={{
          fontSize: 34, fontWeight: 700, letterSpacing: -1.0,
          color: theme.label, margin: 0, lineHeight: 1.05,
        }}>Account</h1>
      </div>

      {/* Identity card */}
      <div style={{ padding: '0 16px 24px' }}>
        <div style={{
          background: theme.bgElev2, borderRadius: 14,
          padding: '20px 18px',
          display: 'flex', alignItems: 'center', gap: 14,
        }}>
          <div style={{
            width: 48, height: 48, borderRadius: 999,
            background: `linear-gradient(135deg, ${theme.accent}, ${theme.mode === 'dark' ? '#000' : '#fff'})`,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            color: theme.mode === 'dark' ? '#000' : '#fff',
            fontSize: 18, fontWeight: 700, letterSpacing: -0.4,
          }}>SM</div>
          <div style={{ flex: 1 }}>
            <div style={{
              fontSize: 16, fontWeight: 600, color: theme.label, letterSpacing: -0.2,
            }}>This device</div>
            <div style={{
              fontFamily: SM_TOKENS.fontMono, fontSize: 11,
              color: theme.label3, marginTop: 3, letterSpacing: 0.5,
            }}>id · 7f3a · 9c12 · ed44</div>
          </div>
        </div>

        {/* Stats grid */}
        <div style={{
          marginTop: 8,
          background: theme.bgElev2, borderRadius: 14,
          padding: '16px 4px',
          display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)',
        }}>
          {[
            { label: 'Watching', value: stats.watching },
            { label: 'Resolved', value: stats.resolved },
            { label: 'Signals',  value: stats.signals },
          ].map((s, i) => (
            <div key={s.label} style={{
              textAlign: 'center',
              borderLeft: i === 0 ? 'none' : `0.5px solid ${theme.separator}`,
              padding: '4px 8px',
            }}>
              <div style={{
                fontSize: 24, fontWeight: 700, color: theme.label,
                letterSpacing: -0.4, fontVariantNumeric: 'tabular-nums',
              }}>{s.value}</div>
              <div style={{
                fontSize: 11, color: theme.label3,
                letterSpacing: 0.3, textTransform: 'uppercase', fontWeight: 600,
                marginTop: 2,
              }}>{s.label}</div>
            </div>
          ))}
        </div>
      </div>

      {/* Settings groups */}
      <div style={{
        padding: '0 22px 8px', fontSize: 11, fontWeight: 600,
        color: theme.label3, letterSpacing: 0.6, textTransform: 'uppercase',
      }}>Agent</div>
      <div style={{ padding: '0 16px 24px' }}>
        <Row icon="bell" label="Notifications" value="On" top/>
        <Row icon="sparkle" label="Agent strictness" value="Balanced"/>
        <Row icon="funnel" label="Quiet hours" value="22:00–08:00" bottom/>
      </div>

      <div style={{
        padding: '0 22px 8px', fontSize: 11, fontWeight: 600,
        color: theme.label3, letterSpacing: 0.6, textTransform: 'uppercase',
      }}>Privacy</div>
      <div style={{ padding: '0 16px 24px' }}>
        <Row icon="shield" label="Data retention" value="30 days" top/>
        <Row icon="globe" label="Region" value="Auto" bottom/>
      </div>

      <div style={{ padding: '0 16px 24px' }}>
        <Row icon="trash" label="Reset all watchers" danger top bottom/>
      </div>

      <div style={{
        textAlign: 'center', padding: '0 20px 12px',
        fontSize: 11, color: theme.label4, letterSpacing: 0.3,
      }}>Signal Monitor · v1.0</div>
    </div>
  );
}

window.AlertsScreen = AlertsScreen;
window.SignalsScreen = SignalsScreen;
window.AccountScreen = AccountScreen;
