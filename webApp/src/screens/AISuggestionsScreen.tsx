import { useEffect, useState } from 'react';
import { Icon } from '../primitives';
import type { Theme } from '../tokens';
import { api, cadenceLabel, type SuggestionDTO } from '../api';

const PREFILL = `San Francisco, Senior Software Engineer at a Series B startup (~150 people). Five years in, three at current company. Backend-leaning full-stack — Python and Go daily, currently leading a Postgres-to-distributed-storage migration. Reads Hacker News in the morning, Pragmatic Engineer on weekends. Vaguely thinking about leaving for a smaller team or starting something herself in 2 years.

Lives in the Mission. Runs 4x/week, training for the SF Marathon. Member of a local run club. Hot yoga twice a week at a studio nearby. Climbs at Mission Cliffs on weekends. Mostly cooks — Whole Foods plus farmers' market on Saturdays. Eats out 2x/week at places with real vegetable programs (Souvla, Reem's, Nopa). Doesn't drink much; will go to a natural wine bar with friends. Sober-curious adjacent, interested in NA cocktails.

Cultural taste: indie and electronic — Bon Iver, Caribou, Floating Points, Mitski. Catches small shows at The Independent, The Chapel, Great American. Avoids arena tours. Watches A24 movies. Genuinely interested in AI/ML developments but skeptical of hype cycles, occasionally reads papers. Goes to maybe one tech meetup a month — picks them carefully, hates recruiting-bait events.

Civic life: watches SF politics closely — housing policy, public transit, Prop measures. Concerned about Mission gentrification, BART funding, street safety. She votes and reads the voter guide.`;

const CHAR_LIMIT = 2000;
const MIN_CHARS = 10;
type Phase = 'input' | 'loading' | 'reveal';

/**
 * AI-from-context watcher suggester. Three phases mirror iOS / Compose:
 * input → loading pulse → reveal cards with toggleable activate.
 */
export function AISuggestionsScreen({
  theme, open, onClose, onCreate, onDelete,
}: {
  theme: Theme;
  open: boolean;
  onClose: () => void;
  onCreate: (q: string, type: 'event' | 'news', cadenceSeconds: number) => Promise<{ id: string } | undefined>;
  onDelete: (id: string) => Promise<unknown>;
}) {
  const [phase, setPhase] = useState<Phase>('input');
  const [text, setText] = useState(PREFILL);
  const [suggestions, setSuggestions] = useState<SuggestionDTO[]>([]);
  const [fallback, setFallback] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activated, setActivated] = useState<Record<string, string>>({}); // localKey → subId
  const [busy, setBusy] = useState<Record<string, boolean>>({});

  // Reset on each open
  useEffect(() => {
    if (open) {
      setPhase('input');
      setText(PREFILL);
      setSuggestions([]);
      setFallback(false);
      setError(null);
      setActivated({});
      setBusy({});
    }
  }, [open]);

  const valid = text.trim().length >= MIN_CHARS;
  const keyFor = (s: SuggestionDTO) => `${s.query}|${s.cadence_seconds}`;

  const submit = async () => {
    if (!valid) return;
    setPhase('loading');
    setError(null);
    const start = Date.now();
    try {
      const resp = await api.suggestFromContext(text.trim());
      const elapsed = Date.now() - start;
      if (elapsed < 1000) await new Promise(r => setTimeout(r, 1000 - elapsed));
      setSuggestions(resp.suggestions);
      setFallback(!!resp.fallback);
      setPhase('reveal');
    } catch (e: any) {
      setError(e?.message ?? 'AI request failed');
      setPhase('input');
    }
  };

  const toggle = async (s: SuggestionDTO) => {
    const k = keyFor(s);
    if (busy[k]) return;
    setBusy(b => ({ ...b, [k]: true }));
    try {
      const existing = activated[k];
      if (existing) {
        await onDelete(existing);
        setActivated(a => { const c = { ...a }; delete c[k]; return c; });
      } else {
        const sub = await onCreate(s.query, s.type, s.cadence_seconds);
        if (sub) setActivated(a => ({ ...a, [k]: sub.id }));
      }
    } finally {
      setBusy(b => { const c = { ...b }; delete c[k]; return c; });
    }
  };

  if (!open) return null;

  return (
    <div style={{
      position: 'absolute', inset: 0, zIndex: 90,
      background: theme.bg,
      color: theme.label,
      display: 'flex', flexDirection: 'column',
      animation: 'fade-in 0.2s',
    }}>
      {/* Top bar */}
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        padding: '14px 16px',
        borderBottom: `0.5px solid ${theme.bgElev3}`,
        background: theme.bg,
      }}>
        <button
          onClick={onClose}
          style={{
            width: 36, height: 36, borderRadius: 999,
            background: theme.bgElev2, border: 'none', cursor: 'pointer',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            color: theme.label,
          }}
        >
          <Icon name="chevron-left" size={18} weight="bold"/>
        </button>
        <div style={{ fontSize: 16, fontWeight: 600 }}>AI Suggestions</div>
        <div style={{ width: 36 }}/>
      </div>

      {phase === 'input' && (
        <InputView
          theme={theme}
          text={text}
          onChange={setText}
          error={error}
          valid={valid}
          onSubmit={submit}
        />
      )}
      {phase === 'loading' && <LoadingView theme={theme}/>}
      {phase === 'reveal' && (
        <RevealView
          theme={theme}
          suggestions={suggestions}
          fallback={fallback}
          activated={activated}
          busy={busy}
          onToggle={toggle}
          onDone={onClose}
          keyFor={keyFor}
        />
      )}
    </div>
  );
}

