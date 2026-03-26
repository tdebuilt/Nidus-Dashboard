import { test, expect } from '@playwright/test'
import { TEST_USER, TEST_PASSWORD, openSidebar, waitForDashboard } from './helpers'

test.describe('Authentication', () => {
  test('login with invalid credentials shows error', async ({ page }) => {
    await page.context().clearCookies()

    await page.goto('/login')
    await expect(page.getByTestId('login-page')).toBeVisible()

    await page.getByTestId('login-username').fill(TEST_USER)
    await page.getByTestId('login-password').fill('wrong-password')
    await page.getByTestId('login-submit').click()

    await expect(page.getByTestId('login-error')).toBeVisible()
  })

  test('login with valid credentials redirects to dashboard', async ({ page }) => {
    await page.context().clearCookies()

    await page.goto('/login')
    await page.getByTestId('login-username').fill(TEST_USER)
    await page.getByTestId('login-password').fill(TEST_PASSWORD)
    await page.getByTestId('login-submit').click()

    await waitForDashboard(page)
  })

  test('logout redirects to login page', async ({ page }) => {
    await page.goto('/')
    await waitForDashboard(page)

    await openSidebar(page)
    await page.getByTestId('logout-button').click()

    await expect(page.getByTestId('login-page')).toBeVisible()
  })

  test('protected route redirects to login without auth', async ({ page }) => {
    await page.context().clearCookies()

    await page.goto('/settings')
    await expect(page.getByTestId('login-page')).toBeVisible()
  })
})
