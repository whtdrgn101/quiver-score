import { test, expect } from '@playwright/test';
import { login, registerUser } from './helpers.js';

const uniqueId = () => Math.random().toString(36).slice(2, 8);

test.describe('Equipment', () => {
  let username;
  let password;

  test.beforeAll(async ({ browser }) => {
    const id = uniqueId();
    username = `e2eeq_${id}`;
    password = 'TestPass123!';
    const page = await browser.newPage();
    await registerUser(page, {
      username,
      email: `e2eeq_${id}@test.com`,
      password,
    });
    await page.waitForURL('/');
    await page.close();
  });

  test('create equipment item and verify it appears', async ({ page }) => {
    await login(page, username, password);
    await page.goto('/equipment');

    await page.getByTestId('add-equipment-btn').click();

    // Fill form
    await page.getByLabel('Name').fill('My Recurve Riser');
    await page.getByLabel('Brand').fill('Hoyt');

    await page.getByRole('button', { name: 'Save' }).click();

    // Verify item appears
    await expect(page.getByText('My Recurve Riser')).toBeVisible();
    await expect(page.getByText('Hoyt')).toBeVisible();
  });

  test('delete equipment item and verify removal', async ({ page }) => {
    await login(page, username, password);
    await page.goto('/equipment');

    // Wait for equipment to load
    await expect(page.getByText('My Recurve Riser')).toBeVisible();

    // Click delete
    page.on('dialog', (dialog) => dialog.accept());
    const deleteBtn = page.getByRole('button', { name: /delete/i }).first();
    await deleteBtn.click();

    // Verify removal
    await expect(page.getByText('My Recurve Riser')).not.toBeVisible();
  });
});
