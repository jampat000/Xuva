package app

import (
	"context"
	"net/http"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/api"
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
	"github.com/vyrdenhq/vyrden/server/internal/probe"
	"github.com/vyrdenhq/vyrden/server/internal/resources"
	"github.com/vyrdenhq/vyrden/server/internal/scanner"
	"github.com/vyrdenhq/vyrden/server/internal/sessions"
	"github.com/vyrdenhq/vyrden/server/internal/streaming"
	"github.com/vyrdenhq/vyrden/server/internal/subtitles"
	"github.com/vyrdenhq/vyrden/server/internal/transcode"
	"github.com/vyrdenhq/vyrden/server/internal/tv"
)

type Application struct {
	Config    config.Config
	StartedAt time.Time
	Database  *database.Service
	Events    *events.Bus
	Resources *resources.Manager
	Jobs      *jobs.Registry

	Libraries *libraries.Service
	Scanner   *scanner.Service
	Probe     *probe.Service
	Media     *media.Service
	Movies    *movies.Service
	TV        *tv.Service
	Playback  *playback.Service
	Streaming *streaming.Service
	Transcode *transcode.Service
	Subtitles *subtitles.Service
	Devices   *devices.Service
	Sessions  *sessions.Service
	Downloads *downloads.Service
}

func New(cfg config.Config) *Application {
	limits := resources.Limits{
		ScanWorkers:      cfg.ScanWorkers,
		ProbeWorkers:     cfg.ProbeWorkers,
		TranscodeWorkers: cfg.TranscodeWorkers,
		GPUWorkers:       cfg.GPUWorkers,
	}
	manager := resources.NewManager(limits)
	bus := events.NewBus(cfg.EventBuffer)

	return &Application{
		Config:    cfg,
		StartedAt: time.Now().UTC(),
		Database:  database.NewService(cfg.DataDir),
		Events:    bus,
		Resources: manager,
		Jobs:      jobs.NewRegistry(manager),
		Libraries: libraries.NewService(),
		Scanner:   scanner.NewService(),
		Probe:     probe.NewService(),
		Media:     media.NewService(),
		Movies:    movies.NewService(),
		TV:        tv.NewService(),
		Playback:  playback.NewService(),
		Streaming: streaming.NewService(),
		Transcode: transcode.NewService(),
		Subtitles: subtitles.NewService(),
		Devices:   devices.NewService(),
		Sessions:  sessions.NewService(),
		Downloads: downloads.NewService(),
	}
}

func (a *Application) Router() http.Handler {
	return api.NewRouter(api.Deps{
		Config:    a.Config,
		StartedAt: a.StartedAt,
		Events:    a.Events,
		Resources: a.Resources,
		Jobs:      a.Jobs,
		Media:     a.Media,
		Movies:    a.Movies,
		TV:        a.TV,
		Playback:  a.Playback,
		Sessions:  a.Sessions,
	})
}

func (a *Application) Shutdown(context.Context) {
	a.Events.Close()
}