function InputView({
  theme, text, onChange, error, valid, onSubmit,
}: {
  theme: Theme; text: string; onChange: (s: string) => void;
  error: string | null; valid: boolean; onSubmit: () => void;
}) {
  return (
    <div style={{ flex: 1, overflow: 'auto', display: 'flex', flexDirection: 'column' }}>
      <div style={{ flex: 1, padding: '16px 22px 24px', display: 'flex', flexDirection: 'column', gap: 18 }}>
        {/* Sparkle badge */}
        <div style={{ alignSelf: 'center' }}>
          <div style={{
            width: 64, height: 64, borderRadius: 14,
            background: theme.bgElev2,
            border: `0.5px solid ${theme.bgElev3}`,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>
            <Icon name="sparkle" size={28} color={theme.accent}/>
          </div>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 10, alignItems: 'center', textAlign: 'center' }}>
          <h1 style={{ margin: 0, fontSize: 28, fontWeight: 700 }}>Tell us what to look for.</h1>
          <p style={{ margin: 0, fontSize: 14, color: theme.label2, lineHeight: 1.5 }}>
            Describe your current interests, projects, or specific signals you want the AI to track.
          </p>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div style={{
            fontSize: 11, fontWeight: 600, letterSpacing: 0.5,
            color: theme.label3, textTransform: 'uppercase',
          }}>CONTEXT &amp; PREFERENCES</div>
          <textarea
            value={text}
            onChange={e => {
              const v = e.target.value;
              onChange(v.length > CHAR_LIMIT ? v.slice(0, CHAR_LIMIT) : v);
            }}
            style={{
              width: '100%', minHeight: 200,
              borderRadius: 16,
              background: theme.bgElev2,
              border: `0.5px solid ${theme.bgElev3}`,
              color: theme.label,
              padding: 14, fontSize: 14, lineHeight: 1.45,
              fontFamily: 'inherit', resize: 'vertical',
              boxSizing: 'border-box',
              outline: 'none',
            }}
          />
          <div style={{ display: 'flex', justifyContent: 'flex-end', fontSize: 12, color: theme.label3, fontFamily: 'monospace' }}>
            {text.length} / {CHAR_LIMIT}
          </div>
        </div>

        {error && (
          <div style={{ fontSize: 13, color: '#ff5b54' }}>{error}</div>
        )}
      </div>

      <div style={{
        padding: '14px 22px 22px',
        borderTop: `0.5px solid ${theme.bgElev3}`,
        background: theme.bg,
      }}>
        <button
          onClick={onSubmit}
          disabled={!valid}
          style={{
            width: '100%', minHeight: 54,
            borderRadius: 999,
            background: theme.accent,
            color: theme.mode === 'dark' ? '#062814' : '#fff',
            border: 'none',
            opacity: valid ? 1 : 0.4,
            cursor: valid ? 'pointer' : 'default',
            fontSize: 16, fontWeight: 600,
            display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 10,
          }}
        >
          <Icon name="magnifier" size={16} weight="bold"/>
          Find Signals
        </button>
      </div>
    </div>
  );
}

