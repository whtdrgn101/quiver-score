import { test, expect } from '@playwright/test';
import { registerUser } from './helpers.js';

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
    await page.goto('/login');
    await page.getByPlaceholder('Username').fill(username);
    await page.getByPlaceholder('Password').fill(password);
    await page.getByRole('button', { name: 'Sign In' }).click();
    await page.waitForURL('/');
    await expect(page.getByText(/welcome/i)).toBeVisible();
  });

  test('dark mode toggle works', async ({ page }) => {
    await page.goto('/login');
    await page.getByPlaceholder('Username').fill(username);
    await page.getByPlaceholder('Password').fill(password);
    await page.getByRole('button', { name: 'Sign In' }).click();
    await page.waitForURL('/');

    // Find and click the dark mode toggle
    const toggle = page.locator('button[title="Dark mode"], button[title="Light mode"]');
    await expect(toggle).toBeVisible();

    const initialTitle = await toggle.getAttribute('title');
    await toggle.click();
    const newTitle = await toggle.getAttribute('title');
    expect(newTitle).not.toBe(initialTitle);
  });

  test('all nav links are accessible', async ({ page }) => {
    await page.goto('/login');
    await page.getByPlaceholder('Username').fill(username);
    await page.getByPlaceholder('Password').fill(password);
    await page.getByRole('button', { name: 'Sign In' }).click();
    await page.waitForURL('/');

    // Check navigation links exist
    const navLinks = [
      { name: 'Dashboard', path: '/' },
      { name: 'Equipment', path: '/equipment' },
      { name: 'Setups', path: '/setups' },
      { name: 'History', path: '/history' },
    ];

    for (const link of navLinks) {
      const navLink = page.locator(`nav a:has-text("${link.name}")`).first();
      await expect(navLink).toBeVisible();
    }
  });
});
