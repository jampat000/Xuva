package database

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

// Helpers for the snapshot tests. Real callers go through catalog.Service, but
// these tests only care about the trigger behaviour, so we INSERT into the raw
// schema directly to avoid coupling to the catalog package.

func newSnapshotTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	dir := t.TempDir()
	svc, err := Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	cleanup := func() { _ = svc.Close() }
	// Seed a library + media_sources we can reference. Trigger maintenance for
	// is_probed requires media_probes to be FK-valid against media_sources.
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := svc.DB().ExecContext(ctx, `
		INSERT INTO libraries(id, kind, name, path, updated_at)
		VALUES('lib1','movies','Movies','/lib1',?)
	`, now); err != nil {
		t.Fatalf("seed library: %v", err)
	}
	return svc.DB(), cleanup
}

func insertMovie(t *testing.T, db *sql.DB, id, title, sortTitle string, year int, needsReview int) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO movies(id, title, year, sort_title, needs_review, review_reason, updated_at)
		VALUES(?,?,?,?,?,?,?)
	`, id, title, year, sortTitle, needsReview, "", now); err != nil {
		t.Fatalf("insert movie %s: %v", id, err)
	}
}

func insertMediaSource(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO media_sources(id, library_id, kind, path, rel_path, name, extension,
		    size_bytes, modified_at, discovered_at, updated_at)
		VALUES(?,'lib1','movie',?,?,?, 'mkv', 0,?,?,?)
	`, id, "/lib1/"+id+".mkv", id+".mkv", id, now, now, now); err != nil {
		t.Fatalf("insert media_source %s: %v", id, err)
	}
}

func insertMovieVersion(t *testing.T, db *sql.DB, movieID, sourceID string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO movie_versions(movie_id, media_source_id, edition, quality_label, updated_at)
		VALUES(?,?, '', '',?)
	`, movieID, sourceID, now); err != nil {
		t.Fatalf("insert movie_version: %v", err)
	}
}

func insertMediaProbe(t *testing.T, db *sql.DB, sourceID string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO media_probes(media_source_id, container, duration_seconds, bitrate,
		    video_codec, width, height, audio_streams, subtitle_streams, raw_json, probed_at)
		VALUES(?, 'mkv', 0, 0, 'h264', 1920, 1080, 1, 0, '{}',?)
	`, sourceID, now); err != nil {
		t.Fatalf("insert media_probe: %v", err)
	}
}

func snapshotRow(t *testing.T, db *sql.DB, movieID string) (title string, year, versionCount, isProbed, needsReview int, found bool) {
	t.Helper()
	err := db.QueryRowContext(context.Background(), `
		SELECT title, year, version_count, is_probed, needs_review
		FROM movies_list_view WHERE movie_id = ?
	`, movieID).Scan(&title, &year, &versionCount, &isProbed, &needsReview)
	if err == sql.ErrNoRows {
		return "", 0, 0, 0, 0, false
	}
	if err != nil {
		t.Fatalf("read snapshot row %s: %v", movieID, err)
	}
	return title, year, versionCount, isProbed, needsReview, true
}

// TestMoviesListView_InsertCreatesSnapshotRow — the AFTER INSERT trigger on
// movies should produce a matching row with zero version_count / is_probed
// (the joins haven't happened yet).
func TestMoviesListView_InsertCreatesSnapshotRow(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertMovie(t, db, "m1", "Inception", "inception", 2010, 0)

	title, year, vc, probed, nr, ok := snapshotRow(t, db, "m1")
	if !ok {
		t.Fatal("movies_list_view row not created by trigger")
	}
	if title != "Inception" || year != 2010 || vc != 0 || probed != 0 || nr != 0 {
		t.Errorf("snapshot row mismatch: title=%q year=%d vc=%d probed=%d nr=%d",
			title, year, vc, probed, nr)
	}
}

// TestMoviesListView_UpdatePropagatesToSnapshot — changing title/year/needs_review
// on movies should flow through to the view.
func TestMoviesListView_UpdatePropagatesToSnapshot(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertMovie(t, db, "m1", "Wrong Title", "wrong", 1999, 1)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.ExecContext(context.Background(), `
		UPDATE movies SET title='Inception', sort_title='inception', year=2010,
		    needs_review=0, updated_at=? WHERE id='m1'
	`, now); err != nil {
		t.Fatalf("update movie: %v", err)
	}

	title, year, _, _, nr, ok := snapshotRow(t, db, "m1")
	if !ok {
		t.Fatal("snapshot row vanished after UPDATE")
	}
	if title != "Inception" || year != 2010 || nr != 0 {
		t.Errorf("snapshot didn't follow UPDATE: title=%q year=%d nr=%d", title, year, nr)
	}
}

