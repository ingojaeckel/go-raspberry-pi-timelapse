import { test as base, expect } from '@playwright/test';

/**
 * Mock API responses for testing without a live backend
 */
export const test = base.extend({
  page: async ({ page }, use) => {
    // Mock the /photos endpoint
    await page.route('**/photos', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          Photos: [
            {
              Name: 'test-photo-1.jpg',
              ModTime: '2024-01-01 12:00:00',
              Size: 1024000
            },
            {
              Name: 'test-photo-2.jpg',
              ModTime: '2024-01-01 13:00:00',
              Size: 2048000
            }
          ]
        })
      });
    });

    // Mock the /monitoring endpoint
    await page.route('**/monitoring', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          Time: '2024-01-01 12:00:00',
          Uptime: '5 days',
          CpuTemperature: '45.0°C',
          CpuTemperatureFloat: 45.0,
          GpuTemperature: '43.0°C',
          GpuTemperatureFloat: 43.0,
          FreeDiskSpace: '10GB'
        })
      });
    });

    // Mock the /configuration endpoint
    await page.route('**/configuration', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          SecondsBetweenCaptures: 300,
          OffsetWithinHour: 0,
          PhotoResolutionWidth: 3280,
          PhotoResolutionHeight: 2464,
          PreviewResolutionWidth: 640,
          PreviewResolutionHeight: 480,
          RotateBy: 0,
          ResolutionSetting: 0,
          Quality: 100,
          DebugEnabled: false
        })
      });
    });

    // Mock the /logs endpoint
    await page.route('**/logs', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          Logs: 'Sample log line 1\nSample log line 2\nSample log line 3'
        })
      });
    });

    // Mock the /capture endpoint for preview image
    await page.route('**/capture', async (route) => {
      // Return a 1x1 transparent PNG
      const transparentPng = Buffer.from(
        'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
        'base64'
      );
      await route.fulfill({
        status: 200,
        contentType: 'image/png',
        body: transparentPng
      });
    });

    await use(page);
  },
});

export { expect };
