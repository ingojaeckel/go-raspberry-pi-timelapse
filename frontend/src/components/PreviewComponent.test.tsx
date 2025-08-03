import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import PreviewComponent from './PreviewComponent';
import axios from 'axios';
import { DetectionResponse, SettingsResponse } from '../models/response';

// Mock axios
jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

// Mock canvas context methods for bounding box testing
const mockContext = {
  clearRect: jest.fn(),
  strokeRect: jest.fn(),
  fillRect: jest.fn(),
  fillText: jest.fn(),
  measureText: jest.fn(),
  beginPath: jest.fn(),
  moveTo: jest.fn(),
  lineTo: jest.fn(),
  stroke: jest.fn(),
  fill: jest.fn(),
  strokeStyle: '',
  fillStyle: '',
  lineWidth: 0,
  font: '',
  globalAlpha: 0,
};

const mockGetContext = jest.fn(() => mockContext);

// Mock HTMLCanvasElement
Object.defineProperty(HTMLCanvasElement.prototype, 'getContext', {
  value: mockGetContext,
});

// Mock HTMLImageElement properties
Object.defineProperty(HTMLImageElement.prototype, 'naturalWidth', {
  get: () => 800,
  configurable: true,
});

Object.defineProperty(HTMLImageElement.prototype, 'naturalHeight', {
  get: () => 600,
  configurable: true,
});

