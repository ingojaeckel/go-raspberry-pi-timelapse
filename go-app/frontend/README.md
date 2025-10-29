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
