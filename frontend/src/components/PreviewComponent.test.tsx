import React from 'react';
import { render } from '@testing-library/react';
import PreviewComponent from './PreviewComponent';

describe('PreviewComponent', () => {
  test('renders preview image', () => {
    const { container } = render(<PreviewComponent />);
    const img = container.querySelector('img[alt="preview"]');
    expect(img).toBeInTheDocument();
  });

  test('preview image has correct source URL', () => {
    const { container } = render(<PreviewComponent />);
    const img = container.querySelector('img[alt="preview"]') as HTMLImageElement;
    expect(img.src).toContain('/capture');
  });

  test('renders tips section', () => {
    const { getByText } = render(<PreviewComponent />);
    expect(getByText(/tips for fine-tuning/i)).toBeInTheDocument();
  });

  test('renders tip about low-latency viewfinder', () => {
    const { getByText } = render(<PreviewComponent />);
    expect(getByText(/low-latency viewfinder/i)).toBeInTheDocument();
  });

  test('renders tip about medium-latency viewfinder', () => {
    const { getByText } = render(<PreviewComponent />);
    expect(getByText(/medium-latency viewfinder/i)).toBeInTheDocument();
  });

  test('renders VLC link', () => {
    const { container } = render(<PreviewComponent />);
    const vlcLink = container.querySelector('a[href="https://www.videolan.org/"]');
    expect(vlcLink).toBeInTheDocument();
  });

  test('renders command examples', () => {
    const { container } = render(<PreviewComponent />);
    const preElements = container.querySelectorAll('pre');
    expect(preElements.length).toBeGreaterThan(0);
  });

  test('renders VLC command example', () => {
    const { getByText } = render(<PreviewComponent />);
    expect(getByText(/vlc udp/i)).toBeInTheDocument();
  });

  test('renders raspivid command example', () => {
    const { getByText } = render(<PreviewComponent />);
    expect(getByText(/raspivid/i)).toBeInTheDocument();
  });

  test('renders Adafruit tutorial links', () => {
    const { container } = render(<PreviewComponent />);
    const links = container.querySelectorAll('a[href*="adafruit"]');
    expect(links.length).toBeGreaterThan(0);
  });

  test('renders tutorial link to learn.adafruit.com', () => {
    const { container } = render(<PreviewComponent />);
    const adafruitLink = container.querySelector('a[href*="learn.adafruit.com"]');
    expect(adafruitLink).toBeInTheDocument();
  });

  test('renders link about power off TFT screen', () => {
    const { container } = render(<PreviewComponent />);
    const powerOffLink = container.querySelector('a[href*="willhaley.com"]');
    expect(powerOffLink).toBeInTheDocument();
  });
});
