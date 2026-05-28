package database

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenCreatesDatabaseAndMigratesSchema(t *testing.T) {
	service, err := Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer service.Close()

	var tableCount int
	err = service.DB().QueryRowContext(context.Background(), `
		SELECT count(*)
		FROM sqlite_master
		WHERE type = 'table'
		AND name IN ('libraries', 'media_sources', 'movies', 'tv_series', 'scan_runs')
	`).Scan(&tableCount)
	if err != nil {
		t.Fatalf("query schema: %v", err)
	}
	if tableCount != 5 {
		t.Fatalf("expected core tables to exist, got %d", tableCount)
	}

	var migrationCount int
	err = service.DB().QueryRowContext(context.Background(), `
		SELECT count(*)
		FROM schema_migrations
		WHERE id = '0001_legacy_inline_schema'
	`).Scan(&migrationCount)
	if err != nil {
		t.Fatalf("query schema migration ledger: %v", err)
	}
	if migrationCount != 1 {
		t.Fatalf("expected legacy migration ledger row, got %d", migrationCount)
	}
}

func TestOpenReusesSchemaMigrationLedger(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	service, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := service.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	reopened, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	defer reopened.Close()

	var migrationCount int
	if err := reopened.DB().QueryRowContext(ctx, `SELECT count(*) FROM schema_migrations`).Scan(&migrationCount); err != nil {
		t.Fatalf("query schema migrations: %v", err)
	}
	if migrationCount != len(schemaMigrations) {
		t.Fatalf("expected migration ledger to stay stable at %d rows, got %d",
			len(schemaMigrations), migrationCount)
	}
}

func TestSchemaVersionReturnsLatestMigrationID(t *testing.T) {
	ctx := context.Background()
	service, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer service.Close()

	version, err := service.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("schema version: %v", err)
	}
	want := schemaMigrations[len(schemaMigrations)-1].ID
	if version != want {
		t.Fatalf("expected latest schema version %q, got %q", want, version)
	}
}

func TestIntegrityCheckPassesForMigratedDatabase(t *testing.T) {
	ctx := context.Background()
	service, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer service.Close()

	if err := service.integrityCheck(ctx, "test"); err != nil {
		t.Fatalf("integrity check: %v", err)
	}
}

func TestOpenBacksUpPreexistingDatabaseBeforeFirstSchemaMigration(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	raw, err := sql.Open("sqlite", sqliteDSN(filepath.Join(dir, "xuva.db")))
	if err != nil {
		t.Fatalf("open raw database: %v", err)
	}
	if _, err := raw.ExecContext(ctx, `CREATE TABLE preexisting_marker(value TEXT)`); err != nil {
		t.Fatalf("create preexisting marker: %v", err)
	}
	if err := raw.Close(); err != nil {
		t.Fatalf("close raw database: %v", err)
	}

	service, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("open migrated database: %v", err)
	}
	defer service.Close()

	backups, err := filepath.Glob(filepath.Join(dir, "backups", "schema-upgrade-*.db"))
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected one schema backup, got %d: %#v", len(backups), backups)
	}

	backupDB, err := sql.Open("sqlite", sqliteDSN(backups[0]))
	if err != nil {
		t.Fatalf("open backup database: %v", err)
	}
	defer backupDB.Close()
	var markerCount int
	if err := backupDB.QueryRowContext(ctx, `
		SELECT count(*)
		FROM sqlite_master
		WHERE type = 'table'
		AND name = 'preexisting_marker'
	`).Scan(&markerCount); err != nil {
		t.Fatalf("query backup marker: %v", err)
	}
	if markerCount != 1 {
		t.Fatalf("expected preexisting marker in backup, got %d", markerCount)
	}
	var ledgerCount int
	if err := backupDB.QueryRowContext(ctx, `
		SELECT count(*)
		FROM sqlite_master
		WHERE type = 'table'
		AND name = 'schema_migrations'
	`).Scan(&ledgerCount); err != nil {
		t.Fatalf("query backup ledger: %v", err)
	}
	if ledgerCount != 0 {
		t.Fatalf("expected backup before migration ledger creation, got %d ledger tables", ledgerCount)
	}
}

func TestOpenDoesNotBackUpBrandNewDatabase(t *testing.T) {
	service, err := Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer service.Close()

	if _, err := os.Stat(filepath.Join(service.DataDir, "backups")); !os.IsNotExist(err) {
		t.Fatalf("expected no backups directory for new database, got err=%v", err)
	}
}

func TestOpenRejectsSchemaMigrationChecksumMismatch(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	service, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if _, err := service.DB().ExecContext(ctx, `
		UPDATE schema_migrations
		SET checksum = 'bad'
		WHERE id = '0001_legacy_inline_schema'
	`); err != nil {
		t.Fatalf("corrupt schema migration checksum: %v", err)
	}
	if err := service.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	_, err = Open(ctx, dir)
	if err == nil {
		t.Fatalf("expected checksum mismatch")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch error, got %v", err)
	}
}
