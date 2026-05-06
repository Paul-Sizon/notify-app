// Design tokens — Signal Monitor
// Dark-first iOS aesthetic, deep amber accent, restrained signal metaphor

const SM_TOKENS = {
  // Dark theme colors (defaults)
  dark: {
    bg:         '#000000',
    bgElev1:    '#0E0E10',     // groupedBackground
    bgElev2:    '#1C1C1E',     // secondarySystemGroupedBackground
    bgElev3:    '#2C2C2E',     // tertiary
    separator:  'rgba(84, 84, 88, 0.45)',
    hairline:   'rgba(255,255,255,0.06)',
    label:      '#FFFFFF',
    label2:     'rgba(235,235,245,0.78)',  // secondary
    label3:     'rgba(235,235,245,0.50)',  // tertiary
    label4:     'rgba(235,235,245,0.30)',  // quaternary
    chipBg:     'rgba(118,118,128,0.24)',
    glassBg:    'rgba(28,28,30,0.72)',
    glassBorder:'rgba(255,255,255,0.08)',
  },
  light: {
    bg:         '#FFFFFF',
    bgElev1:    '#F2F2F7',
    bgElev2:    '#FFFFFF',
    bgElev3:    '#E5E5EA',
    separator:  'rgba(60,60,67,0.18)',
    hairline:   'rgba(0,0,0,0.06)',
    label:      '#000000',
    label2:     'rgba(60,60,67,0.78)',
    label3:     'rgba(60,60,67,0.50)',
    label4:     'rgba(60,60,67,0.30)',
    chipBg:     'rgba(118,118,128,0.12)',
    glassBg:    'rgba(255,255,255,0.72)',
    glassBorder:'rgba(0,0,0,0.06)',
  },
  // Accent palette (deep amber default)
  accents: {
    amber:   { hex: '#FF9F1C', soft: 'rgba(255,159,28,0.16)', glow: 'rgba(255,159,28,0.45)' },
    green:   { hex: '#7CFFB2', soft: 'rgba(124,255,178,0.14)', glow: 'rgba(124,255,178,0.45)' },
    violet:  { hex: '#A78BFA', soft: 'rgba(167,139,250,0.16)', glow: 'rgba(167,139,250,0.45)' },
    coral:   { hex: '#FF7A6B', soft: 'rgba(255,122,107,0.16)', glow: 'rgba(255,122,107,0.45)' },
    mono:    { hex: '#F2F2F7', soft: 'rgba(242,242,247,0.14)', glow: 'rgba(242,242,247,0.4)' },
  },
  // Type scale (mimics SF Pro)
  font: '-apple-system, "SF Pro Text", "SF Pro Display", "Inter", system-ui, sans-serif',
  fontMono: '"SF Mono", ui-monospace, "JetBrains Mono", Menlo, monospace',

  // Radii — 16 cards, 12 inputs, pills full
  radii: { card: 16, input: 12, chip: 999, sheet: 28 },
  // Spacing — 4pt grid
  s: (n) => n * 4,
};

// Hook to read theme
function useSMTheme(mode, accentName) {
  const t = mode === 'light' ? SM_TOKENS.light : SM_TOKENS.dark;
  const a = SM_TOKENS.accents[accentName] || SM_TOKENS.accents.amber;
  return { ...t, accent: a.hex, accentSoft: a.soft, accentGlow: a.glow, mode };
}

window.SM_TOKENS = SM_TOKENS;
window.useSMTheme = useSMTheme;
