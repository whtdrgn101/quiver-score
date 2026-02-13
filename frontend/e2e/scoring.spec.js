import { test, expect } from '@playwright/test';
import { login, registerUser } from './helpers.js';

const uniqueId = () => Math.random().toString(36).slice(2, 8);

test.describe('Scoring Flow', () => {
  let username;
  let password;

  test.beforeAll(async ({ browser }) => {
    const id = uniqueId();
    username = `e2escore_${id}`;
    password = 'TestPass123!';
    const page = await browser.newPage();
    await registerUser(page, {
      username,
      email: `e2escore_${id}@test.com`,
      password,
    });
    await page.waitForURL('/');
    await page.close();
  });

  test('start a new session and submit an end', async ({ page }) => {
    await login(page, username, password);

    await page.goto('/rounds');
    await expect(page.getByText('Select a Round')).toBeVisible();
    await expect(page.getByTestId('round-list')).toBeVisible();

    // Click the first available round template
    const roundCards = page.getByTestId('round-list').locator('button').first();
    await roundCards.click();

    await page.getByRole('button', { name: 'Start Scoring' }).click();
    await page.waitForURL(/\/score\//);

    await expect(page.getByText(/End 1/)).toBeVisible();

    // Get available score buttons
    const scoreButtons = page.locator('.grid button');
    const buttonCount = await scoreButtons.count();
    expect(buttonCount).toBeGreaterThan(0);

    // Click the first score button enough times for one end
    const arrowSlots = page.locator('.border-dashed, .bg-yellow-400, .bg-yellow-300, .bg-red-500, .bg-blue-500, .bg-gray-400, .bg-gray-700');
    const arrowCount = await arrowSlots.count();

    for (let i = 0; i < arrowCount; i++) {
      await scoreButtons.first().click();
    }

    // Submit the end
    await page.getByTestId('submit-end-btn').click();

    // End should now appear in the scorecard
    await expect(page.getByText('Scorecard')).toBeVisible();
  });

  test('view session detail after completing', async ({ page }) => {
    await login(page, username, password);

    await page.goto('/history');
    await expect(page.getByText('Session History')).toBeVisible();
    const inProgress = page.getByText('in_progress').first();
    if (await inProgress.isVisible()) {
      await inProgress.click();
      await expect(page).toHaveURL(/\/score\//);
    }
  });
});
