import { render, screen } from '@testing-library/react';
import { formatClock } from '../format';
import { HourDial } from './HourDial';

const hourly = { SecondsBetweenCaptures: 1800, OffsetWithinHour: 0 };
const now = new Date('2026-06-21T23:10:00Z');

/** The camera's next shot `ms` from now. */
const shotIn = (ms: number, { night = false, zoneOffsetSeconds = 0 } = {}) => ({ at: new Date(now.getTime() + ms), night, zoneOffsetSeconds });

describe('HourDial', () => {
  it("counts down to the camera's next capture", () => {
    render(<HourDial schedule={hourly} now={now} next={shotIn(20 * 60_000)} />);
    expect(screen.getByText('20:00')).toBeInTheDocument();
    expect(screen.getByText('until next capture')).toBeInTheDocument();
  });

  it('says when photos resume after a skipped night instead of counting down', () => {
    const shot = shotIn(290 * 60_000, { night: true });
    render(<HourDial schedule={hourly} now={now} next={shot} />);
    expect(screen.getByText('Night')).toBeInTheDocument();
    expect(screen.getByText(`photos resume at ${formatClock(shot.at)}`)).toBeInTheDocument();
    expect(screen.queryByText('until next capture')).not.toBeInTheDocument();
  });

  it('counts down for captures every N minutes from start-up', () => {
    render(<HourDial schedule={{ SecondsBetweenCaptures: 600, OffsetWithinHour: -1 }} now={now} next={shotIn(4 * 60_000)} />);
    expect(screen.getByText('4:00')).toBeInTheDocument();
  });

  it("places the hand in the camera's time zone", () => {
    // 23:10 UTC is 04:40 in India (+5:30): the hand points at 40 minutes, not 10.
    const { container } = render(<HourDial schedule={hourly} now={now} next={shotIn(60_000, { zoneOffsetSeconds: 19800 })} />);
    const hand = container.querySelector('.dial-hand')!;
    expect(Number(hand.getAttribute('x2'))).toBeLessThan(100); // left half of the face: 30..60 minutes
  });

  it.each([
    [{ SecondsBetweenCaptures: 600, OffsetWithinHour: -1 }, 'Every 10 min', 'since start-up'],
    [hourly, 'Every 30 min', 'at the marked minutes'],
  ])('describes the schedule before the camera has planned a capture (%o)', (schedule, count, caption) => {
    render(<HourDial schedule={schedule} now={now} next={null} />);
    expect(screen.getByText(count)).toBeInTheDocument();
    expect(screen.getByText(caption)).toBeInTheDocument();
    expect(screen.queryByText('until next capture')).not.toBeInTheDocument();
  });
});
