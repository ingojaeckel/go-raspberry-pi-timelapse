import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import PreviewComponent from './PreviewComponent';

describe('PreviewComponent', () => {
  it('renders without crashing', () => {
    render(<PreviewComponent />);
  });

  it('renders preview image', () => {
    const { container } = render(<PreviewComponent />);
    const img = container.querySelector('img');
    
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute('src', 'http://localhost:8080/capture');
    expect(img).toHaveAttribute('alt', 'preview');
  });

  it('displays tips section', () => {
    render(<PreviewComponent />);
    
    expect(screen.getByText(/Tips for fine-tuning the camera position/)).toBeInTheDocument();
  });

  it('contains TFT tutorial link', () => {
    const { container } = render(<PreviewComponent />);
    const links = container.querySelectorAll('a');
    
    const tftLink = Array.from(links).find(link => 
      link.getAttribute('href')?.includes('adafruit-pitft-28-inch-resistive-touchscreen-display-raspberry-pi')
    );
    
    expect(tftLink).toBeInTheDocument();
  });

  it('contains performance tuning link', () => {
    const { container } = render(<PreviewComponent />);
    const links = container.querySelectorAll('a');
    
    const perfLink = Array.from(links).find(link => 
      link.getAttribute('href')?.includes('running-opengl-based-games-and-emulators-on-adafruit-pitft-displays')
    );
    
    expect(perfLink).toBeInTheDocument();
  });

  it('contains power-off link', () => {
    const { container } = render(<PreviewComponent />);
    const links = container.querySelectorAll('a');
    
    const powerLink = Array.from(links).find(link => 
      link.getAttribute('href')?.includes('power-off-raspberry-pi-adafruit-tft-screen-shutdown')
    );
    
    expect(powerLink).toBeInTheDocument();
  });

  it('mentions VLC for streaming', () => {
    const { container } = render(<PreviewComponent />);
    const links = container.querySelectorAll('a');
    
    const vlcLink = Array.from(links).find(link => 
      link.getAttribute('href')?.includes('videolan.org')
    );
    
    expect(vlcLink).toBeInTheDocument();
  });

  it('displays VLC command examples', () => {
    render(<PreviewComponent />);
    
    expect(screen.getByText(/vlc udp/)).toBeInTheDocument();
    expect(screen.getByText(/raspivid -t 60000/)).toBeInTheDocument();
  });

  it('contains two main tip items', () => {
    const { container } = render(<PreviewComponent />);
    const listItems = container.querySelectorAll('ul > li');
    
    expect(listItems.length).toBeGreaterThanOrEqual(2);
  });

  it('displays error message when image fails to load', () => {
    const { container } = render(<PreviewComponent />);
    const img = container.querySelector('img') as HTMLImageElement;
    
    // Trigger error event
    fireEvent.error(img);
    
    expect(screen.getByText(/Failed to load camera preview/)).toBeInTheDocument();
  });

  it('displays tips header', () => {
    render(<PreviewComponent />);
    
    expect(screen.getByText(/Tips for fine-tuning the camera position/)).toBeInTheDocument();
  });
});
