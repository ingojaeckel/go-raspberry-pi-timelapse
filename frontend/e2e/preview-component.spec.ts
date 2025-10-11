import { test, expect } from './fixtures';

test.describe('PreviewComponent', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    // Switch to preview tab
    await page.getByRole('tab', { name: /preview/i }).click();
  });

  test('should display preview image', async ({ page }) => {
    const previewImage = page.getByAltText('preview');
    await expect(previewImage).toBeVisible();
    
    // Check the image source
    const src = await previewImage.getAttribute('src');
    expect(src).toContain('/capture');
  });

  test('should display tips section', async ({ page }) => {
    await expect(page.getByText(/tips for fine-tuning the camera position/i)).toBeVisible();
  });

  test('should display TFT tutorial link', async ({ page }) => {
    // Check for Adafruit tutorial links
    const links = page.locator('a[href*="adafruit"]');
    await expect(links.first()).toBeVisible();
  });

  test('should display VLC streaming information', async ({ page }) => {
    // VLC appears in a link
    await expect(page.getByRole('link', { name: /vlc/i })).toBeVisible();
  });

  test('should display VLC command examples', async ({ page }) => {
    await expect(page.getByText(/vlc udp/i)).toBeVisible();
    await expect(page.getByText(/raspivid/i)).toBeVisible();
  });

  test('should display multiple tip items', async ({ page }) => {
    // Check that there are list items in the tips section
    const listItems = page.locator('ul > li');
    await expect(listItems).toHaveCount(2);
  });

  test('should have external links with proper attributes', async ({ page }) => {
    // Find an external link (should open in new tab/window)
    const externalLinks = page.locator('a[href^="http"]');
    const count = await externalLinks.count();
    expect(count).toBeGreaterThan(0);
  });
});
