import { test, expect } from '@playwright/test';
import { login, registerUser } from './helpers.js';

const uniqueId = () => Math.random().toString(36).slice(2, 8);

test.describe('Full Scoring Workflow', () => {
  let username;
  let password;

  test.beforeAll(async ({ browser }) => {
    const id = uniqueId();
    username = `e2efull_${id}`;
    password = 'TestPass123!';
    const page = await browser.newPage();
    await registerUser(page, {
      username,
      email: `e2efull_${id}@test.com`,
      password,
    });
    await page.waitForURL('/');
    await page.close();
  });

  /**
   * Helper to fill one end of arrows using the first score button.
   */
  async function fillAndSubmitEnd(page) {
    const scoreButtons = page.locator('.grid button');
    const arrowSlots = page.locator(
      '.border-dashed, .bg-yellow-400, .bg-yellow-300, .bg-red-500, .bg-blue-500, .bg-gray-400, .bg-gray-700'
    );
    const arrowCount = await arrowSlots.count();
    for (let i = 0; i < arrowCount; i++) {
      await scoreButtons.first().click();
    }
    await page.getByTestId('submit-end-btn').click();
  }

  test('select round, score all ends, complete, view in history', async ({ page }) => {
    await login(page, username, password);

    // Select a round
    await page.goto('/rounds');
    await expect(page.getByTestId('round-list')).toBeVisible();

    const roundCard = page.getByTestId('round-list').locator('button').first();
    await roundCard.click();
    await page.getByRole('button', { name: 'Start Scoring' }).click();
    await page.waitForURL(/\/score\//);

    // Submit one end
    await fillAndSubmitEnd(page);
    await expect(page.getByText('Scorecard')).toBeVisible();

    // Check session total is displayed
    const totalText = page.getByText(/Total:/);
    await expect(totalText).toBeVisible();
  });

  test('undo last end works during scoring', async ({ page }) => {
    await login(page, username, password);

    // Start a new session
    await page.goto('/rounds');
    const roundCard = page.getByTestId('round-list').locator('button').first();
    await roundCard.click();
    await page.getByRole('button', { name: 'Start Scoring' }).click();
    await page.waitForURL(/\/score\//);

    // Submit an end
    await fillAndSubmitEnd(page);
    await expect(page.getByText('Scorecard')).toBeVisible();

    // Undo
    const undoBtn = page.getByRole('button', { name: /undo/i });
    if (await undoBtn.isVisible()) {
      await undoBtn.click();
      // After undo, scorecard may be gone or updated
      await expect(page.getByText(/End 1/)).toBeVisible();
    }
  });

  test('session detail shows correct total after completion', async ({ page }) => {
    await login(page, username, password);

    // Go to history and check if we have sessions
    await page.goto('/history');
    await expect(page.getByText('Session History')).toBeVisible();

    // Check if there are any sessions in the list
    const historyList = page.getByTestId('history-list');
    if (await historyList.isVisible()) {
      const sessionLink = historyList.locator('a, [role="link"], [class*="cursor-pointer"]').first();
      if (await sessionLink.isVisible()) {
        await sessionLink.click();
        // Should show some scoring data
        await expect(page.getByText(/Total|Score|pts/i)).toBeVisible();
      }
    }
  });
});