// TestMoviesListView_DeleteCascadesViaFK — DELETE FROM movies removes the
// snapshot row via ON DELETE CASCADE.
func TestMoviesListView_DeleteCascadesViaFK(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertMovie(t, db, "m1", "Inception", "inception", 2010, 0)
	if _, err := db.ExecContext(context.Background(), `DELETE FROM movies WHERE id='m1'`); err != nil {
		t.Fatalf("delete movie: %v", err)
	}
	if _, _, _, _, _, ok := snapshotRow(t, db, "m1"); ok {
		t.Fatal("snapshot row should have been deleted with parent movie")
	}
}

// TestMoviesListView_VersionInsertBumpsCount — adding a movie_versions row
// should bump version_count in the snapshot.
func TestMoviesListView_VersionInsertBumpsCount(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertMovie(t, db, "m1", "Inception", "inception", 2010, 0)
	insertMediaSource(t, db, "ms1")
	insertMovieVersion(t, db, "m1", "ms1")

	_, _, vc, probed, _, ok := snapshotRow(t, db, "m1")
	if !ok {
		t.Fatal("snapshot row missing")
	}
	if vc != 1 {
		t.Errorf("expected version_count=1 after one INSERT, got %d", vc)
	}
	if probed != 0 {
		t.Errorf("is_probed should still be 0 (no media_probes row yet), got %d", probed)
	}

	insertMediaSource(t, db, "ms2")
	insertMovieVersion(t, db, "m1", "ms2")
	_, _, vc, _, _, _ = snapshotRow(t, db, "m1")
	if vc != 2 {
		t.Errorf("expected version_count=2 after second INSERT, got %d", vc)
	}
}

// TestMoviesListView_ProbeFlipsIsProbed — inserting a media_probes row for
// any of a movie's versions should flip is_probed to 1.
func TestMoviesListView_ProbeFlipsIsProbed(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertMovie(t, db, "m1", "Inception", "inception", 2010, 0)
	insertMediaSource(t, db, "ms1")
	insertMovieVersion(t, db, "m1", "ms1")

	insertMediaProbe(t, db, "ms1")

	_, _, _, probed, _, _ := snapshotRow(t, db, "m1")
	if probed != 1 {
		t.Errorf("is_probed should be 1 after media_probes INSERT, got %d", probed)
	}
}

// TestMoviesListView_ProbeDeleteRevertsIsProbedWhenLast — deleting the only
// probe row that backs a movie's versions should flip is_probed back to 0.
func TestMoviesListView_ProbeDeleteRevertsIsProbedWhenLast(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertMovie(t, db, "m1", "Inception", "inception", 2010, 0)
	insertMediaSource(t, db, "ms1")
	insertMediaSource(t, db, "ms2")
	insertMovieVersion(t, db, "m1", "ms1")
	insertMovieVersion(t, db, "m1", "ms2")
	insertMediaProbe(t, db, "ms1")
	// Second probe still alive → is_probed should stay 1
	insertMediaProbe(t, db, "ms2")
	if _, err := db.ExecContext(context.Background(), `DELETE FROM media_probes WHERE media_source_id='ms1'`); err != nil {
		t.Fatalf("delete probe: %v", err)
	}
	_, _, _, probed, _, _ := snapshotRow(t, db, "m1")
	if probed != 1 {
		t.Errorf("is_probed should remain 1 while a sibling probe exists, got %d", probed)
	}
	// Now delete the second probe → is_probed should flip back to 0
	if _, err := db.ExecContext(context.Background(), `DELETE FROM media_probes WHERE media_source_id='ms2'`); err != nil {
		t.Fatalf("delete probe: %v", err)
	}
	_, _, _, probed, _, _ = snapshotRow(t, db, "m1")
	if probed != 0 {
		t.Errorf("is_probed should be 0 after last probe removed, got %d", probed)
	}
}

