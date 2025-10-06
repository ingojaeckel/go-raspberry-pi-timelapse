import React from 'react';
import { render, fireEvent, wait, screen } from '@testing-library/react';
import axios from 'axios';
import SetupComponent from './SetupComponent';
import { SettingsResponse } from '../models/response';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('SetupComponent', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  const mockSettings: SettingsResponse = {
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
  };

  test('renders form fields', () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    const { container } = render(<SetupComponent />);
    
    expect(container.querySelector('#tfSecondsBetweenCaptures')).toBeInTheDocument();
    expect(container.querySelector('#tfOffset')).toBeInTheDocument();
    expect(container.querySelector('#tfRotation')).toBeInTheDocument();
    expect(container.querySelector('#tfQuality')).toBeInTheDocument();
  });

  test('renders action buttons', () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    const { getByText } = render(<SetupComponent />);
    
    expect(getByText('Save')).toBeInTheDocument();
    expect(getByText('Restart')).toBeInTheDocument();
    expect(getByText('Shutdown')).toBeInTheDocument();
  });

  test('fetches configuration on mount', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    
    render(<SetupComponent />);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/configuration'));
    });
  });

  test('displays loaded configuration values', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    
    const { container } = render(<SetupComponent />);
    
    await wait(() => {
      const qualityInput = container.querySelector('#tfQuality') as HTMLInputElement;
      expect(qualityInput?.value).toBe('100');
    });
  });

  test('handles save button click', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    mockedAxios.post.mockResolvedValue({ data: {} });
    
    const { getByText } = render(<SetupComponent />);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalled();
    });
    
    const saveButton = getByText('Save');
    fireEvent.click(saveButton);
    
    await wait(() => {
      expect(mockedAxios.post).toHaveBeenCalledWith(
        expect.stringContaining('/configuration'),
        expect.any(Object)
      );
    });
  });

  test('handles restart button click', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    
    const { getByText } = render(<SetupComponent />);
    
    const restartButton = getByText('Restart');
    fireEvent.click(restartButton);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/admin/restart'));
    });
  });

  test('handles shutdown button click', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    
    const { getByText } = render(<SetupComponent />);
    
    const shutdownButton = getByText('Shutdown');
    fireEvent.click(shutdownButton);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/admin/shutdown'));
    });
  });

  test('updates seconds between captures field', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    
    const { container } = render(<SetupComponent />);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalled();
    });
    
    const input = container.querySelector('#tfSecondsBetweenCaptures') as HTMLInputElement;
    fireEvent.change(input, { target: { value: '120' } });
    
    expect(input.value).toBe('120');
  });

  test('updates rotation field', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    
    const { container } = render(<SetupComponent />);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalled();
    });
    
    const input = container.querySelector('#tfRotation') as HTMLInputElement;
    fireEvent.change(input, { target: { value: '90' } });
    
    expect(input.value).toBe('90');
  });

  test('displays version information when available', () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    
    // Set environment variables
    process.env.REACT_APP_GIT_SHA = 'abc123def456';
    process.env.REACT_APP_GIT_SHA_ABBREV = 'abc123';
    process.env.REACT_APP_COMMIT_TIME = '2023-01-01';
    
    const { getByText } = render(<SetupComponent />);
    
    expect(getByText(/version/i)).toBeInTheDocument();
  });

  test('renders labels', () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    
    const { getByText } = render(<SetupComponent />);
    
    expect(getByText(/time between captures/i)).toBeInTheDocument();
    expect(getByText(/quality/i)).toBeInTheDocument();
  });

  test('updates quality field', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    
    const { container } = render(<SetupComponent />);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalled();
    });
    
    const input = container.querySelector('#tfQuality') as HTMLInputElement;
    fireEvent.change(input, { target: { value: '85' } });
    
    expect(input.value).toBe('85');
  });

  test('all numeric fields accept only numbers', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockSettings });
    
    const { container } = render(<SetupComponent />);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalled();
    });
    
    const numericFields = [
      '#tfSecondsBetweenCaptures',
      '#tfOffset',
      '#tfRotation',
      '#tfQuality'
    ];
    
    numericFields.forEach(selector => {
      const input = container.querySelector(selector) as HTMLInputElement;
      expect(input?.type).toBe('number');
    });
  });
});