describe('PreviewComponent drawBoundingBoxes Tests', () => {
  const mockSettings: SettingsResponse = {
    SecondsBetweenCaptures: 60,
    OffsetWithinHour: 0,
    PhotoResolutionWidth: 1920,
    PhotoResolutionHeight: 1080,
    PreviewResolutionWidth: 800,
    PreviewResolutionHeight: 600,
    RotateBy: 0,
    ResolutionSetting: 1,
    Quality: 85,
    DebugEnabled: false,
    ObjectDetectionEnabled: true,
  };

  const mockDetection: DetectionResponse = {
    IsDay: true,
    Objects: ['human', 'vehicle', 'animal'],
    Summary: "It's day time. The photo includes: human, vehicle, animal",
    PhotoPath: '/test/photo.jpg',
    LatencyMs: 1247,
    OverallConfidence: 0.784,
    Details: [
      {
        class: 'person',
        confidence: 0.89,
        category: 'human',
        bbox: { x: 100, y: 50, width: 150, height: 300 }
      },
      {
        class: 'car',
        confidence: 0.75,
        category: 'vehicle',
        bbox: { x: 300, y: 200, width: 200, height: 100 }
      },
      {
        class: 'bird',
        confidence: 0.65,
        category: 'animal',
        bbox: { x: 50, y: 100, width: 80, height: 60 }
      }
    ]
  };

  beforeEach(() => {
    jest.clearAllMocks();
    mockContext.measureText.mockReturnValue({ width: 120 } as TextMetrics);
    
    // Setup default axios mock responses
    mockedAxios.get.mockImplementation((url) => {
      if (url.includes('/configuration')) {
        return Promise.resolve({ data: mockSettings });
      }
      if (url.includes('/detection')) {
        return Promise.resolve({ data: mockDetection });
      }
      return Promise.reject(new Error('Not found'));
    });
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe('drawBoundingBoxes functionality', () => {
    test('should setup canvas dimensions correctly when image loads', (done) => {
      render(<PreviewComponent />);

      // Simulate image loading after component mounts
      setTimeout(() => {
        const img = screen.getByAltText('preview');
        fireEvent.load(img);

        // Wait a bit for effects to run
        setTimeout(() => {
          const canvas = document.querySelector('canvas');
          expect(canvas).toBeInTheDocument();
          expect(canvas?.width).toBe(800);
          expect(canvas?.height).toBe(600);
          done();
        }, 100);
      }, 100);
    });

    test('should clear canvas before drawing bounding boxes', (done) => {
      render(<PreviewComponent />);

      setTimeout(() => {
        const img = screen.getByAltText('preview');
        fireEvent.load(img);

        setTimeout(() => {
          // Verify clearRect is called with correct dimensions
          expect(mockContext.clearRect).toHaveBeenCalledWith(0, 0, 800, 600);
          done();
        }, 100);
      }, 100);
    });

    test('should draw bounding boxes for each detected object', (done) => {
      render(<PreviewComponent />);

      setTimeout(() => {
        const img = screen.getByAltText('preview');
        fireEvent.load(img);

        setTimeout(() => {
          // Verify bounding boxes are drawn for each object
          expect(mockContext.strokeRect).toHaveBeenCalledTimes(3);
          
          // Check specific bounding box coordinates
          expect(mockContext.strokeRect).toHaveBeenCalledWith(100, 50, 150, 300); // person
          expect(mockContext.strokeRect).toHaveBeenCalledWith(300, 200, 200, 100); // car
          expect(mockContext.strokeRect).toHaveBeenCalledWith(50, 100, 80, 60); // bird
          done();
        }, 100);
      }, 100);
    });

    test('should draw labels with confidence percentages', (done) => {
      render(<PreviewComponent />);

      setTimeout(() => {
        const img = screen.getByAltText('preview');
        fireEvent.load(img);

        setTimeout(() => {
          // Verify labels are drawn with correct text
          expect(mockContext.fillText).toHaveBeenCalledTimes(3);
          expect(mockContext.fillText).toHaveBeenCalledWith('person (89.0%)', 104, 45);
          expect(mockContext.fillText).toHaveBeenCalledWith('car (75.0%)', 304, 195);
          expect(mockContext.fillText).toHaveBeenCalledWith('bird (65.0%)', 54, 95);
          done();
        }, 100);
      }, 100);
    });

    test('should draw label backgrounds for each object', (done) => {
      render(<PreviewComponent />);

      setTimeout(() => {
        const img = screen.getByAltText('preview');
        fireEvent.load(img);

        setTimeout(() => {
          // Verify label background rectangles are drawn
          expect(mockContext.fillRect).toHaveBeenCalledTimes(3);
          
          // Check that backgrounds are drawn above the bounding boxes
          expect(mockContext.fillRect).toHaveBeenCalledWith(100, 30, 128, 20); // person label background
          expect(mockContext.fillRect).toHaveBeenCalledWith(300, 180, 128, 20); // car label background
          expect(mockContext.fillRect).toHaveBeenCalledWith(50, 80, 128, 20); // bird label background
          done();
        }, 100);
      }, 100);
    });

    test('should handle objects without bounding boxes gracefully', (done) => {
      const detectionWithoutBbox: DetectionResponse = {
        ...mockDetection,
        Details: [
          {
            class: 'unknown',
            confidence: 0.5,
            category: 'other',
            // No bbox property
          }
        ]
      };

      mockedAxios.get.mockImplementation((url) => {
        if (url.includes('/configuration')) {
          return Promise.resolve({ data: mockSettings });
        }
        if (url.includes('/detection')) {
          return Promise.resolve({ data: detectionWithoutBbox });
        }
        return Promise.reject(new Error('Not found'));
      });

      render(<PreviewComponent />);

      setTimeout(() => {
        const img = screen.getByAltText('preview');
        fireEvent.load(img);

        setTimeout(() => {
          // Should clear canvas but not draw any bounding boxes
          expect(mockContext.clearRect).toHaveBeenCalledWith(0, 0, 800, 600);
          expect(mockContext.strokeRect).not.toHaveBeenCalled();
          expect(mockContext.fillText).not.toHaveBeenCalled();
          done();
        }, 100);
      }, 100);
    });

    test('should handle empty detection details array', (done) => {
      const emptyDetection: DetectionResponse = {
        ...mockDetection,
        Details: []
      };

      mockedAxios.get.mockImplementation((url) => {
        if (url.includes('/configuration')) {
          return Promise.resolve({ data: mockSettings });
        }
        if (url.includes('/detection')) {
          return Promise.resolve({ data: emptyDetection });
        }
        return Promise.reject(new Error('Not found'));
      });

      render(<PreviewComponent />);

      setTimeout(() => {
        const img = screen.getByAltText('preview');
        fireEvent.load(img);

        setTimeout(() => {
          // Should clear canvas but not draw anything
          expect(mockContext.clearRect).toHaveBeenCalledWith(0, 0, 800, 600);
          expect(mockContext.strokeRect).not.toHaveBeenCalled();
          done();
        }, 100);
      }, 100);
    });

    test('should apply correct font and canvas properties', (done) => {
      render(<PreviewComponent />);

      setTimeout(() => {
        const img = screen.getByAltText('preview');
        fireEvent.load(img);

        setTimeout(() => {
          // Verify font and alpha settings are applied
          expect(mockContext.font).toBe('16px Arial');
          expect(mockContext.lineWidth).toBe(3);
          expect(mockContext.globalAlpha).toBe(1.0); // Final alpha setting for text
          done();
        }, 100);
      }, 100);
    });

    test('should not render canvas when object detection is disabled', (done) => {
      const settingsWithDetectionDisabled: SettingsResponse = {
        ...mockSettings,
        ObjectDetectionEnabled: false
      };

      mockedAxios.get.mockImplementation((url) => {
        if (url.includes('/configuration')) {
          return Promise.resolve({ data: settingsWithDetectionDisabled });
        }
        return Promise.reject(new Error('Not found'));
      });

      render(<PreviewComponent />);

      setTimeout(() => {
        const img = screen.getByAltText('preview');
        fireEvent.load(img);

        // Canvas should not be rendered when object detection is disabled
        const canvas = document.querySelector('canvas');
        expect(canvas).not.toBeInTheDocument();
        done();
      }, 100);
    });

    test('should render object detection results with metrics', (done) => {
      render(<PreviewComponent />);

      setTimeout(() => {
        expect(screen.getByText(/Detection time:/)).toBeInTheDocument();
        expect(screen.getByText(/1247ms/)).toBeInTheDocument();
        expect(screen.getByText(/Confidence:/)).toBeInTheDocument();
        expect(screen.getByText(/78.4%/)).toBeInTheDocument();
        done();
      }, 100);
    });
  });
});