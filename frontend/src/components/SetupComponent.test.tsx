import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import SetupComponent from './SetupComponent';
import { apiClient } from '../api-client';
import { vi } from 'vitest';

vi.mock('../api-client', () => ({
  apiClient: {
    GET: vi.fn(),
    POST: vi.fn(),
  }
}));

const mockedApiClient = apiClient as any;

describe('SetupComponent', () => {
  const mockSettings = {
    SecondsBetweenCaptures: 300,
    OffsetWithinHour: 0,
    PhotoResolutionWidth: 3280,
    PhotoResolutionHeight: 2464,
    PreviewResolutionWidth: 640,
    PreviewResolutionHeight: 480,
    RotateBy: 0,
    ResolutionSetting: 0,
    Quality: 100,
    DebugEnabled: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders without crashing', () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    render(<SetupComponent />);
  });

  it('fetches configuration on mount', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    render(<SetupComponent />);

    await waitFor(() => {
      expect(mockedApiClient.GET).toHaveBeenCalledWith('/configuration');
    });
  });

  it('renders Save button', () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    render(<SetupComponent />);
    
    expect(screen.getByText('Save')).toBeInTheDocument();
  });

  it('renders Restart button', () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    render(<SetupComponent />);
    
    expect(screen.getByText('Restart')).toBeInTheDocument();
  });

  it('renders Shutdown button', () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    render(<SetupComponent />);
    
    expect(screen.getByText('Shutdown')).toBeInTheDocument();
  });

  it('displays time between captures field', () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    render(<SetupComponent />);
    
    expect(screen.getByText(/Time between captures/)).toBeInTheDocument();
  });

  it('displays offset field', () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    render(<SetupComponent />);
    
    expect(screen.getByText(/Delay within the hour/)).toBeInTheDocument();
  });

  it('displays photo resolution field', () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    render(<SetupComponent />);
    
    expect(screen.getByText(/Photo Resolution/)).toBeInTheDocument();
  });

  it('displays rotation field', () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    render(<SetupComponent />);
    
    expect(screen.getByText(/Rotation/)).toBeInTheDocument();
  });

  it('displays quality field', () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    render(<SetupComponent />);
    
    expect(screen.getByText(/Photo Quality/)).toBeInTheDocument();
  });

  it('calls save endpoint when Save button is clicked', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    mockedApiClient.POST.mockResolvedValue({ data: mockSettings });
    
    render(<SetupComponent />);

    await waitFor(() => {
      expect(mockedApiClient.GET).toHaveBeenCalled();
    });

    const saveButton = screen.getByText('Save');
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(mockedApiClient.POST).toHaveBeenCalledWith(
        '/configuration',
        expect.any(Object)
      );
    });
  });

  it('calls restart endpoint when Restart button is clicked', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    mockedApiClient.GET.mockResolvedValueOnce({ data: mockSettings });
    
    render(<SetupComponent />);

    await waitFor(() => {
      expect(mockedApiClient.GET).toHaveBeenCalled();
    });

    mockedApiClient.GET.mockResolvedValue({ data: {} });
    const restartButton = screen.getByText('Restart');
    fireEvent.click(restartButton);

    await waitFor(() => {
      expect(mockedApiClient.GET).toHaveBeenCalledWith('/admin/{command}', { params: { path: { command: 'restart' } } });
    });
  });

  it('calls shutdown endpoint when Shutdown button is clicked', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    render(<SetupComponent />);

    await waitFor(() => {
      expect(mockedApiClient.GET).toHaveBeenCalled();
    });

    mockedApiClient.GET.mockResolvedValue({ data: {} });
    const shutdownButton = screen.getByText('Shutdown');
    fireEvent.click(shutdownButton);

    await waitFor(() => {
      expect(mockedApiClient.GET).toHaveBeenCalledWith('/admin/{command}', { params: { path: { command: 'shutdown' } } });
    });
  });

  it('renders version information when environment variables are set', () => {
    process.env.REACT_APP_GIT_SHA = 'abc123';
    process.env.REACT_APP_GIT_SHA_ABBREV = 'abc';
    process.env.REACT_APP_COMMIT_TIME = '2024-01-01';

    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    render(<SetupComponent />);

    expect(screen.getByText(/Version:/)).toBeInTheDocument();
  });

  it('handles time between captures input change', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    const { container } = render(<SetupComponent />);

    await waitFor(() => {
      const input = container.querySelector('#tfSecondsBetweenCaptures') as HTMLInputElement;
      expect(input).toBeInTheDocument();
    });

    const input = container.querySelector('#tfSecondsBetweenCaptures') as HTMLInputElement;
    fireEvent.change(input, { target: { value: '10' } });
    
    expect(input.value).toBe('10');
  });

  it('handles offset input change', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    const { container } = render(<SetupComponent />);

    await waitFor(() => {
      const input = container.querySelector('#tfOffset') as HTMLInputElement;
      expect(input).toBeInTheDocument();
    });

    const input = container.querySelector('#tfOffset') as HTMLInputElement;
    fireEvent.change(input, { target: { value: '5' } });
    
    expect(input.value).toBe('5');
  });

  it('handles rotation input change', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    const { container } = render(<SetupComponent />);

    await waitFor(() => {
      const input = container.querySelector('#tfRotation') as HTMLInputElement;
      expect(input).toBeInTheDocument();
    });

    const input = container.querySelector('#tfRotation') as HTMLInputElement;
    fireEvent.change(input, { target: { value: '90' } });
    
    expect(input.value).toBe('90');
  });

  it('handles quality input change', async () => {
    mockedApiClient.GET.mockResolvedValue({ data: mockSettings });
    const { container } = render(<SetupComponent />);

    await waitFor(() => {
      const input = container.querySelector('#tfQuality') as HTMLInputElement;
      expect(input).toBeInTheDocument();
    });

    const input = container.querySelector('#tfQuality') as HTMLInputElement;
    fireEvent.change(input, { target: { value: '85' } });
    
    expect(input.value).toBe('85');
  });

  it('updates form with received configuration data', async () => {
    const customSettings = {
      ...mockSettings,
      SecondsBetweenCaptures: 600,
      OffsetWithinHour: 10,
      RotateBy: 180,
      Quality: 85,
    };

    mockedApiClient.GET.mockResolvedValue({ data: customSettings });
    const { container } = render(<SetupComponent />);

    await waitFor(() => {
      const timeInput = container.querySelector('#tfSecondsBetweenCaptures') as HTMLInputElement;
      expect(timeInput.value).toBe('10'); // 600 seconds / 60 = 10 minutes
      
      const offsetInput = container.querySelector('#tfOffset') as HTMLInputElement;
      expect(offsetInput.value).toBe('10');
      
      const rotationInput = container.querySelector('#tfRotation') as HTMLInputElement;
      expect(rotationInput.value).toBe('180');
      
      const qualityInput = container.querySelector('#tfQuality') as HTMLInputElement;
      expect(qualityInput.value).toBe('85');
    });
  });
});
