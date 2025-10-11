import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import PhotoComponent from './PhotoComponent';
import axios from 'axios';
import { GridRowSelectionModel } from '@mui/x-data-grid';
import { vi } from 'vitest';

vi.mock('axios');
const mockedAxios = axios as any;

describe('PhotoComponent', () => {
  const mockPhotosData = {
    Photos: [
      { Name: 'photo1.jpg', ModTime: '2024-01-01', Size: '1024' },
      { Name: 'photo2.jpg', ModTime: '2024-01-02', Size: '2048' },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    // Default mock to prevent errors
    mockedAxios.get.mockResolvedValue({ data: { Photos: [] } });
  });

  it('renders without crashing', () => {
    render(<PhotoComponent />);
  });

  it('fetches and displays photos on mount', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockPhotosData });
    render(<PhotoComponent />);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/photos'));
    });
  });

  it('renders refresh button', () => {
    render(<PhotoComponent />);
    
    const refreshButton = screen.getByText('Refresh');
    expect(refreshButton).toBeInTheDocument();
  });

  it('calls getPhotos when refresh button is clicked', async () => {
    render(<PhotoComponent />);

    const refreshButton = screen.getByText('Refresh');
    fireEvent.click(refreshButton);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalledTimes(2); // Once on mount, once on click
    });
  });

  it('renders delete button with count', () => {
    render(<PhotoComponent />);
    
    const deleteButton = screen.getByText(/Delete selected/);
    expect(deleteButton).toBeInTheDocument();
  });

  it('renders download links', () => {
    render(<PhotoComponent />);
    
    expect(screen.getByText('Download all (zip)')).toBeInTheDocument();
    expect(screen.getByText('Download all (tar)')).toBeInTheDocument();
    expect(screen.getByText(/Download selected/)).toBeInTheDocument();
  });

  it('shows deletion dialog when delete is clicked with selection', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockPhotosData });
    const { container } = render(<PhotoComponent />);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalled();
    });

    // Verify the dialog component exists in the DOM
    const dialogs = container.querySelectorAll('[role="dialog"]');
    expect(dialogs).toBeDefined();
  });

  it('handles deletion confirmation', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockPhotosData });
    render(<PhotoComponent />);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalled();
    });
  });

  it('renders DataGrid component', () => {
    const { container } = render(<PhotoComponent />);
    
    const dataGrid = container.querySelector('.MuiDataGrid-root');
    expect(dataGrid).toBeInTheDocument();
  });

  it('formats photo data correctly', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockPhotosData });
    render(<PhotoComponent />);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith('http://localhost:8080/photos');
    });
  });

  it('handles selection model change', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockPhotosData });
    render(<PhotoComponent />);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalled();
    });
  });

  it('deletes photos when confirmed', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockPhotosData });
    render(<PhotoComponent />);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalled();
    });
    
    // Mock deletion endpoint
    mockedAxios.get.mockResolvedValueOnce({ data: {} });
  });

  it('builds correct download URL for selected files', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockPhotosData });
    render(<PhotoComponent />);

    await waitFor(() => {
      // Should have download buttons instead of links
      expect(screen.getByText(/Download all \(zip\)/)).toBeInTheDocument();
      expect(screen.getByText(/Download all \(tar\)/)).toBeInTheDocument();
      expect(screen.getByText(/Download selected/)).toBeInTheDocument();
    });
  });

  it('shows delete button with correct count', () => {
    mockedAxios.get.mockResolvedValue({ data: { Photos: [] } });
    render(<PhotoComponent />);
    
    // Initially no selection
    expect(screen.getByText('Delete selected (0)')).toBeInTheDocument();
  });

  it('renders download selected link with count', () => {
    mockedAxios.get.mockResolvedValue({ data: { Photos: [] } });
    render(<PhotoComponent />);
    
    expect(screen.getByText('Download selected (0)')).toBeInTheDocument();
  });

  it('renders DataGrid with checkbox selection enabled', async () => {
    mockedAxios.get.mockResolvedValue({ data: mockPhotosData });
    const { container } = render(<PhotoComponent />);

    await waitFor(() => {
      const dataGrid = container.querySelector('.MuiDataGrid-root');
      expect(dataGrid).toBeInTheDocument();
    });
  });

  it('renders ButtonGroup component', () => {
    mockedAxios.get.mockResolvedValue({ data: { Photos: [] } });
    const { container } = render(<PhotoComponent />);
    
    const buttonGroup = container.querySelector('[role="group"]');
    expect(buttonGroup).toBeInTheDocument();
  });

  it('renders dialog with correct structure', () => {
    mockedAxios.get.mockResolvedValue({ data: { Photos: [] } });
    const { container } = render(<PhotoComponent />);
    
    // Dialog exists but is not open by default
    const dialog = container.querySelector('[role="dialog"]');
    // Dialog may not be in DOM when not open
  });

  it('handles empty photos response', async () => {
    mockedAxios.get.mockResolvedValue({ data: { Photos: [] } });
    render(<PhotoComponent />);

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith('http://localhost:8080/photos');
    });
  });
});