// TestMoviesListView_VersionDeleteDecrementsCount — DELETE FROM movie_versions
// should decrement version_count.
func TestMoviesListView_VersionDeleteDecrementsCount(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertMovie(t, db, "m1", "Inception", "inception", 2010, 0)
	insertMediaSource(t, db, "ms1")
	insertMediaSource(t, db, "ms2")
	insertMovieVersion(t, db, "m1", "ms1")
	insertMovieVersion(t, db, "m1", "ms2")
	if _, err := db.ExecContext(context.Background(), `DELETE FROM movie_versions WHERE media_source_id='ms1'`); err != nil {
		t.Fatalf("delete version: %v", err)
	}
	_, _, vc, _, _, _ := snapshotRow(t, db, "m1")
	if vc != 1 {
		t.Errorf("expected version_count=1 after one DELETE, got %d", vc)
	}
}

// TestMoviesListView_OrderingIsCorrect — the (sort_title, year) index should
// let us pull a page of rows in expected order without a full table scan.
func TestMoviesListView_OrderingIsCorrect(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	// Insert in scrambled order; query should return them sorted.
	insertMovie(t, db, "m1", "Inception", "inception", 2010, 0)
	insertMovie(t, db, "m2", "Avatar", "avatar", 2009, 0)
	insertMovie(t, db, "m3", "Avatar", "avatar", 2022, 0) // same sort_title, later year

	rows, err := db.QueryContext(context.Background(), `
		SELECT movie_id FROM movies_list_view ORDER BY sort_title, year
	`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, id)
	}
	want := []string{"m2", "m3", "m1"}
	if len(got) != len(want) {
		t.Fatalf("expected %d rows, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ordering mismatch at %d: got %q want %q (full: %v)", i, got[i], want[i], got)
		}
	}
}

// TestMoviesListView_BackfillPopulatesExistingMovies — the migration backfills
// movies that existed before the trigger, so a freshly-migrated DB matches a
// brand-new one.
func TestMoviesListView_BackfillPopulatesExistingMovies(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	// Seed several movies, versions, probes — and then nuke the view to
	// simulate "the migration just ran and needs to backfill". The migration
	// statement is an INSERT OR REPLACE so re-running it is safe.
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("m%d", i)
		sid := fmt.Sprintf("ms%d", i)
		insertMovie(t, db, id, fmt.Sprintf("Movie %d", i), fmt.Sprintf("movie %d", i), 2000+i, 0)
		insertMediaSource(t, db, sid)
		insertMovieVersion(t, db, id, sid)
		if i%2 == 0 {
			insertMediaProbe(t, db, sid)
		}
	}
	if _, err := db.ExecContext(context.Background(), `DELETE FROM movies_list_view`); err != nil {
		t.Fatalf("nuke view: %v", err)
	}
	// Re-run the backfill statement from the migration. We inline it here
	// rather than re-executing the migration (which would fail the checksum
	// guard) — the SQL must match what the migration produces.
	if _, err := db.ExecContext(context.Background(), `
		INSERT OR REPLACE INTO movies_list_view (movie_id, title, year, sort_title, needs_review, version_count, is_probed, created_at)
		SELECT m.id, m.title, m.year, m.sort_title, m.needs_review,
		       (SELECT COUNT(*) FROM movie_versions mv WHERE mv.movie_id = m.id) AS version_count,
		       CASE WHEN EXISTS (
		           SELECT 1 FROM movie_versions mv
		           JOIN media_probes mp ON mp.media_source_id = mv.media_source_id
		           WHERE mv.movie_id = m.id
		       ) THEN 1 ELSE 0 END AS is_probed,
		       m.created_at
		FROM movies m
	`); err != nil {
		t.Fatalf("backfill: %v", err)
	}
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("m%d", i)
		_, _, vc, probed, _, ok := snapshotRow(t, db, id)
		if !ok {
			t.Errorf("%s: snapshot row missing after backfill", id)
			continue
		}
		if vc != 1 {
			t.Errorf("%s: expected version_count=1, got %d", id, vc)
		}
		wantProbed := 0
		if i%2 == 0 {
			wantProbed = 1
		}
		if probed != wantProbed {
			t.Errorf("%s: expected is_probed=%d, got %d", id, wantProbed, probed)
		}
	}
}
