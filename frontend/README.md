This project was bootstrapped with [Create React App](https://github.com/facebook/create-react-app).

## Available Scripts

In the project directory, you can run:

### `npm start`

Runs the app in the development mode.<br />
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

The page will reload if you make edits.<br />
You will also see any lint errors in the console.

### `npm test`

Launches the test runner in the interactive watch mode.<br />
See the section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) for more information.

### `npm run test:e2e`

Runs end-to-end tests using Playwright. These tests run against the actual application in a browser.<br />
**Note:** The application must be running (via `npm start`) before running the E2E tests.

### `npm run test:e2e:ui`

Runs Playwright tests in UI mode, which provides an interactive interface for debugging tests.

### `npm run test:e2e:headed`

Runs Playwright tests in headed mode (browser window is visible).

### `npm run test:e2e:debug`

Runs Playwright tests in debug mode with the Playwright Inspector.

### `npm run build`

Builds the app for production to the `build` folder.<br />
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.<br />
Your app is ready to be deployed!

See the section about [deployment](https://facebook.github.io/create-react-app/docs/deployment) for more information.

### `npm run eject`

**Note: this is a one-way operation. Once you `eject`, you can’t go back!**

If you aren’t satisfied with the build tool and configuration choices, you can `eject` at any time. This command will remove the single build dependency from your project.

Instead, it will copy all the configuration files and the transitive dependencies (webpack, Babel, ESLint, etc) right into your project so you have full control over them. All of the commands except `eject` will still work, but they will point to the copied scripts so you can tweak them. At this point you’re on your own.

You don’t have to ever use `eject`. The curated feature set is suitable for small and middle deployments, and you shouldn’t feel obligated to use this feature. However we understand that this tool wouldn’t be useful if you couldn’t customize it when you are ready for it.

## Learn More

You can learn more in the [Create React App documentation](https://facebook.github.io/create-react-app/docs/getting-started).

To learn React, check out the [React documentation](https://reactjs.org/).

## Testing

This project includes two types of tests:

### Unit Tests

Unit tests use React Testing Library and Jest. They test individual components in isolation with mocked dependencies.

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

**Prerequisites:** The application must be running before executing E2E tests:
```bash
# In one terminal, start the frontend dev server
npm start

# In another terminal, run the E2E tests
npm run test:e2e
```

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
