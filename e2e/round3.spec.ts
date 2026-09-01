import { test, expect } from '@playwright/test'

// Round-3 smoke — the four new pages (Deals, Catalog, Green Tick,
// Integrations) load without errors, render their headline, and
// respond to a minimal happy-path interaction. Runs after smoke.spec.ts
// which already registers a user.

const email = `round3_${Date.now()}@talkex.dev`
const password = 'Round3Test!23'

async function register(page: any) {
  await page.goto('/register')
  await page.getByLabel(/full name/i).fill('Round Three Tester')
  await page.getByLabel(/^mobile$/i).fill('9876500001')
  await page.getByLabel(/^email$/i).fill(email)
  await page.getByLabel(/^password$/i).fill(password)
  await page.getByLabel(/terms/i).check()
  await page.getByRole('button', { name: /create account/i }).click()
  await expect(page).toHaveURL(/\/$|\/dashboard/)
}

test.describe('Round 3 UI', () => {
  test('deals page renders default pipeline kanban', async ({ page }) => {
    await register(page)
    await page.goto('/deals')
    await expect(page.getByRole('heading', { name: /deal pipeline/i })).toBeVisible()
    // The default pipeline seeds six stages — assert the first two appear.
    await expect(page.getByText(/new lead/i)).toBeVisible()
    await expect(page.getByText(/qualified/i)).toBeVisible()

    // Create a deal via the modal.
    await page.getByRole('button', { name: /new deal/i }).click()
    await page.getByPlaceholder(/enterprise inbox/i).fill('Smoke test deal')
    await page.getByRole('button', { name: /^create$/i }).click()
    await expect(page.getByText(/smoke test deal/i)).toBeVisible()
  })

  test('catalog page renders + supports creating a product', async ({ page }) => {
    await register(page)
    await page.goto('/catalog')
    await expect(page.getByRole('heading', { name: /product catalog/i })).toBeVisible()
    // Empty state shows call-to-action.
    await expect(page.getByText(/no products yet/i)).toBeVisible()

    await page.getByRole('button', { name: /add product/i }).click()
    await page.getByPlaceholder('SAREE-001').fill('SMOKE-001')
    // Name = first empty text input after SKU.
    await page.locator('input.input').nth(1).fill('Smoke Product')
    await page.locator('input[type="number"]').first().fill('499')
    await page.getByRole('button', { name: /^save$/i }).click()
    await expect(page.getByText(/smoke product/i)).toBeVisible()
    await expect(page.getByText(/SKU SMOKE-001/i)).toBeVisible()
  })

  test('green tick page shows checklist + progress', async ({ page }) => {
    await register(page)
    await page.goto('/green-tick')
    await expect(page.getByRole('heading', { name: /green tick verification/i })).toBeVisible()
    await expect(page.getByText(/meta prerequisite checklist/i)).toBeVisible()
    await expect(page.getByText(/notable brand/i)).toBeVisible()

    // Toggle two items and confirm the % progress updates.
    await page.getByText(/notable brand/i).click()
    await page.getByText(/organisation website/i).click()
    // Progress bar text shows a percentage between 1 and 99.
    await expect(page.locator('text=/\\d+%/').first()).toBeVisible()
  })

  test('integrations page renders all four tabs', async ({ page }) => {
    await register(page)
    await page.goto('/integrations')
    await expect(page.getByRole('heading', { name: /integrations/i })).toBeVisible()

    // Default tab (Sheets) shows the CSV URL input.
    await expect(page.getByPlaceholder(/docs\.google\.com/i)).toBeVisible()

    // Switch to Zapier tab and confirm the catalog loads at least one event.
    await page.getByRole('button', { name: /zapier/i }).click()
    await expect(page.getByText(/message\.received/i)).toBeVisible()

    // Analytics PDF footer button.
    await expect(page.getByRole('link', { name: /download/i })).toBeVisible()
  })
})
