package metadata

import (
	"context"
	"encoding/json"
	"log/slog"
)

// BackfillDetailsJSON re-parses raw_json for every TMDB metadata record whose
// details_json column is still the migration default '{}'. This fixes people
// search on databases created before the cast/crew indexing was added. It runs
// in the background at startup and is a no-op when all records are current.
func (s *Service) BackfillDetailsJSON(ctx context.Context) error {
	records, err := s.catalog.ListStaleDetailsRecords(ctx, "tmdb")
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}
	slog.Info("details-json backfill: reprocessing stale records", "count", len(records))
	fixed := 0
	for _, record := range records {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		updated := record
		switch updated.Kind {
		case "movie":
			var detail tmdbMovieDetail
			if err := json.Unmarshal([]byte(updated.RawJSON), &detail); err != nil {
				slog.Debug("details-json backfill: skip movie (parse error)", "itemID", updated.ItemID, "err", err)
				continue
			}
			updated.Cast = tmdbCredits(detail.Credits.Cast, "cast")
			updated.Crew = tmdbCredits(detail.Credits.Crew, "crew")
		case "series":
			var detail tmdbTVDetail
			if err := json.Unmarshal([]byte(updated.RawJSON), &detail); err != nil {
				slog.Debug("details-json backfill: skip series (parse error)", "itemID", updated.ItemID, "err", err)
				continue
			}
			updated.Cast = tmdbAggregateCredits(detail.AggregateCredits.Cast)
			updated.Crew = tmdbCredits(detail.AggregateCredits.Crew, "crew")
		case "season":
			var detail tmdbSeasonDetail
			if err := json.Unmarshal([]byte(updated.RawJSON), &detail); err != nil {
				slog.Debug("details-json backfill: skip season (parse error)", "itemID", updated.ItemID, "err", err)
				continue
			}
			updated.Cast = tmdbCredits(detail.Credits.Cast, "cast")
			updated.Crew = tmdbCredits(detail.Credits.Crew, "crew")
		case "episode":
			var detail tmdbEpisode
			if err := json.Unmarshal([]byte(updated.RawJSON), &detail); err != nil {
				slog.Debug("details-json backfill: skip episode (parse error)", "itemID", updated.ItemID, "err", err)
				continue
			}
			updated.GuestCast = tmdbCredits(detail.GuestStars, "cast")
			updated.Crew = tmdbCredits(detail.Crew, "crew")
		default:
			continue
		}
		if err := s.catalog.UpsertMetadataRecord(ctx, updated); err != nil {
			slog.Debug("details-json backfill: upsert failed", "itemID", updated.ItemID, "err", err)
			continue
		}
		fixed++
	}
	slog.Info("details-json backfill: complete", "fixed", fixed, "total", len(records))
	return nil
}
