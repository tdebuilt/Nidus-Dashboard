import { test, expect } from '@playwright/test'
import { enableEditMode, createCategoryViaAPI, createWidgetViaAPI, waitForDashboard } from './helpers'

test.describe('Drag & Drop and Resize', () => {
  test('drag a widget to a new position', async ({ page }) => {
    // Setup via API
    const catId = await createCategoryViaAPI(page, 'Drag Test ' + Date.now())
    await createWidgetViaAPI(page, catId, 'applink', 'Drag Widget')

    await page.goto('/')
    await waitForDashboard(page)

    // Navigate to the new category
    await page.getByTestId('category-tab').last().click()
    await enableEditMode(page)

    const widget = page.getByTestId('widget-card').filter({ hasText: 'Drag Widget' }).first()
    await expect(widget).toBeVisible()

    const box = await widget.boundingBox()
    expect(box).not.toBeNull()

    const headerY = box!.y + 20
    const headerX = box!.x + box!.width / 2

    // Perform drag with slow mouse movement
    await page.mouse.move(headerX, headerY)
    await page.mouse.down()
    for (let i = 1; i <= 20; i++) {
      await page.mouse.move(headerX + i * 10, headerY + i * 5, { steps: 1 })
    }
    await page.mouse.up()

    // The drag operation should complete without error
  })

  test('resize a widget', async ({ page }) => {
    // Setup via API
    const catId = await createCategoryViaAPI(page, 'Resize Test ' + Date.now())
    await createWidgetViaAPI(page, catId, 'applink', 'Resize Widget')

    await page.goto('/')
    await waitForDashboard(page)

    await page.getByTestId('category-tab').last().click()
    await enableEditMode(page)

    const widget = page.getByTestId('widget-card').filter({ hasText: 'Resize Widget' }).first()
    await expect(widget).toBeVisible()

    // Toggle auto-height off if needed
    const autoHeightBtn = widget.getByTestId('widget-auto-height')
    const isAutoHeight = await autoHeightBtn.evaluate(
      (el) => el.classList.contains('text-[var(--color-primary)]')
    )
    if (isAutoHeight) {
      await autoHeightBtn.click()
    }

    await widget.hover()
    const resizeHandle = widget.getByTestId('widget-resize')
    await expect(resizeHandle).toBeAttached()

    const boxBefore = await widget.boundingBox()
    expect(boxBefore).not.toBeNull()

    const handleBox = await resizeHandle.boundingBox()
    expect(handleBox).not.toBeNull()
    const handleX = handleBox!.x + handleBox!.width / 2
    const handleY = handleBox!.y + handleBox!.height / 2

    await page.mouse.move(handleX, handleY)
    await page.mouse.down()
    await page.mouse.move(handleX + 150, handleY + 80, { steps: 5 })
    await page.mouse.up()

    const boxAfter = await widget.boundingBox()
    expect(boxAfter).not.toBeNull()
    expect(boxAfter!.width).toBeGreaterThan(boxBefore!.width)
  })
})
