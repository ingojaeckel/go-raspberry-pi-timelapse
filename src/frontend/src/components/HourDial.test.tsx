import { render, screen } from '@testing-library/react';
import { HourDial } from './HourDial';

const london = { SecondsBetweenCaptures: 1800, OffsetWithinHour: 0, SkipNightPhotos: true, Latitude: 51.5074, Longitude: -0.1278 };

describe('HourDial', () => {
  it('counts down to the next capture during the day', () => {
    render(<HourDial schedule={london} now={new Date('2026-06-21T12:10:00Z')} />);
    expect(screen.getByText('until next capture')).toBeInTheDocument();
  });

  it('says when photos resume at night instead of counting down', () => {
    render(<HourDial schedule={london} now={new Date('2026-06-21T23:10:00Z')} />);
    expect(screen.getByText('Night')).toBeInTheDocument();
    expect(screen.getByText(/^photos resume at \d\d:\d\d$/)).toBeInTheDocument();
    expect(screen.queryByText('until next capture')).not.toBeInTheDocument();
  });

  it('counts down at night when night photos are taken', () => {
    render(<HourDial schedule={{ ...london, SkipNightPhotos: false }} now={new Date('2026-06-21T23:10:00Z')} />);
    expect(screen.getByText('until next capture')).toBeInTheDocument();
  });
});
