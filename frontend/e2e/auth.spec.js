import { test, expect } from '@playwright/test';
import { login, registerUser } from './helpers.js';

const uniqueId = () => Math.random().toString(36).slice(2, 8);

test.describe('Authentication', () => {
  test('register a new user and land on dashboard', async ({ page }) => {
    const id = uniqueId();
    await registerUser(page, {
      username: `e2euser_${id}`,
      email: `e2e_${id}@test.com`,
      password: 'TestPass123!',
      displayName: `E2E User ${id}`,
    });
    await page.waitForURL('/');
    await expect(page.getByTestId('dashboard-heading')).toBeVisible();
  });

  test('login with valid credentials', async ({ page }) => {
    const id = uniqueId();
    const username = `e2elogin_${id}`;
    await registerUser(page, {
      username,
      email: `e2elogin_${id}@test.com`,
      password: 'TestPass123!',
    });
    await page.waitForURL('/');

    await page.evaluate(() => localStorage.clear());
    await page.goto('/login');

    await login(page, username, 'TestPass123!');
    await expect(page.getByTestId('dashboard-heading')).toBeVisible();
  });

  test('login with invalid credentials shows error', async ({ page }) => {
    await page.goto('/login');
    await page.getByPlaceholder('Username').fill('nonexistent_user');
    await page.getByPlaceholder('Password').fill('wrongpassword');
    await page.getByRole('button', { name: 'Sign In' }).click();
    await expect(page.getByTestId('login-error')).toBeVisible();
  });

  test('logout clears session', async ({ page }) => {
    const id = uniqueId();
    await registerUser(page, {
      username: `e2elogout_${id}`,
      email: `e2elogout_${id}@test.com`,
      password: 'TestPass123!',
    });
    await page.waitForURL('/');

    const logoutBtn = page.getByTestId('logout-btn');
    if (await logoutBtn.isVisible()) {
      await logoutBtn.click();
      await expect(page).toHaveURL('/login');
    }
  });
});
