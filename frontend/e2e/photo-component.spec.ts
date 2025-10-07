import { test, expect } from './fixtures';

test.describe('PhotoComponent', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    // Should be on home tab by default
  });

  test('should display photo list with data grid', async ({ page }) => {
    // Wait for the data grid to load
    await expect(page.locator('[role="grid"]')).toBeVisible();
  });

  test('should display refresh and delete buttons', async ({ page }) => {
    await expect(page.getByRole('button', { name: /refresh/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /delete selected/i })).toBeVisible();
  });

  test('should display download links', async ({ page }) => {
    await expect(page.getByRole('link', { name: /download all \(zip\)/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /download all \(tar\)/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /download selected/i })).toBeVisible();
  });

  test('should display photos from API in data grid', async ({ page }) => {
    // Wait for the grid to load
    await page.waitForSelector('[role="grid"]');
    
    // Check if the test photos are displayed
    await expect(page.getByText('test-photo-1.jpg')).toBeVisible();
    await expect(page.getByText('test-photo-2.jpg')).toBeVisible();
  });

  test('should allow selecting photos with checkboxes', async ({ page }) => {
    await page.waitForSelector('[role="grid"]');
    
    // Find and click the first checkbox in the grid
    const checkboxes = page.locator('[role="gridcell"] input[type="checkbox"]');
    await checkboxes.first().click();
    
    // The delete button text should update to show selected count
    await expect(page.getByRole('button', { name: /delete selected \(1\)/i })).toBeVisible();
  });

  test('should show delete confirmation dialog when delete is clicked with selection', async ({ page }) => {
    await page.waitForSelector('[role="grid"]');
    
    // Select a photo
    const checkboxes = page.locator('[role="gridcell"] input[type="checkbox"]');
    await checkboxes.first().click();
    
    // Click delete button
    await page.getByRole('button', { name: /delete selected/i }).click();
    
    // Check if confirmation dialog appears
    await expect(page.getByText(/are you sure you want to delete/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /cancel/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /^delete$/i })).toBeVisible();
  });

  test('should close delete dialog when cancel is clicked', async ({ page }) => {
    await page.waitForSelector('[role="grid"]');
    
    // Select a photo
    const checkboxes = page.locator('[role="gridcell"] input[type="checkbox"]');
    await checkboxes.first().click();
    
    // Click delete button
    await page.getByRole('button', { name: /delete selected/i }).click();
    
    // Click cancel
    await page.getByRole('button', { name: /cancel/i }).click();
    
    // Dialog should be gone
    await expect(page.getByText(/are you sure you want to delete/i)).not.toBeVisible();
  });

  test('should make refresh API call when refresh button is clicked', async ({ page }) => {
    // Set up request listener
    const requestPromise = page.waitForRequest('**/photos');
    
    // Click refresh button
    await page.getByRole('button', { name: /refresh/i }).click();
    
    // Wait for the request
    const request = await requestPromise;
    expect(request.url()).toContain('/photos');
  });

  test('should display photo file names as clickable links', async ({ page }) => {
    await page.waitForSelector('[role="grid"]');
    
    // Check if photo names are links
    const photoLink = page.getByRole('link', { name: 'test-photo-1.jpg' });
    await expect(photoLink).toBeVisible();
    
    // Check href contains the file path
    const href = await photoLink.getAttribute('href');
    expect(href).toContain('/file/test-photo-1.jpg');
  });
});
