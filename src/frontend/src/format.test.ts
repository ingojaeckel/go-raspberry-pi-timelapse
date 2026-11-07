import {
  checkClock,
  estimateRunway,
  formatCountdown,
  formatDuration,
  formatPastHour,
  parseCpuTemp,
  parseDf,
  parseGoTime,
  parseGpuTemp,
  parseSize,
  parseUptime,
  primaryDisk,
} from './format';

describe('parseGoTime', () => {
  it('parses time.Time.String() including the zone offset', () => {
    const d = parseGoTime('2026-11-02 21:15:00.123456789 +0100 CET');
    expect(d?.toISOString()).toBe('2026-11-02T20:15:00.123Z');
  });
  it('parses times without fractional seconds', () => {
    expect(parseGoTime('2026-01-05 07:00:00 +0000 UTC')?.toISOString()).toBe('2026-01-05T07:00:00.000Z');
  });
  it('returns null for anything else', () => {
    expect(parseGoTime('yesterday')).toBeNull();
  });
});

describe('durations', () => {
  it('formats countdowns', () => {
    expect(formatCountdown(65)).toBe('1:05');
    expect(formatCountdown(3725)).toBe('1:02:05');
    expect(formatCountdown(-3)).toBe('0:00');
  });
  it('formats intervals', () => {
    expect(formatDuration(45)).toBe('45 s');
    expect(formatDuration(1800)).toBe('30 min');
    expect(formatDuration(5400)).toBe('1 h 30 min');
  });
  it('formats times past the hour', () => {
    expect(formatPastHour(900)).toBe(':15');
    expect(formatPastHour(450)).toBe(':07:30');
  });
});

describe('sizes', () => {
  it('parses backend and df sizes', () => {
    expect(parseSize('2 MB')).toBe(2 * 1024 ** 2);
    expect(parseSize('800 KB')).toBe(800 * 1024);
    expect(parseSize('12 bytes')).toBe(12);
    expect(parseSize('12G')).toBe(12 * 1024 ** 3);
    expect(parseSize('3,5Gi')).toBe(3.5 * 1024 ** 3);
    expect(parseSize('n/a')).toBeNull();
  });
});

describe('Pi command output', () => {
  it('reads temperatures', () => {
    expect(parseCpuTemp('48312\n')).toBeCloseTo(48.312);
    expect(parseCpuTemp('cat: no such file')).toBeNull();
    expect(parseGpuTemp("temp=51.0'C\n")).toBe(51);
    expect(parseGpuTemp('fork/exec /opt/vc/bin/vcgencmd: no such file or directory')).toBeNull();
  });

  it('reads uptime on Linux and macOS', () => {
    expect(parseUptime(' 21:39:00 up 3 days,  2:01,  1 user,  load average: 0.08, 0.10, 0.09')).toEqual({
      up: '3 days, 2:01',
      load: [0.08, 0.1, 0.09],
    });
    expect(parseUptime('21:39  up 5 mins, 2 users, load averages: 1.52 1.80 1.95').load).toEqual([1.52, 1.8, 1.95]);
  });

  it('reads GNU df -h and keeps block devices only', () => {
    const disks = parseDf(
      [
        'Filesystem      Size  Used Avail Use% Mounted on',
        '/dev/root        29G  4.1G   24G  15% /',
        'devtmpfs        183M     0  183M   0% /dev',
        '/dev/mmcblk0p1  253M   51M  202M  20% /boot/firmware',
      ].join('\n'),
    );
    expect(disks.map((d) => d.mount)).toEqual(['/', '/boot/firmware']);
    expect(primaryDisk(disks)).toMatchObject({ avail: '24G', usedPercent: 15 });
  });

  it('reads BSD df -h with mount points containing spaces', () => {
    const disks = parseDf(
      [
        'Filesystem        Size    Used   Avail Capacity iused ifree %iused  Mounted on',
        '/dev/disk3s1s1   926Gi    10Gi   500Gi     3%    426k  4.3G    0%   /',
        '/dev/disk5s1     100Gi    50Gi    50Gi    50%      1k    1k    1%   /Volumes/My Card',
      ].join('\n'),
    );
    expect(disks.map((d) => [d.mount, d.usedPercent])).toEqual([
      ['/', 3],
      ['/Volumes/My Card', 50],
    ]);
  });

  it('measures clock drift and spots a wrong year', () => {
    const now = new Date(2026, 10, 2, 14, 10, 0);
    expect(checkClock(now.getTime() + 150_000, now)).toEqual({ driftSeconds: 150, wrongYear: null });
    // Behind by three days after a long power cut: more than the old in-hour comparison could see.
    expect(checkClock(now.getTime() - 3 * 86400_000, now)).toEqual({ driftSeconds: -259200, wrongYear: null });
    expect(checkClock(0, now).wrongYear).toBe(new Date(0).getFullYear());
  });

  it('estimates how long until the card is full', () => {
    const storage = {
      FreeBytes: 1024 ** 3 + 256 * 1024 ** 2,
      ReserveBytes: 256 * 1024 ** 2,
      Photos: 4,
      PhotoBytes: 8 * 1024 ** 2,
      StopWhenDiskFull: false,
      DeletedOldest: 0,
    };
    // 1 GB usable / 2 MB per photo, 48 a day
    expect(estimateRunway(storage, 48)).toEqual({ photos: 512, days: 512 / 48 });
    expect(estimateRunway(storage, 0)).toEqual({ photos: 512, days: null });
    expect(estimateRunway({ ...storage, FreeBytes: 0 }, 48)).toEqual({ photos: 0, days: 0 });
    expect(estimateRunway({ ...storage, Photos: 0, PhotoBytes: 0 }, 48)).toBeNull();
    expect(estimateRunway({ ...storage, Error: 'statfs failed' }, 48)).toBeNull();
  });
});
