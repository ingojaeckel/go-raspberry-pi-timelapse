import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import MonitoringComponent from './MonitoringComponent';
import axios from 'axios';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('MonitoringComponent', () => {
  beforeEach(() => {
    jest.clearAllMocks();
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

  it('displays current time label', () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    expect(screen.getByText(/Current Time:/)).toBeInTheDocument();
  });

  it('displays uptime label', () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    expect(screen.getByText(/Uptime:/)).toBeInTheDocument();
  });

  it('displays free disk space label', () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    expect(screen.getByText(/Free Disk Space:/)).toBeInTheDocument();
  });

  it('displays temperature section', () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    expect(screen.getByText(/Temperature:/)).toBeInTheDocument();
  });

  it('displays GPU temperature label', () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
    render(<MonitoringComponent />);
    
    expect(screen.getByText(/GPU:/)).toBeInTheDocument();
  });

  it('displays CPU temperature label', () => {
    mockedAxios.get.mockResolvedValue({ data: {} });
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
});
