import { Outlet } from 'react-router'

/**
 * AuthLayout — split-screen login/register shell matched to the marketing
 * site's Ocean palette (paper #F7F3EE / jade #0EA5A0 / midnight #0A0E27 /
 * Instrument Serif italic display + Inter Tight body + JetBrains Mono
 * timestamps).
 *
 * Left = branded panel with a chat-thread preview so the merchant sees
 * what they're signing into. Right = the auth form itself.
 *
 * The whole panel is a single Google Fonts pull; every other rule is
 * inline so the auth screens don't depend on Tailwind's config
 * survival across theme swaps.
 */
export default function AuthLayout() {
  return (
    <>
      <link
        rel="stylesheet"
        href="https://fonts.googleapis.com/css2?family=Instrument+Serif:ital@0;1&family=Inter+Tight:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap"
      />

      <style>{`
        :root {
          --paper: #F7F3EE;
          --paper-2: #EFEAE1;
          --ink: #0F172A;
          --ink-soft: #334155;
          --muted: #64748B;
          --hair: rgba(15, 23, 42, 0.09);
          --jade: #0EA5A0;
          --jade-deep: #0B7F7B;
          --coral: #F97066;
          --peri: #7C6AF6;
          --midnight: #0A0E27;
          --serif: 'Instrument Serif', 'Georgia', serif;
          --sans: 'Inter Tight', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
          --mono: 'JetBrains Mono', ui-monospace, Menlo, monospace;
        }
        .auth-shell { font-family: var(--sans); background: var(--paper); color: var(--ink); }
        .brand-panel {
          background: radial-gradient(ellipse at top left, #101838 0%, var(--midnight) 60%);
          color: #F7F3EE;
        }
        .brand-mark {
          background: linear-gradient(135deg, var(--jade), var(--jade-deep));
          box-shadow: 0 12px 32px rgba(14, 165, 160, 0.35);
        }
        .thread-bubble-in {
          background: #ffffff10;
          border: 1px solid #ffffff18;
          backdrop-filter: blur(8px);
        }
        .thread-bubble-out {
          background: #DCF8C6;
          color: var(--ink);
        }
        .form-input {
          font-family: var(--sans);
          font-size: 14px;
          background: #FFFFFF;
          border: 1px solid var(--hair);
          border-radius: 10px;
          padding: 11px 14px;
          width: 100%;
          transition: border-color .15s, box-shadow .15s;
          color: var(--ink);
        }
        .form-input:focus { outline: none; border-color: var(--jade); box-shadow: 0 0 0 3px rgba(14, 165, 160, 0.18); }
        .form-label {
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 0.06em;
          color: var(--muted);
          margin-bottom: 6px;
          display: block;
          font-weight: 600;
        }
        .btn-primary {
          background: var(--ink);
          color: var(--paper);
          font-weight: 600;
          padding: 12px 18px;
          border-radius: 10px;
          font-size: 14px;
          transition: transform .1s, background .15s;
          border: none;
          cursor: pointer;
          width: 100%;
          display: inline-flex; align-items: center; justify-content: center; gap: 8px;
        }
        .btn-primary:hover { background: #1a2540; }
        .btn-primary:active { transform: translateY(1px); }
        .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
        .btn-oauth {
          background: #FFFFFF;
          color: var(--ink);
          border: 1px solid var(--hair);
          border-radius: 10px;
          padding: 10px 14px;
          font-size: 13px;
          font-weight: 500;
          display: inline-flex; align-items: center; justify-content: center; gap: 10px;
          transition: border-color .15s, background .15s;
          cursor: pointer;
          font-family: var(--sans);
        }
        .btn-oauth:hover { border-color: var(--ink); background: var(--paper-2); }
        .headline-serif {
          font-family: var(--serif);
          font-style: italic;
          font-weight: 400;
          font-size: 34px;
          line-height: 1.15;
          letter-spacing: -0.01em;
          color: var(--ink);
        }
        .link-jade { color: var(--jade); font-weight: 600; }
        .link-jade:hover { color: var(--jade-deep); }
        .hairline { height: 1px; background: var(--hair); }
        .status-dot {
          width: 8px; height: 8px; border-radius: 50%;
          background: var(--jade);
          box-shadow: 0 0 0 4px rgba(14, 165, 160, 0.2);
          animation: pulse 2.4s ease-in-out infinite;
        }
        @keyframes pulse {
          0%, 100% { box-shadow: 0 0 0 4px rgba(14, 165, 160, 0.2); }
          50%      { box-shadow: 0 0 0 8px rgba(14, 165, 160, 0.05); }
        }
        .channel-chip {
          font-family: var(--mono);
          font-size: 10px;
          padding: 2px 8px;
          border-radius: 999px;
          background: #ffffff10;
          border: 1px solid #ffffff20;
          color: #F7F3EE;
        }
      `}</style>

      <div className="auth-shell" style={{ display: 'flex', minHeight: '100vh' }}>
        {/* Left — brand + chat preview (hidden on mobile) */}
        <div
          className="brand-panel"
          style={{
            display: 'none',
            width: '52%',
            padding: '48px 56px',
            flexDirection: 'column',
            justifyContent: 'space-between',
          }}
          data-brand-panel
        >
          <BrandPanelInner />
        </div>

        {/* Right — form */}
        <div
          style={{
            flex: 1,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '32px 24px',
            background: 'var(--paper)',
          }}
        >
          <div style={{ width: '100%', maxWidth: 420 }}>
            <Outlet />
            <p style={{
              marginTop: 40,
              textAlign: 'center',
              fontSize: 11,
              color: 'var(--muted)',
              letterSpacing: '0.03em',
            }}>
              by <span style={{ fontFamily: 'var(--serif)', fontStyle: 'italic' }}>CoreAxis Ventures</span>
            </p>
          </div>
        </div>

        {/* Show the brand panel on lg+ via a media query in a style tag —
            React doesn't do @media in inline style. */}
        <style>{`
          @media (min-width: 1024px) {
            [data-brand-panel] { display: flex !important; }
          }
        `}</style>
      </div>
    </>
  )
}

