import { test, expect } from '@playwright/test'

// Mock API responses for unauthenticated state
async function mockAPI(page: import('@playwright/test').Page) {
  await page.route('**/api/auth/status', (route) =>
    route.fulfill({ json: { setup_completed: true } })
  )
  await page.route('**/api/auth/me', (route) =>
    route.fulfill({ json: { id: 1, username: 'admin', role: 'admin' } })
  )
  await page.route('**/api/settings', (route) =>
    route.fulfill({
      json: {
        theme: 'dark',
        language: 'en',
        refresh_interval: 30,
        accent_color: '',
        custom_css: '',
        enable_keyboard_shortcuts: true,
      },
    })
  )
  await page.route('**/api/categories', (route) =>
    route.fulfill({
      json: [
        { id: 1, name: 'Home', icon: 'home', sort_order: 0 },
        { id: 2, name: 'Media', icon: 'tv', sort_order: 1 },
      ],
    })
  )
  await page.route('**/api/categories/*/widgets', (route) =>
    route.fulfill({
      json: [
        { id: 1, category_id: 1, type: 'applink', title: 'Links', config: '{}', collapsed: false, pos_x: 0, pos_y: 0, width: 4, height: 0, created_at: '', updated_at: '' },
        { id: 2, category_id: 1, type: 'applink', title: 'Tools', config: '{}', collapsed: false, pos_x: 4, pos_y: 0, width: 4, height: 0, created_at: '', updated_at: '' },
        { id: 3, category_id: 1, type: 'applink', title: 'Monitoring', config: '{}', collapsed: false, pos_x: 8, pos_y: 0, width: 4, height: 0, created_at: '', updated_at: '' },
      ],
    })
  )
  await page.route('**/api/version', (route) =>
    route.fulfill({ json: { version: 'test' } })
  )
  // Catch-all for other API calls
  await page.route('**/api/**', (route) => {
    if (!route.request().url().includes('/api/auth/') && !route.request().url().includes('/api/settings') && !route.request().url().includes('/api/categories') && !route.request().url().includes('/api/version')) {
      route.fulfill({ json: [] })
    }
  })
}

test.describe('Responsive grid layout', () => {
  test.beforeEach(async ({ page }) => {
    await mockAPI(page)
  })

  test('mobile shows single column layout', async ({ page, browserName }, testInfo) => {
    test.skip(testInfo.project.name !== 'mobile', 'Mobile only')
    await page.goto('/')
    await page.waitForSelector('[data-testid="widget-grid"]')
    const grid = page.locator('.widget-grid')
    const style = await grid.evaluate((el) => window.getComputedStyle(el).gridTemplateColumns)
    // Single column: should be one value (the full width)
    const columns = style.split(' ').filter((s) => s.trim())
    expect(columns.length).toBe(1)
  })

  test('tablet shows 6-column grid', async ({ page }, testInfo) => {
    test.skip(testInfo.project.name !== 'tablet', 'Tablet only')
    await page.goto('/')
    await page.waitForSelector('[data-testid="widget-grid"]')
    const grid = page.locator('.widget-grid')
    const style = await grid.evaluate((el) => window.getComputedStyle(el).gridTemplateColumns)
    const columns = style.split(' ').filter((s) => s.trim())
    expect(columns.length).toBe(6)
  })

  test('desktop shows 12-column grid', async ({ page }, testInfo) => {
    test.skip(testInfo.project.name !== 'desktop', 'Desktop only')
    await page.goto('/')
    await page.waitForSelector('[data-testid="widget-grid"]')
    const grid = page.locator('.widget-grid')
    const style = await grid.evaluate((el) => window.getComputedStyle(el).gridTemplateColumns)
    const columns = style.split(' ').filter((s) => s.trim())
    expect(columns.length).toBe(12)
  })

  test('widget cards render without horizontal overflow on mobile', async ({ page }, testInfo) => {
    test.skip(testInfo.project.name !== 'mobile', 'Mobile only')
    await page.goto('/')
    await page.waitForSelector('[data-testid="widget-card"]')
    const cards = page.locator('[data-testid="widget-card"]')
    const count = await cards.count()
    for (let i = 0; i < count; i++) {
      const card = cards.nth(i)
      const box = await card.boundingBox()
      if (box) {
        const viewport = page.viewportSize()!
        expect(box.x).toBeGreaterThanOrEqual(0)
        expect(box.x + box.width).toBeLessThanOrEqual(viewport.width + 2) // 2px tolerance
      }
    }
  })

  test('TV viewport has larger base font', async ({ page }, testInfo) => {
    test.skip(testInfo.project.name !== 'tv', 'TV only')
    await page.goto('/')
    const fontSize = await page.evaluate(() =>
      window.getComputedStyle(document.documentElement).fontSize
    )
    const size = parseInt(fontSize)
    expect(size).toBeGreaterThanOrEqual(20)
  })
})

test.describe('Sidebar responsive', () => {
  test.beforeEach(async ({ page }) => {
    await mockAPI(page)
  })

  test('sidebar is hidden on mobile', async ({ page }, testInfo) => {
    test.skip(testInfo.project.name !== 'mobile', 'Mobile only')
    await page.goto('/')
    await page.waitForSelector('[data-testid="main-content"]')
    const sidebar = page.locator('[data-testid="sidebar"]')
    // Sidebar should exist but be translated offscreen
    await expect(sidebar).toBeAttached()
    const transform = await sidebar.evaluate((el) => window.getComputedStyle(el).transform)
    expect(transform).not.toBe('none')
  })
})
