import React from 'react';
import { render, wait } from '@testing-library/react';
import axios from 'axios';
import MonitoringComponent from './MonitoringComponent';
import { MonitoringResponse } from '../models/response';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('MonitoringComponent', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders monitoring data labels', () => {
    mockedAxios.get.mockResolvedValue({
      data: {
        Time: '',
        Uptime: '',
        CpuTemperature: '',
        GpuTemperature: '',
        FreeDiskSpace: ''
      }
    });
    
    const { getByText } = render(<MonitoringComponent />);
    expect(getByText(/current time/i)).toBeInTheDocument();
    expect(getByText(/uptime/i)).toBeInTheDocument();
    expect(getByText(/free disk space/i)).toBeInTheDocument();
    expect(getByText(/temperature/i)).toBeInTheDocument();
    expect(getByText(/gpu/i)).toBeInTheDocument();
    expect(getByText(/cpu/i)).toBeInTheDocument();
  });

  test('fetches monitoring data on mount', async () => {
    const mockData: MonitoringResponse = {
      Time: '2023-01-01T10:00:00Z',
      Uptime: '5 days',
      CpuTemperature: '45.5°C',
      GpuTemperature: '42.0°C',
      FreeDiskSpace: '10GB'
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockData });
    
    render(<MonitoringComponent />);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/monitoring'));
    });
  });

  test('displays monitoring data', async () => {
    const mockData: MonitoringResponse = {
      Time: '2023-01-01T10:00:00Z',
      Uptime: '5 days',
      CpuTemperature: '45.5°C',
      GpuTemperature: '42.0°C',
      FreeDiskSpace: '10GB'
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockData });
    
    const { getByText } = render(<MonitoringComponent />);
    
    await wait(() => {
      expect(getByText('2023-01-01T10:00:00Z')).toBeInTheDocument();
      expect(getByText('5 days')).toBeInTheDocument();
      expect(getByText('45.5°C')).toBeInTheDocument();
      expect(getByText('42.0°C')).toBeInTheDocument();
      expect(getByText('10GB')).toBeInTheDocument();
    });
  });

  test('handles empty monitoring data', async () => {
    const mockData: MonitoringResponse = {
      Time: '',
      Uptime: '',
      CpuTemperature: '',
      GpuTemperature: '',
      FreeDiskSpace: ''
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockData });
    
    const { container } = render(<MonitoringComponent />);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalled();
    });
    
    // Component should still render without errors
    expect(container.querySelector('ul')).toBeInTheDocument();
  });
});
