# Timelapse Frontend

This project is built with [Vite](https://vitejs.dev/) and React 18.

## Prerequisites

This project uses Node.js 22.20.0 LTS (as specified in `.nvmrc`).

If you have [nvm](https://github.com/nvm-sh/nvm) installed, you can use the correct Node version by running:

```bash
nvm use
```

## Available Scripts

In the project directory, you can run:

### `npm start`

Runs the app in development mode using Vite dev server.<br />
Open [http://localhost:5173](http://localhost:5173) to view it in the browser.

The page will automatically reload when you make changes.<br />
You will see build errors and lint warnings in the console.

### `npm test`

Runs the test suite using [Vitest](https://vitest.dev/).<br />
Tests run in watch mode by default.

### `npm run test:e2e`

Runs end-to-end tests using Playwright. These tests run against the actual application in a browser.<br />
**Note:** The application dev server is started automatically by Playwright.

### `npm run test:e2e:ui`

Runs Playwright tests in UI mode, which provides an interactive interface for debugging tests.

### `npm run test:e2e:headed`

Runs Playwright tests in headed mode (browser window is visible).

### `npm run test:e2e:debug`

Runs Playwright tests in debug mode with the Playwright Inspector.

### `npm run build`

Builds the app for production to the `build` folder.<br />
It correctly bundles React in production mode and optimizes the build for best performance.

The build is minified and the filenames include content hashes.<br />
Your app is ready to be deployed!

### `npm run preview`

Preview the production build locally.<br />
This serves the built files from the `build` folder.

### `npm run analyze`

Analyzes the bundle size using source-map-explorer.<br />
This helps identify large dependencies in your bundle.

## Building for Production

Use the `build.sh` script to build with proper git metadata:

```bash
./build.sh
```

This script:
- Automatically loads the correct Node version via nvm (if available)
- Sets git commit information as environment variables
- Runs the production build

## Learn More

- [Vite Documentation](https://vitejs.dev/)
- [React Documentation](https://reactjs.org/)
- [Vitest Documentation](https://vitest.dev/)

## Testing

This project includes two types of tests:

### Unit Tests

Unit tests use Vitest and React Testing Library. They test individual components in isolation with mocked dependencies.

Run unit tests with:
```bash
npm test
```

### End-to-End Tests

End-to-end (E2E) tests use [Playwright](https://playwright.dev) to test the application as a whole in a real browser environment. These tests:

- Test core user flows and component interactions
- Run against the actual running application
- Verify integration between frontend and backend
- Complement unit tests by catching integration issues

**Prerequisites:** The application dev server is started automatically by Playwright when running tests.

**Available E2E test commands:**
- `npm run test:e2e` - Run all E2E tests headlessly
- `npm run test:e2e:ui` - Run tests in interactive UI mode
- `npm run test:e2e:headed` - Run tests with visible browser
- `npm run test:e2e:debug` - Run tests in debug mode

**E2E Test Coverage:**
- App navigation and tab switching
- Photo list, selection, and deletion
- Preview image display
- System monitoring metrics
- Settings form interactions
- Log viewing

The E2E tests use mocked API responses by default, making them suitable for rapid development and CI/CD pipelines. To test against a real backend, update the `baseURL` in `playwright.config.ts` to point to your running Go service.

For more details, see `e2e/README.md` and `e2e/QUICKSTART.md`.
