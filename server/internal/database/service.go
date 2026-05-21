package database

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"

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

	dbPath := filepath.Join(dataDir, "xuva.db")
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
			if strings.Contains(err.Error(), "duplicate column name") {
				continue
			}
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
		storage_type TEXT NOT NULL DEFAULT 'unknown',
		metadata_sources_json TEXT NOT NULL DEFAULT '[]',
		artwork_sources_json TEXT NOT NULL DEFAULT '[]',
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at TEXT NOT NULL
	)`,
	`ALTER TABLE libraries ADD COLUMN storage_type TEXT NOT NULL DEFAULT 'unknown'`,
	`ALTER TABLE libraries ADD COLUMN metadata_sources_json TEXT NOT NULL DEFAULT '[]'`,
	`ALTER TABLE libraries ADD COLUMN artwork_sources_json TEXT NOT NULL DEFAULT '[]'`,
	`CREATE TABLE IF NOT EXISTS app_settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
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
	`CREATE INDEX IF NOT EXISTS idx_media_sources_library_rel_path ON media_sources(library_id, rel_path)`,
	`CREATE INDEX IF NOT EXISTS idx_media_sources_updated ON media_sources(updated_at DESC)`,
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
	`CREATE INDEX IF NOT EXISTS idx_movies_sort ON movies(sort_title, year)`,
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
	`CREATE INDEX IF NOT EXISTS idx_tv_series_sort ON tv_series(sort_title)`,
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
	`CREATE INDEX IF NOT EXISTS idx_tv_episodes_series_season_episode ON tv_episodes(series_id, season_number, episode_number)`,
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
	`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		display_name TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at TEXT NOT NULL
	)`,
	`ALTER TABLE users ADD COLUMN username TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'standard'`,
	`ALTER TABLE users ADD COLUMN password_hash TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE users ADD COLUMN password_updated_at TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE users ADD COLUMN locked_until TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE users ADD COLUMN failed_login_count INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE users ADD COLUMN last_failed_at TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE users ADD COLUMN avatar_url TEXT NOT NULL DEFAULT ''`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE username <> ''`,
	`INSERT OR IGNORE INTO users(id, display_name, updated_at) VALUES('local', 'Local User', strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))`,
	`UPDATE users SET role = 'standard' WHERE id = 'local' AND role = ''`,
	`CREATE TABLE IF NOT EXISTS approved_devices (
		id TEXT PRIMARY KEY,
		device_id TEXT NOT NULL UNIQUE,
		device_name TEXT NOT NULL,
		client_profile TEXT NOT NULL,
		display_name TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'approved',
		approved_at TEXT NOT NULL,
		approved_by TEXT NOT NULL DEFAULT '',
		auth_session_id TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`ALTER TABLE approved_devices ADD COLUMN auth_session_id TEXT NOT NULL DEFAULT ''`,
	`CREATE INDEX IF NOT EXISTS idx_approved_devices_status_updated ON approved_devices(status, updated_at DESC)`,
	`CREATE TABLE IF NOT EXISTS pairing_requests (
		id TEXT PRIMARY KEY,
		code TEXT NOT NULL,
		device_name TEXT NOT NULL,
		client_profile TEXT NOT NULL,
		device_id TEXT NOT NULL DEFAULT '',
		auth_method TEXT NOT NULL DEFAULT '',
		auth_session_token TEXT NOT NULL DEFAULT '',
		auth_expires_at TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL,
		approved_by TEXT NOT NULL DEFAULT '',
		expires_at TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_pairing_requests_status_created ON pairing_requests(status, created_at DESC)`,
	`CREATE TABLE IF NOT EXISTS playback_states (
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		media_source_id TEXT NOT NULL REFERENCES media_sources(id) ON DELETE CASCADE,
		watched INTEGER NOT NULL DEFAULT 0,
		progress_seconds REAL NOT NULL DEFAULT 0,
		duration_seconds REAL NOT NULL DEFAULT 0,
		last_played_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		PRIMARY KEY(user_id, media_source_id)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_playback_states_recent ON playback_states(user_id, last_played_at DESC)`,
	`CREATE TABLE IF NOT EXISTS metadata_overrides (
		kind TEXT NOT NULL,
		item_id TEXT NOT NULL,
		title TEXT NOT NULL,
		year INTEGER NOT NULL DEFAULT 0,
		review_resolved INTEGER NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL,
		PRIMARY KEY(kind, item_id)
	)`,
	`CREATE TABLE IF NOT EXISTS metadata_records (
		kind TEXT NOT NULL,
		item_id TEXT NOT NULL,
		provider TEXT NOT NULL,
		external_id TEXT NOT NULL DEFAULT '',
		title TEXT NOT NULL,
		year INTEGER NOT NULL DEFAULT 0,
		overview TEXT NOT NULL DEFAULT '',
		poster_url TEXT NOT NULL DEFAULT '',
		backdrop_url TEXT NOT NULL DEFAULT '',
		thumbnail_url TEXT NOT NULL DEFAULT '',
		logo_url TEXT NOT NULL DEFAULT '',
		banner_url TEXT NOT NULL DEFAULT '',
		video_key TEXT NOT NULL DEFAULT '',
		trailer_path TEXT NOT NULL DEFAULT '',
		artwork_json TEXT NOT NULL DEFAULT '{}',
		details_json TEXT NOT NULL DEFAULT '{}',
		confidence REAL NOT NULL DEFAULT 0,
		raw_json TEXT NOT NULL DEFAULT '',
		fetched_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		PRIMARY KEY(kind, item_id, provider)
	)`,
	`ALTER TABLE metadata_records ADD COLUMN thumbnail_url TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE metadata_records ADD COLUMN logo_url TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE metadata_records ADD COLUMN banner_url TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE metadata_records ADD COLUMN artwork_json TEXT NOT NULL DEFAULT '{}'`,
	`ALTER TABLE metadata_records ADD COLUMN details_json TEXT NOT NULL DEFAULT '{}'`,
	`ALTER TABLE metadata_records ADD COLUMN video_key TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE metadata_records ADD COLUMN trailer_path TEXT NOT NULL DEFAULT ''`,
	`CREATE INDEX IF NOT EXISTS idx_metadata_records_item ON metadata_records(kind, item_id, updated_at DESC)`,
	`CREATE INDEX IF NOT EXISTS idx_metadata_records_best ON metadata_records(kind, item_id, provider, confidence DESC, updated_at DESC)`,
	`CREATE TABLE IF NOT EXISTS metadata_external_ids (
		kind TEXT NOT NULL,
		item_id TEXT NOT NULL,
		provider TEXT NOT NULL,
		external_id TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		PRIMARY KEY(kind, item_id, provider)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_metadata_external_ids_item ON metadata_external_ids(kind, item_id)`,
	`CREATE TABLE IF NOT EXISTS auth_sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		secret_hash TEXT NOT NULL,
		csrf_token TEXT NOT NULL,
		remote_addr TEXT NOT NULL DEFAULT '',
		user_agent TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		last_seen_at TEXT NOT NULL,
		expires_at TEXT NOT NULL,
		revoked_at TEXT NOT NULL DEFAULT ''
	)`,
	`CREATE INDEX IF NOT EXISTS idx_auth_sessions_user ON auth_sessions(user_id, expires_at DESC)`,
	`CREATE TABLE IF NOT EXISTS metadata_ratings (
		kind TEXT NOT NULL,
		item_id TEXT NOT NULL,
		provider TEXT NOT NULL,
		rating_type TEXT NOT NULL,
		value REAL NOT NULL DEFAULT 0,
		display_value TEXT NOT NULL DEFAULT '',
		scale REAL NOT NULL DEFAULT 0,
		votes INTEGER NOT NULL DEFAULT 0,
		source_url TEXT NOT NULL DEFAULT '',
		fetched_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		PRIMARY KEY(kind, item_id, provider, rating_type)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_metadata_ratings_item ON metadata_ratings(kind, item_id)`,
	`CREATE TABLE IF NOT EXISTS migration_runs (
		id TEXT PRIMARY KEY,
		source TEXT NOT NULL,
		schema TEXT NOT NULL,
		status TEXT NOT NULL,
		scopes_json TEXT NOT NULL,
		summary_json TEXT NOT NULL,
		verification_json TEXT NOT NULL,
		created_at TEXT NOT NULL,
		completed_at TEXT NOT NULL DEFAULT '',
		rolled_back_at TEXT NOT NULL DEFAULT '',
		error_text TEXT NOT NULL DEFAULT ''
	)`,
	`CREATE INDEX IF NOT EXISTS idx_migration_runs_created ON migration_runs(created_at DESC)`,
	`CREATE TABLE IF NOT EXISTS migration_run_items (
		run_id TEXT NOT NULL REFERENCES migration_runs(id) ON DELETE CASCADE,
		import_key TEXT NOT NULL,
		item_json TEXT NOT NULL,
		PRIMARY KEY(run_id, import_key)
	)`,
	`CREATE TABLE IF NOT EXISTS migration_backups (
		run_id TEXT NOT NULL REFERENCES migration_runs(id) ON DELETE CASCADE,
		scope TEXT NOT NULL,
		target_kind TEXT NOT NULL,
		target_id TEXT NOT NULL,
		provider TEXT NOT NULL DEFAULT '',
		media_source_id TEXT NOT NULL DEFAULT '',
		backup_json TEXT NOT NULL,
		PRIMARY KEY(run_id, scope, target_kind, target_id, provider, media_source_id)
	)`,
	`CREATE TABLE IF NOT EXISTS scan_file_state (
		library_id TEXT NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
		rel_path TEXT NOT NULL,
		size_bytes INTEGER NOT NULL,
		modified_at TEXT NOT NULL,
		last_seen_at TEXT NOT NULL,
		changed_at TEXT NOT NULL,
		PRIMARY KEY(library_id, rel_path)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_scan_file_state_seen ON scan_file_state(library_id, last_seen_at DESC)`,
	`CREATE TABLE IF NOT EXISTS runtime_entities (
		entity_type TEXT NOT NULL,
		id TEXT NOT NULL,
		status TEXT NOT NULL,
		payload_json TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		heartbeat_at TEXT NOT NULL,
		completed_at TEXT NOT NULL DEFAULT '',
		PRIMARY KEY(entity_type, id)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_runtime_entities_type_status ON runtime_entities(entity_type, status, updated_at DESC)`,
	`CREATE INDEX IF NOT EXISTS idx_runtime_entities_heartbeat ON runtime_entities(entity_type, heartbeat_at)`,
	`INSERT OR IGNORE INTO metadata_records(kind, item_id, provider, title, year, confidence, fetched_at, updated_at)
		SELECT 'movie', id, 'filename', title, year, CASE needs_review WHEN 1 THEN 0.35 ELSE 0.7 END, updated_at, updated_at
		FROM movies`,
	`INSERT OR IGNORE INTO metadata_records(kind, item_id, provider, title, confidence, fetched_at, updated_at)
		SELECT 'series', id, 'filename', title, 0.7, updated_at, updated_at
		FROM tv_series`,
	`INSERT OR IGNORE INTO metadata_records(kind, item_id, provider, title, confidence, fetched_at, updated_at)
		SELECT 'episode', id, 'filename', title, CASE needs_review WHEN 1 THEN 0.35 ELSE 0.7 END, updated_at, updated_at
		FROM tv_episodes`,
	// issue #56: enrich probe data with profile/level/bit-depth/HDR/frame-rate
	// so the playback decision tree can do more than coarse codec-name matching.
	`ALTER TABLE media_probes ADD COLUMN video_profile TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE media_probes ADD COLUMN video_level TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE media_probes ADD COLUMN video_bit_depth INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE media_probes ADD COLUMN video_frame_rate REAL NOT NULL DEFAULT 0`,
	`ALTER TABLE media_probes ADD COLUMN pixel_format TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE media_probes ADD COLUMN color_primaries TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE media_probes ADD COLUMN color_transfer TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE media_probes ADD COLUMN color_space TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE media_probes ADD COLUMN hdr_format TEXT NOT NULL DEFAULT ''`,
	// issue #61: Dolby Vision detection + HDR10 MaxCLL/MaxFALL metadata
	`ALTER TABLE media_probes ADD COLUMN dovi_profile INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE media_probes ADD COLUMN max_cll INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE media_probes ADD COLUMN max_fall INTEGER NOT NULL DEFAULT 0`,
	// issue #89: in-app notification feed
	`CREATE TABLE IF NOT EXISTS notifications (
		id TEXT PRIMARY KEY,
		kind TEXT NOT NULL,
		title TEXT NOT NULL,
		message TEXT NOT NULL DEFAULT '',
		link TEXT NOT NULL DEFAULT '',
		dismissed INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	)`,
	`CREATE INDEX IF NOT EXISTS idx_notifications_dismissed_created ON notifications(dismissed, created_at DESC)`,
	// issue #85: materialized people index for fast search
	`CREATE TABLE IF NOT EXISTS people (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		name_lower TEXT NOT NULL,
		profile_url TEXT NOT NULL DEFAULT '',
		department TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	)`,
	`CREATE TABLE IF NOT EXISTS people_credits (
		person_id TEXT NOT NULL,
		item_kind TEXT NOT NULL,
		item_id TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT '',
		character TEXT NOT NULL DEFAULT '',
		PRIMARY KEY (person_id, item_kind, item_id, role)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_people_name_lower ON people(name_lower)`,
	// issue #82: chapter markers (intro/credits) per media source
	`CREATE TABLE IF NOT EXISTS media_source_chapters (
		media_source_id TEXT PRIMARY KEY REFERENCES media_sources(id) ON DELETE CASCADE,
		intro_start REAL NOT NULL DEFAULT -1,
		intro_end REAL NOT NULL DEFAULT -1,
		credits_start REAL NOT NULL DEFAULT -1,
		analyzed_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	// issue #82: per-user preferences (auto-skip intros etc.)
	`ALTER TABLE users ADD COLUMN preferences_json TEXT NOT NULL DEFAULT '{}'`,
	// issue #83: household profiles — PIN, restriction flag, rating ceiling, avatar preset/colour
	`ALTER TABLE users ADD COLUMN pin_hash TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE users ADD COLUMN is_restricted INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE users ADD COLUMN max_rating TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE users ADD COLUMN avatar_preset TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE users ADD COLUMN avatar_color TEXT NOT NULL DEFAULT ''`,
	// issue #83: profile sessions — lightweight switch-profile tokens scoped to a main session
	`CREATE TABLE IF NOT EXISTS profile_sessions (
		token TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		profile_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TEXT NOT NULL,
		expires_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_profile_sessions_session ON profile_sessions(session_id)`,
	`CREATE INDEX IF NOT EXISTS idx_profile_sessions_user ON profile_sessions(profile_user_id)`,
}
