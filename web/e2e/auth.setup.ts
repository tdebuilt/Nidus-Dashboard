import { test as setup, expect } from '@playwright/test'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { TEST_USER, TEST_PASSWORD } from './helpers'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const AUTH_FILE = path.join(__dirname, '.auth', 'user.json')

setup('complete setup wizard and save auth state', async ({ page }) => {
  // Ensure .auth directory exists
  fs.mkdirSync(path.dirname(AUTH_FILE), { recursive: true })

  // Navigate to app — should redirect to /setup since DB is fresh
  await page.goto('/')
  await page.waitForURL('**/setup')
  await expect(page.getByTestId('setup-page')).toBeVisible()

  // Step 0: Admin account
  await expect(page.getByTestId('setup-step-admin')).toBeVisible()
  await page.getByTestId('setup-username').fill(TEST_USER)
  await page.getByTestId('setup-password').fill(TEST_PASSWORD)
  await page.getByTestId('setup-password-confirm').fill(TEST_PASSWORD)
  await page.getByTestId('setup-next').click()

  // Step 1: First category
  await expect(page.getByTestId('setup-step-category')).toBeVisible()
  await page.getByTestId('setup-category-name').fill('Test')
  await page.getByTestId('setup-next').click()

  // Should redirect to dashboard
  await expect(page.getByTestId('dashboard-page')).toBeVisible({ timeout: 10000 })

  // Save authenticated state
  await page.context().storageState({ path: AUTH_FILE })
})
