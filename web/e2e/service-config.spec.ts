import { test, expect } from '@playwright/test'
import { waitForSettings } from './helpers'

test.describe('Service configuration', () => {
  test.describe.configure({ mode: 'serial' })

  const serviceType = 'uptimekuma'
  const serviceUrl = 'http://fake-uptime:3001'

  test('add a service', async ({ page }) => {
    await page.goto('/settings')
    await waitForSettings(page)

    // Navigate to Services tab
    await page.getByTestId('settings-pill-services').click()
    await expect(page.getByTestId('settings-services')).toBeVisible()

    // Add a new service via dialog (empty-btn in empty state, add-btn otherwise)
    const emptyBtn = page.getByTestId('service-add-empty-btn')
    const addBtn = page.getByTestId('service-add-btn')
    if (await emptyBtn.isVisible().catch(() => false)) {
      await emptyBtn.click()
    } else {
      await addBtn.click()
    }
    await expect(page.getByTestId('add-service-dialog')).toBeVisible()

    await page.getByTestId(`service-type-${serviceType}`).click()

    // Fill service URL
    await page.getByTestId('service-url-input').fill(serviceUrl)
    await page.getByTestId('service-save-btn').click()

    // Verify service row appears
    await expect(page.getByTestId(`service-row-${serviceType}`)).toBeVisible()
  })

  test('delete a service', async ({ page }) => {
    await page.goto('/settings')
    await waitForSettings(page)

    await page.getByTestId('settings-pill-services').click()
    await expect(page.getByTestId('settings-services')).toBeVisible()

    const serviceRow = page.getByTestId(`service-row-${serviceType}`)
    await expect(serviceRow).toBeVisible()

    // Open overflow menu then delete
    await serviceRow.getByTestId('service-overflow-btn').click()
    await serviceRow.getByTestId('service-delete-btn').click()

    // Confirm deletion
    await expect(page.getByTestId('confirm-ok')).toBeVisible()
    await page.getByTestId('confirm-ok').click()

    // Service should be removed
    await expect(serviceRow).toBeHidden()
  })
})
