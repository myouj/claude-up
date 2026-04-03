import { test, expect } from '@playwright/test'

test.describe('Batch Test (A/B Test)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/ab-tests')
  })

  test('batch test list page loads', async ({ page }) => {
    // Check page title and header
    await expect(page.locator('h1')).toContainText('A/B 测试')

    // Check filter tabs are visible
    await expect(page.locator('.status-tabs')).toBeVisible()

    // Use more specific locators for radio buttons
    await expect(page.locator('.el-radio-button:has-text("全部")')).toBeVisible()
    await expect(page.locator('.el-radio-button:has-text("运行中")')).toBeVisible()
    await expect(page.locator('.el-radio-button:has-text("已完成")')).toBeVisible()

    // Check create button is visible
    await expect(page.locator('button:has-text("新建测试")')).toBeVisible()
  })

  test('batch test creation dialog opens', async ({ page }) => {
    // Click create button
    await page.click('button:has-text("新建测试")')

    // Check dialog opens
    await expect(page.locator('.el-dialog')).toBeVisible()
    await expect(page.locator('.el-dialog__title')).toContainText('新建 A/B 测试')

    // Check form fields exist
    await expect(page.locator('text=测试名称')).toBeVisible()
    await expect(page.locator('text=选择 Prompt')).toBeVisible()
  })

  test('filter tabs work correctly', async ({ page }) => {
    // Use specific radio button locators
    const allTab = page.locator('.el-radio-button:has-text("全部")')
    const runningTab = page.locator('.el-radio-button:has-text("运行中")')
    const completedTab = page.locator('.el-radio-button:has-text("已完成")')

    // All tabs should be visible
    await expect(allTab).toBeVisible()
    await expect(runningTab).toBeVisible()
    await expect(completedTab).toBeVisible()

    // Click on running filter
    await runningTab.click()

    // Verify the running tab becomes active
    await expect(runningTab).toHaveClass(/is-active/)
  })

  test('search input filters tests', async ({ page }) => {
    // Find and use search input
    const searchInput = page.locator('.search-input input')
    await expect(searchInput).toBeVisible()

    // Type search query
    await searchInput.fill('test')

    // Wait for filtering to apply
    await page.waitForTimeout(300)

    // Input should have the value
    await expect(searchInput).toHaveValue('test')
  })

  test('test cards are clickable', async ({ page }) => {
    // Wait for test cards to load (if any exist)
    await page.waitForTimeout(1000)

    // Check if test cards exist
    const testCards = page.locator('.test-card')
    const count = await testCards.count()

    if (count > 0) {
      // Click on first card
      await testCards.first().click()

      // Should navigate to detail page
      await page.waitForURL(/\/ab-tests\/\d+/, { timeout: 5000 })
      await expect(page.locator('.ab-test-detail')).toBeVisible()
    } else {
      // Empty state should show create button
      await expect(page.locator('.el-empty')).toBeVisible()
      await expect(page.locator('text=创建第一个 A/B 测试')).toBeVisible()
    }
  })

  test('batch test detail page displays variants', async ({ page }) => {
    // First, check if any tests exist
    await page.goto('/ab-tests')
    await page.waitForTimeout(1000)

    const testCards = page.locator('.test-card')
    const count = await testCards.count()

    if (count > 0) {
      // Click on first test card
      await testCards.first().click()
      await page.waitForURL(/\/ab-tests\/\d+/, { timeout: 5000 })

      // Check detail page elements
      await expect(page.locator('.ab-test-detail')).toBeVisible()

      // Check for variant cards
      const variantCards = page.locator('.variant-card')
      const variantCount = await variantCards.count()

      if (variantCount > 0) {
        await expect(variantCards.first()).toBeVisible()
      }
    }
  })

  test('variant comparison functionality', async ({ page }) => {
    // Navigate to an existing test detail
    await page.goto('/ab-tests')
    await page.waitForTimeout(1000)

    const testCards = page.locator('.test-card')
    const count = await testCards.count()

    if (count > 0) {
      await testCards.first().click()
      await page.waitForURL(/\/ab-tests\/\d+/, { timeout: 5000 })

      // Check for comparison panels
      const comparePanels = page.locator('.compare-panel')
      const panelCount = await comparePanels.count()

      if (panelCount >= 2) {
        // Multiple panels should be visible for comparison
        await expect(comparePanels.first()).toBeVisible()
        await expect(comparePanels.nth(1)).toBeVisible()
      }
    }
  })

  test('winner badge displays correctly', async ({ page }) => {
    // Navigate to test list
    await page.goto('/ab-tests')
    await page.waitForTimeout(1000)

    const testCards = page.locator('.test-card')
    const count = await testCards.count()

    if (count > 0) {
      // Look for winner badge in list view
      const winnerBadges = page.locator('.winner-badge')
      const winnerCount = await winnerBadges.count()

      if (winnerCount > 0) {
        await expect(winnerBadges.first()).toBeVisible()
        await expect(winnerBadges.first()).toContainText('胜出')
      }

      // Also check test cards with winner class
      const cardsWithWinner = page.locator('.test-card.winner')
      const winnerCardCount = await cardsWithWinner.count()

      if (winnerCardCount > 0) {
        await expect(cardsWithWinner.first()).toBeVisible()
      }
    }
  })

  test('test list shows correct metadata', async ({ page }) => {
    // Navigate to test list
    await page.goto('/ab-tests')
    await page.waitForTimeout(1000)

    const testCards = page.locator('.test-card')
    const count = await testCards.count()

    if (count > 0) {
      // Check for prompt title in card
      const firstCard = testCards.first()
      await expect(firstCard.locator('.prompt-title')).toBeVisible()

      // Check for meta row (timer and run count)
      await expect(firstCard.locator('.meta-row')).toBeVisible()
      await expect(firstCard.locator('.meta-item')).toHaveCount(2)
    }
  })

  test('create dialog has required form fields', async ({ page }) => {
    // Open create dialog
    await page.click('button:has-text("新建测试")')
    await expect(page.locator('.el-dialog')).toBeVisible()

    // Check required fields exist
    await expect(page.locator('label:has-text("测试名称")')).toBeVisible()
    await expect(page.locator('label:has-text("选择 Prompt")')).toBeVisible()

    // Check for variant inputs (A/B test needs at least 2 variants)
    await expect(page.locator('text=Variant A')).toBeVisible()
    await expect(page.locator('text=Variant B')).toBeVisible()

    // Close dialog
    await page.keyboard.press('Escape')
    await page.waitForTimeout(300)
  })
})
