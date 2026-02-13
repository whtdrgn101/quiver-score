import { test, expect } from '@playwright/test';
import { login, registerUser } from './helpers.js';

const uniqueId = () => Math.random().toString(36).slice(2, 8);

test.describe('Navigation', () => {
  let username;
  let password;

  test.beforeAll(async ({ browser }) => {
    const id = uniqueId();
    username = `e2enav_${id}`;
    password = 'TestPass123!';
    const page = await browser.newPage();
    await registerUser(page, {
      username,
      email: `e2enav_${id}@test.com`,
      password,
    });
    await page.waitForURL('/');
    await page.close();
  });

  test('dashboard loads for logged-in user', async ({ page }) => {
    await login(page, username, password);
    await expect(page.getByTestId('dashboard-heading')).toBeVisible();
  });

  test('dark mode toggle works', async ({ page }) => {
    await login(page, username, password);

    const toggle = page.locator('button[title="Dark mode"], button[title="Light mode"]');
    await expect(toggle).toBeVisible();

    const initialTitle = await toggle.getAttribute('title');
    await toggle.click();
    const newTitle = await toggle.getAttribute('title');
    expect(newTitle).not.toBe(initialTitle);
  });

  test('all nav links are accessible', async ({ page }) => {
    await login(page, username, password);

    const navLinks = ['Dashboard', 'Equipment', 'Clubs', 'History'];

    for (const name of navLinks) {
      const navLink = page.getByRole('link', { name }).first();
      await expect(navLink).toBeVisible();
    }
  });
});
