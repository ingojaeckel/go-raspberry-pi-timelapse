# Dark Mode Feature

## Overview

The frontend now supports automatic dark mode that respects the user's device/browser color scheme preference. This provides a more eye-friendly and power-saving experience in low-light environments.

## How It Works

The implementation uses:
- **MUI's `useMediaQuery` hook** to detect `prefers-color-scheme: dark` 
- **MUI's `createTheme()`** to create a theme based on the detected mode
- **MUI's `ThemeProvider`** to apply the theme to all components
- **MUI's default color palettes** for both light and dark modes

The app automatically switches between light and dark mode based on the user's system preferences, with no manual configuration needed.

## Testing Dark Mode

### Method 1: System-Wide Settings

**Windows:**
1. Go to Settings → Personalization → Colors
2. Choose "Dark" under "Choose your color"

**macOS:**
1. Go to System Preferences → General
2. Select "Dark" under Appearance

**Linux:**
1. Varies by desktop environment (GNOME, KDE, etc.)
2. Usually found in System Settings → Appearance

### Method 2: Browser DevTools (Chrome/Edge)

1. Open the page in your browser
2. Open DevTools (F12 or Cmd+Option+I)
3. Press `Ctrl+Shift+P` (Windows/Linux) or `Cmd+Shift+P` (Mac)
4. Type "Rendering" and select "Show Rendering"
5. Find "Emulate CSS media feature prefers-color-scheme"
6. Select either:
   - `prefers-color-scheme: dark` - to see dark mode
   - `prefers-color-scheme: light` - to see light mode

### Method 3: Firefox DevTools

1. Open the page in Firefox
2. Open DevTools (F12)
3. Click the three-dot menu (⋯) in DevTools
4. Select "Settings"
5. Scroll to "Emulate CSS media feature prefers-color-scheme"
6. Select your preferred mode

## Implementation Details

### Code Changes

The implementation required minimal changes to `App.tsx`:

```tsx
// Added imports
import { ThemeProvider, createTheme, useMediaQuery } from '@mui/material';

// In App component
const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)');

const theme = React.useMemo(
  () =>
    createTheme({
      palette: {
        mode: prefersDarkMode ? 'dark' : 'light',
      },
    }),
  [prefersDarkMode],
);

// Wrapped the app with ThemeProvider
return (
  <ThemeProvider theme={theme}>
    <CssBaseline />
    {/* ... rest of app ... */}
  </ThemeProvider>
);
```

### Features

✅ **Automatic Detection** - No user action required  
✅ **MUI Integration** - Uses Material-UI's built-in theming  
✅ **Responsive** - Changes dynamically when system preference changes  
✅ **Power Saving** - Dark mode reduces power consumption on OLED screens  
✅ **Eye Friendly** - Reduces eye strain in low-light conditions  
✅ **Default Colors** - Uses MUI's carefully crafted default color palettes

### Dark Mode Colors (MUI Defaults)

**Light Mode:**
- Background: `#fff` (white)
- Text: `rgba(0, 0, 0, 0.87)` (dark gray)
- Primary: `#1976d2` (blue)
- Secondary: `#dc004e` (pink)

**Dark Mode:**
- Background: `#121212` (very dark gray, not pure black)
- Text: `#fff` (white)
- Primary: `#90caf9` (light blue)
- Secondary: `#f48fb1` (light pink)

## Browser Support

The `prefers-color-scheme` media query is supported in:
- Chrome 76+
- Firefox 67+
- Safari 12.1+
- Edge 79+
- Opera 62+

For older browsers, the app defaults to light mode.

## Future Enhancements

Potential improvements for future versions:
- Manual toggle to override system preference
- Remember user's manual preference in localStorage
- Smooth transitions when switching modes
- Per-component theme customization
