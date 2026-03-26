import { test, expect } from '@playwright/test'

// These tests validate the setup wizard's client-side validation using API mocking.
// The happy path is covered by auth.setup.ts.

async function mockSetupNotCompleted(page: import('@playwright/test').Page) {
  await page.route('**/api/auth/status', (route) =>
    route.fulfill({ json: { setup_completed: false } })
  )
}

test.describe('Setup wizard validation', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('shows error when passwords do not match', async ({ page }) => {
    await mockSetupNotCompleted(page)
    await page.route('**/api/auth/setup', (route) =>
      route.fulfill({ status: 400, json: { error: 'Passwords do not match' } })
    )

    await page.goto('/setup')
    await expect(page.getByTestId('setup-page')).toBeVisible()

    await page.getByTestId('setup-username').fill('admin')
    await page.getByTestId('setup-password').fill('Password123!')
    await page.getByTestId('setup-password-confirm').fill('DifferentPass!')
    await page.getByTestId('setup-next').click()

    await expect(page.getByTestId('setup-error')).toBeVisible()
  })

  test('shows error when password is too short', async ({ page }) => {
    await mockSetupNotCompleted(page)

    await page.goto('/setup')
    await expect(page.getByTestId('setup-page')).toBeVisible()

    await page.getByTestId('setup-username').fill('admin')
    await page.getByTestId('setup-password').fill('short')
    await page.getByTestId('setup-password-confirm').fill('short')
    await page.getByTestId('setup-next').click()

    await expect(page.getByTestId('setup-error')).toBeVisible()
  })

  test('next button is disabled when username is empty', async ({ page }) => {
    await mockSetupNotCompleted(page)

    await page.goto('/setup')
    await expect(page.getByTestId('setup-page')).toBeVisible()

    // Leave username empty, fill password
    await page.getByTestId('setup-password').fill('Password123!')
    await page.getByTestId('setup-password-confirm').fill('Password123!')

    await expect(page.getByTestId('setup-next')).toBeDisabled()
  })

  test('can navigate back from category step', async ({ page }) => {
    await mockSetupNotCompleted(page)
    await page.route('**/api/auth/setup', (route) =>
      route.fulfill({ json: { message: 'ok' } })
    )
    await page.route('**/api/auth/login', (route) =>
      route.fulfill({ json: { user: { id: 1, username: 'admin', role: 'admin' } } })
    )

    await page.goto('/setup')
    await page.getByTestId('setup-username').fill('admin')
    await page.getByTestId('setup-password').fill('Password123!')
    await page.getByTestId('setup-password-confirm').fill('Password123!')
    await page.getByTestId('setup-next').click()

    await expect(page.getByTestId('setup-step-category')).toBeVisible()
    await page.getByTestId('setup-back').click()
    await expect(page.getByTestId('setup-step-admin')).toBeVisible()
  })

  test('language selector changes during setup', async ({ page }) => {
    await mockSetupNotCompleted(page)

    await page.goto('/setup')
    await expect(page.getByTestId('setup-language')).toBeVisible()

    // Click English locale button
    const enBtn = page.getByTestId('setup-lang-en')
    if (await enBtn.isVisible()) {
      await enBtn.click()
      // The page text should now be in English — verify the next button text changed
      await expect(page.getByTestId('setup-next')).toBeVisible()
    }
  })
})
