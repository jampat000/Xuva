package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Service struct {
	DataDir     string
	Path        string
	db          *sql.DB
	preexisting bool
}

func Open(ctx context.Context, dataDir string) (*Service, error) {
	if dataDir == "" {
		dataDir = "data"
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dataDir, "xuva.db")
	_, statErr := os.Stat(dbPath)
	preexisting := statErr == nil
	if statErr != nil && !os.IsNotExist(statErr) {
		return nil, statErr
	}
	db, err := sql.Open("sqlite", sqliteDSN(dbPath))
	if err != nil {
		return nil, err
	}

	service := &Service{
		DataDir:     dataDir,
		Path:        dbPath,
		db:          db,
		preexisting: preexisting,
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

func sqliteDSN(dbPath string) string {
	q := url.Values{}
	q.Add("_pragma", "busy_timeout=5000")
	q.Add("_pragma", "foreign_keys=ON")
	q.Add("_pragma", "journal_mode=WAL")
	q.Add("_pragma", "synchronous=NORMAL")
	return dbPath + "?" + q.Encode()
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

func (s *Service) SchemaVersion(ctx context.Context) (string, error) {
	if s == nil || s.db == nil {
		return "", nil
	}
	var id string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM schema_migrations ORDER BY id DESC LIMIT 1`).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return id, err
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
	if err := s.integrityCheck(ctx, "before schema migrations"); err != nil {
		return err
	}
	backupCreated := false
	if s.preexisting {
		ledgerExists, err := s.schemaMigrationLedgerExists(ctx)
		if err != nil {
			return err
		}
		if !ledgerExists {
			if err := s.backupBeforeSchemaMigration(ctx); err != nil {
				return err
			}
			backupCreated = true
		}
	}
	if err := s.ensureMigrationLedger(ctx); err != nil {
		return err
	}
	for _, migration := range schemaMigrations {
		checksum := migration.checksum()
		applied, err := s.migrationApplied(ctx, migration.ID, checksum)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if s.preexisting && !backupCreated {
			if err := s.backupBeforeSchemaMigration(ctx); err != nil {
				return err
			}
			backupCreated = true
		}
		if err := s.applyMigration(ctx, migration, checksum); err != nil {
			return err
		}
	}
	if err := s.integrityCheck(ctx, "after schema migrations"); err != nil {
		return err
	}
	return nil
}

func (s *Service) schemaMigrationLedgerExists(ctx context.Context) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM sqlite_master
		WHERE type = 'table'
		AND name = 'schema_migrations'
	`).Scan(&count)
	return count > 0, err
}

type schemaMigration struct {
	ID                   string
	Name                 string
	Statements           []string
	AllowDuplicateColumn bool
}

func (m schemaMigration) checksum() string {
	hash := sha256.New()
	for _, statement := range m.Statements {
		_, _ = hash.Write([]byte(statement))
		_, _ = hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *Service) ensureMigrationLedger(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		checksum TEXT NOT NULL,
		applied_at TEXT NOT NULL,
		duration_ms INTEGER NOT NULL
	)`)
	return err
}

func (s *Service) migrationApplied(ctx context.Context, id string, checksum string) (bool, error) {
	var stored string
	err := s.db.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE id = ?`, id).Scan(&stored)
	if err == nil {
		if stored != checksum {
			return false, fmt.Errorf("schema migration %s checksum mismatch", id)
		}
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}

func (s *Service) applyMigration(ctx context.Context, migration schemaMigration, checksum string) error {
	start := time.Now()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, statement := range migration.Statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			if migration.AllowDuplicateColumn && strings.Contains(err.Error(), "duplicate column name") {
				continue
			}
			return fmt.Errorf("apply schema migration %s: %w", migration.ID, err)
		}
	}
	duration := time.Since(start).Milliseconds()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO schema_migrations(id, name, checksum, applied_at, duration_ms)
		VALUES(?, ?, ?, ?, ?)
	`, migration.ID, migration.Name, checksum, time.Now().UTC().Format(time.RFC3339Nano), duration); err != nil {
		return fmt.Errorf("record schema migration %s: %w", migration.ID, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit schema migration %s: %w", migration.ID, err)
	}
	return nil
}

func (s *Service) backupBeforeSchemaMigration(ctx context.Context) error {
	backupDir := filepath.Join(s.DataDir, "backups")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return err
	}
	backupPath := filepath.Join(backupDir, "schema-upgrade-"+time.Now().UTC().Format("20060102T150405Z")+".db")
	if _, err := s.db.ExecContext(ctx, "VACUUM INTO ?", backupPath); err != nil {
		return fmt.Errorf("backup database before schema migration: %w", err)
	}
	return nil
}

func (s *Service) integrityCheck(ctx context.Context, phase string) error {
	rows, err := s.db.QueryContext(ctx, `PRAGMA integrity_check`)
	if err != nil {
		return fmt.Errorf("sqlite integrity check %s: %w", phase, err)
	}
	defer rows.Close()

	results := []string{}
	for rows.Next() {
		var result string
		if err := rows.Scan(&result); err != nil {
			return fmt.Errorf("sqlite integrity check %s: %w", phase, err)
		}
		results = append(results, strings.TrimSpace(result))
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("sqlite integrity check %s: %w", phase, err)
	}
	if len(results) == 1 && strings.EqualFold(results[0], "ok") {
		return nil
	}
	return fmt.Errorf("sqlite integrity check %s failed: %s", phase, strings.Join(results, "; "))
}

var schemaMigrations = []schemaMigration{
	{
		ID:                   "0001_legacy_inline_schema",
		Name:                 "Legacy inline schema",
		Statements:           migrations,
		AllowDuplicateColumn: true,
	},
	{
		ID:         "0002_movies_list_view_snapshot",
		Name:       "Denormalized movies list-view + maintenance triggers",
		Statements: moviesListViewMigration,
	},
	{
		ID:         "0003_tv_series_list_view_snapshot",
		Name:       "Denormalized tv_series list-view + maintenance triggers",
		Statements: tvSeriesListViewMigration,
	},
	{
		ID:         "0004_metadata_collection_index",
		Name:       "Partial expression index for collection lookups in metadata_records",
		Statements: metadataCollectionIndexMigration,
	},
}

// metadataCollectionIndexMigration adds a partial expression index on
// metadata_records for the collection JSON path used by ListCollections /
// ListMoviesByCollection. Before this index, both queries did a full table
// scan with json_extract evaluated per-row on every metadata_records row
// (≈3× the movie count once filename + TMDB + wiki providers all land). The
// partial WHERE clause limits the index to rows that actually have a
// collection — typically ~10% of metadata_records — keeping index size small.
//
// The SQLite planner matches expression indexes by syntactic equality, so the
// query text must use the same `json_extract(details_json, '$.collection.id')`
// form. It does in both call sites.
var metadataCollectionIndexMigration = []string{
	`CREATE INDEX IF NOT EXISTS idx_metadata_records_collection
		ON metadata_records (
			kind,
			json_extract(details_json, '$.collection.id')
		)
		WHERE json_extract(details_json, '$.collection.id') IS NOT NULL`,
}

// tvSeriesListViewMigration is the sibling of moviesListViewMigration for
// tv_series. ListSeries had the same join-and-group-everything-then-LIMIT
// pattern (tv_series × tv_seasons × tv_episodes × episode_versions ×
// playback_states), so it gets the same fix: a snapshot table indexed on
// (sort_title, id) with triggers maintaining season_count + episode_count.
//
// As with movies_list_view, per-user watched state stays out of the snapshot
// and is layered in by ListSeries via a correlated subquery.
var tvSeriesListViewMigration = []string{
	`CREATE TABLE IF NOT EXISTS tv_series_list_view (
		series_id     TEXT PRIMARY KEY REFERENCES tv_series(id) ON DELETE CASCADE,
		title         TEXT NOT NULL,
		sort_title    TEXT NOT NULL,
		season_count  INTEGER NOT NULL DEFAULT 0,
		episode_count INTEGER NOT NULL DEFAULT 0,
		created_at    TEXT NOT NULL DEFAULT ''
	)`,
	`CREATE INDEX IF NOT EXISTS idx_tv_series_list_view_sort ON tv_series_list_view(sort_title, series_id)`,
	`DROP TRIGGER IF EXISTS tv_series_list_view_series_ai`,
	`CREATE TRIGGER tv_series_list_view_series_ai
		AFTER INSERT ON tv_series
		BEGIN
			INSERT INTO tv_series_list_view(series_id, title, sort_title, season_count, episode_count, created_at)
			VALUES (NEW.id, NEW.title, NEW.sort_title, 0, 0, NEW.created_at)
			ON CONFLICT(series_id) DO UPDATE SET
				title = excluded.title,
				sort_title = excluded.sort_title,
				created_at = excluded.created_at;
		END`,
	`DROP TRIGGER IF EXISTS tv_series_list_view_series_au`,
	`CREATE TRIGGER tv_series_list_view_series_au
		AFTER UPDATE ON tv_series
		BEGIN
			INSERT INTO tv_series_list_view(series_id, title, sort_title, season_count, episode_count, created_at)
			VALUES (NEW.id, NEW.title, NEW.sort_title, 0, 0, NEW.created_at)
			ON CONFLICT(series_id) DO UPDATE SET
				title = excluded.title,
				sort_title = excluded.sort_title;
		END`,
	// tv_seasons → snapshot.season_count
	`DROP TRIGGER IF EXISTS tv_series_list_view_seasons_ai`,
	`CREATE TRIGGER tv_series_list_view_seasons_ai
		AFTER INSERT ON tv_seasons
		BEGIN
			UPDATE tv_series_list_view
			SET season_count = (SELECT COUNT(*) FROM tv_seasons WHERE series_id = NEW.series_id)
			WHERE series_id = NEW.series_id;
		END`,
	`DROP TRIGGER IF EXISTS tv_series_list_view_seasons_ad`,
	`CREATE TRIGGER tv_series_list_view_seasons_ad
		AFTER DELETE ON tv_seasons
		BEGIN
			UPDATE tv_series_list_view
			SET season_count = (SELECT COUNT(*) FROM tv_seasons WHERE series_id = OLD.series_id)
			WHERE series_id = OLD.series_id;
		END`,
	// tv_episodes → snapshot.episode_count
	`DROP TRIGGER IF EXISTS tv_series_list_view_episodes_ai`,
	`CREATE TRIGGER tv_series_list_view_episodes_ai
		AFTER INSERT ON tv_episodes
		BEGIN
			UPDATE tv_series_list_view
			SET episode_count = (SELECT COUNT(*) FROM tv_episodes WHERE series_id = NEW.series_id)
			WHERE series_id = NEW.series_id;
		END`,
	`DROP TRIGGER IF EXISTS tv_series_list_view_episodes_ad`,
	`CREATE TRIGGER tv_series_list_view_episodes_ad
		AFTER DELETE ON tv_episodes
		BEGIN
			UPDATE tv_series_list_view
			SET episode_count = (SELECT COUNT(*) FROM tv_episodes WHERE series_id = OLD.series_id)
			WHERE series_id = OLD.series_id;
		END`,
	// Backfill from existing data.
	`INSERT OR REPLACE INTO tv_series_list_view (series_id, title, sort_title, season_count, episode_count, created_at)
		SELECT s.id, s.title, s.sort_title,
		       (SELECT COUNT(*) FROM tv_seasons ts WHERE ts.series_id = s.id) AS season_count,
		       (SELECT COUNT(*) FROM tv_episodes te WHERE te.series_id = s.id) AS episode_count,
		       s.created_at
		FROM tv_series s`,
}

// moviesListViewMigration creates the denormalized snapshot table that backs
// /api/movies. The old ListMovies query joined movies × movie_versions ×
// media_probes × playback_states with a GROUP BY m.id over the entire table
// before LIMIT could clip it — ~2.4s on a 4000-item library.
//
// The new shape:
//  1. One row per movie in movies_list_view, sorted by (sort_title, year) via
//     an index, so the LIMIT-driven page can be served as a single seq read.
//  2. AFTER INSERT/UPDATE/DELETE triggers on movies/movie_versions/media_probes
//     keep the snapshot in lock-step with the source-of-truth tables inside
//     the same transaction as the originating write. Zero writer code changes.
//  3. Per-user watched state stays out of the snapshot (it's per-user, can't
//     denormalize cleanly). ListMovies layers it in via a correlated subquery
//     after LIMIT, so we only compute it for the page being returned.
//
// Backfill runs as the final statement: for an empty view, populate from the
// existing data. Idempotent — re-running the migration would be a no-op since
// the migration ledger ID already exists, but the body uses INSERT OR REPLACE
// regardless so a manual rebuild stays safe.
var moviesListViewMigration = []string{
	`CREATE TABLE IF NOT EXISTS movies_list_view (
		movie_id      TEXT PRIMARY KEY REFERENCES movies(id) ON DELETE CASCADE,
		title         TEXT NOT NULL,
		year          INTEGER NOT NULL DEFAULT 0,
		sort_title    TEXT NOT NULL,
		needs_review  INTEGER NOT NULL DEFAULT 0,
		version_count INTEGER NOT NULL DEFAULT 0,
		is_probed     INTEGER NOT NULL DEFAULT 0,
		created_at    TEXT NOT NULL DEFAULT ''
	)`,
	`CREATE INDEX IF NOT EXISTS idx_movies_list_view_sort ON movies_list_view(sort_title, year)`,
	// movies → snapshot: keep title / year / sort_title / needs_review in sync.
	// ON DELETE CASCADE via FK handles row removal automatically.
	`DROP TRIGGER IF EXISTS movies_list_view_movies_ai`,
	`CREATE TRIGGER movies_list_view_movies_ai
		AFTER INSERT ON movies
		BEGIN
			INSERT INTO movies_list_view(movie_id, title, year, sort_title, needs_review, version_count, is_probed, created_at)
			VALUES (NEW.id, NEW.title, NEW.year, NEW.sort_title, NEW.needs_review, 0, 0, NEW.created_at)
			ON CONFLICT(movie_id) DO UPDATE SET
				title = excluded.title,
				year = excluded.year,
				sort_title = excluded.sort_title,
				needs_review = excluded.needs_review,
				created_at = excluded.created_at;
		END`,
	`DROP TRIGGER IF EXISTS movies_list_view_movies_au`,
	`CREATE TRIGGER movies_list_view_movies_au
		AFTER UPDATE ON movies
		BEGIN
			INSERT INTO movies_list_view(movie_id, title, year, sort_title, needs_review, version_count, is_probed, created_at)
			VALUES (NEW.id, NEW.title, NEW.year, NEW.sort_title, NEW.needs_review, 0, 0, NEW.created_at)
			ON CONFLICT(movie_id) DO UPDATE SET
				title = excluded.title,
				year = excluded.year,
				sort_title = excluded.sort_title,
				needs_review = excluded.needs_review;
		END`,
	// movie_versions → snapshot: recompute version_count + is_probed for the
	// affected movie. is_probed depends on media_probes existing for ANY of the
	// movie's versions, so it needs a full recompute on each version change.
	`DROP TRIGGER IF EXISTS movies_list_view_versions_ai`,
	`CREATE TRIGGER movies_list_view_versions_ai
		AFTER INSERT ON movie_versions
		BEGIN
			UPDATE movies_list_view
			SET version_count = (SELECT COUNT(*) FROM movie_versions WHERE movie_id = NEW.movie_id),
			    is_probed = CASE WHEN EXISTS (
			        SELECT 1 FROM movie_versions mv2
			        JOIN media_probes mp ON mp.media_source_id = mv2.media_source_id
			        WHERE mv2.movie_id = NEW.movie_id
			    ) THEN 1 ELSE 0 END
			WHERE movie_id = NEW.movie_id;
		END`,
	`DROP TRIGGER IF EXISTS movies_list_view_versions_ad`,
	`CREATE TRIGGER movies_list_view_versions_ad
		AFTER DELETE ON movie_versions
		BEGIN
			UPDATE movies_list_view
			SET version_count = (SELECT COUNT(*) FROM movie_versions WHERE movie_id = OLD.movie_id),
			    is_probed = CASE WHEN EXISTS (
			        SELECT 1 FROM movie_versions mv2
			        JOIN media_probes mp ON mp.media_source_id = mv2.media_source_id
			        WHERE mv2.movie_id = OLD.movie_id
			    ) THEN 1 ELSE 0 END
			WHERE movie_id = OLD.movie_id;
		END`,
	// media_probes → snapshot: any movie whose movie_versions reference this
	// media_source_id may need its is_probed flag flipped.
	`DROP TRIGGER IF EXISTS movies_list_view_probes_ai`,
	`CREATE TRIGGER movies_list_view_probes_ai
		AFTER INSERT ON media_probes
		BEGIN
			UPDATE movies_list_view
			SET is_probed = 1
			WHERE movie_id IN (SELECT movie_id FROM movie_versions WHERE media_source_id = NEW.media_source_id);
		END`,
	`DROP TRIGGER IF EXISTS movies_list_view_probes_ad`,
	`CREATE TRIGGER movies_list_view_probes_ad
		AFTER DELETE ON media_probes
		BEGIN
			UPDATE movies_list_view
			SET is_probed = CASE WHEN EXISTS (
				SELECT 1 FROM movie_versions mv2
				JOIN media_probes mp ON mp.media_source_id = mv2.media_source_id
				WHERE mv2.movie_id = movies_list_view.movie_id
			) THEN 1 ELSE 0 END
			WHERE movie_id IN (SELECT movie_id FROM movie_versions WHERE media_source_id = OLD.media_source_id);
		END`,
	// Backfill: populate the view from existing data. INSERT OR REPLACE so a
	// manual rebuild (DELETE FROM movies_list_view; re-run this stmt) stays safe.
	`INSERT OR REPLACE INTO movies_list_view (movie_id, title, year, sort_title, needs_review, version_count, is_probed, created_at)
		SELECT m.id, m.title, m.year, m.sort_title, m.needs_review,
		       (SELECT COUNT(*) FROM movie_versions mv WHERE mv.movie_id = m.id) AS version_count,
		       CASE WHEN EXISTS (
		           SELECT 1 FROM movie_versions mv
		           JOIN media_probes mp ON mp.media_source_id = mv.media_source_id
		           WHERE mv.movie_id = m.id
		       ) THEN 1 ELSE 0 END AS is_probed,
		       m.created_at
		FROM movies m`,
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
		device_id TEXT NOT NULL DEFAULT '',
		remote_addr TEXT NOT NULL DEFAULT '',
		user_agent TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		last_seen_at TEXT NOT NULL,
		expires_at TEXT NOT NULL,
		revoked_at TEXT NOT NULL DEFAULT ''
	)`,
	`ALTER TABLE auth_sessions ADD COLUMN device_id TEXT NOT NULL DEFAULT ''`,
	`CREATE INDEX IF NOT EXISTS idx_auth_sessions_device ON auth_sessions(device_id, expires_at DESC)`,
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
	// issue #184: QR pair tokens — short-lived pre-approved pairing codes
	`CREATE TABLE IF NOT EXISTS pair_tokens (
		token       TEXT PRIMARY KEY,
		created_by  TEXT NOT NULL DEFAULT '',
		claimed_by  TEXT NOT NULL DEFAULT '',
		claimed_at  TEXT NOT NULL DEFAULT '',
		expires_at  TEXT NOT NULL,
		created_at  TEXT NOT NULL
	)`,
	// issue #174: server-side watchlist sync
	`CREATE TABLE IF NOT EXISTS watchlist_items (
		user_id      TEXT NOT NULL,
		media_id     TEXT NOT NULL,
		kind         TEXT NOT NULL CHECK(kind IN ('movie','series')),
		title        TEXT NOT NULL,
		year         INTEGER,
		poster_url   TEXT NOT NULL DEFAULT '',
		backdrop_url TEXT NOT NULL DEFAULT '',
		genres_json  TEXT NOT NULL DEFAULT '',
		added_at     TEXT NOT NULL,
		PRIMARY KEY (user_id, media_id, kind)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_watchlist_user ON watchlist_items(user_id, added_at DESC)`,
}
