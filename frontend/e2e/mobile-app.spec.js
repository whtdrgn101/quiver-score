import { test, expect } from '@playwright/test';

test.describe('Mobile app page', () => {
  test('is reachable unauthenticated and shows private beta info', async ({ page }) => {
    await page.goto('/mobile');
    await expect(page.getByRole('heading', { name: 'QuiverScore Mobile' })).toBeVisible();
    await expect(page.getByText(/private beta/i)).toBeVisible();
    await expect(page.getByRole('link', { name: 'info@quiverscore.com' }).first()).toBeVisible();
  });

  test('is linked from the landing footer', async ({ page }) => {
    await page.goto('/');
    const link = page.getByRole('link', { name: 'Mobile App' });
    await expect(link).toBeVisible();
    await link.click();
    await page.waitForURL('**/mobile');
    await expect(page.getByRole('heading', { name: 'QuiverScore Mobile' })).toBeVisible();
  });

  test('hides store buttons when build env vars are unset', async ({ page }) => {
    await page.goto('/mobile');
    // With no VITE_ANDROID_APP_URL / VITE_IOS_APP_URL set, the download buttons should not render.
    await expect(page.getByRole('link', { name: /Download for Android/i })).toHaveCount(0);
    await expect(page.getByRole('link', { name: /Download for iOS/i })).toHaveCount(0);
  });
});
