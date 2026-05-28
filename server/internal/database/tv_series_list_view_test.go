package database

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func insertSeries(t *testing.T, db *sql.DB, id, title, sortTitle string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO tv_series(id, title, sort_title, updated_at) VALUES(?,?,?,?)
	`, id, title, sortTitle, now); err != nil {
		t.Fatalf("insert tv_series %s: %v", id, err)
	}
}

func insertSeason(t *testing.T, db *sql.DB, id, seriesID string, num int) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO tv_seasons(id, series_id, season_number, updated_at) VALUES(?,?,?,?)
	`, id, seriesID, num, now); err != nil {
		t.Fatalf("insert tv_season %s: %v", id, err)
	}
}

func insertEpisode(t *testing.T, db *sql.DB, id, seriesID, seasonID string, seasonNum, epNum int) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO tv_episodes(id, series_id, season_id, season_number, episode_number, episode_end,
		    title, needs_review, review_reason, updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)
	`, id, seriesID, seasonID, seasonNum, epNum, 0, "", 0, "", now); err != nil {
		t.Fatalf("insert tv_episode %s: %v", id, err)
	}
}

func seriesSnapshotRow(t *testing.T, db *sql.DB, seriesID string) (title string, seasonCount, episodeCount int, found bool) {
	t.Helper()
	err := db.QueryRowContext(context.Background(), `
		SELECT title, season_count, episode_count FROM tv_series_list_view WHERE series_id = ?
	`, seriesID).Scan(&title, &seasonCount, &episodeCount)
	if err == sql.ErrNoRows {
		return "", 0, 0, false
	}
	if err != nil {
		t.Fatalf("read snapshot %s: %v", seriesID, err)
	}
	return title, seasonCount, episodeCount, true
}

func TestTVSeriesListView_InsertCreatesRow(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertSeries(t, db, "s1", "Breaking Bad", "breaking bad")

	title, sc, ec, ok := seriesSnapshotRow(t, db, "s1")
	if !ok {
		t.Fatal("snapshot row missing after INSERT")
	}
	if title != "Breaking Bad" || sc != 0 || ec != 0 {
		t.Errorf("unexpected row: title=%q sc=%d ec=%d", title, sc, ec)
	}
}

func TestTVSeriesListView_UpdatePropagates(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertSeries(t, db, "s1", "Wrong", "wrong")
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.ExecContext(context.Background(), `
		UPDATE tv_series SET title='Breaking Bad', sort_title='breaking bad', updated_at=?
		WHERE id='s1'
	`, now); err != nil {
		t.Fatalf("update tv_series: %v", err)
	}
	title, _, _, _ := seriesSnapshotRow(t, db, "s1")
	if title != "Breaking Bad" {
		t.Errorf("UPDATE didn't propagate: title=%q", title)
	}
}

func TestTVSeriesListView_DeleteCascades(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertSeries(t, db, "s1", "Breaking Bad", "breaking bad")
	if _, err := db.ExecContext(context.Background(), `DELETE FROM tv_series WHERE id='s1'`); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, _, _, ok := seriesSnapshotRow(t, db, "s1"); ok {
		t.Fatal("snapshot row should have cascaded away")
	}
}

func TestTVSeriesListView_SeasonInsertBumpsCount(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertSeries(t, db, "s1", "Breaking Bad", "breaking bad")
	insertSeason(t, db, "se1", "s1", 1)
	_, sc, _, _ := seriesSnapshotRow(t, db, "s1")
	if sc != 1 {
		t.Errorf("expected season_count=1, got %d", sc)
	}
	insertSeason(t, db, "se2", "s1", 2)
	_, sc, _, _ = seriesSnapshotRow(t, db, "s1")
	if sc != 2 {
		t.Errorf("expected season_count=2, got %d", sc)
	}
}

func TestTVSeriesListView_EpisodeInsertBumpsCount(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertSeries(t, db, "s1", "Breaking Bad", "breaking bad")
	insertSeason(t, db, "se1", "s1", 1)
	insertEpisode(t, db, "ep1", "s1", "se1", 1, 1)
	insertEpisode(t, db, "ep2", "s1", "se1", 1, 2)
	insertEpisode(t, db, "ep3", "s1", "se1", 1, 3)
	_, _, ec, _ := seriesSnapshotRow(t, db, "s1")
	if ec != 3 {
		t.Errorf("expected episode_count=3, got %d", ec)
	}
}

func TestTVSeriesListView_SeasonDeleteDecrementsAndCascadesEpisodes(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertSeries(t, db, "s1", "Breaking Bad", "breaking bad")
	insertSeason(t, db, "se1", "s1", 1)
	insertSeason(t, db, "se2", "s1", 2)
	insertEpisode(t, db, "ep1", "s1", "se1", 1, 1)
	insertEpisode(t, db, "ep2", "s1", "se2", 2, 1)
	// Delete season 1 — FK cascades drop ep1, the episode-delete trigger fires
	// for that, and the season-delete trigger fires for the season itself.
	if _, err := db.ExecContext(context.Background(), `DELETE FROM tv_seasons WHERE id='se1'`); err != nil {
		t.Fatalf("delete season: %v", err)
	}
	_, sc, ec, _ := seriesSnapshotRow(t, db, "s1")
	if sc != 1 {
		t.Errorf("expected season_count=1, got %d", sc)
	}
	if ec != 1 {
		t.Errorf("expected episode_count=1 after season cascade, got %d", ec)
	}
}

func TestTVSeriesListView_OrderingByCustomSortTitle(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	insertSeries(t, db, "s1", "The Wire", "wire")
	insertSeries(t, db, "s2", "Breaking Bad", "breaking bad")
	insertSeries(t, db, "s3", "Avenue 5", "avenue 5")
	rows, err := db.QueryContext(context.Background(), `
		SELECT series_id FROM tv_series_list_view ORDER BY sort_title, series_id
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
	want := []string{"s3", "s2", "s1"}
	if len(got) != len(want) {
		t.Fatalf("expected %d rows, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ordering mismatch at %d: got %q want %q (full: %v)", i, got[i], want[i], got)
		}
	}
}

