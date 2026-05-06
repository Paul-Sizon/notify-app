// Signal Monitor — primitives & icons
// SF Symbols recreated as inline SVG for fidelity

function Icon({ name, size = 17, color = 'currentColor', weight = 'regular', style }) {
  const sw = weight === 'bold' ? 2.4 : weight === 'medium' ? 2.0 : 1.7;
  const props = {
    width: size, height: size, viewBox: '0 0 24 24', fill: 'none',
    stroke: color, strokeWidth: sw, strokeLinecap: 'round', strokeLinejoin: 'round',
    style,
  };
  switch (name) {
    case 'plus':
      return <svg {...props}><path d="M12 5v14M5 12h14"/></svg>;
    case 'check':
      return <svg {...props}><path d="M5 12.5l5 5 9-11"/></svg>;
    case 'chevron-right':
      return <svg {...props}><path d="M9 6l6 6-6 6"/></svg>;
    case 'chevron-left':
      return <svg {...props}><path d="M15 6l-6 6 6 6"/></svg>;
    case 'ellipsis':
      return <svg {...props} fill={color} stroke="none">
        <circle cx="5" cy="12" r="1.7"/><circle cx="12" cy="12" r="1.7"/><circle cx="19" cy="12" r="1.7"/></svg>;
    case 'magnifier':
      return <svg {...props}><circle cx="11" cy="11" r="7"/><path d="M16 16l5 5"/></svg>;
    case 'arrow-up-right':
      return <svg {...props}><path d="M7 17L17 7M9 7h8v8"/></svg>;
    case 'bell':
      return <svg {...props}><path d="M18 16v-5a6 6 0 10-12 0v5l-2 3h16l-2-3z"/><path d="M10 21a2 2 0 004 0"/></svg>;
    case 'bell-fill':
      return <svg {...props} fill={color} stroke="none"><path d="M18 16v-5a6 6 0 10-12 0v5l-2 3h16l-2-3z"/><path d="M10 21a2 2 0 004 0" stroke={color} strokeWidth="1.5"/></svg>;
    case 'trash':
      return <svg {...props}><path d="M4 7h16M9 7V4h6v3M6 7l1 13h10l1-13"/></svg>;
    case 'play':
      return <svg {...props} fill={color} stroke="none"><path d="M7 5l12 7-12 7V5z"/></svg>;
    case 'sparkle':
      return <svg {...props}><path d="M12 3l1.7 5.3L19 10l-5.3 1.7L12 17l-1.7-5.3L5 10l5.3-1.7L12 3zM19 17l.7 1.5L21 19l-1.3.5L19 21l-.7-1.5L17 19l1.3-.5L19 17z"/></svg>;
    case 'waveform':
      return <svg {...props}>
        <path d="M3 12h2M7 12h0M9 8v8M13 5v14M17 9v6M21 12h-2"/>
      </svg>;
    case 'globe':
      return <svg {...props}><circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a14 14 0 010 18M12 3a14 14 0 000 18"/></svg>;
    case 'doc-text':
      return <svg {...props}><path d="M7 3h7l5 5v13H7V3z"/><path d="M14 3v5h5M9 13h7M9 17h7"/></svg>;
    case 'calendar':
      return <svg {...props}><rect x="4" y="5" width="16" height="16" rx="2"/><path d="M4 9h16M9 3v4M15 3v4"/></svg>;
    case 'newspaper':
      return <svg {...props}><rect x="3" y="5" width="18" height="14" rx="1"/><path d="M7 9h6M7 13h6M7 17h6M16 9h2M16 13h2M16 17h2"/></svg>;
    case 'arrow-clockwise':
      return <svg {...props}><path d="M3 12a9 9 0 1015-6.7L21 8M21 3v5h-5"/></svg>;
    case 'pause':
      return <svg {...props} fill={color} stroke="none"><rect x="6" y="5" width="4" height="14"/><rect x="14" y="5" width="4" height="14"/></svg>;
    case 'x':
      return <svg {...props}><path d="M6 6l12 12M18 6L6 18"/></svg>;
    case 'gear':
      return <svg {...props}><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 00.3 1.8l.1.1a2 2 0 11-2.8 2.8l-.1-.1a1.7 1.7 0 00-1.8-.3 1.7 1.7 0 00-1 1.5V21a2 2 0 11-4 0v-.1a1.7 1.7 0 00-1.1-1.5 1.7 1.7 0 00-1.8.3l-.1.1a2 2 0 11-2.8-2.8l.1-.1a1.7 1.7 0 00.3-1.8 1.7 1.7 0 00-1.5-1H3a2 2 0 110-4h.1a1.7 1.7 0 001.5-1.1 1.7 1.7 0 00-.3-1.8l-.1-.1a2 2 0 112.8-2.8l.1.1a1.7 1.7 0 001.8.3h0a1.7 1.7 0 001-1.5V3a2 2 0 114 0v.1a1.7 1.7 0 001 1.5 1.7 1.7 0 001.8-.3l.1-.1a2 2 0 112.8 2.8l-.1.1a1.7 1.7 0 00-.3 1.8v0a1.7 1.7 0 001.5 1H21a2 2 0 110 4h-.1a1.7 1.7 0 00-1.5 1z"/></svg>;
    case 'flame':
      return <svg {...props}><path d="M12 22a7 7 0 007-7c0-3-2-5-2-7s-2-4-5-6c0 4-3 5-3 9 0-2-1-3-2-3-1 2-2 3-2 7a7 7 0 007 7z"/></svg>;
    case 'eye':
      return <svg {...props}><path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z"/><circle cx="12" cy="12" r="3"/></svg>;
    case 'shield':
      return <svg {...props}><path d="M12 3l8 3v6c0 5-3.5 8-8 9-4.5-1-8-4-8-9V6l8-3z"/></svg>;
    case 'funnel':
      return <svg {...props}><path d="M3 5h18l-7 9v6l-4-2v-4L3 5z"/></svg>;
    case 'pin':
      return <svg {...props}><path d="M12 2l3 5 5 1-4 4 1 6-5-3-5 3 1-6-4-4 5-1 3-5z"/></svg>;
    case 'antenna':
      return <svg {...props}>
        <path d="M5 9a8 8 0 0114 0M8 11a5 5 0 018 0"/>
        <circle cx="12" cy="13" r="1.5" fill={color}/>
        <path d="M12 14v7M9 21h6"/>
      </svg>;
    default:
      return <svg {...props}><circle cx="12" cy="12" r="8"/></svg>;
  }
}

