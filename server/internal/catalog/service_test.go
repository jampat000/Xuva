package catalog

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/jampat000/Xuva/server/internal/database"
	"github.com/jampat000/Xuva/server/internal/libraries"
	"github.com/jampat000/Xuva/server/internal/movies"
	"github.com/jampat000/Xuva/server/internal/scanner"
	"github.com/jampat000/Xuva/server/internal/tv"
)

func TestSaveMovieScanIsIdempotent(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t)
	library := libraries.Library{ID: "movies", Name: "Movies", Path: `X:\Movies`, Kind: libraries.KindMovies}
	result := scanResult(scanner.KindMovies, library.Path, []scanner.FileCandidate{fileCandidate(`X:\Movies\Heat (1995)\Heat.1995.1080p.mkv`, "Heat (1995)/Heat.1995.1080p.mkv")})
	candidates := []movies.Candidate{{
		Title:        "Heat",
		Year:         1995,
		QualityLabel: "1080p",
		Media:        result.Files[0],
	}}

	if _, err := service.SaveMovieScan(ctx, library, result, candidates); err != nil {
		t.Fatalf("save movie scan: %v", err)
	}
	if _, err := service.SaveMovieScan(ctx, library, result, candidates); err != nil {
		t.Fatalf("save movie scan again: %v", err)
	}
	summary, err := service.Summary(ctx)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	if summary.Libraries != 1 {
		t.Fatalf("expected one library, got %d", summary.Libraries)
	}
	if summary.MediaSources != 1 {
		t.Fatalf("expected one media source, got %d", summary.MediaSources)
	}
	if summary.Movies != 1 {
		t.Fatalf("expected one movie, got %d", summary.Movies)
	}
	if summary.ScanRuns != 2 {
		t.Fatalf("expected two scan runs, got %d", summary.ScanRuns)
	}

	movies, err := service.ListMovies(ctx, 10)
	if err != nil {
		t.Fatalf("list movies: %v", err)
	}
	if len(movies) != 1 || movies[0].Title != "Heat" || movies[0].VersionCount != 1 {
		t.Fatalf("unexpected movies: %#v", movies)
	}
	detail, ok, err := service.GetMovie(ctx, movies[0].ID)
	if err != nil {
		t.Fatalf("get movie: %v", err)
	}
	if !ok || len(detail.Versions) != 1 {
		t.Fatalf("unexpected movie detail: %#v", detail)
	}
	if detail.Metadata == nil || detail.Metadata.Provider != "filename" || detail.Metadata.Title != "Heat" || detail.Metadata.Year != 1995 {
		t.Fatalf("expected filename metadata on movie detail, got %#v", detail.Metadata)
	}
	records, err := service.ListMetadataRecords(ctx, "movie", movies[0].ID)
	if err != nil {
		t.Fatalf("list metadata records: %v", err)
	}
	if len(records) != 1 || records[0].Provider != "filename" {
		t.Fatalf("expected filename metadata record, got %#v", records)
	}
	if err := service.ApplyMetadata(ctx, MetadataUpdate{
		Kind:     "movie",
		ID:       movies[0].ID,
		Title:    "Heat",
		Year:     1995,
		Provider: "manual",
		Overview: "A professional thief weighs one last score.",
	}); err != nil {
		t.Fatalf("apply metadata: %v", err)
	}
	best, ok, err := service.GetBestMetadata(ctx, "movie", movies[0].ID)
	if err != nil {
		t.Fatalf("best metadata: %v", err)
	}
	if !ok || best.Provider != "manual" || best.Overview == "" {
		t.Fatalf("expected manual metadata to win, got %#v", best)
	}
	sources, err := service.ListMediaSources(ctx, 10, true)
	if err != nil {
		t.Fatalf("list media sources: %v", err)
	}
	if len(sources) != 1 || sources[0].Probed {
		t.Fatalf("expected one unprobed source, got %#v", sources)
	}
	if err := service.SaveProbe(ctx, sources[0].ID, ProbeResult{
		Container:       "matroska,webm",
		DurationSeconds: 120,
		Bitrate:         1000,
		VideoCodec:      "h264",
		Width:           1920,
		Height:          1080,
		AudioStreams:    1,
		SubtitleStreams: 1,
		RawJSON:         "{}",
	}); err != nil {
		t.Fatalf("save probe: %v", err)
	}
	source, ok, err := service.GetMediaSource(ctx, sources[0].ID)
	if err != nil {
		t.Fatalf("get media source: %v", err)
	}
	if !ok || !source.Probed || source.VideoCodec != "h264" {
		t.Fatalf("unexpected probed source: %#v", source)
	}
}

