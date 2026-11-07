import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PhotosView } from './PhotosView';
import { callsTo, json, mockServer, settings } from '../test/mockServer';
import { OFFLINE_MESSAGE } from '../api';

describe('PhotosView', () => {
  it('lists photos newest first, grouped by day', async () => {
    mockServer();
    render(<PhotosView settings={settings} />);
    expect(await screen.findByText('20261102-091500.jpg')).toBeInTheDocument();
    const groups = screen.getAllByRole('group');
    expect(groups).toHaveLength(2);
    expect(screen.getByText('3 photos', { exact: false })).toBeInTheDocument();
  });

  it('says when night photos are skipped', async () => {
    mockServer();
    render(<PhotosView settings={{ ...settings, SkipNightPhotos: true, Latitude: 51.5, Longitude: -0.1 }} />);
    expect(await screen.findByText(/Every 30 min.*, daytime only/)).toBeInTheDocument();
  });

  it('loads no photo until one is opened: the list only', async () => {
    const f = mockServer();
    const { container } = render(<PhotosView settings={settings} />);
    expect(await screen.findByText('20261102-091500.jpg')).toBeInTheDocument();
    expect(container.querySelector('img')).toBeNull();
    expect(f.mock.calls.filter(([u]) => String(u).includes('/file/'))).toHaveLength(0);
  });

  it('offers archives of everything, then of the selection', async () => {
    mockServer();
    render(<PhotosView settings={settings} />);
    const user = userEvent.setup();
    await screen.findByText('20261102-091500.jpg');

    expect(screen.getByRole('link', { name: 'ZIP' })).toHaveAttribute('href', '/archive/zip');
    expect(screen.getByRole('link', { name: 'TAR' })).toHaveAttribute('href', '/archive/tar');

    await user.click(screen.getByRole('checkbox', { name: 'Select 20261102-091500.jpg' }));
    expect(screen.getByText('1 photo selected')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'ZIP' })).toHaveAttribute('href', '/archive/zip?f=20261102-091500.jpg');
    expect(screen.getByRole('link', { name: 'Download 20261102-091500.jpg' })).toHaveAttribute(
      'href',
      '/file/20261102-091500.jpg',
    );
  });

  it('selects a whole day at once', async () => {
    mockServer();
    render(<PhotosView settings={settings} />);
    const user = userEvent.setup();
    await screen.findByText('20261102-091500.jpg');
    const dayBoxes = screen.getAllByRole('checkbox', { name: /Select all photos from/ });
    await user.click(dayBoxes[1]!);
    expect(screen.getByText('2 photos selected')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'TAR' })).toHaveAttribute(
      'href',
      '/archive/tar?f=20261101-084500.jpg&f=20261101-081500.jpg',
    );
  });

  it('deletes the selection after confirmation and refreshes the list', async () => {
    const f = mockServer();
    render(<PhotosView settings={settings} />);
    const user = userEvent.setup();
    await screen.findByText('20261102-091500.jpg');
    await user.click(screen.getByRole('checkbox', { name: 'Select 20261102-091500.jpg' }));
    await user.click(screen.getByRole('button', { name: 'Delete' }));

    const dialog = screen.getByRole('dialog', { name: 'Delete 1 photo?' });
    expect(callsTo(f, 'DELETE', '/file')).toHaveLength(0);
    await user.click(within(dialog).getByRole('button', { name: 'Delete' }));

    await waitFor(() => expect(callsTo(f, 'DELETE', '/file')).toHaveLength(1));
    const url = new URL(String(callsTo(f, 'DELETE', '/file')[0]![0]), 'http://x');
    expect(url.searchParams.getAll('f')).toEqual(['20261102-091500.jpg']);
    expect(await screen.findByText('Deleted 1 photo.')).toBeInTheDocument();
    await waitFor(() => expect(callsTo(f, 'GET', '/photos')).toHaveLength(2));
  });

  it('encodes unusual file names', async () => {
    mockServer({
      'GET /photos': () => json({ Photos: [{ Name: 'a b&c.jpg', ModTime: '2026-11-02 09:15:00 +0000 UTC', Size: '1 MB' }] }),
    });
    render(<PhotosView settings={settings} />);
    const user = userEvent.setup();
    await user.click(await screen.findByRole('checkbox', { name: 'Select a b&c.jpg' }));
    expect(screen.getByRole('link', { name: 'ZIP' })).toHaveAttribute('href', '/archive/zip?f=a+b%26c.jpg');
    expect(screen.getByRole('link', { name: 'Download a b&c.jpg' })).toHaveAttribute('href', '/file/a%20b%26c.jpg');
  });

  it('opens a photo in the viewer', async () => {
    mockServer();
    render(<PhotosView settings={settings} />);
    const user = userEvent.setup();
    await screen.findByText('20261102-091500.jpg');
    // Older days start collapsed
    await user.click(within(screen.getAllByRole('group')[1]!).getByText(/photos/));
    await user.click(await screen.findByText('20261101-084500.jpg'));
    const viewer = screen.getByRole('dialog', { name: 'Photo 20261101-084500.jpg' });
    expect(within(viewer).getByText('2 of 3')).toBeInTheDocument();
    await user.click(within(viewer).getByRole('button', { name: /Older/ }));
    expect(within(viewer).getByText('3 of 3')).toBeInTheDocument();
  });

  it('explains when the camera is unreachable', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));
    render(<PhotosView settings={settings} />);
    expect(await screen.findByText(OFFLINE_MESSAGE)).toBeInTheDocument();
  });

  it('shows an empty state when nothing was captured yet', async () => {
    mockServer({ 'GET /photos': () => json({ Photos: null }) });
    render(<PhotosView settings={settings} />);
    expect(await screen.findByText(/hasn't saved any photos yet/)).toBeInTheDocument();
  });
});
