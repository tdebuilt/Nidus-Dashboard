import { test, expect } from '@playwright/test'
import { enableEditMode, createWidget, createCategory, createCategoryViaAPI, createWidgetViaAPI, waitForDashboard } from './helpers'

test.describe('Categories and Widgets CRUD', () => {
  test('create a new category via UI', async ({ page }) => {
    await page.goto('/')
    await waitForDashboard(page)

    const name = 'E2E Cat ' + Date.now()
    await createCategory(page, name)

    await expect(
      page.getByTestId('category-tab').filter({ hasText: name })
    ).toBeVisible()
  })

  test('add and edit a widget', async ({ page }) => {
    // Setup via API
    const catId = await createCategoryViaAPI(page, 'Widget Test ' + Date.now())

    await page.goto(`/dashboard/widget-test-${catId}`)
    await waitForDashboard(page)

    // Select the category tab (navigate may not have selected it)
    const tab = page.getByTestId('category-tab').nth(1)
    if (await tab.isVisible()) await tab.click()

    // Add widget via UI
    const title = 'Test Widget'
    await createWidget(page, 'applink', title)
    await expect(
      page.getByTestId('widget-card').filter({ hasText: title })
    ).toBeVisible()

    // Edit widget title
    const widget = page.getByTestId('widget-card').filter({ hasText: title })
    await widget.getByTestId('widget-edit').click()
    await expect(page.getByTestId('edit-widget-dialog')).toBeVisible()

    const newTitle = 'Renamed Widget'
    await page.getByTestId('edit-widget-title-input').fill(newTitle)
    await page.getByTestId('edit-widget-save').click()

    await expect(page.getByTestId('edit-widget-dialog')).toBeHidden()
    await expect(
      page.getByTestId('widget-card').filter({ hasText: newTitle })
    ).toBeVisible()
  })

  test('delete a widget', async ({ page }) => {
    // Setup via API
    const catId = await createCategoryViaAPI(page, 'Delete Widget ' + Date.now())
    await createWidgetViaAPI(page, catId, 'applink', 'To Delete')

    await page.goto('/')
    await waitForDashboard(page)

    // Navigate to the new category
    const tabs = page.getByTestId('category-tab')
    await tabs.last().click()

    await enableEditMode(page)

    const widget = page.getByTestId('widget-card').filter({ hasText: 'To Delete' })
    await expect(widget).toBeVisible()
    await widget.getByTestId('widget-delete').click()

    await expect(page.getByTestId('confirm-ok')).toBeVisible()
    await page.getByTestId('confirm-ok').click()

    await expect(widget).toBeHidden()
  })

  test('delete a category', async ({ page }) => {
    // Setup via API
    const catId = await createCategoryViaAPI(page, 'To Delete ' + Date.now())

    // Delete via API and verify UI updates
    await page.goto('/')
    await waitForDashboard(page)

    // Wait for tabs to be rendered
    await expect(page.getByTestId('category-tab').first()).toBeVisible()
    const tabsBefore = await page.getByTestId('category-tab').count()
    expect(tabsBefore).toBeGreaterThan(0)

    // Delete category via API
    await page.request.delete(`/api/categories/${catId}`)

    // Reload and verify the category is gone
    await page.reload()
    await waitForDashboard(page)

    const tabsAfter = await page.getByTestId('category-tab').count()
    expect(tabsAfter).toBeLessThan(tabsBefore)
  })
})