func TestTVSeriesListView_BackfillMatchesTriggerOutput(t *testing.T) {
	db, cleanup := newSnapshotTestDB(t)
	defer cleanup()

	// Build a small library, then nuke the view and re-run the backfill SQL
	// from the migration. The result must match what the triggers produced.
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("s%d", i)
		insertSeries(t, db, id, fmt.Sprintf("Series %d", i), fmt.Sprintf("series %d", i))
		insertSeason(t, db, fmt.Sprintf("sea%d", i), id, 1)
		insertEpisode(t, db, fmt.Sprintf("ep%da", i), id, fmt.Sprintf("sea%d", i), 1, 1)
		insertEpisode(t, db, fmt.Sprintf("ep%db", i), id, fmt.Sprintf("sea%d", i), 1, 2)
	}
	// Capture trigger-generated state
	beforeMap := map[string][2]int{}
	{
		rows, _ := db.QueryContext(context.Background(), `SELECT series_id, season_count, episode_count FROM tv_series_list_view`)
		for rows.Next() {
			var id string
			var sc, ec int
			_ = rows.Scan(&id, &sc, &ec)
			beforeMap[id] = [2]int{sc, ec}
		}
		rows.Close()
	}
	if _, err := db.ExecContext(context.Background(), `DELETE FROM tv_series_list_view`); err != nil {
		t.Fatalf("nuke view: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
		INSERT OR REPLACE INTO tv_series_list_view (series_id, title, sort_title, season_count, episode_count, created_at)
		SELECT s.id, s.title, s.sort_title,
		       (SELECT COUNT(*) FROM tv_seasons ts WHERE ts.series_id = s.id),
		       (SELECT COUNT(*) FROM tv_episodes te WHERE te.series_id = s.id),
		       s.created_at
		FROM tv_series s
	`); err != nil {
		t.Fatalf("backfill: %v", err)
	}
	for id, want := range beforeMap {
		_, sc, ec, ok := seriesSnapshotRow(t, db, id)
		if !ok {
			t.Errorf("%s missing after backfill", id)
			continue
		}
		if sc != want[0] || ec != want[1] {
			t.Errorf("%s: backfill mismatch sc=%d/%d ec=%d/%d", id, sc, want[0], ec, want[1])
		}
	}
}
