import { expect, test } from '@playwright/test';

test('the camera serves the app, plans a shot and saves settings', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveURL(/\/static\/frontend\/build\/$/);
  await expect(page.getByRole('heading', { name: 'Photos' })).toBeVisible();
  await expect(page.locator('.next-pill')).toContainText(/Next shot\s*(now|\d+:\d\d)/);

  await page.getByRole('link', { name: 'Settings' }).click();
  const quality = page.getByRole('spinbutton', { name: 'JPEG quality' });
  await quality.fill('77');
  await page.getByRole('button', { name: 'Save changes' }).click();
  await expect(page.getByText('Settings saved. They apply from the next capture.')).toBeVisible();

  await page.reload();
  await expect(page.getByRole('spinbutton', { name: 'JPEG quality' })).toHaveValue('77');
});
