package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jampat000/Xuva/server/internal/api"
	"github.com/jampat000/Xuva/server/internal/auth"
	"github.com/jampat000/Xuva/server/internal/catalog"
	"github.com/jampat000/Xuva/server/internal/config"
	"github.com/jampat000/Xuva/server/internal/database"
	"github.com/jampat000/Xuva/server/internal/devices"
	"github.com/jampat000/Xuva/server/internal/discovery"
	"github.com/jampat000/Xuva/server/internal/downloads"
	"github.com/jampat000/Xuva/server/internal/events"
	"github.com/jampat000/Xuva/server/internal/jobs"
	"github.com/jampat000/Xuva/server/internal/libraries"
	"github.com/jampat000/Xuva/server/internal/media"
	"github.com/jampat000/Xuva/server/internal/metadata"
	"github.com/jampat000/Xuva/server/internal/metasources"
	"github.com/jampat000/Xuva/server/internal/migration"
	"github.com/jampat000/Xuva/server/internal/movies"
	"github.com/jampat000/Xuva/server/internal/observability"
	"github.com/jampat000/Xuva/server/internal/pairing"
	"github.com/jampat000/Xuva/server/internal/playback"
	"github.com/jampat000/Xuva/server/internal/playstate"
	"github.com/jampat000/Xuva/server/internal/probe"
	"github.com/jampat000/Xuva/server/internal/probes"
	"github.com/jampat000/Xuva/server/internal/resources"
	runtimestore "github.com/jampat000/Xuva/server/internal/runtime"
	"github.com/jampat000/Xuva/server/internal/scanner"
	"github.com/jampat000/Xuva/server/internal/scans"
	"github.com/jampat000/Xuva/server/internal/sessions"
	"github.com/jampat000/Xuva/server/internal/streaming"
	"github.com/jampat000/Xuva/server/internal/subtitles"
	"github.com/jampat000/Xuva/server/internal/transcode"
	"github.com/jampat000/Xuva/server/internal/tv"
)

