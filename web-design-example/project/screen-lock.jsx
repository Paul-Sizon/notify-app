// Lockscreen with stacked Signal Monitor notifications
function LockScreen({ theme, accentName, onTap }) {
  const t = theme;
  const wallpaper = theme.mode === 'dark'
    ? 'radial-gradient(120% 80% at 50% 0%, #1a1a1f 0%, #050507 60%, #000 100%)'
    : 'linear-gradient(180deg, #c9d6e8 0%, #e8d8c2 100%)';

  return (
    <div style={{
      position: 'absolute', inset: 0, zIndex: 110,
      background: wallpaper,
      borderRadius: 48, overflow: 'hidden',
      display: 'flex', flexDirection: 'column',
    }}>
      {/* Time */}
      <div style={{
        textAlign: 'center', paddingTop: 80,
        color: theme.mode === 'dark' ? '#fff' : '#000',
      }}>
        <div style={{
          fontSize: 18, fontWeight: 500, letterSpacing: 0.3, opacity: 0.85,
        }}>Tuesday, May 4</div>
        <div style={{
          fontSize: 96, fontWeight: 300, letterSpacing: -4,
          lineHeight: 1, marginTop: -2,
          fontFamily: '-apple-system, "SF Pro Display", system-ui',
        }}>9:41</div>
      </div>

      {/* Stacked notifications — thread-id grouping */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'flex-end', padding: '0 12px 110px' }}>
        <div style={{
          fontSize: 12, fontWeight: 600,
          color: theme.mode === 'dark' ? 'rgba(255,255,255,0.7)' : 'rgba(0,0,0,0.5)',
          padding: '0 8px 6px', letterSpacing: 0.3,
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        }}>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
            <Icon name="antenna" size={11}
              color={theme.mode === 'dark' ? 'rgba(255,255,255,0.7)' : 'rgba(0,0,0,0.5)'}/>
            Notifications
          </span>
          <span>3 new</span>
        </div>

        {/* Stack — Signal Monitor (grouped by thread-id) */}
        <div style={{ position: 'relative', height: 92, marginBottom: 8 }} onClick={onTap}>
          {/* Back card 2 */}
          <div style={{
            position: 'absolute', left: 12, right: 12, bottom: 0, height: 70,
            borderRadius: 18, background: 'rgba(255,255,255,0.08)',
            backdropFilter: 'blur(20px) saturate(180%)',
            WebkitBackdropFilter: 'blur(20px) saturate(180%)',
            border: '0.5px solid rgba(255,255,255,0.06)',
            transform: 'scale(0.95)', transformOrigin: 'top center',
          }}/>
          {/* Back card 1 */}
          <div style={{
            position: 'absolute', left: 6, right: 6, bottom: 4, height: 76,
            borderRadius: 19, background: 'rgba(255,255,255,0.10)',
            backdropFilter: 'blur(20px) saturate(180%)',
            WebkitBackdropFilter: 'blur(20px) saturate(180%)',
            border: '0.5px solid rgba(255,255,255,0.06)',
            transform: 'scale(0.97)', transformOrigin: 'top center',
          }}/>
          {/* Front card */}
          <div style={{
            position: 'absolute', left: 0, right: 0, bottom: 8,
            borderRadius: 20, padding: '12px 14px',
            background: theme.mode === 'dark' ? 'rgba(28,28,30,0.78)' : 'rgba(255,255,255,0.78)',
            backdropFilter: 'blur(28px) saturate(180%)',
            WebkitBackdropFilter: 'blur(28px) saturate(180%)',
            border: '0.5px solid rgba(255,255,255,0.10)',
            cursor: 'pointer',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <div style={{
                width: 38, height: 38, borderRadius: 9,
                background: theme.mode === 'dark' ? '#0E0E10' : '#fff',
                border: `1px solid ${theme.hairline}`,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                position: 'relative', flexShrink: 0,
              }}>
                <Icon name="antenna" size={20} color={theme.accent}/>
                <div style={{
                  position: 'absolute', top: -2, right: -2,
                  background: theme.accent, color: theme.mode === 'dark' ? '#000' : '#fff',
                  fontSize: 10, fontWeight: 700,
                  width: 16, height: 16, borderRadius: 999,
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  border: theme.mode === 'dark' ? '2px solid #000' : '2px solid #fff',
                }}>3</div>
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{
                  display: 'flex', alignItems: 'baseline', justifyContent: 'space-between',
                  marginBottom: 2,
                }}>
                  <div style={{
                    fontSize: 13, fontWeight: 600,
                    color: theme.mode === 'dark' ? '#fff' : '#000',
                    letterSpacing: -0.1,
                  }}>blockchain meetups curitiba</div>
                  <div style={{
                    fontSize: 11,
                    color: theme.mode === 'dark' ? 'rgba(255,255,255,0.5)' : 'rgba(0,0,0,0.45)',
                  }}>now</div>
                </div>
                <div style={{
                  fontSize: 13, lineHeight: 1.3,
                  color: theme.mode === 'dark' ? 'rgba(255,255,255,0.85)' : 'rgba(0,0,0,0.78)',
                  textWrap: 'pretty',
                  overflow: 'hidden',
                  display: '-webkit-box',
                  WebkitLineClamp: 2,
                  WebkitBoxOrient: 'vertical',
                }}>ETH Curitiba — June Builders Night. Solidity workshop + lightning talks.</div>
              </div>
            </div>
          </div>
        </div>

        {/* Singleton notification — different thread */}
        <div style={{
          borderRadius: 18, padding: '12px 14px',
          background: theme.mode === 'dark' ? 'rgba(28,28,30,0.78)' : 'rgba(255,255,255,0.78)',
          backdropFilter: 'blur(28px) saturate(180%)',
          WebkitBackdropFilter: 'blur(28px) saturate(180%)',
          border: '0.5px solid rgba(255,255,255,0.10)',
          marginTop: 12,
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <div style={{
              width: 38, height: 38, borderRadius: 9,
              background: theme.mode === 'dark' ? '#0E0E10' : '#fff',
              border: `1px solid ${theme.hairline}`,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              flexShrink: 0,
            }}>
              <Icon name="antenna" size={20} color={theme.accent}/>
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{
                display: 'flex', alignItems: 'baseline', justifyContent: 'space-between',
                marginBottom: 2,
              }}>
                <div style={{
                  fontSize: 13, fontWeight: 600,
                  color: theme.mode === 'dark' ? '#fff' : '#000',
                }}>phoebe bridgers tour announcements</div>
                <div style={{
                  fontSize: 11,
                  color: theme.mode === 'dark' ? 'rgba(255,255,255,0.5)' : 'rgba(0,0,0,0.45)',
                }}>4h ago</div>
              </div>
              <div style={{
                fontSize: 13, lineHeight: 1.3,
                color: theme.mode === 'dark' ? 'rgba(255,255,255,0.85)' : 'rgba(0,0,0,0.78)',
              }}>Phoebe Bridgers adds South American leg for fall 2026.</div>
            </div>
          </div>
        </div>

        {/* "Tap to unlock" hint */}
        <div style={{
          textAlign: 'center', marginTop: 16,
          fontSize: 12,
          color: theme.mode === 'dark' ? 'rgba(255,255,255,0.5)' : 'rgba(0,0,0,0.45)',
        }}>Tap a signal to open</div>
      </div>
    </div>
  );
}

window.LockScreen = LockScreen;
