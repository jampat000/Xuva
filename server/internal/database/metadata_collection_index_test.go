package database

import (
	"context"
	"strings"
	"testing"
)

// TestMetadataCollectionIndex_PlannerUsesIt verifies that SQLite's query
// planner actually picks up the partial expression index for the collection
// GROUP BY shape used by ListCollections. If a future schema change drifts
// the index column list or partial-WHERE predicate, EXPLAIN QUERY PLAN will
// stop mentioning the index by name and this test will catch it before the
// performance regression hits production.
func TestMetadataCollectionIndex_PlannerUsesIt(t *testing.T) {
	svc, err := Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer svc.Close()
	rows, err := svc.DB().QueryContext(context.Background(), `
		EXPLAIN QUERY PLAN
		SELECT json_extract(mr.details_json, '$.collection.id') AS coll_id,
		       COUNT(*) AS movie_count
		FROM metadata_records mr
		WHERE mr.kind = 'movie'
		  AND json_extract(mr.details_json, '$.collection.id') IS NOT NULL
		GROUP BY coll_id
	`)
	if err != nil {
		t.Fatalf("EXPLAIN QUERY PLAN: %v", err)
	}
	defer rows.Close()
	var plan []string
	for rows.Next() {
		var id, parent, notused int
		var detail string
		if err := rows.Scan(&id, &parent, &notused, &detail); err != nil {
			t.Fatalf("scan: %v", err)
		}
		plan = append(plan, detail)
	}
	joined := strings.Join(plan, "\n")
	if !strings.Contains(joined, "idx_metadata_records_collection") {
		t.Errorf("expected ListCollections-shaped query to use idx_metadata_records_collection; plan was:\n%s", joined)
	}
}
