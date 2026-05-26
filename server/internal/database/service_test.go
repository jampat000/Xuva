package database

import (
	"context"
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
	if migrationCount != 1 {
		t.Fatalf("expected migration ledger to stay stable, got %d rows", migrationCount)
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
	if version != "0001_legacy_inline_schema" {
		t.Fatalf("expected latest schema version, got %q", version)
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