type Application struct {
	Config    config.Config
	StartedAt time.Time
	cancel    context.CancelFunc
	Database  *database.Service
	Catalog   *catalog.Service
	Auth      *auth.Service
	Events    *events.Bus
	Observe   *observability.Service
	Resources *resources.Manager
	Jobs      *jobs.Registry
	Discovery *discovery.Service

	Libraries *libraries.Service
	Scanner   *scanner.Service
	Scans     *scans.Service
	Probe     *probe.Service
	Probes    *probes.Service
	Media     *media.Service
	Metadata  *metadata.Service
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
	Pairing   *pairing.Service
	Migration *migration.Service
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
	observe := observability.NewService()
	observe.Subscribe(appCtx, bus)
	databaseService, err := database.Open(ctx, cfg.DataDir)
	if err != nil {
		return nil, err
	}
	authService := auth.NewService(databaseService, cfg.AuthDisabled)
	if cfg.AdminPassword != "" {
		if err := authService.Bootstrap(ctx, auth.BootstrapOptions{
			Username: cfg.AdminUsername,
			Password: cfg.AdminPassword,
		}); err != nil {
			return nil, err
		}
	} else if required, err := authService.RequiresBootstrap(ctx); err == nil && required {
		slog.Info("auth bootstrap pending", "reason", "no admin password configured", "next_step", "create first admin from sign-in screen")
	} else if err != nil {
		return nil, err
	}
	catalogService := catalog.NewService(databaseService)
	runtimeStore := runtimestore.NewStore(databaseService)
	libraryService := libraries.NewService()
	if savedLibraries, err := catalogService.ListLibraries(ctx); err == nil {
		for _, library := range savedLibraries {
			libraryService.Set(library)
		}
	}
	if cfg.MovieLibraryPath != "" {
		library, err := catalogService.SaveLibrary(ctx, libraries.Library{
			ID:              "movies",
			Name:            "Movies",
			Path:            cfg.MovieLibraryPath,
			Kind:            libraries.KindMovies,
			MetadataSources: metasources.NormalizeRequestedSourceOrder("movie", cfg.MovieMetadataSources),
		})
		if err == nil {
			libraryService.Set(library)
		}
	}
	if cfg.TVLibraryPath != "" {
		library, err := catalogService.SaveLibrary(ctx, libraries.Library{
			ID:              "tv",
			Name:            "TV",
			Path:            cfg.TVLibraryPath,
			Kind:            libraries.KindTV,
			MetadataSources: metasources.NormalizeRequestedSourceOrder("series", cfg.SeriesMetadataSources),
		})
		if err == nil {
			libraryService.Set(library)
		}
	}
	_ = config.SaveFile(cfg.DataDir, cfg)
	_ = ensureRuntimeDirs(cfg)
	_ = catalogService.SaveSettings(ctx, catalog.RuntimeSettings{
		HTTPAddr:          cfg.HTTPAddr,
		DataDir:           cfg.DataDir,
		TranscodeDir:      cfg.TranscodeDir,
		DownloadsDir:      cfg.DownloadsDir,
		MetadataDir:       cfg.MetadataDir,
		CacheDir:          cfg.CacheDir,
		TempDir:           cfg.TempDir,
		FFmpegPath:        cfg.FFmpegPath,
		FFprobePath:       cfg.FFprobePath,
		ScanWorkers:       cfg.ScanWorkers,
		ProbeWorkers:      cfg.ProbeWorkers,
		TranscodeWorkers:  cfg.TranscodeWorkers,
		GPUWorkers:        cfg.GPUWorkers,
		LibrarySyncMode:   cfg.LibrarySyncMode,
		SyncIntervalMins:  cfg.SyncIntervalMins,
		WatchDebounceSecs: cfg.WatchDebounceSecs,
		ProbeBatchLimit:   cfg.ProbeBatchLimit,
	})

	jobRegistry := jobs.NewRegistry(manager)
	jobRegistry.Start(appCtx)
	scannerService := scanner.NewService()
	movieService := movies.NewService()
	tvService := tv.NewService()
	metadataService := metadata.NewService(cfg, catalogService, bus)
	scanService := scans.NewService(cfg, bus, jobRegistry.Scan, libraryService, scannerService, catalogService, metadataService, movieService, tvService)
	probeService := probe.NewService(cfg.FFprobePath)
	probesService, err := probes.NewPersistentService(ctx, bus, jobRegistry.Probe, catalogService, probeService, runtimeStore)
	if err != nil {
		return nil, err
	}
	startLibraryAutomation(appCtx, cfg, bus, scanService, probesService)

	playStateService := playstate.NewService(databaseService, bus)
	transcodeService, err := transcode.NewPersistentService(ctx, bus, jobRegistry.Transcode, cfg.FFmpegPath, cfg.TranscodeDir, runtimeStore)
	if err != nil {
		return nil, err
	}
	downloadService, err := downloads.NewPersistentService(ctx, bus, jobRegistry.Transcode, cfg.FFmpegPath, cfg.DownloadsDir, runtimeStore)
	if err != nil {
		return nil, err
	}
	sessionService, err := sessions.NewPersistentService(ctx, bus, runtimeStore)
	if err != nil {
		return nil, err
	}
	startRuntimeMaintenance(appCtx, sessionService, transcodeService, probesService, downloadService)
	discoveryService := discovery.NewService(cfg)
	discoveryService.Start(appCtx)

	return &Application{
		Config:    cfg,
		StartedAt: time.Now().UTC(),
		cancel:    cancel,
		Database:  databaseService,
		Catalog:   catalogService,
		Auth:      authService,
		Events:    bus,
		Observe:   observe,
		Resources: manager,
		Jobs:      jobRegistry,
		Discovery: discoveryService,
		Libraries: libraryService,
		Scanner:   scannerService,
		Scans:     scanService,
		Probe:     probeService,
		Probes:    probesService,
		Media:     media.NewService(),
		Metadata:  metadataService,
		Movies:    movieService,
		TV:        tvService,
		Playback:  playback.NewService(),
		PlayState: playStateService,
		Streaming: streaming.NewService(),
		Transcode: transcodeService,
		Subtitles: subtitles.NewService(),
		Devices:   devices.NewPersistentService(databaseService),
		Sessions:  sessionService,
		Downloads: downloadService,
		Pairing:   pairing.NewService(),
		Migration: migration.NewService(databaseService, bus),
	}, nil
}

