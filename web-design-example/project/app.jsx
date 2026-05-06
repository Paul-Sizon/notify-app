// Signal Monitor — main app shell
// Wires home → detail, sheet, agent, lockscreen.

function SignalMonitorApp({ tweaks, setTweak }) {
  const theme = useSMTheme(tweaks.mode, tweaks.accent);
  const [subs, setSubs] = React.useState(SM_SEED.subscriptions);
  const [route, setRoute] = React.useState({ name: 'home' });
  const [tab, setTab] = React.useState('watchers');
  const [addOpen, setAddOpen] = React.useState(false);
  const [agentOpen, setAgentOpen] = React.useState(false);
  const [showLock, setShowLock] = React.useState(tweaks.startOn === 'lock');

  React.useEffect(() => {
    setShowLock(tweaks.startOn === 'lock');
    if (tweaks.startOn === 'home') setRoute({ name: 'home' });
    if (tweaks.startOn === 'detail') setRoute({ name: 'detail', subId: subs[0]?.id });
    if (tweaks.startOn === 'add') { setRoute({ name: 'home' }); setAddOpen(true); }
    if (tweaks.startOn === 'agent') { setRoute({ name: 'home' }); setAgentOpen(true); }
    if (tweaks.startOn === 'empty') setSubs([]);
    else setSubs(SM_SEED.subscriptions);
  }, [tweaks.startOn]);

  const activeSub = route.name === 'detail' ? subs.find(s => s.id === route.subId) : null;

  const onCreate = (data) => {
    const id = 'new-' + Date.now();
    setSubs(prev => [{
      id, query: data.query, type: data.type, cadence: data.cadence,
      lastRun: 'just now', unread: false, signals: [],
    }, ...prev]);
    setAddOpen(false);
  };

  const onDelete = (id) => setSubs(prev => prev.filter(s => s.id !== id));

  const onTap = () => {
    setShowLock(false);
    setRoute({ name: 'detail', subId: subs[0]?.id });
  };

  return (
    <div style={{
      width: '100%', height: '100%', position: 'relative',
      background: theme.bg, overflow: 'hidden',
      color: theme.label,
      fontFamily: SM_TOKENS.font,
    }}>
      {/* Screens */}
      <div style={{
        position: 'absolute', inset: 0,
        opacity: route.name === 'home' && tab === 'watchers' ? 1 : 0,
        pointerEvents: route.name === 'home' && tab === 'watchers' ? 'auto' : 'none',
        transition: 'opacity 0.25s',
      }}>
        <HomeScreen
          theme={theme}
          subs={subs}
          density={tweaks.density}
          onOpen={(s) => setRoute({ name: 'detail', subId: s.id })}
          onAdd={() => setAddOpen(true)}
          onDelete={onDelete}
          onTriggerAgent={() => setAgentOpen(true)}
          showAgentInline={agentOpen}
        />
      </div>

      <div style={{
        position: 'absolute', inset: 0,
        opacity: route.name === 'home' && tab === 'alerts' ? 1 : 0,
        pointerEvents: route.name === 'home' && tab === 'alerts' ? 'auto' : 'none',
        transition: 'opacity 0.25s',
      }}>
        <AlertsScreen theme={theme} subs={subs}/>
      </div>

      <div style={{
        position: 'absolute', inset: 0,
        opacity: route.name === 'home' && tab === 'signals' ? 1 : 0,
        pointerEvents: route.name === 'home' && tab === 'signals' ? 'auto' : 'none',
        transition: 'opacity 0.25s',
      }}>
        <SignalsScreen theme={theme} subs={subs}/>
      </div>

      <div style={{
        position: 'absolute', inset: 0,
        opacity: route.name === 'home' && tab === 'account' ? 1 : 0,
        pointerEvents: route.name === 'home' && tab === 'account' ? 'auto' : 'none',
        transition: 'opacity 0.25s',
      }}>
        <AccountScreen theme={theme} subs={subs}/>
      </div>

      <div style={{
        position: 'absolute', inset: 0,
        opacity: route.name === 'detail' ? 1 : 0,
        pointerEvents: route.name === 'detail' ? 'auto' : 'none',
        transform: route.name === 'detail' ? 'translateX(0)' : 'translateX(40px)',
        transition: 'opacity 0.3s, transform 0.32s cubic-bezier(0.32,0.72,0,1)',
      }}>
        {activeSub && (
          <DetailScreen
            theme={theme}
            sub={activeSub}
            onBack={() => setRoute({ name: 'home' })}
            onOpenSignal={() => {}}
            onTriggerAgent={() => setAgentOpen(true)}
          />
        )}
      </div>

      {/* + button (visible only on watchers tab) */}
      {route.name === 'home' && tab === 'watchers' && subs.length > 0 && (
        <div style={{
          position: 'absolute', bottom: 92, right: 22, zIndex: 80,
        }} onClick={() => setAddOpen(true)}>
          <GlassSurface theme={theme} radius={999} intense style={{
            width: 56, height: 56,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            cursor: 'pointer',
            background: theme.accent,
            boxShadow: `0 12px 32px ${theme.accentGlow}, 0 4px 12px rgba(0,0,0,0.3)`,
          }}>
            <Icon name="plus" size={22} weight="bold"
              color={theme.mode === 'dark' ? '#000' : '#fff'}/>
          </GlassSurface>
        </div>
      )}

      {/* Tab bar — visible whenever on home route (any tab) */}
      {route.name === 'home' && (
        <TabBar theme={theme} active={tab} onChange={setTab}/>
      )}

      <AddSheet
        theme={theme}
        open={addOpen}
        onClose={() => setAddOpen(false)}
        onCreate={onCreate}
      />

      <LiveAgent
        theme={theme}
        open={agentOpen}
        variant={tweaks.agentVariant}
        onClose={() => setAgentOpen(false)}
        onView={() => { setAgentOpen(false); }}
      />

      {showLock && (
        <LockScreen theme={theme} accentName={tweaks.accent} onTap={onTap}/>
      )}
    </div>
  );
}

window.SignalMonitorApp = SignalMonitorApp;