func TestSaveTVScanStoresSeriesEpisodeAndSource(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t)
	library := libraries.Library{ID: "tv", Name: "TV", Path: `X:\TV`, Kind: libraries.KindTV}
	result := scanResult(scanner.KindTV, library.Path, []scanner.FileCandidate{fileCandidate(`X:\TV\The Bear\Season 02\The.Bear.S02E03.Sundae.mkv`, "The Bear/Season 02/The.Bear.S02E03.Sundae.mkv")})
	candidates := []tv.EpisodeCandidate{{
		SeriesTitle:   "The Bear",
		SeasonNumber:  2,
		EpisodeNumber: 3,
		EpisodeTitle:  "Sundae",
		QualityLabel:  "1080p",
		Media:         result.Files[0],
	}}

	if _, err := service.SaveTVScan(ctx, library, result, candidates); err != nil {
		t.Fatalf("save tv scan: %v", err)
	}
	summary, err := service.Summary(ctx)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	if summary.Series != 1 {
		t.Fatalf("expected one series, got %d", summary.Series)
	}
	if summary.Episodes != 1 {
		t.Fatalf("expected one episode, got %d", summary.Episodes)
	}
	if summary.MediaSources != 1 {
		t.Fatalf("expected one media source, got %d", summary.MediaSources)
	}

	series, err := service.ListSeries(ctx, 10)
	if err != nil {
		t.Fatalf("list series: %v", err)
	}
	if len(series) != 1 || series[0].Title != "The Bear" || series[0].EpisodeCount != 1 {
		t.Fatalf("unexpected series: %#v", series)
	}
	detail, ok, err := service.GetSeries(ctx, series[0].ID)
	if err != nil {
		t.Fatalf("get series: %v", err)
	}
	if !ok || len(detail.Seasons) != 1 || len(detail.Seasons[0].Episodes) != 1 {
		t.Fatalf("unexpected series detail: %#v", detail)
	}
	if detail.Metadata == nil || detail.Metadata.Provider != "filename" || detail.Metadata.Title != "The Bear" {
		t.Fatalf("expected filename metadata on series detail, got %#v", detail.Metadata)
	}
}

func TestSaveTVScanKeepsUnmatchedEpisodesSeparate(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t)
	library := libraries.Library{ID: "tv", Name: "TV", Path: `X:\TV`, Kind: libraries.KindTV}
	files := []scanner.FileCandidate{
		fileCandidate(`X:\TV\Series\Special.Video.A.mkv`, "Series/Special.Video.A.mkv"),
		fileCandidate(`X:\TV\Series\Special.Video.B.mkv`, "Series/Special.Video.B.mkv"),
	}
	result := scanResult(scanner.KindTV, library.Path, files)
	candidates := []tv.EpisodeCandidate{
		{SeriesTitle: "Series", NeedsReview: true, ReviewReason: "unable to infer episode number", Media: files[0]},
		{SeriesTitle: "Series", NeedsReview: true, ReviewReason: "unable to infer episode number", Media: files[1]},
	}

	persisted, err := service.SaveTVScan(ctx, library, result, candidates)
	if err != nil {
		t.Fatalf("save tv scan: %v", err)
	}
	if persisted.Episodes != 2 {
		t.Fatalf("expected two separate review episodes, got %d", persisted.Episodes)
	}
	summary, err := service.Summary(ctx)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Episodes != 2 {
		t.Fatalf("expected two stored review episodes, got %d", summary.Episodes)
	}
	reviewItems, err := service.ReviewItems(ctx, 10)
	if err != nil {
		t.Fatalf("review items: %v", err)
	}
	if len(reviewItems) != 2 {
		t.Fatalf("expected two review items, got %#v", reviewItems)
	}
}

