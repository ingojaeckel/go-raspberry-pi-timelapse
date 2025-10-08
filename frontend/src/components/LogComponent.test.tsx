import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import LogComponent from './LogComponent';
import { apiClient } from '../api-client';
import { vi } from 'vitest';

vi.mock('../api-client', () => ({
  apiClient: {
    GET: vi.fn(),
  }
}));

const mockedApiClient = apiClient as any;

describe('LogComponent', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders without crashing', () => {
    mockedApiClient.GET.mockResolvedValue({ data: { Logs: '' } });
    render(<LogComponent />);
  });

  it('fetches logs on mount', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: { Logs: 'Test logs' } });
    render(<LogComponent />);

    await waitFor(() => {
      expect(mockedApiClient.GET).toHaveBeenCalledWith('/logs');
    });
  });

  it('renders refresh button', () => {
    mockedApiClient.GET.mockResolvedValue({ data: { Logs: '' } });
    render(<LogComponent />);
    
    const refreshButton = screen.getByText('Refresh');
    expect(refreshButton).toBeInTheDocument();
  });

  it('calls getLogs when refresh button is clicked', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: { Logs: '' } });
    render(<LogComponent />);

    const refreshButton = screen.getByText('Refresh');
    fireEvent.click(refreshButton);

    await waitFor(() => {
      expect(mockedApiClient.GET).toHaveBeenCalledTimes(2); // Once on mount, once on click
    });
  });

  it('displays logs in a pre element', async () => {
    const testLogs = 'Line 1\nLine 2\nLine 3';
    mockedApiClient.GET.mockResolvedValue({ data: { Logs: testLogs } });
    
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
    
    mockedApiClient.GET.mockResolvedValueOnce({ data: { Logs: initialLogs } });
    render(<LogComponent />);

    await waitFor(() => {
      expect(screen.getByText(initialLogs)).toBeInTheDocument();
    });

    mockedApiClient.GET.mockResolvedValueOnce({ data: { Logs: updatedLogs } });
    const refreshButton = screen.getByText('Refresh');
    fireEvent.click(refreshButton);

    await waitFor(() => {
      expect(screen.getByText(updatedLogs)).toBeInTheDocument();
    });
  });

  it('uses correct API endpoint', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: { Logs: '' } });
    render(<LogComponent />);

    await waitFor(() => {
      expect(mockedApiClient.GET).toHaveBeenCalledWith('/logs');
    });
  });

  it('handles empty logs', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: { Logs: '' } });
    const { container } = render(<LogComponent />);

    await waitFor(() => {
      const preElement = container.querySelector('pre');
      expect(preElement).toBeInTheDocument();
      expect(preElement?.textContent).toBe('');
    });
  });
});
