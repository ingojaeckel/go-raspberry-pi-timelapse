# Testing Against Live Backend

This guide explains how to run Playwright tests against a running Go backend service instead of using mocked APIs.

## Prerequisites

1. Go backend service running on port 8080
2. Frontend built and embedded in the backend (or running separately)

## Option 1: Testing Against Production Build

If you have the Go service running with the embedded frontend:

1. Start the Go service:
   ```bash
   # From repository root
   ./timelapsepi
   ```

2. Update `playwright.config.ts` to point to the Go service:
   ```typescript
   use: {
     baseURL: 'http://localhost:8080/static/frontend/build',
   }
   ```

3. Disable the webServer config (comment out the `webServer` section in `playwright.config.ts`)

4. Comment out the API mocks in `e2e/fixtures.ts` (or create a separate fixture file)

5. Run tests:
   ```bash
   npm run test:e2e
   ```

## Option 2: Testing Development Frontend with Live Backend

If you want to test the development frontend against a live backend:

1. Start the Go backend service on port 8080

2. Configure the frontend to use the backend API (already configured via `src/conf/config.tsx`)

3. Keep the default `playwright.config.ts` configuration (baseURL: `http://localhost:5173`)

4. Comment out API route mocks in `e2e/fixtures.ts`:
   ```typescript
   export const test = base.extend({
     page: async ({ page }, use) => {
       // Comment out all page.route() calls to use real backend
       await use(page);
     },
   });
   ```

5. Start the dev server and run tests:
   ```bash
   # Terminal 1: Start frontend dev server
   npm start
   
   # Terminal 2: Run Playwright tests
   npm run test:e2e
   ```

## Option 3: Selective API Mocking

You can also mock some APIs while using real ones for others. Simply comment out specific route mocks in `fixtures.ts`:

```typescript
export const test = base.extend({
  page: async ({ page }, use) => {
    // Mock /photos endpoint
    await page.route('**/photos', async (route) => {
      await route.fulfill({ /* ... */ });
    });
    
    // Let /monitoring use real backend - don't add route for it
    // await page.route('**/monitoring', ...) // COMMENTED OUT
    
    await use(page);
  },
});
```

## Expected Behavior Differences

When testing against a real backend:

### Photos Component
- Real photo data from your storage folder
- Actual file deletion (be careful!)
- Real file downloads

### Monitoring Component
- Actual system metrics
- Real CPU/GPU temperatures
- Actual disk space

### Settings Component
- Real configuration saved to backend
- Actual system shutdown (be careful!)

### Preview Component
- Real camera preview image

### Logs Component
- Actual application logs

## Safety Considerations

⚠️ **Warning**: When testing against a live backend:

1. **Don't use production data** - Use a test instance
2. **Photo deletion is real** - Deleted files are actually removed
3. **Shutdown button works** - Will actually shutdown the system
4. **Configuration changes persist** - Settings are saved

## Debugging

To debug tests against the live backend:

```bash
# Run in debug mode with visible browser
npm run test:e2e:debug

# Or run in headed mode
npm run test:e2e:headed
```

## Troubleshooting

### Tests fail with "ERR_CONNECTION_REFUSED"
- Ensure the Go backend is running on port 8080
- Check that the baseURL in `playwright.config.ts` is correct

### Tests time out
- Backend might be slow to respond
- Increase timeout in `playwright.config.ts`:
  ```typescript
  use: {
    timeout: 30000, // 30 seconds per action
  }
  ```

### API responses don't match expected format
- Check backend is returning correct data structures
- Update test expectations to match actual API responses
- Verify API endpoints match between frontend and backend
