import React from 'react';
import { render, fireEvent, wait } from '@testing-library/react';
import axios from 'axios';
import LogComponent from './LogComponent';
import { LogResponse } from '../models/response';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('LogComponent', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders refresh button', () => {
    mockedAxios.get.mockResolvedValue({ data: { Logs: '' } });
    const { getByText } = render(<LogComponent />);
    expect(getByText(/refresh/i)).toBeInTheDocument();
  });

  test('fetches logs on mount', async () => {
    const mockLogs: LogResponse = {
      Logs: 'Sample log line 1\nSample log line 2'
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockLogs });
    
    render(<LogComponent />);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/logs'));
    });
  });

  test('displays logs', async () => {
    const mockLogs: LogResponse = {
      Logs: 'Sample log line 1\nSample log line 2'
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockLogs });
    
    const { getByText } = render(<LogComponent />);
    
    await wait(() => {
      expect(getByText(/sample log line 1/i)).toBeInTheDocument();
      expect(getByText(/sample log line 2/i)).toBeInTheDocument();
    });
  });

  test('handles refresh button click', async () => {
    const mockLogs: LogResponse = {
      Logs: 'Initial logs'
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockLogs });
    
    const { getByText } = render(<LogComponent />);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalledTimes(1);
    });
    
    const refreshButton = getByText(/refresh/i);
    fireEvent.click(refreshButton);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalledTimes(2);
    });
  });

  test('updates logs when refresh is clicked', async () => {
    const initialLogs: LogResponse = {
      Logs: 'Initial logs'
    };
    
    const updatedLogs: LogResponse = {
      Logs: 'Updated logs'
    };
    
    mockedAxios.get
      .mockResolvedValueOnce({ data: initialLogs })
      .mockResolvedValueOnce({ data: updatedLogs });
    
    const { getByText } = render(<LogComponent />);
    
    await wait(() => {
      expect(getByText('Initial logs')).toBeInTheDocument();
    });
    
    const refreshButton = getByText(/refresh/i);
    fireEvent.click(refreshButton);
    
    await wait(() => {
      expect(getByText('Updated logs')).toBeInTheDocument();
    });
  });

  test('handles empty logs', async () => {
    const mockLogs: LogResponse = {
      Logs: ''
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockLogs });
    
    const { container } = render(<LogComponent />);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalled();
    });
    
    const preElement = container.querySelector('pre');
    expect(preElement).toBeInTheDocument();
    expect(preElement?.textContent).toBe('');
  });

  test('preserves log formatting with pre tag', async () => {
    const mockLogs: LogResponse = {
      Logs: 'Line 1\n  Indented line\n    More indent'
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockLogs });
    
    const { container } = render(<LogComponent />);
    
    await wait(() => {
      const preElement = container.querySelector('pre');
      expect(preElement?.textContent).toContain('Line 1');
      expect(preElement?.textContent).toContain('  Indented line');
      expect(preElement?.textContent).toContain('    More indent');
    });
  });
});
