import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/svelte';
import type { PlaybackRouteResponse, PlaybackStateResponse } from '$lib/api/details';

// ─── Mocks ────────────────────────────────────────────────────────────────────

vi.mock('hls.js', () => ({
  default: {
    isSupported: vi.fn().mockReturnValue(false),
    Events: {},
    ErrorTypes: {},
  },
}));

vi.mock('$lib/api/details', () => ({
  getMediaSourceTracks: vi.fn().mockResolvedValue({ audioTracks: [], subtitleTracks: [] }),
  getPlaybackRoute: vi.fn(),
  getStreamToken: vi.fn().mockResolvedValue({ query: '?sessionId=s&deviceId=web&token=t' }),
  heartbeatClientPlayback: vi.fn().mockResolvedValue(undefined),
  stopClientPlayback: vi.fn().mockResolvedValue(undefined),
  setPlaybackState: vi.fn().mockResolvedValue(undefined),
}));

// Stub fetch so VTT requests never throw
global.fetch = vi.fn().mockResolvedValue({ ok: false } as Response);

// ─── Test fixtures ────────────────────────────────────────────────────────────

const BASE_ROUTE: PlaybackRouteResponse = {
  protocol: 'direct',
  url: '/api/stream/test.mp4',
  decision: { reason: 'direct-play', container: 'mp4' } as PlaybackRouteResponse['decision'],
};

const NO_PROGRESS: PlaybackStateResponse = {
  progressSeconds: 0,
  durationSeconds: 0,
  watched: false,
  updatedAt: '',
};

// ─── Tests ────────────────────────────────────────────────────────────────────

describe('Player component', async () => {
  // Lazy import so mock registrations above are in place before module loads
  const { default: Player } = await import('./Player.svelte');

  beforeEach(() => {
    vi.clearAllMocks();
    global.fetch = vi.fn().mockResolvedValue({ ok: false } as Response);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('mounts without throwing', () => {
    expect(() => {
      render(Player, {
        props: {
          mediaSourceId: 'ms-1',
          title: 'Test Movie',
          initialRoute: BASE_ROUTE,
          initialState: NO_PROGRESS,
        },
      });
    }).not.toThrow();
  });

  it('calls getMediaSourceTracks on mount', async () => {
    const { getMediaSourceTracks } = await import('$lib/api/details');
    render(Player, {
      props: {
        mediaSourceId: 'ms-1',
        title: 'Test Movie',
        initialRoute: BASE_ROUTE,
        initialState: NO_PROGRESS,
      },
    });
    await waitFor(() => {
      expect(getMediaSourceTracks).toHaveBeenCalledWith('ms-1');
    });
  });

  it('shows resume toast when savedProgress > 10s and < 95%', async () => {
    const withProgress: PlaybackStateResponse = {
      progressSeconds: 300,   // 5 minutes in
      durationSeconds: 3600,  // 1 hour total → ~8% watched
      watched: false,
      updatedAt: '',
    };
    render(Player, {
      props: {
        mediaSourceId: 'ms-1',
        title: 'Test Movie',
        initialRoute: BASE_ROUTE,
        initialState: withProgress,
      },
    });
    await waitFor(() => {
      expect(screen.getByText(/Resuming from/i)).toBeInTheDocument();
    });
  });

  it('does NOT show resume toast when savedProgress <= 10s', async () => {
    const tooEarly: PlaybackStateResponse = {
      progressSeconds: 5,
      durationSeconds: 3600,
      watched: false,
      updatedAt: '',
    };
    render(Player, {
      props: {
        mediaSourceId: 'ms-1',
        title: 'Test Movie',
        initialRoute: BASE_ROUTE,
        initialState: tooEarly,
      },
    });
    // Wait a tick for onMount to run
    await waitFor(() => {
      expect(screen.queryByText(/Resuming from/i)).not.toBeInTheDocument();
    });
  });

  it('does NOT show resume toast when progress >= 95%', async () => {
    const nearEnd: PlaybackStateResponse = {
      progressSeconds: 3500,  // ~97% of 3600
      durationSeconds: 3600,
      watched: true,
      updatedAt: '',
    };
    render(Player, {
      props: {
        mediaSourceId: 'ms-1',
        title: 'Test Movie',
        initialRoute: BASE_ROUTE,
        initialState: nearEnd,
      },
    });
    await waitFor(() => {
      expect(screen.queryByText(/Resuming from/i)).not.toBeInTheDocument();
    });
  });

  it('auto-selects first subtitle track when defaultSubtitlesEnabled=true', async () => {
    const { getMediaSourceTracks, getPlaybackRoute } = await import('$lib/api/details');
    vi.mocked(getMediaSourceTracks).mockResolvedValueOnce({
      audioTracks: [{ index: 0, default: true, language: 'eng', title: 'English', forced: false }],
      subtitleTracks: [{ index: 0, default: false, language: 'eng', title: 'English', forced: false }],
    });
    vi.mocked(getPlaybackRoute).mockResolvedValueOnce(BASE_ROUTE);

    render(Player, {
      props: {
        mediaSourceId: 'ms-1',
        title: 'Test Movie',
        initialRoute: BASE_ROUTE,
        initialState: NO_PROGRESS,
        defaultSubtitlesEnabled: true,
      },
    });

    await waitFor(() => {
      expect(getPlaybackRoute).toHaveBeenCalled();
    });
  });

  it('renders the back-to-home link', async () => {
    render(Player, {
      props: {
        mediaSourceId: 'ms-1',
        title: 'Test Movie',
        initialRoute: BASE_ROUTE,
        initialState: NO_PROGRESS,
        backHref: '/movies',
      },
    });
    const link = await screen.findByRole('link', { name: /back/i });
    expect(link).toHaveAttribute('href', '/movies');
  });
});
