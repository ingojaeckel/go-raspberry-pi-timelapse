import { test, expect } from './fixtures';

test.describe('MonitoringComponent', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    // Switch to monitoring tab
    await page.getByRole('tab', { name: /monitoring/i }).click();
  });

  test('should display monitoring data labels', async ({ page }) => {
    await expect(page.getByText(/Temperature:/i)).toBeVisible();
    await expect(page.getByText(/GPU:/i)).toBeVisible();
    await expect(page.getByText(/CPU:/i)).toBeVisible();
  });

  test('should fetch and display monitoring data from API', async ({ page }) => {
    // Wait for the API call to complete with timeout
    await page.waitForResponse('**/monitoring', { timeout: 10000 });
    
    // Check if the mocked data is displayed
    await expect(page.getByText('2024-01-01 12:00:00')).toBeVisible();
    await expect(page.getByText('5 days')).toBeVisible();
    await expect(page.getByText('45.0°C')).toBeVisible();
    await expect(page.getByText('43.0°C')).toBeVisible();
    await expect(page.getByText('10GB')).toBeVisible();
  });

  test('should make API request to monitoring endpoint', async ({ page }) => {
    const requestPromise = page.waitForRequest('**/monitoring', { timeout: 10000 });
    
    // Reload the component by switching tabs
    await page.getByRole('tab', { name: /home/i }).click();
    await page.getByRole('tab', { name: /monitoring/i }).click();
    
    const request = await requestPromise;
    expect(request.url()).toContain('/monitoring');
  });

  test('should display system uptime', async ({ page }) => {
    await page.waitForResponse('**/monitoring', { timeout: 10000 });
    await expect(page.getByText(/5 days/i)).toBeVisible();
  });

  test('should display temperature information', async ({ page }) => {
    await page.waitForResponse('**/monitoring', { timeout: 10000 });
    
    // Check for temperature values
    const cpuTemp = page.getByText('45.0°C');
    const gpuTemp = page.getByText('43.0°C');
    
    await expect(cpuTemp).toBeVisible();
    await expect(gpuTemp).toBeVisible();
  });

  test('should display disk space information', async ({ page }) => {
    await page.waitForResponse('**/monitoring', { timeout: 10000 });
    await expect(page.getByText('10GB')).toBeVisible();
  });
});