func TestSaveMovieScanPersistsIncrementalState(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t)
	library := libraries.Library{ID: "movies", Name: "Movies", Path: `X:\Movies`, Kind: libraries.KindMovies}
	file := fileCandidate(`X:\Movies\Heat (1995)\Heat.1995.1080p.mkv`, "Heat (1995)/Heat.1995.1080p.mkv")
	result := scanResult(scanner.KindMovies, library.Path, []scanner.FileCandidate{file})
	candidates := []movies.Candidate{{Title: "Heat", Year: 1995, QualityLabel: "1080p", Media: file}}
	persisted, err := service.SaveMovieScan(ctx, library, result, candidates)
	if err != nil {
		t.Fatalf("save movie scan: %v", err)
	}
	if persisted.ChangedSources != 1 || persisted.UnchangedSources != 0 {
		t.Fatalf("expected changed source count, got %#v", persisted)
	}
	state, err := service.ScanState(ctx, library.ID)
	if err != nil {
		t.Fatalf("scan state: %v", err)
	}
	signature, ok := state[file.RelPath]
	if !ok {
		t.Fatalf("expected scan state for %q, got %#v", file.RelPath, state)
	}
	if signature.Size != file.Size || !signature.ModifiedAt.Equal(file.ModifiedAt) {
		t.Fatalf("unexpected signature: %#v", signature)
	}
	file.Changed = false
	result = scanResult(scanner.KindMovies, library.Path, []scanner.FileCandidate{file})
	candidates[0].Media = file
	persisted, err = service.SaveMovieScan(ctx, library, result, candidates)
	if err != nil {
		t.Fatalf("save incremental movie scan: %v", err)
	}
	if persisted.ChangedSources != 0 || persisted.UnchangedSources != 1 {
		t.Fatalf("expected unchanged source count, got %#v", persisted)
	}
}

func TestBestMetadataHonorsLibrarySourceOrder(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t)
	library := libraries.Library{
		ID:              "movies",
		Name:            "Movies",
		Path:            `X:\Movies`,
		Kind:            libraries.KindMovies,
		MetadataSources: []string{"wikipedia", "filename", "nfo"},
	}
	file := fileCandidate(`X:\Movies\Arrival (2016)\Arrival.2016.1080p.mkv`, "Arrival (2016)/Arrival.2016.1080p.mkv")
	result := scanResult(scanner.KindMovies, library.Path, []scanner.FileCandidate{file})
	candidates := []movies.Candidate{{
		Title:        "Arrival",
		Year:         2016,
		QualityLabel: "1080p",
		Media:        file,
	}}
	if _, err := service.SaveMovieScan(ctx, library, result, candidates); err != nil {
		t.Fatalf("save movie scan: %v", err)
	}
	movieList, err := service.ListMovies(ctx, 10)
	if err != nil {
		t.Fatalf("list movies: %v", err)
	}
	if len(movieList) != 1 {
		t.Fatalf("expected one movie, got %#v", movieList)
	}
	movieID := movieList[0].ID
	if err := service.UpsertMetadataRecord(ctx, MetadataRecord{
		Kind:       "movie",
		ItemID:     movieID,
		Provider:   "filename",
		Title:      "Arrival",
		Year:       2016,
		Confidence: 0.98,
	}); err != nil {
		t.Fatalf("upsert filename metadata: %v", err)
	}
	if err := service.UpsertMetadataRecord(ctx, MetadataRecord{
		Kind:       "movie",
		ItemID:     movieID,
		Provider:   "wikipedia",
		Title:      "Arrival (Film)",
		Year:       2016,
		Confidence: 0.45,
	}); err != nil {
		t.Fatalf("upsert wikipedia metadata: %v", err)
	}
	best, ok, err := service.GetBestMetadata(ctx, "movie", movieID)
	if err != nil {
		t.Fatalf("get best metadata: %v", err)
	}
	if !ok {
		t.Fatalf("expected metadata record")
	}
	if best.Provider != "wikipedia" {
		t.Fatalf("expected library source order to prioritize wikipedia, got %#v", best)
	}
	records, err := service.ListMetadataRecords(ctx, "movie", movieID)
	if err != nil {
		t.Fatalf("list metadata records: %v", err)
	}
	if len(records) < 2 {
		t.Fatalf("expected at least two metadata records, got %#v", records)
	}
	if records[0].Provider != "wikipedia" {
		t.Fatalf("expected ordered metadata records to start with wikipedia, got %#v", records[0])
	}
}

