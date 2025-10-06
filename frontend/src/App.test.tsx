import React from 'react';
import { render, fireEvent, screen } from '@testing-library/react';
import axios from 'axios';
import App from './App';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('App Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Mock all axios calls to prevent network errors in child components
    mockedAxios.get.mockResolvedValue({ 
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
        DebugEnabled: false
      } 
    });
  });

  test('renders home tab', () => {
    const { getByText } = render(<App />);
    const linkElement = getByText(/home/i);
    expect(linkElement).toBeInTheDocument();
  });

  test('renders monitoring tab', () => {
    const { getByText } = render(<App />);
    const linkElement = getByText(/monitoring/i);
    expect(linkElement).toBeInTheDocument();
  });

  test('renders settings tab', () => {
    const { getByText } = render(<App />);
    const linkElement = getByText(/settings/i);
    expect(linkElement).toBeInTheDocument();
  });

  test('renders preview tab', () => {
    const { getByText } = render(<App />);
    const linkElement = getByText(/preview/i);
    expect(linkElement).toBeInTheDocument();
  });

  test('renders logs tab', () => {
    const { getByText } = render(<App />);
    const linkElement = getByText(/logs/i);
    expect(linkElement).toBeInTheDocument();
  });

  test('switches tabs on click', () => {
    render(<App />);
    const previewTab = screen.getByText(/preview/i);
    fireEvent.click(previewTab);
    // After clicking preview tab, verify it's selected by checking aria-selected
    expect(previewTab.closest('button')).toHaveAttribute('aria-selected', 'true');
  });

  test('displays PhotoComponent on home tab', () => {
    const { container } = render(<App />);
    // Home tab is active by default (index 0)
    const tabPanel = container.querySelector('#simple-tabpanel-0');
    expect(tabPanel).toBeInTheDocument();
    expect(tabPanel).not.toHaveAttribute('hidden');
  });

  test('TabPanel shows children when value matches index', () => {
    const { getByText } = render(<App />);
    // Click on monitoring tab (index 2)
    const monitoringTab = getByText(/monitoring/i);
    fireEvent.click(monitoringTab);
    // Check that the correct panel is visible
    const monitoringPanel = document.querySelector('#simple-tabpanel-2');
    expect(monitoringPanel).not.toHaveAttribute('hidden');
  });

  test('TabPanel hides children when value does not match index', () => {
    const { container } = render(<App />);
    // Check that non-active panels are hidden
    const previewPanel = container.querySelector('#simple-tabpanel-1');
    expect(previewPanel).toHaveAttribute('hidden');
  });
});