import { expect, type Page } from '@playwright/test'

export const TEST_USER = 'testadmin'
export const TEST_PASSWORD = 'TestPassword123!'

export async function waitForDashboard(page: Page) {
  await expect(page.getByTestId('dashboard-page')).toBeVisible()
}

export async function waitForSettings(page: Page) {
  await expect(page.getByTestId('settings-page')).toBeVisible()
}

export async function enableEditMode(page: Page) {
  const toggle = page.getByTestId('edit-mode-toggle')
  await expect(toggle).toBeVisible()
  const isPressed = await toggle.getAttribute('aria-pressed')
  if (isPressed !== 'true') {
    await toggle.click()
  }
}

export async function disableEditMode(page: Page) {
  const toggle = page.getByTestId('edit-mode-toggle')
  await expect(toggle).toBeVisible()
  const isPressed = await toggle.getAttribute('aria-pressed')
  if (isPressed === 'true') {
    await toggle.click()
  }
}

export async function openSidebar(page: Page) {
  await page.getByTestId('burger-button').click()
  await page.getByTestId('sidebar').waitFor({ state: 'visible' })
}

export async function createCategory(page: Page, name: string) {
  await enableEditMode(page)
  await page.getByTestId('category-add-btn').click()
  await page.getByTestId('category-create-input').fill(name)
  await page.getByTestId('category-create-input').press('Enter')
  await page.getByTestId('category-tab').filter({ hasText: name }).waitFor({ state: 'visible' })
}

export async function createWidget(page: Page, type: string, title: string) {
  await enableEditMode(page)

  // Click add button — use the visible one (multiple grids may exist for different categories)
  const emptyBtn = page.getByTestId('widget-add-empty').first()
  const headerBtn = page.getByTestId('widget-add-header-btn')
  if (await emptyBtn.isVisible().catch(() => false)) {
    await emptyBtn.click()
  } else {
    await headerBtn.click()
  }

  await page.getByTestId('add-widget-dialog').waitFor({ state: 'visible' })
  await page.getByTestId(`widget-type-${type}`).click()
  await page.getByTestId('widget-title-input').fill(title)
  await page.getByTestId('widget-create').click()
  await page.getByTestId('add-widget-dialog').waitFor({ state: 'hidden' })
}

// Create a category via API (faster than UI for test setup)
export async function createCategoryViaAPI(page: Page, name: string): Promise<number> {
  const response = await page.request.post('/api/categories', { data: { name, icon: 'folder' } })
  const body = await response.json()
  return body.id
}

// Create a widget via API (faster than UI for test setup)
export async function createWidgetViaAPI(page: Page, categoryId: number, type: string, title: string): Promise<number> {
  const response = await page.request.post(`/api/categories/${categoryId}/widgets`, {
    data: { type, title, config: '{}', width: 4, height: 0, pos_x: 0, pos_y: 0 },
  })
  const body = await response.json()
  return body.id
}

export async function deleteWidget(page: Page, widgetTitle: string) {
  await enableEditMode(page)
  const card = page.getByTestId('widget-card').filter({ hasText: widgetTitle })
  await card.getByTestId('widget-delete').click()
  await page.getByTestId('confirm-ok').click()
  await card.waitFor({ state: 'detached' })
}
