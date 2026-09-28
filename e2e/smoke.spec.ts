import { test, expect } from '@playwright/test'

// One randomised email per run so we never clash with an existing account.
const email = `smoke_${Date.now()}@talkex.dev`
const password = 'SmokeTest12345!' // ≥12 chars — matches server policy

/**
 * Smoke test — the shortest path that proves the stack is alive:
 *   register → land on Dashboard → open Contacts → Templates →
 *   Campaigns → Conversations → Flows → Live-chat → Settings.
 *
 * Run against local dev:      E2E_BASE_URL=http://localhost:5173 npx playwright test
 * Or against production:      E2E_BASE_URL=https://businessapp.talkex.in npx playwright test
 *
 * Deliberately shallow (no data creation, no downstream assertions).
 * Doubles as a health probe for the frontend build against a running
 * backend and catches broken routes, missing chunks, wiring holes.
 *
 * The current Register form uses placeholder-based inputs (no <label>
 * associations) so we address them by placeholder text. It renders a
 * mobile+email OTP UI too, but handleSubmit does not gate on OTP
 * verification — a name+email+password+terms submission is enough to
 * land in the dashboard.
 */
test.describe('TalkEx smoke', () => {
  test('register and navigate every top nav', async ({ page }) => {
    await page.goto('/register')

    // Fill by placeholder to match the current design (labels are
    // visual-only, not <label htmlFor=…>-bound).
    await page.getByPlaceholder(/Aditi Sharma/i).fill('Smoke Tester')
    await page.getByPlaceholder(/9876543210/i).fill('9876543210')
    await page.getByPlaceholder(/you@company\.com/i).fill(email)

    // The password input is a custom PasswordInput with an eye toggle;
    // it's still a plain <input type="password">. Match by type since
    // there's no unique placeholder.
    await page.locator('input[type="password"]').first().fill(password)

    // Terms checkbox — first checkbox on the page.
    await page.locator('input[type="checkbox"]').first().check()

    await page.getByRole('button', { name: /create account/i }).click()

    // Dashboard reachable
    await expect(page).toHaveURL(/\/$|\/dashboard/, { timeout: 15_000 })

    // Sample of navigation targets — one assertion per page proves
    // route + shell + data-fetch bootstrap all work. Keep the
    // assertions lenient (text anywhere in body) so a copy tweak
    // doesn't break the smoke.
    const stops: [string, RegExp][] = [
      ['/contacts', /contact/i],
      ['/templates', /template/i],
      ['/campaigns', /campaign/i],
      ['/conversations', /conversation|inbox/i],
      ['/flows', /flow|chatbot/i],
      ['/live-chat', /live chat/i],
      ['/settings', /setting/i],
    ]
    for (const [url, expected] of stops) {
      await page.goto(url)
      await expect(page.locator('body')).toContainText(expected, { timeout: 8_000 })
    }
  })
})
