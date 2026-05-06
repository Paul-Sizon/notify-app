// Design tokens — Signal Monitor
// Direct port of web-design-example/project/tokens.jsx

export type ThemeMode = 'dark' | 'light';
export type AccentName = 'amber' | 'green' | 'violet' | 'coral' | 'mono';

export const TOKENS = {
  dark: {
    bg:          '#0A0A0C',
    bgElev1:     '#141417',
    bgElev2:     '#18181C',
    bgElev3:     '#1F1F24',
    separator:   'rgba(255,255,255,0.12)',
    hairline:    'rgba(255,255,255,0.06)',
    label:       '#FFFFFF',
    label2:      'rgba(255,255,255,0.62)',
    label3:      'rgba(255,255,255,0.38)',
    label4:      'rgba(255,255,255,0.22)',
    chipBg:      'rgba(255,255,255,0.08)',
    glassBg:     'rgba(20,20,23,0.72)',
    glassBorder: 'rgba(255,255,255,0.08)',
  },
  light: {
    bg:          '#FFFFFF',
    bgElev1:     '#F2F2F7',
    bgElev2:     '#FFFFFF',
    bgElev3:     '#E5E5EA',
    separator:   'rgba(60,60,67,0.18)',
    hairline:    'rgba(0,0,0,0.06)',
    label:       '#000000',
    label2:      'rgba(60,60,67,0.78)',
    label3:      'rgba(60,60,67,0.50)',
    label4:      'rgba(60,60,67,0.30)',
    chipBg:      'rgba(118,118,128,0.12)',
    glassBg:     'rgba(255,255,255,0.72)',
    glassBorder: 'rgba(0,0,0,0.06)',
  },
  accents: {
    amber:  { hex: '#FF9F1C', soft: 'rgba(255,159,28,0.16)', glow: 'rgba(255,159,28,0.45)' },
    green:  { hex: '#3DD68C', soft: 'rgba(61,214,140,0.16)', glow: 'rgba(61,214,140,0.45)' },
    violet: { hex: '#A78BFA', soft: 'rgba(167,139,250,0.16)', glow: 'rgba(167,139,250,0.45)' },
    coral:  { hex: '#FF7A6B', soft: 'rgba(255,122,107,0.16)', glow: 'rgba(255,122,107,0.45)' },
    mono:   { hex: '#F2F2F7', soft: 'rgba(242,242,247,0.14)', glow: 'rgba(242,242,247,0.4)' },
  },
  font:     '-apple-system, "SF Pro Text", "SF Pro Display", "Inter", system-ui, sans-serif',
  fontMono: '"SF Mono", ui-monospace, "JetBrains Mono", Menlo, monospace',
  radii:    { card: 16, input: 12, chip: 999, sheet: 28 },
} as const;

export interface Theme {
  mode: ThemeMode;
  bg: string; bgElev1: string; bgElev2: string; bgElev3: string;
  separator: string; hairline: string;
  label: string; label2: string; label3: string; label4: string;
  chipBg: string; glassBg: string; glassBorder: string;
  accent: string; accentSoft: string; accentGlow: string;
}

export function buildTheme(mode: ThemeMode = 'dark', accentName: AccentName = 'amber'): Theme {
  const t = TOKENS[mode];
  const a = TOKENS.accents[accentName] ?? TOKENS.accents.amber;
  return { ...t, mode, accent: a.hex, accentSoft: a.soft, accentGlow: a.glow };
}
