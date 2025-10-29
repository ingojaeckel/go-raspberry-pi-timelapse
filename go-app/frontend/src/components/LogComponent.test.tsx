import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import LogComponent from './LogComponent';
import axios from 'axios';
import { vi } from 'vitest';

vi.mock('axios');
const mockedAxios = axios as any;

describe('LogComponent', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders without crashing', () => {
    mockedAxios.get.mockResolvedValue({ data: { Logs: '' } });
    render(<LogComponent />);
  });

  it('fetches logs on mount', async () => {
    mockedAxios.get.mockResolvedValue({ data: { Logs: 'Test logs' } });
    render(<LogComponent />);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/logs'));
    });
  });

  it('renders refresh button', () => {
    mockedAxios.get.mockResolvedValue({ data: { Logs: '' } });
    render(<LogComponent />);
    
    const refreshButton = screen.getByText('Refresh');
    expect(refreshButton).toBeInTheDocument();
  });

  it('calls getLogs when refresh button is clicked', async () => {
    mockedAxios.get.mockResolvedValue({ data: { Logs: '' } });
    render(<LogComponent />);

    const refreshButton = screen.getByText('Refresh');
    fireEvent.click(refreshButton);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalledTimes(2); // Once on mount, once on click
    });
  });

  it('displays logs in a pre element', async () => {
    const testLogs = 'Line 1\nLine 2\nLine 3';
    mockedAxios.get.mockResolvedValue({ data: { Logs: testLogs } });
    
    const { container } = render(<LogComponent />);

    await waitFor(() => {
      const preElement = container.querySelector('pre');
      expect(preElement).toBeInTheDocument();
      expect(preElement?.textContent).toContain('Line 1');
      expect(preElement?.textContent).toContain('Line 2');
      expect(preElement?.textContent).toContain('Line 3');
    });
  });

  it('updates logs when refresh is clicked', async () => {
    const initialLogs = 'Initial logs';
    const updatedLogs = 'Updated logs';
    
    mockedAxios.get.mockResolvedValueOnce({ data: { Logs: initialLogs } });
    render(<LogComponent />);

    await waitFor(() => {
      expect(screen.getByText(initialLogs)).toBeInTheDocument();
    });

    mockedAxios.get.mockResolvedValueOnce({ data: { Logs: updatedLogs } });
    const refreshButton = screen.getByText('Refresh');
    fireEvent.click(refreshButton);

    await waitFor(() => {
      expect(screen.getByText(updatedLogs)).toBeInTheDocument();
    });
  });

  it('uses correct API endpoint', async () => {
    mockedAxios.get.mockResolvedValue({ data: { Logs: '' } });
    render(<LogComponent />);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith('http://localhost:8080/logs');
    });
  });

  it('handles empty logs', async () => {
    mockedAxios.get.mockResolvedValue({ data: { Logs: '' } });
    const { container } = render(<LogComponent />);

    await waitFor(() => {
      const preElement = container.querySelector('pre');
      expect(preElement).toBeInTheDocument();
      expect(preElement?.textContent).toBe('');
    });
  });
});