function BrandPanelInner() {
  return (
    <>
      <div>
        {/* Header row: brand-mark + live status */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 56 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div
              className="brand-mark"
              style={{
                width: 40, height: 40, borderRadius: 12,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}
            >
              <span style={{ color: '#F7F3EE', fontWeight: 800, fontSize: 18, letterSpacing: '-0.03em' }}>T</span>
            </div>
            <div style={{ fontSize: 15, fontWeight: 600, letterSpacing: '-0.01em' }}>
              TalkEx <span style={{ opacity: 0.6 }}>Business</span>
            </div>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 11, opacity: 0.7 }}>
            <span className="status-dot"></span>
            <span style={{ fontFamily: 'var(--mono)' }}>All systems normal</span>
          </div>
        </div>

        <h1 style={{
          fontFamily: 'var(--serif)',
          fontStyle: 'italic',
          fontWeight: 400,
          fontSize: 46,
          lineHeight: 1.05,
          letterSpacing: '-0.015em',
          marginBottom: 20,
        }}>
          One inbox.<br />
          <span style={{ color: '#7DD3D0' }}>Every</span> messaging channel.
        </h1>

        <p style={{ fontSize: 15, lineHeight: 1.6, opacity: 0.75, maxWidth: 380 }}>
          TalkEx, WhatsApp, SMS, Email, Telegram, Instagram, Messenger, RCS — one place to send from, one place to read replies. Chatbots, campaigns, and a wallet that meters every rupee.
        </p>

        <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', marginTop: 20 }}>
          {['TalkEx', 'WhatsApp', 'SMS', 'Email', 'Telegram', '+3'].map(c => (
            <span key={c} className="channel-chip">{c}</span>
          ))}
        </div>
      </div>

      {/* Chat-thread preview */}
      <div style={{ maxWidth: 380, marginTop: 40 }}>
        <ChatBubble side="in" name="Aditi (WhatsApp)" time="09:42" text="Hi — quick question about the campaign scheduler, can it fire at 8pm IST?" />
        <ChatBubble side="out" time="09:43" text="Yes — every campaign has a scheduled_at field. Fires ±30 seconds." />
        <ChatBubble side="in" name="Aditi" time="09:43" text="Perfect. Sending you the contact list now ✨" />
      </div>

      <div style={{ fontSize: 11, opacity: 0.5, fontFamily: 'var(--mono)', marginTop: 40 }}>
        v2.4.0 · built {new Date().getFullYear()} · Made in Bengaluru
      </div>
    </>
  )
}

function ChatBubble({
  side,
  name,
  time,
  text,
}: {
  side: 'in' | 'out'
  name?: string
  time: string
  text: string
}) {
  const isOut = side === 'out'
  return (
    <div style={{
      display: 'flex',
      justifyContent: isOut ? 'flex-end' : 'flex-start',
      marginBottom: 10,
    }}>
      <div style={{ maxWidth: '78%' }}>
        {name && (
          <div style={{
            fontSize: 10,
            opacity: 0.55,
            letterSpacing: '0.03em',
            marginBottom: 4,
            paddingLeft: 4,
            fontFamily: 'var(--mono)',
          }}>
            {name}
          </div>
        )}
        <div
          className={isOut ? 'thread-bubble-out' : 'thread-bubble-in'}
          style={{
            padding: '10px 14px',
            borderRadius: isOut ? '16px 16px 4px 16px' : '16px 16px 16px 4px',
            fontSize: 13.5,
            lineHeight: 1.45,
            color: isOut ? '#0F172A' : '#F7F3EE',
          }}
        >
          {text}
          <span style={{
            display: 'inline-block',
            marginLeft: 8,
            fontSize: 10,
            fontFamily: 'var(--mono)',
            opacity: 0.5,
          }}>
            {time}
          </span>
        </div>
      </div>
    </div>
  )
}
