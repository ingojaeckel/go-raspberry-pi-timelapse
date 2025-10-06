import React from 'react';
import { render, fireEvent, wait, screen } from '@testing-library/react';
import axios from 'axios';
import PhotoComponent from './PhotoComponent';
import { PhotosResponse } from '../models/response';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('PhotoComponent', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders refresh button', () => {
    mockedAxios.get.mockResolvedValue({ data: { Photos: [] } });
    const { getByText } = render(<PhotoComponent />);
    expect(getByText(/refresh/i)).toBeInTheDocument();
  });

  test('renders delete selected button', () => {
    mockedAxios.get.mockResolvedValue({ data: { Photos: [] } });
    const { getByText } = render(<PhotoComponent />);
    expect(getByText(/delete selected/i)).toBeInTheDocument();
  });

  test('renders download links', () => {
    mockedAxios.get.mockResolvedValue({ data: { Photos: [] } });
    const { getByText } = render(<PhotoComponent />);
    expect(getByText(/download all \(zip\)/i)).toBeInTheDocument();
    expect(getByText(/download all \(tar\)/i)).toBeInTheDocument();
  });

  test('fetches photos on mount', async () => {
    const mockPhotos: PhotosResponse = {
      Photos: [
        { Name: 'photo1.jpg', ModTime: '2023-01-01', Size: '1024' },
        { Name: 'photo2.jpg', ModTime: '2023-01-02', Size: '2048' }
      ]
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockPhotos });
    
    render(<PhotoComponent />);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/photos'));
    });
  });

  test('displays photos in data grid', async () => {
    const mockPhotos: PhotosResponse = {
      Photos: [
        { Name: 'photo1.jpg', ModTime: '2023-01-01T10:00:00', Size: '1024' }
      ]
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockPhotos });
    
    const { getByText } = render(<PhotoComponent />);
    
    await wait(() => {
      expect(getByText('photo1.jpg')).toBeInTheDocument();
    });
  });

  test('handles refresh button click', async () => {
    mockedAxios.get.mockResolvedValue({ data: { Photos: [] } });
    
    const { getByText } = render(<PhotoComponent />);
    
    const refreshButton = getByText(/refresh/i);
    fireEvent.click(refreshButton);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalledTimes(2); // Once on mount, once on click
    });
  });

  test('shows deletion dialog when delete button clicked with selection', async () => {
    const mockPhotos: PhotosResponse = {
      Photos: [
        { Name: 'photo1.jpg', ModTime: '2023-01-01', Size: '1024' }
      ]
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockPhotos });
    
    render(<PhotoComponent />);
    
    await wait(() => {
      expect(screen.getByText('photo1.jpg')).toBeInTheDocument();
    });
    
    // Select a photo by clicking checkbox
    const checkbox = screen.getAllByRole('checkbox')[1]; // First is header checkbox
    fireEvent.click(checkbox);
    
    // Click delete button
    const deleteButton = screen.getByText(/delete selected/i);
    fireEvent.click(deleteButton);
    
    // Check for dialog
    await wait(() => {
      expect(screen.getByText(/are you sure/i)).toBeInTheDocument();
    });
  });

  test('handles deletion confirmation', async () => {
    const mockPhotos: PhotosResponse = {
      Photos: [
        { Name: 'photo1.jpg', ModTime: '2023-01-01', Size: '1024' }
      ]
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockPhotos });
    
    render(<PhotoComponent />);
    
    await wait(() => {
      expect(screen.getByText('photo1.jpg')).toBeInTheDocument();
    });
    
    // Select and delete
    const checkbox = screen.getAllByRole('checkbox')[1];
    fireEvent.click(checkbox);
    
    const deleteButton = screen.getByText(/delete selected/i);
    fireEvent.click(deleteButton);
    
    await wait(() => {
      expect(screen.getByText(/are you sure/i)).toBeInTheDocument();
    });
    
    // Confirm deletion
    const confirmButton = screen.getByText('Delete');
    fireEvent.click(confirmButton);
    
    await wait(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/file/delete'));
    });
  });

  test('handles deletion cancellation', async () => {
    const mockPhotos: PhotosResponse = {
      Photos: [
        { Name: 'photo1.jpg', ModTime: '2023-01-01', Size: '1024' }
      ]
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockPhotos });
    
    render(<PhotoComponent />);
    
    await wait(() => {
      expect(screen.getByText('photo1.jpg')).toBeInTheDocument();
    });
    
    // Select and try to delete
    const checkbox = screen.getAllByRole('checkbox')[1];
    fireEvent.click(checkbox);
    
    const deleteButton = screen.getByText(/delete selected/i);
    fireEvent.click(deleteButton);
    
    await wait(() => {
      expect(screen.getByText(/are you sure/i)).toBeInTheDocument();
    });
    
    // Cancel deletion
    const cancelButton = screen.getByText('Cancel');
    fireEvent.click(cancelButton);
    
    await wait(() => {
      expect(screen.queryByText(/are you sure/i)).not.toBeInTheDocument();
    });
  });

  test('builds correct download URL for selected files', async () => {
    const mockPhotos: PhotosResponse = {
      Photos: [
        { Name: 'photo1.jpg', ModTime: '2023-01-01', Size: '1024' },
        { Name: 'photo2.jpg', ModTime: '2023-01-02', Size: '2048' }
      ]
    };
    
    mockedAxios.get.mockResolvedValue({ data: mockPhotos });
    
    const { container } = render(<PhotoComponent />);
    
    await wait(() => {
      expect(screen.getByText('photo1.jpg')).toBeInTheDocument();
    });
    
    // Select both photos
    const checkboxes = screen.getAllByRole('checkbox');
    fireEvent.click(checkboxes[1]); // photo1
    fireEvent.click(checkboxes[2]); // photo2
    
    // Find download link for selected files
    const downloadLink = container.querySelector('a[href*="f=photo1.jpg"]');
    expect(downloadLink).toBeInTheDocument();
  });
});
