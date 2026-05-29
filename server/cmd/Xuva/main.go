package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jampat000/Xuva/server/internal/app"
	"github.com/jampat000/Xuva/server/internal/backup"
	"github.com/jampat000/Xuva/server/internal/config"
	"github.com/jampat000/Xuva/server/internal/firewall"
	xuvalogging "github.com/jampat000/Xuva/server/internal/logging"
	"github.com/jampat000/Xuva/server/internal/tlscert"
)

// portFromHTTPAddr extracts the TCP port from a Go ListenAndServe addr.
// Addrs come in three flavours: ":8097", "0.0.0.0:8097", "192.168.1.5:8097".
// Returns 0 on any parse failure (caller treats that as "skip firewall").
func portFromHTTPAddr(addr string) int {
	// Handle pure ":N" form first — net.SplitHostPort wants a non-empty host.
	if len(addr) > 0 && addr[0] == ':' {
		var port int
		if _, err := fmt.Sscanf(addr[1:], "%d", &port); err == nil {
			return port
		}
		return 0
	}
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0
	}
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		return 0
	}
	return port
}

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
		"httpsEnabled", cfg.TLSEnabled,
		"httpsAddr", cfg.HTTPSAddr,
		"serverName", cfg.ServerName,
	)
	application, err := app.New(context.Background(), cfg)
	if err != nil {
		slog.Error("server startup failed", "error", err)
		os.Exit(1)
	}

	router := application.Router()

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Ensure the host firewall lets other LAN devices reach us. On Windows
	// this is the single biggest reason "the server is fine on localhost but
	// my phone can't connect": Windows Defender Firewall blocks unsolicited
	// inbound TCP by default, even for processes listening on 0.0.0.0. On
	// non-Windows platforms this is a no-op (see firewall_other.go).
	//
	// Failure here is non-fatal — the warn log includes the exact manual
	// netsh command the user can run as Administrator.
	if port := portFromHTTPAddr(cfg.HTTPAddr); port > 0 {
		fwCtx, fwCancel := context.WithTimeout(context.Background(), 5*time.Second)
		firewall.LogResult(fwCtx, slog.Default(), port, "Xuva Server")
		fwCancel()
	}

	go func() {
		slog.Info("xuva server listening", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	// Optional HTTPS listener — enabled via XUVA_TLS_ENABLED=true or
	// tlsEnabled in settings.json. When no cert/key files are configured,
	// a self-signed certificate is auto-generated in <DataDir>/tls/ and its
	// SHA-256 fingerprint is logged for browser-trust verification.
	var httpsServer *http.Server
	if cfg.TLSEnabled {
		certFile := cfg.TLSCertFile
		keyFile := cfg.TLSKeyFile
		if certFile == "" {
			certFile = filepath.Join(cfg.DataDir, "tls", "cert.pem")
		}
		if keyFile == "" {
			keyFile = filepath.Join(cfg.DataDir, "tls", "key.pem")
		}

		// Collect SANs: the bind host (if a specific IP was given) plus hostname.
		var hosts []string
		if h, _, err := net.SplitHostPort(cfg.HTTPSAddr); err == nil && h != "" && h != "0.0.0.0" && h != "::" {
			hosts = append(hosts, h)
		}
		if hostname, err := os.Hostname(); err == nil {
			hosts = append(hosts, hostname)
		}

		tlsCert, fingerprint, err := tlscert.Ensure(certFile, keyFile, hosts)
		if err != nil {
			slog.Error("tls setup failed — HTTPS listener not started", "error", err)
		} else {
			if fingerprint != "" {
				slog.Info("tls certificate fingerprint (SHA-256)",
					"fingerprint", fingerprint,
					"certFile", certFile,
				)
			}
			httpsServer = &http.Server{
				Addr:              cfg.HTTPSAddr,
				Handler:           router,
				ReadHeaderTimeout: 5 * time.Second,
				TLSConfig: &tls.Config{
					Certificates: []tls.Certificate{tlsCert},
					MinVersion:   tls.VersionTLS12,
				},
			}
			if port := portFromHTTPAddr(cfg.HTTPSAddr); port > 0 {
				fwCtx, fwCancel := context.WithTimeout(context.Background(), 5*time.Second)
				firewall.LogResult(fwCtx, slog.Default(), port, "Xuva Server (HTTPS)")
				fwCancel()
			}
			go func() {
				slog.Info("xuva server listening (https)", "addr", cfg.HTTPSAddr)
				if err := httpsServer.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
					slog.Error("https server stopped unexpectedly", "error", err)
				}
			}()
		}
	}

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
	if httpsServer != nil {
		if err := httpsServer.Shutdown(ctx); err != nil {
			slog.Error("https server shutdown failed", "error", err)
		}
	}
	slog.Info("xuva server stopped")
}
