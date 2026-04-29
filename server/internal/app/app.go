package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/api"
	"github.com/vyrdenhq/vyrden/server/internal/auth"
	"github.com/vyrdenhq/vyrden/server/internal/catalog"
	"github.com/vyrdenhq/vyrden/server/internal/config"
	"github.com/vyrdenhq/vyrden/server/internal/database"
	"github.com/vyrdenhq/vyrden/server/internal/devices"
	"github.com/vyrdenhq/vyrden/server/internal/downloads"
	"github.com/vyrdenhq/vyrden/server/internal/events"
	"github.com/vyrdenhq/vyrden/server/internal/jobs"
	"github.com/vyrdenhq/vyrden/server/internal/libraries"
	"github.com/vyrdenhq/vyrden/server/internal/media"
	"github.com/vyrdenhq/vyrden/server/internal/metadata"
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
	Auth      *auth.Service
	Events    *events.Bus
	Resources *resources.Manager
	Jobs      *jobs.Registry

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
	authService := auth.NewService(databaseService, cfg.AuthDisabled)
	if err := authService.Bootstrap(ctx, auth.BootstrapOptions{
		Username: cfg.AdminUsername,
		Password: cfg.AdminPassword,
	}); err != nil {
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
	metadataService := metadata.NewService(cfg, catalogService, bus)
	scanService := scans.NewService(cfg, bus, jobRegistry.Scan, libraryService, scannerService, catalogService, metadataService, movieService, tvService)
	probeService := probe.NewService(cfg.FFprobePath)
	probesService := probes.NewService(bus, jobRegistry.Probe, catalogService, probeService)
	startLibraryAutomation(appCtx, cfg, bus, scanService, probesService)

	playStateService := playstate.NewService(databaseService, bus)

	return &Application{
		Config:    cfg,
		StartedAt: time.Now().UTC(),
		cancel:    cancel,
		Database:  databaseService,
		Catalog:   catalogService,
		Auth:      authService,
		Events:    bus,
		Resources: manager,
		Jobs:      jobRegistry,
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
		Transcode: transcode.NewService(bus, jobRegistry.Transcode, cfg.FFmpegPath, cfg.TranscodeDir),
		Subtitles: subtitles.NewService(),
		Devices:   devices.NewService(),
		Sessions:  sessions.NewService(bus),
		Downloads: downloads.NewService(bus, jobRegistry.Transcode, cfg.FFmpegPath, cfg.DownloadsDir),
	}, nil
}

func startLibraryAutomation(ctx context.Context, cfg config.Config, bus *events.Bus, scanService *scans.Service, probesService *probes.Service) {
	if cfg.LibrarySyncMode == "manual" {
		return
	}
	interval := time.Duration(cfg.SyncIntervalMins) * time.Minute
	if interval < 15*time.Minute {
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
		Resources: a.Resources,
		Jobs:      a.Jobs,
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
	})
}

func (a *Application) Shutdown(context.Context) {
	a.cancel()
	a.Events.Close()
	_ = a.Database.Close()
}
