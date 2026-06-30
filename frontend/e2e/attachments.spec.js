import { test, expect } from '@playwright/test';
import { registerUser, login } from './helpers.js';

const uniqueId = () => Math.random().toString(36).slice(2, 8);

// Same valid 32x24 JPEG used by the Go contract tests, base64-encoded so the
// fixture stays self-contained without a binary asset. Wrapping kept identical
// to the Python source to avoid the kind of cut-and-paste corruption that
// silently produces a slightly-shorter b64 string and an undecodable image.
const SAMPLE_JPEG_B64 =
  '/9j/2wCEAAYEBQYFBAYGBQYHBwYIChAKCgkJChQODwwQFxQYGBcUFhYaHSUfGhsjHBYWICwgIyYnKSopGR' +
  '8tMC0oMCUoKSgBBwcHCggKEwoKEygaFhooKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgo' +
  'KCgoKCgoKCgoKCgoKP/AABEIABgAIAMBIgACEQEDEQH/xAGiAAABBQEBAQEBAQAAAAAAAAAAAQIDBAUGBw' +
  'gJCgsQAAIBAwMCBAMFBQQEAAABfQECAwAEEQUSITFBBhNRYQcicRQygZGhCCNCscEVUtHwJDNicoIJChYX' +
  'GBkaJSYnKCkqNDU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6g4SFhoeIiYqSk5SVlp' +
  'eYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2drh4uPk5ebn6Onq8fLz9PX29/j5+gEA' +
  'AwEBAQEBAQEBAQAAAAAAAAECAwQFBgcICQoLEQACAQIEBAMEBwUEBAABAncAAQIDEQQFITEGEkFRB2FxEy' +
  'IygQgUQpGhscEJIzNS8BVictEKFiQ04SXxFxgZGiYnKCkqNTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZn' +
  'aGlqc3R1dnd4eXqCg4SFhoeIiYqSk5SVlpeYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1t' +
  'fY2dri4+Tl5ufo6ery8/T19vf4+fr/2gAMAwEAAhEDEQA/APCrTSuny1tWmldPlrpbTSuny1s2mldPlr0a' +
  '+Y+ZyZVm22pzVppXT5a2rTSuny10tppXT5a2bTSuny149fMfM/ScqzbbUo2mldPlratNK6fLTrTtWzaV4d' +
  'fETP51yrFVNNRtppXT5a2rTSuny060rZtO1ePXxEz9JyrFVNNT/9k=';

test.describe('Attachments — equipment photos', () => {
  let username;
  let password;

  test.beforeAll(async ({ browser }) => {
    const id = uniqueId();
    username = `e2eatt_${id}`;
    password = 'TestPass123!';
    const page = await browser.newPage();
    await registerUser(page, {
      username,
      email: `e2eatt_${id}@test.com`,
      password,
    });
    await page.waitForURL('/');
    await page.close();
  });

  test('upload, view, and delete a photo on an equipment item', async ({ page }) => {
    await login(page, username, password);
    await page.goto('/equipment');

    // Create a piece of equipment to attach a photo to. Labels in this form
    // aren't htmlFor-associated, so we identify the field by placeholder.
    await page.getByTestId('add-equipment-btn').click();
    await page.getByPlaceholder('e.g. Hoyt Formula Xi').fill('Photo Test Riser');
    await page.getByRole('button', { name: 'Add', exact: true }).click();
    const list = page.getByTestId('equipment-list');
    await expect(list.getByText('Photo Test Riser')).toBeVisible();

    // Open the photos pane on the new equipment card.
    await page.getByRole('button', { name: 'Photos' }).first().click();
    await expect(page.getByRole('heading', { name: 'Photos' })).toBeVisible();

    // Upload a JPEG via the hidden file input.
    const fileInput = page.locator('input[type="file"][accept*="image"]');
    await fileInput.setInputFiles({
      name: 'target.jpg',
      mimeType: 'image/jpeg',
      buffer: Buffer.from(SAMPLE_JPEG_B64, 'base64'),
    });

    // Wait for the thumbnail to render — when the AttachmentImage finishes
    // loading, it swaps from the placeholder div to an <img>.
    await expect(page.getByRole('button', { name: 'View photo' })).toBeVisible({ timeout: 10000 });

    // Open the modal and delete.
    page.on('dialog', (d) => d.accept());
    await page.getByRole('button', { name: 'View photo' }).click();
    await expect(page.getByRole('button', { name: 'Delete photo' })).toBeVisible();
    await page.getByRole('button', { name: 'Delete photo' }).click();

    // Thumbnail is gone after delete.
    await expect(page.getByRole('button', { name: 'View photo' })).toHaveCount(0);
  });
});
