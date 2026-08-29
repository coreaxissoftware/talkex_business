/**
 * i18n bootstrap — lightweight react-i18next setup with English +
 * Hindi bundles. The rest of the app opts in gradually via useTranslation()
 * on new copy; legacy hard-coded strings keep working until they're migrated.
 *
 * Locale preference is stored in localStorage under `talkex_lang` so it
 * survives reloads. Language selector lives in Settings.
 */
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

import en from './locales/en.json'
import hi from './locales/hi.json'

const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('talkex_lang') : null

i18n
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: en },
      hi: { translation: hi },
    },
    lng: stored || 'en',
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false, // React already escapes
    },
  })

export function setLanguage(lang: 'en' | 'hi') {
  i18n.changeLanguage(lang)
  try { localStorage.setItem('talkex_lang', lang) } catch { /* private mode */ }
}

export default i18n