func TestBestMetadataUsesSeparateArtworkSourceOrder(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t)
	library := libraries.Library{
		ID:              "movies",
		Name:            "Movies",
		Path:            `X:\Movies`,
		Kind:            libraries.KindMovies,
		MetadataSources: []string{"tmdb", "wikipedia", "filename"},
		ArtworkSources:  []string{"artwork", "fanart", "tmdb"},
	}
	file := fileCandidate(`X:\Movies\Arrival (2016)\Arrival.2016.1080p.mkv`, "Arrival (2016)/Arrival.2016.1080p.mkv")
	result := scanResult(scanner.KindMovies, library.Path, []scanner.FileCandidate{file})
	candidates := []movies.Candidate{{
		Title:        "Arrival",
		Year:         2016,
		QualityLabel: "1080p",
		Media:        file,
	}}
	if _, err := service.SaveMovieScan(ctx, library, result, candidates); err != nil {
		t.Fatalf("save movie scan: %v", err)
	}
	movieList, err := service.ListMovies(ctx, 10)
	if err != nil {
		t.Fatalf("list movies: %v", err)
	}
	if len(movieList) != 1 {
		t.Fatalf("expected one movie, got %#v", movieList)
	}
	movieID := movieList[0].ID
	if err := service.UpsertMetadataRecord(ctx, MetadataRecord{
		Kind:       "movie",
		ItemID:     movieID,
		Provider:   "tmdb",
		Title:      "Arrival",
		Year:       2016,
		Overview:   "TMDB overview",
		Genres:     []string{"Science Fiction", "Drama"},
		PosterURL:  "https://images.example/tmdb-poster.jpg",
		Confidence: 0.95,
	}); err != nil {
		t.Fatalf("upsert tmdb metadata: %v", err)
	}
	if err := service.UpsertMetadataRecord(ctx, MetadataRecord{
		Kind:        "movie",
		ItemID:      movieID,
		Provider:    "artwork",
		Title:       "Arrival",
		PosterURL:   `X:\Movies\Arrival (2016)\poster.jpg`,
		BackdropURL: `X:\Movies\Arrival (2016)\fanart.jpg`,
		Confidence:  1,
	}); err != nil {
		t.Fatalf("upsert local artwork: %v", err)
	}
	if err := service.UpsertMetadataRecord(ctx, MetadataRecord{
		Kind:       "movie",
		ItemID:     movieID,
		Provider:   "fanart",
		Title:      "Arrival",
		LogoURL:    "https://images.example/fanart-logo.png",
		BannerURL:  "https://images.example/fanart-banner.jpg",
		Confidence: 0.7,
	}); err != nil {
		t.Fatalf("upsert fanart artwork: %v", err)
	}

	best, ok, err := service.GetBestMetadata(ctx, "movie", movieID)
	if err != nil {
		t.Fatalf("get best metadata: %v", err)
	}
	if !ok {
		t.Fatalf("expected merged metadata")
	}
	if best.Provider != "tmdb" || best.Overview != "TMDB overview" {
		t.Fatalf("expected tmdb to drive metadata fields, got %#v", best)
	}
	if best.PosterURL != `X:\Movies\Arrival (2016)\poster.jpg` || best.BackdropURL != `X:\Movies\Arrival (2016)\fanart.jpg` {
		t.Fatalf("expected local artwork to win for poster/backdrop, got %#v", best)
	}
	if best.LogoURL != "https://images.example/fanart-logo.png" || best.BannerURL != "https://images.example/fanart-banner.jpg" {
		t.Fatalf("expected fanart artwork enhancement to fill logo/banner, got %#v", best)
	}
}

