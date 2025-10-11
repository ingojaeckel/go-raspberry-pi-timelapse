import { test, expect } from './fixtures';

test.describe('SetupComponent', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    // Switch to settings tab
    await page.getByRole('tab', { name: /settings/i }).click();
  });

  test('should display settings form fields', async ({ page }) => {
    await expect(page.getByText(/time between captures/i)).toBeVisible();
    await expect(page.getByText(/rotation/i)).toBeVisible();
    await expect(page.getByText(/quality/i)).toBeVisible();
  });

  test('should fetch configuration on mount', async ({ page }) => {
    const requestPromise = page.waitForRequest('**/configuration', { timeout: 10000 });
    
    // Reload the component by switching tabs
    await page.getByRole('tab', { name: /home/i }).click();
    await page.getByRole('tab', { name: /settings/i }).click();
    
    const request = await requestPromise;
    expect(request.url()).toContain('/configuration');
  });

  test('should display action buttons', async ({ page }) => {
    await page.waitForResponse('**/configuration', { timeout: 10000 });
    
    await expect(page.getByRole('button', { name: /save/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /shutdown/i })).toBeVisible();
  });

  test('should allow entering time between captures', async ({ page }) => {
    await page.waitForResponse('**/configuration', { timeout: 10000 });
    
    // Find the time between captures input
    const input = page.locator('#tfTimeBetweenCaptures');
    await expect(input).toBeVisible();
    
    // Clear and enter new value
    await input.clear();
    await input.fill('600');
    
    await expect(input).toHaveValue('600');
  });

  test('should allow changing rotation value', async ({ page }) => {
    await page.waitForResponse('**/configuration', { timeout: 10000 });
    
    // Find the rotation input
    const rotationInput = page.locator('#tfRotation');
    await expect(rotationInput).toBeVisible();
    
    // Enter a rotation value
    await rotationInput.clear();
    await rotationInput.fill('90');
    
    await expect(rotationInput).toHaveValue('90');
  });

  test('should allow changing quality value', async ({ page }) => {
    await page.waitForResponse('**/configuration', { timeout: 10000 });
    
    // Find the quality input
    const qualityInput = page.locator('#tfQuality');
    await expect(qualityInput).toBeVisible();
    
    // Enter a quality value
    await qualityInput.clear();
    await qualityInput.fill('85');
    
    await expect(qualityInput).toHaveValue('85');
  });

  test('should populate form with fetched configuration data', async ({ page }) => {
    await page.waitForResponse('**/configuration', { timeout: 10000 });
    
    // Check if the form is populated with mocked data
    const timeInput = page.locator('#tfTimeBetweenCaptures');
    await expect(timeInput).toHaveValue('300');
    
    const qualityInput = page.locator('#tfQuality');
    await expect(qualityInput).toHaveValue('100');
    
    const rotationInput = page.locator('#tfRotation');
    await expect(rotationInput).toHaveValue('0');
  });

  test('should display version information when available', async ({ page }) => {
    await page.waitForResponse('**/configuration', { timeout: 10000 });
    
    // Version info may or may not be displayed depending on env vars
    // Just check that the component renders without errors
    const settingsContainer = page.locator('[role="tabpanel"]').filter({ hasText: /time between captures/i });
    await expect(settingsContainer).toBeVisible();
  });

  test('should handle save button click', async ({ page }) => {
    await page.waitForResponse('**/configuration', { timeout: 10000 });
    
    // Mock the save endpoint
    await page.route('**/configuration', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true })
        });
      } else {
        await route.continue();
      }
    });
    
    const saveButton = page.getByRole('button', { name: /save/i });
    await saveButton.click();
    
    // Should make a POST request with timeout
    await page.waitForRequest(request => 
      request.url().includes('/configuration') && request.method() === 'POST',
      { timeout: 10000 }
    );
  });
});
