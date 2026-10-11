import { test, expect } from '@playwright/test';

test.describe('Auth', () => {
  test('register → login → trips list', async ({ page }) => {
    const email = `test${Date.now()}@example.com`;
    await page.goto('/register');
    await page.fill('input[name="name"]', 'Test User');
    await page.fill('input[name="email"]', email);
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/login');
    await page.fill('input[name="email"]', email);
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/');
    await expect(page.locator('h1, h2, h3')).toContainText(/MyTrip|Поездки/i);
  });
});