func TestGetSeriesAttachesSeasonAndEpisodeMetadata(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t)
	library := libraries.Library{ID: "tv", Name: "TV", Path: `X:\TV`, Kind: libraries.KindTV}
	file := fileCandidate(`X:\TV\The Bear\Season 01\The.Bear.S01E01.System.mkv`, "The Bear/Season 01/The.Bear.S01E01.System.mkv")
	result := scanResult(scanner.KindTV, library.Path, []scanner.FileCandidate{file})
	candidates := []tv.EpisodeCandidate{{
		SeriesTitle:   "The Bear",
		SeasonNumber:  1,
		EpisodeNumber: 1,
		EpisodeTitle:  "System",
		QualityLabel:  "1080p",
		Media:         file,
	}}

	if _, err := service.SaveTVScan(ctx, library, result, candidates); err != nil {
		t.Fatalf("save tv scan: %v", err)
	}
	seriesList, err := service.ListSeries(ctx, 10)
	if err != nil {
		t.Fatalf("list series: %v", err)
	}
	if len(seriesList) != 1 {
		t.Fatalf("expected one series, got %#v", seriesList)
	}
	seeded, ok, err := service.GetSeries(ctx, seriesList[0].ID)
	if err != nil {
		t.Fatalf("get seeded series: %v", err)
	}
	if !ok || len(seeded.Seasons) != 1 || len(seeded.Seasons[0].Episodes) != 1 {
		t.Fatalf("expected one season and one episode, got %#v", seeded)
	}

	if err := service.UpsertMetadataRecord(ctx, MetadataRecord{
		Kind:       "series",
		ItemID:     seeded.ID,
		Provider:   "tmdb",
		Title:      "The Bear",
		Overview:   "A chef returns home to run the family restaurant.",
		Genres:     []string{"Drama"},
		PosterURL:  "https://images.example/bear-poster.jpg",
		Confidence: 0.95,
	}); err != nil {
		t.Fatalf("upsert series metadata: %v", err)
	}
	if err := service.UpsertMetadataRecord(ctx, MetadataRecord{
		Kind:         "season",
		ItemID:       seeded.Seasons[0].ID,
		Provider:     "tmdb",
		Title:        "Season 1",
		Overview:     "Season overview",
		PosterURL:    "https://images.example/bear-season-1.jpg",
		SeasonNumber: 1,
		EpisodeCount: 1,
		Confidence:   0.9,
	}); err != nil {
		t.Fatalf("upsert season metadata: %v", err)
	}
	if err := service.UpsertMetadataRecord(ctx, MetadataRecord{
		Kind:          "episode",
		ItemID:        seeded.Seasons[0].Episodes[0].ID,
		Provider:      "tmdb",
		Title:         "System",
		Overview:      "Episode overview",
		ThumbnailURL:  "https://images.example/bear-s01e01.jpg",
		SeasonNumber:  1,
		EpisodeNumber: 1,
		RuntimeMinutes: 42,
		Confidence:    0.9,
	}); err != nil {
		t.Fatalf("upsert episode metadata: %v", err)
	}

	detail, ok, err := service.GetSeries(ctx, seeded.ID)
	if err != nil {
		t.Fatalf("get enriched series: %v", err)
	}
	if !ok {
		t.Fatalf("expected enriched series detail")
	}
	if detail.Metadata == nil || detail.Metadata.Provider != "tmdb" || detail.Metadata.Overview == "" {
		t.Fatalf("expected show metadata on detail, got %#v", detail.Metadata)
	}
	season := detail.Seasons[0]
	if season.Metadata == nil || season.Metadata.PosterURL != "https://images.example/bear-season-1.jpg" || season.Metadata.EpisodeCount != 1 {
		t.Fatalf("expected season metadata on detail, got %#v", season.Metadata)
	}
	episode := season.Episodes[0]
	if episode.Metadata == nil || episode.Metadata.ThumbnailURL != "https://images.example/bear-s01e01.jpg" || episode.Metadata.RuntimeMinutes != 42 {
		t.Fatalf("expected episode metadata on detail, got %#v", episode.Metadata)
	}
}

