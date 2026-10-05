import { test, expect } from '@playwright/test';

test.describe('Trips', () => {
  test.beforeEach(async ({ page }) => {
    // логин через API (быстрее)
    await page.request.post('/api/v1/auth/login', {
      data: { email: 'e2e@example.com', password: 'password123' },
    });
    await page.goto('/');
  });

  test('create trip → detail → add stop → export', async ({ page }) => {
    await page.click('text=+ Новая поездка');
    await page.waitForURL(/\/trips\/.+\//);
    await page.fill('input[name="title"]', 'E2E Trip');
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/trips\/.+(?<!\/edit)/);
    await page.click('text=+ Стоп');
    await page.fill('input[name="title"]', 'E2E Stop');
    await page.fill('input[name="lat"]', '55.75');
    await page.fill('input[name="lng"]', '37.6');
    await page.fill('input[name="visit_date"]', '2025-01-15');
    await page.fill('input[name="order"]', '0');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=E2E Stop')).toBeVisible();
    // экспорт
    const [download] = await Promise.all([
      page.waitForEvent('download'),
      page.click('text=Экспорт CSV'),
    ]);
    expect(download.suggestedFilename()).toContain('.csv');
  });
});