function LoadingView({ theme }: { theme: Theme }) {
  return (
    <div style={{
      flex: 1, display: 'flex', flexDirection: 'column',
      alignItems: 'center', justifyContent: 'center', gap: 20,
    }}>
      <div style={{
        width: 80, height: 80, borderRadius: 18,
        background: theme.bgElev2,
        border: `0.5px solid ${theme.bgElev3}`,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        animation: 'pulse 1s ease-in-out infinite alternate',
        boxShadow: `0 0 40px ${theme.accentGlow}`,
      }}>
        <Icon name="sparkle" size={32} color={theme.accent}/>
      </div>
      <div style={{ fontSize: 14, color: theme.label2 }}>Scanning the web for signals…</div>
      <style>{'@keyframes pulse{from{transform:scale(0.92)}to{transform:scale(1.04)}}'}</style>
    </div>
  );
}

function RevealView({
  theme, suggestions, fallback, activated, busy, onToggle, onDone, keyFor,
}: {
  theme: Theme;
  suggestions: SuggestionDTO[];
  fallback: boolean;
  activated: Record<string, string>;
  busy: Record<string, boolean>;
  onToggle: (s: SuggestionDTO) => void;
  onDone: () => void;
  keyFor: (s: SuggestionDTO) => string;
}) {
  const count = Object.keys(activated).length;
  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <div style={{ padding: '12px 22px 8px' }}>
        <div style={{ fontSize: 22, fontWeight: 700 }}>
          {fallback ? 'Popular watchers for you' : 'Suggested watchers'}
        </div>
        <div style={{ marginTop: 4, fontSize: 13, color: theme.label2 }}>
          Tap to add. Edit cadence later.
        </div>
      </div>
      <div style={{ flex: 1, overflowY: 'auto', padding: '8px 22px 16px', display: 'flex', flexDirection: 'column', gap: 10 }}>
        {suggestions.map(s => {
          const k = keyFor(s);
          const active = !!activated[k];
          return (
            <div
              key={k}
              onClick={() => onToggle(s)}
              style={{
                padding: 14,
                borderRadius: 16,
                background: theme.bgElev2,
                border: `${active ? 1.2 : 0.5}px solid ${active ? theme.accent : theme.bgElev3}`,
                cursor: busy[k] ? 'wait' : 'pointer',
                opacity: busy[k] ? 0.6 : 1,
                display: 'flex', gap: 12, alignItems: 'flex-start',
              }}
            >
              <div style={{
                width: 36, height: 36, borderRadius: 999,
                background: active ? theme.accent : `${theme.accent}28`,
                color: active ? (theme.mode === 'dark' ? '#062814' : '#fff') : theme.accent,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                flexShrink: 0,
              }}>
                <Icon name={active ? 'check' : (s.type === 'event' ? 'calendar' : 'newspaper')} size={16} weight="bold"/>
              </div>
              <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 4 }}>
                <div style={{ fontSize: 16, fontWeight: 600 }}>{s.query}</div>
                <div style={{ fontSize: 13, color: theme.label3 }}>{s.reason}</div>
                <div style={{ fontSize: 12, color: theme.label3, fontWeight: 500 }}>
                  {s.type === 'event' ? 'Event' : 'News'} · {cadenceLabel(s.cadence_seconds)}
                </div>
              </div>
            </div>
          );
        })}
      </div>
      <div style={{ padding: '12px 22px 22px', borderTop: `0.5px solid ${theme.bgElev3}` }}>
        <button
          onClick={onDone}
          style={{
            width: '100%', minHeight: 54,
            borderRadius: 999,
            background: theme.accent,
            color: theme.mode === 'dark' ? '#062814' : '#fff',
            border: 'none', cursor: 'pointer',
            fontSize: 16, fontWeight: 600,
          }}
        >
          {count === 0 ? 'Done' : `Done · ${count} added`}
        </button>
      </div>
    </div>
  );
}
