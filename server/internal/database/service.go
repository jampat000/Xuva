package database

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Service struct {
	DataDir string
	Path    string
	db      *sql.DB
}

func Open(ctx context.Context, dataDir string) (*Service, error) {
	if dataDir == "" {
		dataDir = "data"
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dataDir, "vyrden.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	service := &Service{
		DataDir: dataDir,
		Path:    dbPath,
		db:      db,
	}
	if err := service.configure(ctx); err != nil {
		db.Close()
		return nil, err
	}
	if err := service.migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return service, nil
}

func (s *Service) DB() *sql.DB {
	return s.db
}

func (s *Service) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Service) configure(ctx context.Context) error {
	statements := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) migrate(ctx context.Context) error {
	for _, statement := range migrations {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS libraries (
		id TEXT PRIMARY KEY,
		kind TEXT NOT NULL,
		name TEXT NOT NULL,
		path TEXT NOT NULL UNIQUE,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS media_sources (
		id TEXT PRIMARY KEY,
		library_id TEXT NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
		kind TEXT NOT NULL,
		path TEXT NOT NULL UNIQUE,
		rel_path TEXT NOT NULL,
		name TEXT NOT NULL,
		extension TEXT NOT NULL,
		size_bytes INTEGER NOT NULL,
		modified_at TEXT NOT NULL,
		discovered_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_media_sources_library_kind ON media_sources(library_id, kind)`,
	`CREATE TABLE IF NOT EXISTS media_probes (
		media_source_id TEXT PRIMARY KEY REFERENCES media_sources(id) ON DELETE CASCADE,
		container TEXT NOT NULL,
		duration_seconds REAL NOT NULL,
		bitrate INTEGER NOT NULL,
		video_codec TEXT NOT NULL,
		width INTEGER NOT NULL,
		height INTEGER NOT NULL,
		audio_streams INTEGER NOT NULL,
		subtitle_streams INTEGER NOT NULL,
		raw_json TEXT NOT NULL,
		probed_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS movies (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		year INTEGER NOT NULL,
		sort_title TEXT NOT NULL,
		needs_review INTEGER NOT NULL,
		review_reason TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS movie_versions (
		movie_id TEXT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
		media_source_id TEXT NOT NULL REFERENCES media_sources(id) ON DELETE CASCADE,
		edition TEXT NOT NULL,
		quality_label TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at TEXT NOT NULL,
		PRIMARY KEY(movie_id, media_source_id)
	)`,
	`CREATE TABLE IF NOT EXISTS tv_series (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		sort_title TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at TEXT NOT NULL
	)`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_tv_series_title ON tv_series(title)`,
	`CREATE TABLE IF NOT EXISTS tv_seasons (
		id TEXT PRIMARY KEY,
		series_id TEXT NOT NULL REFERENCES tv_series(id) ON DELETE CASCADE,
		season_number INTEGER NOT NULL,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at TEXT NOT NULL,
		UNIQUE(series_id, season_number)
	)`,
	`CREATE TABLE IF NOT EXISTS tv_episodes (
		id TEXT PRIMARY KEY,
		series_id TEXT NOT NULL REFERENCES tv_series(id) ON DELETE CASCADE,
		season_id TEXT NOT NULL REFERENCES tv_seasons(id) ON DELETE CASCADE,
		season_number INTEGER NOT NULL,
		episode_number INTEGER NOT NULL,
		episode_end INTEGER NOT NULL,
		title TEXT NOT NULL,
		needs_review INTEGER NOT NULL,
		review_reason TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS episode_versions (
		episode_id TEXT NOT NULL REFERENCES tv_episodes(id) ON DELETE CASCADE,
		media_source_id TEXT NOT NULL REFERENCES media_sources(id) ON DELETE CASCADE,
		quality_label TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at TEXT NOT NULL,
		PRIMARY KEY(episode_id, media_source_id)
	)`,
	`CREATE TABLE IF NOT EXISTS scan_runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		kind TEXT NOT NULL,
		root TEXT NOT NULL,
		started_at TEXT NOT NULL,
		completed_at TEXT NOT NULL,
		duration_ms INTEGER NOT NULL,
		total_files INTEGER NOT NULL,
		media_files INTEGER NOT NULL,
		ignored_files INTEGER NOT NULL,
		error_count INTEGER NOT NULL
	)`,
}
