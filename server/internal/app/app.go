package app

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/api"
	"github.com/vyrdenhq/vyrden/server/internal/catalog"
	"github.com/vyrdenhq/vyrden/server/internal/config"
	"github.com/vyrdenhq/vyrden/server/internal/database"
	"github.com/vyrdenhq/vyrden/server/internal/devices"
	"github.com/vyrdenhq/vyrden/server/internal/downloads"
	"github.com/vyrdenhq/vyrden/server/internal/events"
	"github.com/vyrdenhq/vyrden/server/internal/jobs"
	"github.com/vyrdenhq/vyrden/server/internal/libraries"
	"github.com/vyrdenhq/vyrden/server/internal/media"
	"github.com/vyrdenhq/vyrden/server/internal/movies"
	"github.com/vyrdenhq/vyrden/server/internal/playback"
	"github.com/vyrdenhq/vyrden/server/internal/playstate"
	"github.com/vyrdenhq/vyrden/server/internal/probe"
	"github.com/vyrdenhq/vyrden/server/internal/probes"
	"github.com/vyrdenhq/vyrden/server/internal/resources"
	"github.com/vyrdenhq/vyrden/server/internal/scanner"
	"github.com/vyrdenhq/vyrden/server/internal/scans"
	"github.com/vyrdenhq/vyrden/server/internal/sessions"
	"github.com/vyrdenhq/vyrden/server/internal/streaming"
	"github.com/vyrdenhq/vyrden/server/internal/subtitles"
	"github.com/vyrdenhq/vyrden/server/internal/transcode"
	"github.com/vyrdenhq/vyrden/server/internal/tv"
)

type Application struct {
	Config    config.Config
	StartedAt time.Time
	cancel    context.CancelFunc
	Database  *database.Service
	Catalog   *catalog.Service
	Events    *events.Bus
	Resources *resources.Manager
	Jobs      *jobs.Registry

	Libraries *libraries.Service
	Scanner   *scanner.Service
	Scans     *scans.Service
	Probe     *probe.Service
	Probes    *probes.Service
	Media     *media.Service
	Movies    *movies.Service
	TV        *tv.Service
	Playback  *playback.Service
	PlayState *playstate.Service
	Streaming *streaming.Service
	Transcode *transcode.Service
	Subtitles *subtitles.Service
	Devices   *devices.Service
	Sessions  *sessions.Service
	Downloads *downloads.Service
}

func New(ctx context.Context, cfg config.Config) (*Application, error) {
	appCtx, cancel := context.WithCancel(ctx)
	limits := resources.Limits{
		ScanWorkers:      cfg.ScanWorkers,
		ProbeWorkers:     cfg.ProbeWorkers,
		TranscodeWorkers: cfg.TranscodeWorkers,
		GPUWorkers:       cfg.GPUWorkers,
	}
	manager := resources.NewManager(limits)
	bus := events.NewBus(cfg.EventBuffer)
	databaseService, err := database.Open(ctx, cfg.DataDir)
	if err != nil {
		return nil, err
	}
	catalogService := catalog.NewService(databaseService)
	libraryService := libraries.NewService()
	if savedLibraries, err := catalogService.ListLibraries(ctx); err == nil {
		for _, library := range savedLibraries {
			libraryService.Set(library)
		}
	}
	if cfg.MovieLibraryPath != "" {
		library, err := catalogService.SaveLibrary(ctx, libraries.Library{
			ID:   "movies",
			Name: "Movies",
			Path: cfg.MovieLibraryPath,
			Kind: libraries.KindMovies,
		})
		if err == nil {
			libraryService.Set(library)
		}
	}
	if cfg.TVLibraryPath != "" {
		library, err := catalogService.SaveLibrary(ctx, libraries.Library{
			ID:   "tv",
			Name: "TV",
			Path: cfg.TVLibraryPath,
			Kind: libraries.KindTV,
		})
		if err == nil {
			libraryService.Set(library)
		}
	}
	_ = config.SaveFile(cfg.DataDir, cfg)
	_ = ensureRuntimeDirs(cfg)
	_ = catalogService.SaveSettings(ctx, catalog.RuntimeSettings{
		HTTPAddr:         cfg.HTTPAddr,
		DataDir:          cfg.DataDir,
		TranscodeDir:     cfg.TranscodeDir,
		DownloadsDir:     cfg.DownloadsDir,
		MetadataDir:      cfg.MetadataDir,
		CacheDir:         cfg.CacheDir,
		TempDir:          cfg.TempDir,
		FFmpegPath:       cfg.FFmpegPath,
		FFprobePath:      cfg.FFprobePath,
		ScanWorkers:      cfg.ScanWorkers,
		ProbeWorkers:     cfg.ProbeWorkers,
		TranscodeWorkers: cfg.TranscodeWorkers,
		GPUWorkers:       cfg.GPUWorkers,
	})

	jobRegistry := jobs.NewRegistry(manager)
	jobRegistry.Start(appCtx)
	scannerService := scanner.NewService()
	movieService := movies.NewService()
	tvService := tv.NewService()
	scanService := scans.NewService(cfg, bus, jobRegistry.Scan, libraryService, scannerService, catalogService, movieService, tvService)
	probeService := probe.NewService(cfg.FFprobePath)
	probesService := probes.NewService(bus, jobRegistry.Probe, catalogService, probeService)

	playStateService := playstate.NewService(databaseService, bus)

	return &Application{
		Config:    cfg,
		StartedAt: time.Now().UTC(),
		cancel:    cancel,
		Database:  databaseService,
		Catalog:   catalogService,
		Events:    bus,
		Resources: manager,
		Jobs:      jobRegistry,
		Libraries: libraryService,
		Scanner:   scannerService,
		Scans:     scanService,
		Probe:     probeService,
		Probes:    probesService,
		Media:     media.NewService(),
		Movies:    movieService,
		TV:        tvService,
		Playback:  playback.NewService(),
		PlayState: playStateService,
		Streaming: streaming.NewService(),
		Transcode: transcode.NewService(bus, jobRegistry.Transcode, cfg.FFmpegPath, cfg.TranscodeDir),
		Subtitles: subtitles.NewService(),
		Devices:   devices.NewService(),
		Sessions:  sessions.NewService(bus),
		Downloads: downloads.NewService(bus, jobRegistry.Transcode, cfg.FFmpegPath, cfg.DownloadsDir),
	}, nil
}

func ensureRuntimeDirs(cfg config.Config) error {
	for _, dir := range []string{cfg.DataDir, cfg.TranscodeDir, cfg.DownloadsDir, cfg.MetadataDir, cfg.CacheDir, cfg.TempDir} {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) Router() http.Handler {
	return api.NewRouter(api.Deps{
		Config:    a.Config,
		StartedAt: a.StartedAt,
		Events:    a.Events,
		Resources: a.Resources,
		Jobs:      a.Jobs,
		Libraries: a.Libraries,
		Scanner:   a.Scanner,
		Scans:     a.Scans,
		Catalog:   a.Catalog,
		Media:     a.Media,
		Movies:    a.Movies,
		TV:        a.TV,
		Probe:     a.Probe,
		Probes:    a.Probes,
		Playback:  a.Playback,
		PlayState: a.PlayState,
		Transcode: a.Transcode,
		Downloads: a.Downloads,
		Devices:   a.Devices,
		Sessions:  a.Sessions,
	})
}

func (a *Application) Shutdown(context.Context) {
	a.cancel()
	a.Events.Close()
	_ = a.Database.Close()
}
