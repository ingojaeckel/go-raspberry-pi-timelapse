import { test, expect } from './fixtures';

test.describe('App Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display all navigation tabs', async ({ page }) => {
    await expect(page.getByRole('tab', { name: /home/i })).toBeVisible();
    await expect(page.getByRole('tab', { name: /preview/i })).toBeVisible();
    await expect(page.getByRole('tab', { name: /monitoring/i })).toBeVisible();
    await expect(page.getByRole('tab', { name: /settings/i })).toBeVisible();
    await expect(page.getByRole('tab', { name: /logs/i })).toBeVisible();
  });

  test('should start on home tab by default', async ({ page }) => {
    const homeTab = page.getByRole('tab', { name: /home/i });
    await expect(homeTab).toHaveAttribute('aria-selected', 'true');
  });

  test('should switch to preview tab when clicked', async ({ page }) => {
    const previewTab = page.getByRole('tab', { name: /preview/i });
    await previewTab.click();
    await expect(previewTab).toHaveAttribute('aria-selected', 'true');
  });

  test('should switch to monitoring tab when clicked', async ({ page }) => {
    const monitoringTab = page.getByRole('tab', { name: /monitoring/i });
    await monitoringTab.click();
    await expect(monitoringTab).toHaveAttribute('aria-selected', 'true');
  });

  test('should switch to settings tab when clicked', async ({ page }) => {
    const settingsTab = page.getByRole('tab', { name: /settings/i });
    await settingsTab.click();
    await expect(settingsTab).toHaveAttribute('aria-selected', 'true');
  });

  test('should switch to logs tab when clicked', async ({ page }) => {
    const logsTab = page.getByRole('tab', { name: /logs/i });
    await logsTab.click();
    await expect(logsTab).toHaveAttribute('aria-selected', 'true');
  });

  test('should display correct tab panel content when switching tabs', async ({ page }) => {
    // Check home tab content is visible by default
    await expect(page.getByRole('tabpanel', { name: /home/i }).or(page.locator('#simple-tabpanel-0'))).toBeVisible();

    // Switch to preview and verify
    await page.getByRole('tab', { name: /preview/i }).click();
    await expect(page.getByRole('tabpanel').filter({ has: page.getByAltText('preview') })).toBeVisible();

    // Switch to monitoring and verify
    await page.getByRole('tab', { name: /monitoring/i }).click();
    await expect(page.getByText(/CPU:/i)).toBeVisible();
    
    // Switch to settings and verify
    await page.getByRole('tab', { name: /settings/i }).click();
    await expect(page.getByText(/Time between captures/i)).toBeVisible();
  });
});
