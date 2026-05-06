// Bottom tab bar — flat, no liquid glass (matches the user's reference)
function TabBar({ theme, active, onChange }) {
  const tabs = [
    { id: 'watchers', label: 'Watchers', icon: 'eye-line',     iconActive: 'eye-fill' },
    { id: 'alerts',   label: 'Alerts',   icon: 'bell-line',    iconActive: 'bell-active' },
    { id: 'signals',  label: 'Signals',  icon: 'spark-line',   iconActive: 'spark-line' },
    { id: 'account',  label: 'Account',  icon: 'person-line',  iconActive: 'person-fill' },
  ];

  return (
    <div style={{
      position: 'absolute', left: 0, right: 0, bottom: 0,
      paddingBottom: 22, paddingTop: 10,
      background: theme.mode === 'dark'
        ? 'linear-gradient(180deg, rgba(0,0,0,0) 0%, rgba(0,0,0,0.85) 30%, #000 70%)'
        : 'linear-gradient(180deg, rgba(255,255,255,0) 0%, rgba(255,255,255,0.9) 40%, #fff 100%)',
      zIndex: 70,
      pointerEvents: 'auto',
    }}>
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(4, 1fr)',
        padding: '0 8px',
      }}>
        {tabs.map(t => {
          const isActive = active === t.id;
          const color = isActive ? theme.accent : theme.label3;
          return (
            <div key={t.id} onClick={() => onChange?.(t.id)} style={{
              display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4,
              padding: '8px 4px', cursor: 'pointer', userSelect: 'none',
              transition: 'transform 0.15s', transform: isActive ? 'scale(1)' : 'scale(0.98)',
            }}>
              <TabIcon name={isActive ? t.iconActive : t.icon} color={color} size={22}/>
              <div style={{
                fontSize: 11, fontWeight: 600, letterSpacing: 0.1,
                color, fontFamily: 'inherit',
              }}>{t.label}</div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function TabIcon({ name, color, size = 22 }) {
  const sw = 1.7;
  const props = {
    width: size, height: size, viewBox: '0 0 24 24', fill: 'none',
    stroke: color, strokeWidth: sw, strokeLinecap: 'round', strokeLinejoin: 'round',
    style: { display: 'block' },
  };
  switch (name) {
    case 'eye-line':
      return <svg {...props}>
        <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z"/>
        <circle cx="12" cy="12" r="3"/>
      </svg>;
    case 'eye-fill':
      return <svg {...props}>
        <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z" fill={color} fillOpacity="0.18"/>
        <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z"/>
        <circle cx="12" cy="12" r="3" fill={color}/>
      </svg>;
    case 'bell-line':
      return <svg {...props}>
        <path d="M18 16v-5a6 6 0 10-12 0v5l-2 3h16l-2-3z"/>
        <path d="M10 21a2 2 0 004 0"/>
      </svg>;
    case 'bell-active':
      return <svg {...props}>
        <path d="M18 16v-5a6 6 0 10-12 0v5l-2 3h16l-2-3z" fill={color} fillOpacity="0.18"/>
        <path d="M18 16v-5a6 6 0 10-12 0v5l-2 3h16l-2-3z"/>
        <path d="M10 21a2 2 0 004 0"/>
        <path d="M3 6c1-1.5 2.5-2.5 4-3M21 6c-1-1.5-2.5-2.5-4-3" strokeWidth={sw - 0.2}/>
      </svg>;
    case 'spark-line':
      return <svg {...props}>
        <path d="M3 17l5-6 4 3 5-7 4 5"/>
        <circle cx="17" cy="7" r="1.4" fill={color}/>
      </svg>;
    case 'person-line':
      return <svg {...props}>
        <circle cx="12" cy="9" r="4"/>
        <path d="M4 21c0-4 4-6 8-6s8 2 8 6"/>
      </svg>;
    case 'person-fill':
      return <svg {...props}>
        <circle cx="12" cy="9" r="4" fill={color} fillOpacity="0.18"/>
        <circle cx="12" cy="9" r="4"/>
        <path d="M4 21c0-4 4-6 8-6s8 2 8 6"/>
      </svg>;
    default:
      return <svg {...props}><circle cx="12" cy="12" r="6"/></svg>;
  }
}

window.TabBar = TabBar;
