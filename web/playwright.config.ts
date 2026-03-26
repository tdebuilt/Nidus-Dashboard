import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  forbidOnly: !!process.env.CI,
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['html'], ['github']] : 'html',
  globalSetup: './e2e/global-setup.ts',
  timeout: process.env.CI ? 90000 : 30000,
  expect: {
    timeout: process.env.CI ? 30000 : 5000,
  },
  use: {
    baseURL: 'http://localhost:3777',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    actionTimeout: process.env.CI ? 15000 : 10000,
  },
  projects: [
    // Auth setup — runs first, completes setup wizard and saves storage state
    {
      name: 'auth-setup',
      testMatch: /auth\.setup\.ts/,
      use: { ...devices['Desktop Chrome'] },
    },
    // Functional E2E tests — depend on auth setup
    {
      name: 'e2e',
      dependencies: ['auth-setup'],
      fullyParallel: false,
      testMatch: /.*\.spec\.ts/,
      testIgnore: /responsive\.spec\.ts/,
      use: {
        ...devices['Desktop Chrome'],
        storageState: 'e2e/.auth/user.json',
      },
    },
    // Responsive tests (mocked, no real backend needed)
    {
      name: 'mobile',
      testMatch: /responsive\.spec\.ts/,
      use: { ...devices['iPhone 13'] },
    },
    {
      name: 'tablet',
      testMatch: /responsive\.spec\.ts/,
      use: {
        viewport: { width: 768, height: 1024 },
        userAgent: 'Mozilla/5.0 (iPad; CPU OS 15_0 like Mac OS X)',
      },
    },
    {
      name: 'desktop',
      testMatch: /responsive\.spec\.ts/,
      use: {
        viewport: { width: 1280, height: 800 },
      },
    },
    {
      name: 'tv',
      testMatch: /responsive\.spec\.ts/,
      use: {
        viewport: { width: 1920, height: 1080 },
      },
    },
  ],
  webServer: {
    command: 'cd .. && NIDUS_DB_PATH=./data/e2e-test.db NIDUS_AUTH_RATE_LIMIT=1000 ./nidus 2>&1 | tee /tmp/nidus-e2e.log',
    url: 'http://localhost:3777/api/health',
    reuseExistingServer: !process.env.CI,
    stdout: 'pipe',
    stderr: 'pipe',
    timeout: 60000,
  },
})
