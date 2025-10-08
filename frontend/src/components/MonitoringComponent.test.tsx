import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import MonitoringComponent from './MonitoringComponent';
import { apiClient } from '../api-client';
import { vi } from 'vitest';

vi.mock('../api-client', () => ({
  apiClient: {
    GET: vi.fn(),
  }
}));

const mockedApiClient = apiClient as any;

describe('MonitoringComponent', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders without crashing', () => {
    mockedApiClient.GET.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
  });

  it('fetches monitoring data on mount', async () => {
    const mockData = {
      Time: '2024-01-01 12:00:00',
      Uptime: '5 days',
      CpuTemperature: '45.0°C',
      GpuTemperature: '43.0°C',
      FreeDiskSpace: '10GB',
    };

    mockedApiClient.GET.mockResolvedValue({ data: mockData });
    render(<MonitoringComponent />);

    await waitFor(() => {
      expect(mockedApiClient.GET).toHaveBeenCalledWith('/monitoring');
    });
  });

  it('displays current time label', () => {
    mockedApiClient.GET.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    expect(screen.getByText(/Current Time:/)).toBeInTheDocument();
  });

  it('displays uptime label', () => {
    mockedApiClient.GET.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    expect(screen.getByText(/Uptime:/)).toBeInTheDocument();
  });

  it('displays free disk space label', () => {
    mockedApiClient.GET.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    expect(screen.getByText(/Free Disk Space:/)).toBeInTheDocument();
  });

  it('displays temperature section', () => {
    mockedApiClient.GET.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    expect(screen.getByText(/Temperature:/)).toBeInTheDocument();
  });

  it('displays GPU temperature label', () => {
    mockedApiClient.GET.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    expect(screen.getByText(/GPU:/)).toBeInTheDocument();
  });

  it('displays CPU temperature label', () => {
    mockedApiClient.GET.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    expect(screen.getByText(/CPU:/)).toBeInTheDocument();
  });

  it('renders monitoring data when received', async () => {
    const mockData = {
      Time: '2024-01-01 12:00:00',
      Uptime: '5 days',
      CpuTemperature: '45.0°C',
      GpuTemperature: '43.0°C',
      FreeDiskSpace: '10GB',
    };

    mockedApiClient.GET.mockResolvedValue({ data: mockData });
    render(<MonitoringComponent />);

    await waitFor(() => {
      expect(screen.getByText('2024-01-01 12:00:00')).toBeInTheDocument();
      expect(screen.getByText('5 days')).toBeInTheDocument();
      expect(screen.getByText('45.0°C')).toBeInTheDocument();
      expect(screen.getByText('43.0°C')).toBeInTheDocument();
      expect(screen.getByText('10GB')).toBeInTheDocument();
    });
  });

  it('uses correct API endpoint', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);

    await waitFor(() => {
      expect(mockedApiClient.GET).toHaveBeenCalledWith('/monitoring');
    });
  });
});
