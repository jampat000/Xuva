package database

import (
	"context"
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
}
