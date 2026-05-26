package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jampat000/Xuva/server/internal/app"
	"github.com/jampat000/Xuva/server/internal/backup"
	"github.com/jampat000/Xuva/server/internal/config"
	xuvalogging "github.com/jampat000/Xuva/server/internal/logging"
)

func main() {
	cfg := config.FromEnv()
	logCloser, err := xuvalogging.Configure(xuvalogging.Config{
		Format:   cfg.LogFormat,
		Level:    cfg.LogLevel,
		Dir:      cfg.LogDir,
		MaxMB:    cfg.LogMaxMB,
		MaxFiles: cfg.LogMaxFiles,
	})
	if err != nil {
		slog.Warn("structured file logging disabled", "logDir", cfg.LogDir, "error", err)
	}
	defer logCloser.Close()
	if restored, err := backup.ApplyIfPending(cfg.DataDir); err != nil {
		slog.Error("backup restore failed", "error", err)
		os.Exit(1)
	} else if restored {
		slog.Info("backup restore applied — starting with restored database")
	}
	slog.Info("xuva server starting",
		"runtimeHome", cfg.RuntimeHome,
		"runtimeScope", cfg.RuntimeScope,
		"dataDir", cfg.DataDir,
		"logDir", cfg.LogDir,
		"logMaxMB", cfg.LogMaxMB,
		"logMaxFiles", cfg.LogMaxFiles,
		"httpAddr", cfg.HTTPAddr,
		"serverName", cfg.ServerName,
	)
	application, err := app.New(context.Background(), cfg)
	if err != nil {
		slog.Error("server startup failed", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           application.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("xuva server listening", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	application.Shutdown(ctx)
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}
	slog.Info("xuva server stopped")
}
