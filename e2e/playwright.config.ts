import { defineConfig } from '@playwright/test'

/**
 * End-to-end smoke tests against a running dev stack:
 *   docker compose up
 *   cd e2e && npm install
 *   npx playwright test
 *
 * The CI job (in .github/workflows/ci.yml) can add a Playwright stage
 * once test coverage grows enough to justify running it on every push.
 */
export default defineConfig({
  testDir: '.',
  timeout: 30_000,
  expect: { timeout: 5_000 },
  fullyParallel: false, // serial keeps auth flows deterministic
  retries: 0,
  reporter: [['list'], ['html', { outputFolder: 'playwright-report', open: 'never' }]],
  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { browserName: 'chromium' },
    },
  ],
})
