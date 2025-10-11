# End-to-End Tests

This directory contains Playwright-based end-to-end tests for the timelapse application frontend.

## Test Files

- **fixtures.ts** - Shared test fixtures and API mocking setup
- **app-navigation.spec.ts** - Tests for app navigation and tab switching
- **photo-component.spec.ts** - Tests for photo list, selection, and deletion
- **preview-component.spec.ts** - Tests for camera preview display
- **monitoring-component.spec.ts** - Tests for system monitoring display
- **setup-component.spec.ts** - Tests for settings form interactions
- **log-component.spec.ts** - Tests for log viewing

## Timeout Configuration

Tests are configured with the following timeouts to prevent hanging:

- **Test timeout**: 30 seconds (maximum time for a single test)
- **Action timeout**: 10 seconds (maximum time for page actions like click, fill)
- **Navigation timeout**: 10 seconds (maximum time for page navigation)
- **API wait timeout**: 10 seconds (explicit timeout on waitForRequest/waitForResponse)
- **Web server startup**: 60 seconds (Vite dev/preview server)

These timeouts ensure tests fail quickly if something goes wrong, rather than relying on external CI timeout.

**CI Optimization**: In CI environments, tests use `vite preview` to serve pre-built files instead of the dev server. This is much faster and more reliable:
- Local development: Uses `npm start` (Vite dev server on port 5173)
- CI environment: Uses `npm run preview` (Vite preview server on port 4173 serving pre-built files)

The frontend is built in development mode before running E2E tests in CI, ensuring the base path is `/` for testing.

## Running Tests

### Prerequisites

The tests can run in two modes:

1. **With mocked APIs (default)**: Tests run independently without a backend
2. **Against live backend**: Update `playwright.config.ts` baseURL to point to your Go service

### Commands

```bash
# Run all tests headlessly
npm run test:e2e

# Run tests in UI mode (interactive)
npm run test:e2e:ui

# Run tests with browser visible
npm run test:e2e:headed

# Run tests in debug mode
npm run test:e2e:debug
```

## Test Structure

Each test file follows this pattern:

1. Import test fixtures with API mocking
2. Navigate to the appropriate section of the app
3. Test user interactions and verify expected outcomes
4. Check API calls are made correctly

## API Mocking

The `fixtures.ts` file provides automatic API mocking for all tests. This ensures tests:

- Run quickly and reliably
- Don't depend on backend availability
- Can be run in CI/CD pipelines
- Test frontend behavior in isolation

Mocked endpoints:
- `/photos` - Returns sample photo data
- `/monitoring` - Returns sample system metrics
- `/configuration` - Returns sample configuration
- `/logs` - Returns sample log content
- `/capture` - Returns a placeholder image

## Writing New Tests

To add new tests:

1. Create a new `.spec.ts` file in this directory
2. Import `test` and `expect` from `./fixtures`
3. Use `test.describe()` to group related tests
4. Use `test.beforeEach()` to set up each test
5. Write tests using Playwright's assertions

Example:
```typescript
import { test, expect } from './fixtures';

test.describe('My Component', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should do something', async ({ page }) => {
    await expect(page.getByText('Hello')).toBeVisible();
  });
});
```

## Testing Against Live Backend

To test against a running Go service:

1. Start your Go backend service on port 8080
2. Update `playwright.config.ts`:
   ```typescript
   use: {
     baseURL: 'http://localhost:8080/static/frontend/build',
     // ...
   }
   ```
3. Comment out or remove the route mocks in `fixtures.ts`
4. Run tests: `npm run test:e2e`

See `TESTING_LIVE_BACKEND.md` for detailed instructions.

## CI/CD Integration

These tests are designed to run in continuous integration:

- Fast execution with mocked APIs
- Retry logic on CI (configured in `playwright.config.ts`)
- HTML report generation
- Screenshot/video capture on failure

## Learn More

- [Playwright Documentation](https://playwright.dev)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [Playwright Test API](https://playwright.dev/docs/api/class-test)
