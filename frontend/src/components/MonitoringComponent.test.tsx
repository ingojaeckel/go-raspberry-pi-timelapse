import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import MonitoringComponent from './MonitoringComponent';
import axios from 'axios';
import { vi } from 'vitest';

vi.mock('axios');
const mockedAxios = axios as any;

describe('MonitoringComponent', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders without crashing', () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
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

    mockedAxios.get.mockResolvedValue({ data: mockData });
    render(<MonitoringComponent />);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/monitoring'));
    });
  });

  it('displays current time label', async () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    await waitFor(() => {
      expect(screen.getByText(/System Time/)).toBeInTheDocument();
    });
  });

  it('displays uptime label', async () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    await waitFor(() => {
      expect(screen.getByText(/Uptime/)).toBeInTheDocument();
    });
  });

  it('displays free disk space label', async () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    await waitFor(() => {
      expect(screen.getByText(/Free Disk Space/)).toBeInTheDocument();
    });
  });

  it('displays temperature section', async () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    await waitFor(() => {
      expect(screen.getByText(/CPU Temperature/)).toBeInTheDocument();
    });
  });

  it('displays GPU temperature label', async () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    await waitFor(() => {
      expect(screen.getByText(/GPU Temperature/)).toBeInTheDocument();
    });
  });

  it('displays CPU temperature label', async () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    await waitFor(() => {
      expect(screen.getByText(/CPU Temperature/)).toBeInTheDocument();
    });
  });

  it('renders monitoring data when received', async () => {
    const mockData = {
      Time: '2024-01-01 12:00:00',
      Uptime: '5 days',
      CpuTemperature: '45.0°C',
      GpuTemperature: '43.0°C',
      FreeDiskSpace: '10GB',
    };

    mockedAxios.get.mockResolvedValue({ data: mockData });
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
    mockedAxios.get.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith('http://localhost:8080/monitoring');
    });
  });

  it('displays loading state initially', () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
    const { container } = render(<MonitoringComponent />);
    
    expect(container.querySelector('.MuiCircularProgress-root')).toBeInTheDocument();
  });

  it('displays error when network request fails', async () => {
    mockedAxios.get.mockRejectedValue({ code: 'ERR_NETWORK', message: 'Network Error' });
    render(<MonitoringComponent />);

    await waitFor(() => {
      expect(screen.getByText(/Unable to connect to server/)).toBeInTheDocument();
    });
  });

  it('displays error for other failures', async () => {
    mockedAxios.get.mockRejectedValue({ message: 'Server error' });
    render(<MonitoringComponent />);

    await waitFor(() => {
      expect(screen.getByText(/Failed to fetch monitoring data/)).toBeInTheDocument();
    });
  });
});