func TestListSeriesCollapsesDuplicatesBySharedExternalIdentity(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t)
	library := libraries.Library{ID: "tv", Name: "TV", Path: `X:\TV`, Kind: libraries.KindTV}
	files := []scanner.FileCandidate{
		fileCandidate(`X:\TV\Dangerously Obese\Season 01\Dangerously.Obese.S01E01.mkv`, "Dangerously Obese/Season 01/Dangerously.Obese.S01E01.mkv"),
		fileCandidate(`X:\TV\Dangerously Obese UK\Season 02\Dangerously.Obese.UK.S02E01.mkv`, "Dangerously Obese UK/Season 02/Dangerously.Obese.UK.S02E01.mkv"),
	}
	result := scanResult(scanner.KindTV, library.Path, files)
	candidates := []tv.EpisodeCandidate{
		{SeriesTitle: "Dangerously Obese", SeasonNumber: 1, EpisodeNumber: 1, EpisodeTitle: "Episode 1", Media: files[0]},
		{SeriesTitle: "Dangerously Obese UK", SeasonNumber: 2, EpisodeNumber: 1, EpisodeTitle: "Episode 1", Media: files[1]},
	}
	if _, err := service.SaveTVScan(ctx, library, result, candidates); err != nil {
		t.Fatalf("save tv scan: %v", err)
	}

	series, err := service.ListSeries(ctx, 10)
	if err != nil {
		t.Fatalf("list raw series: %v", err)
	}
	if len(series) != 2 {
		t.Fatalf("expected two raw series before metadata grouping assumptions, got %#v", series)
	}
	for _, item := range series {
		if err := service.UpsertMetadataRecord(ctx, MetadataRecord{
			Kind:       "series",
			ItemID:     item.ID,
			Provider:   "tmdb",
			Title:      "Dangerously Obese",
			Overview:   "Merged by shared metadata identity.",
			PosterURL:  "https://images.example/dangerously-obese.jpg",
			Confidence: 0.9,
		}); err != nil {
			t.Fatalf("upsert series metadata: %v", err)
		}
		if err := service.UpsertExternalID(ctx, ExternalID{
			Kind:       "series",
			ItemID:     item.ID,
			Provider:   "tmdb",
			ExternalID: "9991",
		}); err != nil {
			t.Fatalf("upsert shared tmdb id: %v", err)
		}
	}

	grouped, err := service.ListSeries(ctx, 10)
	if err != nil {
		t.Fatalf("list grouped series: %v", err)
	}
	if len(grouped) != 1 {
		t.Fatalf("expected duplicate series to collapse into one list item, got %#v", grouped)
	}
	if grouped[0].SeasonCount != 2 || grouped[0].EpisodeCount != 2 {
		t.Fatalf("expected grouped counts across both records, got %#v", grouped[0])
	}
	if grouped[0].Metadata == nil || grouped[0].Metadata.Title != "Dangerously Obese" {
		t.Fatalf("expected grouped metadata title, got %#v", grouped[0].Metadata)
	}
}

