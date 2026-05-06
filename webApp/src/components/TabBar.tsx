import React from 'react';
import type { Theme } from '../tokens';

export type TabId = 'watchers' | 'alerts' | 'signals' | 'account';

const TABS: { id: TabId; label: string }[] = [
  { id: 'watchers', label: 'Watchers' },
  { id: 'alerts',   label: 'Alerts' },
  { id: 'signals',  label: 'Signals' },
  { id: 'account',  label: 'Account' },
];

export function TabBar({
  theme, active, onChange,
}: { theme: Theme; active: TabId; onChange: (t: TabId) => void }) {
  return (
    <div style={{
      position: 'absolute', left: 0, right: 0, bottom: 0,
      paddingBottom: 22, paddingTop: 10,
      background: theme.mode === 'dark'
        ? 'linear-gradient(180deg, rgba(0,0,0,0) 0%, rgba(0,0,0,0.85) 30%, #000 70%)'
        : 'linear-gradient(180deg, rgba(255,255,255,0) 0%, rgba(255,255,255,0.9) 40%, #fff 100%)',
      zIndex: 70, pointerEvents: 'auto',
    }}>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', padding: '0 8px' }}>
        {TABS.map(t => {
          const isActive = active === t.id;
          const color = isActive ? theme.accent : theme.label3;
          return (
            <div
              key={t.id}
              onClick={() => onChange(t.id)}
              style={{
                display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4,
                padding: '8px 4px', cursor: 'pointer', userSelect: 'none',
                transition: 'transform 0.15s',
                transform: isActive ? 'scale(1)' : 'scale(0.98)',
              }}
            >
              <TabIcon id={t.id} active={isActive} color={color}/>
              <div style={{ fontSize: 11, fontWeight: 600, letterSpacing: 0.1, color }}>{t.label}</div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function TabIcon({ id, active, color }: { id: TabId; active: boolean; color: string }) {
  const sw = 1.7;
  const props: React.SVGProps<SVGSVGElement> = {
    width: 22, height: 22, viewBox: '0 0 24 24', fill: 'none',
    stroke: color, strokeWidth: sw, strokeLinecap: 'round', strokeLinejoin: 'round',
    style: { display: 'block' },
  };
  switch (id) {
    case 'watchers': return active ? (
      <svg {...props}>
        <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z" fill={color} fillOpacity={0.18}/>
        <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z"/>
        <circle cx={12} cy={12} r={3} fill={color}/>
      </svg>
    ) : (
      <svg {...props}>
        <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z"/>
        <circle cx={12} cy={12} r={3}/>
      </svg>
    );
    case 'alerts': return active ? (
      <svg {...props}>
        <path d="M18 16v-5a6 6 0 10-12 0v5l-2 3h16l-2-3z" fill={color} fillOpacity={0.18}/>
        <path d="M18 16v-5a6 6 0 10-12 0v5l-2 3h16l-2-3z"/>
        <path d="M10 21a2 2 0 004 0"/>
        <path d="M3 6c1-1.5 2.5-2.5 4-3M21 6c-1-1.5-2.5-2.5-4-3" strokeWidth={sw - 0.2}/>
      </svg>
    ) : (
      <svg {...props}>
        <path d="M18 16v-5a6 6 0 10-12 0v5l-2 3h16l-2-3z"/>
        <path d="M10 21a2 2 0 004 0"/>
      </svg>
    );
    case 'signals': return (
      <svg {...props}>
        <path d="M3 17l5-6 4 3 5-7 4 5"/>
        <circle cx={17} cy={7} r={1.4} fill={color}/>
      </svg>
    );
    case 'account': return active ? (
      <svg {...props}>
        <circle cx={12} cy={9} r={4} fill={color} fillOpacity={0.18}/>
        <circle cx={12} cy={9} r={4}/>
        <path d="M4 21c0-4 4-6 8-6s8 2 8 6"/>
      </svg>
    ) : (
      <svg {...props}>
        <circle cx={12} cy={9} r={4}/>
        <path d="M4 21c0-4 4-6 8-6s8 2 8 6"/>
      </svg>
    );
  }
}
