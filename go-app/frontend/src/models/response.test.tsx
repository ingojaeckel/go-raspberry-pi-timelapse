import {
  MonitoringResponse,
  SettingsResponse,
  LogResponse,
  PhotosResponse,
  Photo,
} from './response';

describe('response types', () => {
  it('defines MonitoringResponse interface', () => {
    const monitoring: MonitoringResponse = {
      Time: '2024-01-01',
      Uptime: '5 days',
      CpuTemperature: '45°C',
      GpuTemperature: '43°C',
      FreeDiskSpace: '10GB',
    };
    expect(monitoring).toBeDefined();
    expect(monitoring.Time).toBe('2024-01-01');
  });

  it('defines SettingsResponse interface', () => {
    const settings: SettingsResponse = {
      SecondsBetweenCaptures: 300,
      OffsetWithinHour: 0,
      PhotoResolutionWidth: 3280,
      PhotoResolutionHeight: 2464,
      PreviewResolutionWidth: 640,
      PreviewResolutionHeight: 480,
      RotateBy: 0,
      ResolutionSetting: 0,
      Quality: 100,
      DebugEnabled: false,
    };
    expect(settings).toBeDefined();
    expect(settings.SecondsBetweenCaptures).toBe(300);
  });

  it('defines LogResponse interface', () => {
    const logs: LogResponse = {
      Logs: 'Test log content',
    };
    expect(logs).toBeDefined();
    expect(logs.Logs).toBe('Test log content');
  });

  it('defines Photo interface', () => {
    const photo: Photo = {
      Name: 'test.jpg',
      ModTime: '2024-01-01',
      Size: '1024',
    };
    expect(photo).toBeDefined();
    expect(photo.Name).toBe('test.jpg');
  });

  it('defines PhotosResponse interface', () => {
    const photos: PhotosResponse = {
      Photos: [
        { Name: 'photo1.jpg', ModTime: '2024-01-01', Size: '1024' },
        { Name: 'photo2.jpg', ModTime: '2024-01-02', Size: '2048' },
      ],
    };
    expect(photos).toBeDefined();
    expect(photos.Photos.length).toBe(2);
  });
});
