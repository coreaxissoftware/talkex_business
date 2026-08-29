import { useEffect, useState } from 'react'

type Theme = 'light' | 'dark' | 'system'

const KEY = 'talkex-theme'

function apply(theme: Theme) {
  const root = document.documentElement
  root.classList.remove('light', 'dark')
  if (theme === 'system') {
    // Let @media (prefers-color-scheme: dark) rule handle it
    return
  }
  root.classList.add(theme)
}

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(() => {
    try {
      return (localStorage.getItem(KEY) as Theme) || 'system'
    } catch {
      return 'system'
    }
  })

  useEffect(() => {
    apply(theme)
    try { localStorage.setItem(KEY, theme) } catch {}
  }, [theme])

  return {
    theme,
    setTheme: setThemeState,
    toggle: () => {
      // system → dark → light → dark cycle for quick toggle
      setThemeState((t) => (t === 'dark' ? 'light' : 'dark'))
    },
  }
}
