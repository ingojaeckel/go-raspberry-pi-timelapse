import { render, screen } from '@testing-library/react';
import { HourDial } from './HourDial';

const hourly = { SecondsBetweenCaptures: 1800, OffsetWithinHour: 0 };
const now = new Date('2026-06-21T23:10:00Z');

describe('HourDial', () => {
  it("counts down to the camera's next capture", () => {
    render(<HourDial schedule={hourly} now={now} next={{ at: new Date(now.getTime() + 20 * 60_000), night: false, zoneOffsetSeconds: 0 }} />);
    expect(screen.getByText('20:00')).toBeInTheDocument();
    expect(screen.getByText('until next capture')).toBeInTheDocument();
  });

  it('says when photos resume after a skipped night instead of counting down', () => {
    render(<HourDial schedule={hourly} now={now} next={{ at: new Date('2026-06-22T04:00:00Z'), night: true, zoneOffsetSeconds: 0 }} />);
    expect(screen.getByText('Night')).toBeInTheDocument();
    expect(screen.getByText(/^photos resume at \d\d:\d\d$/)).toBeInTheDocument();
    expect(screen.queryByText('until next capture')).not.toBeInTheDocument();
  });

  it('counts down for captures every N minutes from start-up', () => {
    render(
      <HourDial
        schedule={{ SecondsBetweenCaptures: 600, OffsetWithinHour: -1 }}
        now={now}
        next={{ at: new Date(now.getTime() + 4 * 60_000), night: false, zoneOffsetSeconds: 0 }}
      />,
    );
    expect(screen.getByText('4:00')).toBeInTheDocument();
  });

  it("places the hand in the camera's time zone", () => {
    // 23:10 UTC is 04:40 in India (+5:30): the hand points at 40 minutes, not 10.
    const { container } = render(
      <HourDial schedule={hourly} now={now} next={{ at: new Date(now.getTime() + 60_000), night: false, zoneOffsetSeconds: 19800 }} />,
    );
    const hand = container.querySelector('.dial-hand')!;
    expect(Number(hand.getAttribute('x2'))).toBeLessThan(100); // left half of the face: 30..60 minutes
  });

  it('shows no countdown before the camera has planned a capture', () => {
    render(<HourDial schedule={hourly} now={now} next={null} />);
    expect(screen.queryByText('until next capture')).not.toBeInTheDocument();
  });
});