func TestGetSeriesAggregatesDuplicateSeriesMembers(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t)
	library := libraries.Library{ID: "tv", Name: "TV", Path: `X:\TV`, Kind: libraries.KindTV}
	files := []scanner.FileCandidate{
		fileCandidate(`X:\TV\Dangerously Obese\Season 01\Dangerously.Obese.S01E01.mkv`, "Dangerously Obese/Season 01/Dangerously.Obese.S01E01.mkv"),
		fileCandidate(`X:\TV\Dangerously Obese UK\Season 02\Dangerously.Obese.UK.S02E01.mkv`, "Dangerously Obese UK/Season 02/Dangerously.Obese.UK.S02E01.mkv"),
	}
	result := scanResult(scanner.KindTV, library.Path, files)
	candidates := []tv.EpisodeCandidate{
		{SeriesTitle: "Dangerously Obese", SeasonNumber: 1, EpisodeNumber: 1, EpisodeTitle: "Episode 1", Media: files[0]},
		{SeriesTitle: "Dangerously Obese UK", SeasonNumber: 2, EpisodeNumber: 1, EpisodeTitle: "Episode 1", Media: files[1]},
	}
	if _, err := service.SaveTVScan(ctx, library, result, candidates); err != nil {
		t.Fatalf("save tv scan: %v", err)
	}

	rawSeries, err := service.ListSeries(ctx, 10)
	if err != nil {
		t.Fatalf("list series: %v", err)
	}
	if len(rawSeries) != 2 {
		t.Fatalf("expected two seeded series, got %#v", rawSeries)
	}
	for _, item := range rawSeries {
		if err := service.UpsertMetadataRecord(ctx, MetadataRecord{
			Kind:       "series",
			ItemID:     item.ID,
			Provider:   "tmdb",
			Title:      "Dangerously Obese",
			Overview:   "Merged by shared metadata identity.",
			PosterURL:  "https://images.example/dangerously-obese.jpg",
			Confidence: 0.9,
		}); err != nil {
			t.Fatalf("upsert series metadata: %v", err)
		}
		if err := service.UpsertExternalID(ctx, ExternalID{
			Kind:       "series",
			ItemID:     item.ID,
			Provider:   "tmdb",
			ExternalID: "9991",
		}); err != nil {
			t.Fatalf("upsert shared tmdb id: %v", err)
		}
	}

	detail, ok, err := service.GetSeries(ctx, rawSeries[0].ID)
	if err != nil {
		t.Fatalf("get aggregated series: %v", err)
	}
	if !ok {
		t.Fatalf("expected series detail")
	}
	if detail.SeasonCount != 2 || detail.EpisodeCount != 2 {
		t.Fatalf("expected aggregated season and episode counts, got %#v", detail)
	}
	if len(detail.Seasons) != 2 {
		t.Fatalf("expected both seasons to appear in one detail view, got %#v", detail.Seasons)
	}
	if detail.Metadata == nil || detail.Metadata.Title != "Dangerously Obese" {
		t.Fatalf("expected aggregate metadata on detail, got %#v", detail.Metadata)
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()

	db, err := database.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
	})
	return NewService(db)
}

func scanResult(kind scanner.LibraryKind, root string, files []scanner.FileCandidate) scanner.Result {
	startedAt := time.Now().UTC()
	seen := make([]string, 0, len(files))
	changed := 0
	unchanged := 0
	for _, file := range files {
		seen = append(seen, file.RelPath)
		if file.Changed {
			changed++
		} else {
			unchanged++
		}
	}
	return scanner.Result{
		Summary: scanner.Summary{
			Kind:         kind,
			Root:         root,
			StartedAt:    startedAt,
			CompletedAt:  startedAt.Add(time.Millisecond),
			DurationMS:   1,
			TotalFiles:   len(files),
			MediaFiles:   len(files),
			ChangedFiles: changed,
			Unchanged:    unchanged,
			Extensions:   map[string]int{".mkv": len(files)},
		},
		Files:        files,
		SeenRelPaths: seen,
	}
}

func fileCandidate(path string, relPath string) scanner.FileCandidate {
	return scanner.FileCandidate{
		Path:       filepath.Clean(path),
		RelPath:    filepath.Clean(relPath),
		Name:       filepath.Base(path),
		Extension:  ".mkv",
		Size:       1024,
		ModifiedAt: time.Now().UTC(),
		Changed:    true,
	}
}
