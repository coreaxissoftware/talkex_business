import { test, expect } from '@playwright/test'

// One randomised email per run so we never clash with an existing account.
const email = `smoke_${Date.now()}@talkex.dev`
const password = 'Smoke1234!'

/**
 * Smoke test — the shortest path that proves the stack is alive:
 *   register → land on Dashboard → open Contacts → open Templates →
 *   open Conversations → open Settings.
 *
 * Deliberately shallow (no data creation, no downstream assertions).
 * Doubles as a health probe for the frontend build against a fresh
 * backend, catches broken routes, missing chunks, missing widgets.
 */
test.describe('TalkEx smoke', () => {
  test('register and navigate every top nav', async ({ page }) => {
    // Register
    await page.goto('/register')
    await page.getByLabel(/full name/i).fill('Smoke Tester')
    await page.getByLabel(/^mobile$/i).fill('9876543210')
    await page.getByLabel(/^email$/i).fill(email)
    await page.getByLabel(/^password$/i).fill(password)
    await page.getByLabel(/terms/i).check()
    await page.getByRole('button', { name: /create account/i }).click()

    // Dashboard reachable
    await expect(page).toHaveURL(/\/$|\/dashboard/)
    await expect(page.getByRole('heading', { name: /dashboard/i })).toBeVisible()

    // Sample of navigation targets — one assertion per page proves
    // route + shell + data-fetch bootstrap all work.
    const stops: [string, RegExp][] = [
      ['/contacts', /contacts/i],
      ['/templates', /templates/i],
      ['/campaigns', /campaigns/i],
      ['/conversations', /inbox/i],
      ['/flows', /chatbot/i],
      ['/live-chat', /live chat/i],
      ['/team/activity', /team activity/i],
      ['/settings', /settings/i],
    ]
    for (const [url, expected] of stops) {
      await page.goto(url)
      await expect(page.locator('body')).toContainText(expected, { timeout: 5000 })
    }
  })
})
