import { useState } from 'react';
import { buildTheme } from './tokens';
import { TabBar, type TabId } from './components/TabBar';
import { HomeScreen } from './screens/HomeScreen';
import { AlertsScreen } from './screens/AlertsScreen';
import { SignalsScreen } from './screens/SignalsScreen';
import { AccountScreen } from './screens/AccountScreen';
import { DetailScreen } from './screens/DetailScreen';
import { AddSheet } from './screens/AddSheet';
import { AISuggestionsScreen } from './screens/AISuggestionsScreen';
import { LiveAgent } from './screens/LiveAgent';
import { Icon } from './primitives';
import { useApp } from './state';
import { api } from './api';

export function App() {
  const theme = buildTheme('dark', 'green');
  const app = useApp();
  const [tab, setTab] = useState<TabId>('watchers');
  const [route, setRoute] = useState<{ name: 'home' } | { name: 'detail'; subId: string }>({ name: 'home' });
  const [addOpen, setAddOpen] = useState(false);
  const [aiOpen, setAiOpen] = useState(false);
  const [agentOpen, setAgentOpen] = useState(false);

  const activeSub = route.name === 'detail' ? app.subs.find(s => s.id === route.subId) ?? null : null;

  const onResetAll = async () => {
    if (!confirm('Delete all watchers? This cannot be undone.')) return;
    for (const s of app.subs) await app.remove(s.id);
  };

  const onRefreshSub = async () => {
    if (!activeSub) return;
    setAgentOpen(true);
    await app.runOne(activeSub.id);
    // LiveAgent overlay autoplays its phases independently; close after agent returns.
    setTimeout(() => setAgentOpen(false), 200);
  };

  return (
    <div style={{
      minHeight: '100vh', display: 'flex', justifyContent: 'center',
      background: '#050507',
    }}>
      <div style={{
        position: 'relative',
        width: '100%', maxWidth: 440,
        minHeight: '100vh',
        background: theme.bg,
        color: theme.label,
        overflow: 'hidden',
        boxShadow: '0 0 80px rgba(0,0,0,0.6)',
      }}>
        <Screen visible={route.name === 'home' && tab === 'watchers'}>
          <HomeScreen
            theme={theme}
            subs={app.subs}
            refreshing={app.refreshing}
            onOpen={s => { app.markViewed(s.id); setRoute({ name: 'detail', subId: s.id }); }}
            onAdd={() => setAddOpen(true)}
            onAI={() => setAiOpen(true)}
            onDelete={app.remove}
          />
        </Screen>
        <Screen visible={route.name === 'home' && tab === 'alerts'}>
          <AlertsScreen theme={theme} subs={app.subs}/>
        </Screen>
        <Screen visible={route.name === 'home' && tab === 'signals'}>
          <SignalsScreen theme={theme} subs={app.subs}/>
        </Screen>
        <Screen visible={route.name === 'home' && tab === 'account'}>
          <AccountScreen theme={theme} subs={app.subs} deviceId={api.deviceId} onResetAll={onResetAll}/>
        </Screen>

        <Screen visible={route.name === 'detail'} slide>
          {activeSub && (
            <DetailScreen
              theme={theme}
              sub={activeSub}
              refreshing={app.refreshing}
              onBack={() => setRoute({ name: 'home' })}
              onRun={onRefreshSub}
            />
          )}
        </Screen>

        {route.name === 'home' && tab === 'watchers' && app.subs.length > 0 && (
          <div
            onClick={() => setAddOpen(true)}
            style={{
              position: 'absolute', bottom: 92, right: 22, zIndex: 80,
              width: 56, height: 56, borderRadius: 999,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              background: theme.accent, cursor: 'pointer',
              boxShadow: `0 12px 32px ${theme.accentGlow}, 0 4px 12px rgba(0,0,0,0.3)`,
            }}
          >
            <Icon name="plus" size={22} weight="bold" color={theme.mode === 'dark' ? '#000' : '#fff'}/>
          </div>
        )}

        {route.name === 'home' && (
          <TabBar theme={theme} active={tab} onChange={setTab}/>
        )}

        <AddSheet
          theme={theme}
          open={addOpen}
          onClose={() => setAddOpen(false)}
          onCreate={app.create}
          onAI={() => { setAddOpen(false); setAiOpen(true); }}
        />

        <AISuggestionsScreen
          theme={theme}
          open={aiOpen}
          onClose={() => setAiOpen(false)}
          onCreate={app.create}
          onDelete={app.remove}
        />

        <LiveAgent
          theme={theme}
          open={agentOpen}
          onClose={() => setAgentOpen(false)}
        />

        {app.toast && (
          <div style={{
            position: 'absolute', top: 12, left: 16, right: 16, zIndex: 100,
            padding: '10px 14px', borderRadius: 12,
            background: app.toast.kind === 'error' ? 'rgba(255,59,48,0.92)' : theme.bgElev2,
            color: app.toast.kind === 'error' ? '#fff' : theme.label,
            fontSize: 13, fontWeight: 500,
            boxShadow: '0 8px 24px rgba(0,0,0,0.3)',
            animation: 'fade-in 0.2s',
          }}>{app.toast.text}</div>
        )}

        {app.loading && (
          <div style={{
            position: 'absolute', inset: 0,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            background: theme.bg, zIndex: 200,
          }}>
            <div style={{
              width: 32, height: 32, borderRadius: 999,
              border: `2px solid ${theme.bgElev3}`,
              borderTopColor: theme.accent,
              animation: 'spin 0.8s linear infinite',
            }}/>
            <style>{'@keyframes spin{to{transform:rotate(360deg)}}'}</style>
          </div>
        )}
      </div>
    </div>
  );
}

function Screen({
  visible, slide, children,
}: { visible: boolean; slide?: boolean; children: React.ReactNode }) {
  return (
    <div style={{
      position: 'absolute', inset: 0,
      opacity: visible ? 1 : 0,
      pointerEvents: visible ? 'auto' : 'none',
      transform: slide ? (visible ? 'translateX(0)' : 'translateX(40px)') : undefined,
      transition: slide
        ? 'opacity 0.3s, transform 0.32s cubic-bezier(0.32,0.72,0,1)'
        : 'opacity 0.25s',
    }}>{children}</div>
  );
}