func startRuntimeMaintenance(ctx context.Context, sessionService *sessions.Service, transcodeService *transcode.Service, probesService *probes.Service, downloadService *downloads.Service) {
	runRuntimeMaintenance(ctx, sessionService, transcodeService, probesService, downloadService)
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runRuntimeMaintenance(ctx, sessionService, transcodeService, probesService, downloadService)
			}
		}
	}()
}

func runRuntimeMaintenance(ctx context.Context, sessionService *sessions.Service, transcodeService *transcode.Service, probesService *probes.Service, downloadService *downloads.Service) {
	const sessionTTL = 15 * time.Minute
	const terminalRetention = 24 * time.Hour
	_, _ = sessionService.Cleanup(ctx, sessionTTL, terminalRetention)
	_, _ = transcodeService.Cleanup(ctx, terminalRetention)
	_, _ = probesService.Cleanup(ctx, terminalRetention)
	_, _ = downloadService.Cleanup(ctx, terminalRetention)
}

func startLibraryAutomation(ctx context.Context, cfg config.Config, bus *events.Bus, scanService *scans.Service, probesService *probes.Service) {
	if cfg.LibrarySyncMode == "manual" {
		return
	}
	interval := time.Duration(cfg.SyncIntervalMins) * time.Minute
	if cfg.LibrarySyncMode == "watch" {
		interval = time.Duration(cfg.WatchDebounceSecs) * time.Second
		if interval < 5*time.Second {
			interval = 5 * time.Second
		}
		if interval > 5*time.Minute {
			interval = 5 * time.Minute
		}
	} else if interval < 15*time.Minute {
		interval = 15 * time.Minute
	}
	probeLimit := cfg.ProbeBatchLimit
	if probeLimit <= 0 {
		probeLimit = 50
	}
	go func() {
		timer := time.NewTimer(interval)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				runAutomatedLibrarySync(ctx, bus, scanService, probesService, probeLimit)
				timer.Reset(interval)
			}
		}
	}()
}

func runAutomatedLibrarySync(ctx context.Context, bus *events.Bus, scanService *scans.Service, probesService *probes.Service, probeLimit int) {
	bus.Publish("automation.sync.started", map[string]any{"probeLimit": probeLimit})
	job, err := scanService.Start(ctx, scans.Request{Kind: scans.KindAll})
	if err != nil {
		bus.Publish("automation.sync.failed", map[string]string{"stage": "scan", "error": err.Error()})
		slog.Debug("automated library scan skipped", "error", err)
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			current, ok := scanService.Get(job.ID)
			if !ok {
				return
			}
			switch current.Status {
			case scans.StatusCompleted:
				if _, err := probesService.Start(ctx, probes.Request{Limit: probeLimit}); err != nil {
					bus.Publish("automation.sync.failed", map[string]string{"stage": "probe", "error": err.Error()})
				}
				bus.Publish("automation.sync.completed", map[string]any{"scanId": job.ID, "probeLimit": probeLimit})
				return
			case scans.StatusFailed:
				bus.Publish("automation.sync.failed", map[string]string{"stage": "scan", "error": current.Error})
				return
			}
		}
	}
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
		Auth:      a.Auth,
		Events:    a.Events,
		Observe:   a.Observe,
		Resources: a.Resources,
		Jobs:      a.Jobs,
		Discovery: a.Discovery,
		Libraries: a.Libraries,
		Scanner:   a.Scanner,
		Scans:     a.Scans,
		Catalog:   a.Catalog,
		Media:     a.Media,
		Metadata:  a.Metadata,
		Movies:    a.Movies,
		TV:        a.TV,
		Probe:     a.Probe,
		Probes:    a.Probes,
		Playback:  a.Playback,
		PlayState: a.PlayState,
		Streaming: a.Streaming,
		Transcode: a.Transcode,
		Downloads: a.Downloads,
		Devices:   a.Devices,
		Sessions:  a.Sessions,
		Subtitles: a.Subtitles,
		Pairing:   a.Pairing,
	})
}

func (a *Application) Shutdown(context.Context) {
	a.cancel()
	if a.Discovery != nil {
		a.Discovery.Shutdown()
	}
	a.Events.Close()
	_ = a.Database.Close()
}