// Liquid glass surface — used for floating elements
function GlassSurface({ children, theme, style, radius = 28, intense = false }) {
  return (
    <div style={{
      position: 'relative',
      borderRadius: radius,
      overflow: 'hidden',
      ...style,
    }}>
      <div style={{
        position: 'absolute', inset: 0,
        background: theme.glassBg,
        backdropFilter: `blur(${intense ? 40 : 24}px) saturate(180%)`,
        WebkitBackdropFilter: `blur(${intense ? 40 : 24}px) saturate(180%)`,
      }}/>
      <div style={{
        position: 'absolute', inset: 0, borderRadius: radius,
        boxShadow: theme.mode === 'dark'
          ? 'inset 0 1px 0 rgba(255,255,255,0.08), inset 0 0 0 0.5px rgba(255,255,255,0.06)'
          : 'inset 0 1px 0 rgba(255,255,255,0.6), inset 0 0 0 0.5px rgba(0,0,0,0.04)',
        pointerEvents: 'none',
      }}/>
      <div style={{ position: 'relative', zIndex: 1 }}>{children}</div>
    </div>
  );
}

// Subtle chip — system gray translucent
function Chip({ children, theme, accent = false, small = false, style }) {
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 4,
      height: small ? 18 : 22,
      padding: small ? '0 7px' : '0 9px',
      borderRadius: 999,
      background: accent ? theme.accentSoft : theme.chipBg,
      color: accent ? theme.accent : theme.label2,
      fontSize: small ? 11 : 12,
      fontWeight: 510,
      letterSpacing: 0.1,
      whiteSpace: 'nowrap',
      ...style,
    }}>{children}</span>
  );
}

// Mini waveform glyph — appears as the "signal metaphor" cue
function WaveGlyph({ color = 'currentColor', size = 14, animated = false }) {
  const bars = [0.4, 0.7, 1.0, 0.6, 0.85, 0.45];
  return (
    <svg width={size * 1.5} height={size} viewBox="0 0 24 16" style={{ display: 'block' }}>
      {bars.map((h, i) => (
        <rect
          key={i}
          x={i * 4 + 1} y={8 - (h * 7)}
          width="2" height={h * 14}
          rx="1" fill={color}
          style={animated ? {
            animation: `wave-bar 1.2s ${i * 0.12}s ease-in-out infinite`,
            transformOrigin: 'center',
          } : undefined}
        />
      ))}
    </svg>
  );
}

window.Icon = Icon;
window.GlassSurface = GlassSurface;
window.Chip = Chip;
window.WaveGlyph = WaveGlyph;
