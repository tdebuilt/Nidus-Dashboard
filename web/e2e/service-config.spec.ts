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

    // Add a new service
    await page.getByTestId('service-add-btn').click()
    await expect(page.getByTestId('service-add-panel')).toBeVisible()

    await page.getByTestId(`service-add-option-${serviceType}`).click()

    // Fill service URL
    await page.getByTestId('service-url-input').fill(serviceUrl)
    await page.getByTestId('service-save').click()

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

    // Delete the service
    await serviceRow.getByTestId('service-delete-btn').click()

    // Confirm deletion
    await expect(page.getByTestId('confirm-ok')).toBeVisible()
    await page.getByTestId('confirm-ok').click()

    // Service should be removed
    await expect(serviceRow).toBeHidden()
  })
})
