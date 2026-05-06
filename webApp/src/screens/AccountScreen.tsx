import React from 'react';
import { Icon, type IconName } from '../primitives';
import { TOKENS, type Theme } from '../tokens';
import type { EnrichedSub } from '../state';

export function AccountScreen({
  theme, subs, deviceId, onResetAll,
}: {
  theme: Theme;
  subs: EnrichedSub[];
  deviceId: string | null;
  onResetAll: () => void;
}) {
  const stats = {
    watching: subs.length,
    signals:  subs.reduce((n, s) => n + s.signals.length, 0),
    sources:  new Set(subs.flatMap(s => s.signals.flatMap(sg => sg.source_domains))).size,
  };

  const idShort = (deviceId ?? '').replace(/-/g, '').slice(0, 12).match(/.{1,4}/g)?.join(' · ') ?? '—';

  return (
    <div style={{ height: '100%', overflow: 'auto', paddingBottom: 140 }}>
      <div style={{ padding: '60px 20px 24px' }}>
        <h1 style={{
          fontSize: 34, fontWeight: 700, letterSpacing: -1.0,
          color: theme.label, margin: 0, lineHeight: 1.05,
        }}>Account</h1>
      </div>

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
              fontFamily: TOKENS.fontMono, fontSize: 11,
              color: theme.label3, marginTop: 3, letterSpacing: 0.5,
            }}>id · {idShort}</div>
          </div>
        </div>

        <div style={{
          marginTop: 8,
          background: theme.bgElev2, borderRadius: 14,
          padding: '16px 4px',
          display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)',
        }}>
          {[
            { label: 'Watching', value: stats.watching },
            { label: 'Signals',  value: stats.signals },
            { label: 'Sources',  value: stats.sources },
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

      <SectionTitle theme={theme}>Agent</SectionTitle>
      <div style={{ padding: '0 16px 24px' }}>
        <Row theme={theme} icon="bell"    label="Notifications"     value="On"        top/>
        <Row theme={theme} icon="sparkle" label="Agent strictness"  value="Balanced"/>
        <Row theme={theme} icon="funnel"  label="Quiet hours"       value="22:00–08:00" bottom/>
      </div>

      <SectionTitle theme={theme}>Privacy</SectionTitle>
      <div style={{ padding: '0 16px 24px' }}>
        <Row theme={theme} icon="shield" label="Data retention" value="30 days" top/>
        <Row theme={theme} icon="globe"  label="Region"         value="Auto"    bottom/>
      </div>

      <div style={{ padding: '0 16px 24px' }}>
        <Row theme={theme} icon="trash" label="Reset all watchers" danger top bottom onClick={onResetAll}/>
      </div>

      <div style={{
        textAlign: 'center', padding: '0 20px 12px',
        fontSize: 11, color: theme.label4, letterSpacing: 0.3,
      }}>notify · web · v1.0</div>
    </div>
  );
}

function SectionTitle({ theme, children }: { theme: Theme; children: React.ReactNode }) {
  return (
    <div style={{
      padding: '0 22px 8px', fontSize: 11, fontWeight: 600,
      color: theme.label3, letterSpacing: 0.6, textTransform: 'uppercase',
    }}>{children}</div>
  );
}

function Row({
  theme, icon, label, value, danger, top, bottom, onClick,
}: {
  theme: Theme;
  icon: IconName;
  label: string;
  value?: string;
  danger?: boolean;
  top?: boolean;
  bottom?: boolean;
  onClick?: () => void;
}) {
  return (
    <div onClick={onClick} style={{
      display: 'flex', alignItems: 'center', gap: 12,
      padding: '14px 16px',
      background: theme.bgElev2,
      borderTopLeftRadius:    top    ? 14 : 0,
      borderTopRightRadius:   top    ? 14 : 0,
      borderBottomLeftRadius: bottom ? 14 : 0,
      borderBottomRightRadius:bottom ? 14 : 0,
      borderBottom: bottom ? 'none' : `0.5px solid ${theme.separator}`,
      cursor: onClick ? 'pointer' : 'default',
    }}>
      <div style={{
        width: 28, height: 28, borderRadius: 7,
        background: danger ? 'rgba(255,93,110,0.14)' : theme.chipBg,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        flexShrink: 0,
      }}>
        <Icon name={icon} size={15} color={danger ? '#FF5D6E' : theme.label2}/>
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{
          fontSize: 15, color: danger ? '#FF5D6E' : theme.label, fontWeight: 500,
          letterSpacing: -0.1,
        }}>{label}</div>
      </div>
      {value && <div style={{ fontSize: 13, color: theme.label3 }}>{value}</div>}
      {!danger && <Icon name="chevron-right" size={14} color={theme.label4}/>}
    </div>
  );
}
