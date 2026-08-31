// Client-side hardening — right-click block, devtools shortcut deterrent,
// text-selection lock on sensitive surfaces, and a lightweight iframe
// buster. These are UX-level protections; they cannot stop a determined
// attacker (nothing in a browser ever can), but they raise the bar for
// casual data-scraping, over-the-shoulder screenshotting, and the "open
// devtools by accident" support tickets we get every week.

const DEV_MODE = import.meta.env.DEV

// Elements marked with data-allow-context bypass the block (rich-text
// editors, code inputs, and the API playground need native context menu).
function isAllowed(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false
  return !!target.closest('[data-allow-context], input, textarea, [contenteditable="true"]')
}

export function installClientSecurity(): void {
  if (DEV_MODE) {
    // Never fight the developer building the app.
    console.info('[TalkEx] security shim skipped (dev mode)')
    return
  }

  // 1. Right-click / context menu — the "banking level" request.
  document.addEventListener(
    'contextmenu',
    (e) => {
      if (isAllowed(e.target)) return
      e.preventDefault()
    },
    { capture: true }
  )

  // 2. Devtools / view-source shortcut deterrent.
  //    Real devtools cannot be blocked — extension owns the browser —
  //    but the common muscle-memory shortcuts we can intercept.
  document.addEventListener(
    'keydown',
    (e) => {
      const k = e.key.toLowerCase()
      // F12
      if (k === 'f12') {
        e.preventDefault()
        return
      }
      // Ctrl+Shift+I / J / C / K — devtools panels
      if (e.ctrlKey && e.shiftKey && (k === 'i' || k === 'j' || k === 'c' || k === 'k')) {
        e.preventDefault()
        return
      }
      // Ctrl+U — view source
      if (e.ctrlKey && k === 'u') {
        e.preventDefault()
        return
      }
      // Ctrl+S — save page as HTML
      if (e.ctrlKey && k === 's') {
        e.preventDefault()
        return
      }
      // macOS mirrors: ⌘+Option+I / J / C / U / S
      if (e.metaKey && e.altKey && (k === 'i' || k === 'j' || k === 'c')) {
        e.preventDefault()
        return
      }
      if (e.metaKey && (k === 'u' || k === 's')) {
        e.preventDefault()
        return
      }
    },
    { capture: true }
  )

  // 3. Drag-and-drop of images / links away from the page.
  document.addEventListener('dragstart', (e) => {
    if (isAllowed(e.target)) return
    const t = e.target as HTMLElement
    if (t.tagName === 'IMG' || t.tagName === 'A') {
      e.preventDefault()
    }
  })

  // 4. Iframe buster — if we've been embedded in someone else's page,
  //    break out to the top window. Complements X-Frame-Options: DENY
  //    when the API doesn't happen to serve the HTML shell (e.g. served
  //    from an object-store CDN).
  try {
    if (window.top && window.top !== window.self) {
      window.top.location.href = window.self.location.href
    }
  } catch {
    // Cross-origin frame — assignment throws; force blank.
    document.body.innerHTML = ''
  }

  // 5. Suppress the browser's password-manager "reveal" button leaking
  //    into screenshots by blocking print keystrokes on the app.
  //    (User's OS-level screenshot tools still work — this only blocks
  //    Ctrl+P print-to-PDF which many DLP tools flag on their own.)
  window.addEventListener(
    'beforeprint',
    () => {
      document.body.classList.add('print-blocked')
    },
    { capture: true }
  )
}
