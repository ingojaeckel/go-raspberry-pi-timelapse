# Playwright E2E Testing - Quick Start Guide

This is a quick reference for developers working with the Playwright E2E tests.

## Quick Commands

```bash
# Install dependencies (first time only)
cd frontend
npm install

# Install Playwright browsers (first time only)
npx playwright install chromium

# Run all E2E tests (headless)
npm run test:e2e

# Run tests with UI (recommended for development)
npm run test:e2e:ui

# Run tests with browser visible
npm run test:e2e:headed

# Debug a specific test
npm run test:e2e:debug

# Run a specific test file
npx playwright test e2e/app-navigation.spec.ts

# Run tests matching a pattern
npx playwright test --grep "should display"
```

## Test Files Overview

```
frontend/e2e/
├── fixtures.ts                    # Shared test fixtures & API mocking
├── app-navigation.spec.ts         # Tab navigation (8 tests)
├── photo-component.spec.ts        # Photo management (9 tests)
├── preview-component.spec.ts      # Camera preview (7 tests)
├── monitoring-component.spec.ts   # System monitoring (6 tests)
├── setup-component.spec.ts        # Settings (9 tests)
├── log-component.spec.ts          # Logs viewing (4 tests)
├── README.md                      # Detailed documentation
└── TESTING_LIVE_BACKEND.md        # Guide for live backend testing
```

## Development Workflow

### 1. Adding New Tests

```typescript
// e2e/my-component.spec.ts
import { test, expect } from './fixtures';

test.describe('MyComponent', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    // Navigate to your component
  });

  test('should do something', async ({ page }) => {
    await expect(page.getByText('Hello')).toBeVisible();
  });
});
```

### 2. Running Your New Tests

```bash
# Run in UI mode for interactive development
npm run test:e2e:ui

# Or run specific file
npx playwright test e2e/my-component.spec.ts --headed
```

### 3. Debugging Failures

```bash
# Run in debug mode with inspector
npm run test:e2e:debug

# Show trace for failed tests
npx playwright show-trace test-results/trace.zip
```

## Common Patterns

### Wait for API Response
```typescript
await page.waitForResponse('**/photos');
```

### Click and Verify
```typescript
await page.getByRole('button', { name: /refresh/i }).click();
await expect(page.getByText('Success')).toBeVisible();
```

### Find Elements
```typescript
// By role
page.getByRole('button', { name: /save/i })

// By text
page.getByText(/hello world/i)

// By label
page.getByLabel('Username')

// By test ID
page.getByTestId('submit-button')

// By placeholder
page.getByPlaceholder('Enter email')
```

### Form Interactions
```typescript
await page.locator('#myInput').fill('value');
await page.locator('#mySelect').selectOption('option1');
await page.locator('#myCheckbox').check();
```

## Testing Modes

### Mode 1: Mocked APIs (Default)
- Fast, reliable, no backend needed
- Great for CI/CD
- Tests frontend logic in isolation

### Mode 2: Live Backend
- Tests full integration
- Requires running Go service
- See `TESTING_LIVE_BACKEND.md` for setup

## Tips

1. **Use UI Mode for Development**
   ```bash
   npm run test:e2e:ui
   ```
   - See tests as they run
   - Time travel through test steps
   - Easy debugging

2. **Use Headed Mode for Debugging**
   ```bash
   npm run test:e2e:headed
   ```
   - Watch browser in real-time
   - See what's happening

3. **Use Debug Mode for Stepping Through**
   ```bash
   npm run test:e2e:debug
   ```
   - Pause execution
   - Step through test
   - Inspect page state

4. **Keep Tests Independent**
   - Each test should work on its own
   - Use `beforeEach` for setup
   - Don't rely on test order

5. **Use Good Selectors**
   - Prefer semantic selectors (role, label)
   - Avoid CSS selectors when possible
   - Make tests resilient to UI changes

## Troubleshooting

### "Target closed" errors
- App might have crashed
- Check console for errors
- Run in headed mode to see what happened

### Tests timeout
- Increase timeout in config
- Check if API is responding
- Verify element selectors are correct

### Can't find elements
- Use Playwright Inspector
- Check page.screenshot() output
- Verify element is visible (not hidden)

### Flaky tests
- Add proper waits (`waitForResponse`, `waitForSelector`)
- Don't use hardcoded timeouts
- Check for race conditions

## CI/CD Integration

Tests are configured to run in CI with:
- Automatic retries (2x on failure)
- HTML report generation
- Screenshots on failure
- Video recording on failure

## Learn More

- [Playwright Docs](https://playwright.dev)
- [Best Practices](https://playwright.dev/docs/best-practices)
- [Selectors Guide](https://playwright.dev/docs/selectors)
- [Assertions](https://playwright.dev/docs/test-assertions)
