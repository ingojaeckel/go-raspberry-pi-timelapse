import { test, expect } from './fixtures';

test.describe('LogComponent', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    // Switch to logs tab
    await page.getByRole('tab', { name: /logs/i }).click();
  });

  test('should display logs section', async ({ page }) => {
    // Wait for the logs to load with timeout
    await page.waitForResponse('**/logs', { timeout: 10000 });
    
    // Check that the logs tab panel is visible
    const logsPanel = page.locator('[role="tabpanel"]').filter({ hasText: /sample log line/i });
    await expect(logsPanel).toBeVisible();
  });

  test('should fetch logs from API', async ({ page }) => {
    const requestPromise = page.waitForRequest('**/logs', { timeout: 10000 });
    
    // Reload the component by switching tabs
    await page.getByRole('tab', { name: /home/i }).click();
    await page.getByRole('tab', { name: /logs/i }).click();
    
    const request = await requestPromise;
    expect(request.url()).toContain('/logs');
  });

  test('should display log content from API', async ({ page }) => {
    await page.waitForResponse('**/logs', { timeout: 10000 });
    
    // Check if the mocked log content is displayed
    await expect(page.getByText(/sample log line 1/i)).toBeVisible();
    await expect(page.getByText(/sample log line 2/i)).toBeVisible();
    await expect(page.getByText(/sample log line 3/i)).toBeVisible();
  });

  test('should display logs in a text format', async ({ page }) => {
    await page.waitForResponse('**/logs', { timeout: 10000 });
    
    // Check if there's a text element containing logs
    const logsContainer = page.locator('text=/sample log line/i');
    await expect(logsContainer).toBeVisible();
  });
});
