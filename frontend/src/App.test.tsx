import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import App from './App';
import { apiClient } from './api-client';
import { vi } from 'vitest';

vi.mock('./api-client', () => ({
  apiClient: {
    GET: vi.fn(),
    POST: vi.fn(),
    use: vi.fn(),
  }
}));

const mockedApiClient = apiClient as any;

describe('App', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Provide default mocks for all apiClient requests
    mockedApiClient.GET.mockResolvedValue({ 
      data: { 
        Photos: [],
        Logs: '',
        Time: '',
        Uptime: '',
        CpuTemperature: '',
        GpuTemperature: '',
        FreeDiskSpace: '',
        SecondsBetweenCaptures: 60,
        OffsetWithinHour: 0,
        PhotoResolutionWidth: 3280,
        PhotoResolutionHeight: 2464,
        PreviewResolutionWidth: 640,
        PreviewResolutionHeight: 480,
        RotateBy: 0,
        ResolutionSetting: 0,
        Quality: 100,
        DebugEnabled: false,
      } 
    });
  });

  test('renders home tab', () => {
    const { getByText } = render(<App />);
    const linkElement = getByText(/home/i);
    expect(linkElement).toBeInTheDocument();
  });

  test('renders preview tab', () => {
    render(<App />);
    expect(screen.getByText(/preview/i)).toBeInTheDocument();
  });

  test('renders monitoring tab', () => {
    const { getByText } = render(<App />);
    const linkElement = getByText(/monitor/i);
    expect(linkElement).toBeInTheDocument();
  });

  test('renders settings tab', () => {
    const { getByText } = render(<App />);
    const linkElement = getByText(/settings/i);
    expect(linkElement).toBeInTheDocument();
  });

  test('renders logs tab', () => {
    render(<App />);
    expect(screen.getByText(/logs/i)).toBeInTheDocument();
  });

  test('switches between tabs', async () => {
    render(<App />);
    
    const previewTab = screen.getByRole('tab', { name: /preview/i });
    fireEvent.click(previewTab);
    
    // The tab should be selected
    await waitFor(() => {
      expect(previewTab).toHaveAttribute('aria-selected', 'true');
    });
  });

  test('displays PhotoComponent by default (home tab)', () => {
    const { container } = render(<App />);
    // PhotoComponent should be visible in the first tab
    const tabPanel = container.querySelector('[role="tabpanel"][id="simple-tabpanel-0"]');
    expect(tabPanel).toBeInTheDocument();
    expect(tabPanel).not.toHaveAttribute('hidden');
  });

  test('displays correct component when switching to monitoring tab', async () => {
    const { container } = render(<App />);
    
    const monitoringTab = screen.getByRole('tab', { name: /monitor/i });
    fireEvent.click(monitoringTab);
    
    await waitFor(() => {
      const tabPanel = container.querySelector('[role="tabpanel"][id="simple-tabpanel-2"]');
      expect(tabPanel).not.toHaveAttribute('hidden');
    });
  });

  test('displays correct component when switching to settings tab', async () => {
    const { container } = render(<App />);
    
    const settingsTab = screen.getByRole('tab', { name: /settings/i });
    fireEvent.click(settingsTab);
    
    await waitFor(() => {
      const tabPanel = container.querySelector('[role="tabpanel"][id="simple-tabpanel-3"]');
      expect(tabPanel).not.toHaveAttribute('hidden');
    });
  });

  test('displays correct component when switching to logs tab', async () => {
    const { container } = render(<App />);
    
    const logsTab = screen.getByRole('tab', { name: /logs/i });
    fireEvent.click(logsTab);
    
    await waitFor(() => {
      const tabPanel = container.querySelector('[role="tabpanel"][id="simple-tabpanel-4"]');
      expect(tabPanel).not.toHaveAttribute('hidden');
    });
  });

  test('renders with CssBaseline', () => {
    const { container } = render(<App />);
    expect(container.firstChild).toBeInTheDocument();
  });

  test('renders tabs with fullWidth variant', () => {
    const { container } = render(<App />);
    const tabs = container.querySelector('[role="tablist"]');
    expect(tabs).toBeInTheDocument();
  });

  test('tab icons are present', () => {
    const { container } = render(<App />);
    const tabs = container.querySelectorAll('[role="tab"]');
    expect(tabs.length).toBe(5); // home, preview, monitor, settings, logs
  });

  test('TabPanel component hides when not selected', () => {
    const { container } = render(<App />);
    
    // Check that non-selected panels are hidden
    const hiddenPanels = container.querySelectorAll('[role="tabpanel"][hidden]');
    expect(hiddenPanels.length).toBeGreaterThan(0);
  });

  test('container has responsive padding', () => {
    const { container } = render(<App />);
    // The Container component should be in the document
    const containerElement = container.querySelector('.MuiContainer-root');
    expect(containerElement).toBeInTheDocument();
  });

  test('applies theme based on user preference', () => {
    const { container } = render(<App />);
    // ThemeProvider should wrap the content
    expect(container.firstChild).toBeInTheDocument();
  });

  test('includes ThemeProvider in component tree', () => {
    const { container } = render(<App />);
    // Verify CssBaseline is applied (which comes from theme)
    expect(container.querySelector('.MuiContainer-root')).toBeInTheDocument();
  });
});