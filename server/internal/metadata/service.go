package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/catalog"
	"github.com/vyrdenhq/vyrden/server/internal/config"
	"github.com/vyrdenhq/vyrden/server/internal/events"
)

type Service struct {
	cfg     config.Config
	catalog *catalog.Service
	events  *events.Bus
	client  *http.Client
}

type RefreshRequest struct {
	Kind  string `json:"kind"`
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
	Year  int    `json:"year,omitempty"`
}

type RefreshResult struct {
	Kind        string                   `json:"kind"`
	ID          string                   `json:"id"`
	Configured  map[string]bool          `json:"configured"`
	Records     []catalog.MetadataRecord `json:"records"`
	Ratings     []catalog.Rating         `json:"ratings"`
	ExternalIDs []catalog.ExternalID     `json:"externalIds"`
	Warnings    []string                 `json:"warnings,omitempty"`
}

func NewService(cfg config.Config, catalogService *catalog.Service, eventBus *events.Bus) *Service {
	return &Service{
		cfg:     cfg,
		catalog: catalogService,
		events:  eventBus,
		client:  &http.Client{Timeout: 12 * time.Second},
	}
}

func (s *Service) Refresh(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
	if request.Kind == "" || request.ID == "" {
		return RefreshResult{}, errors.New("kind and id are required")
	}
	if request.Title == "" {
		record, ok, err := s.catalog.GetBestMetadata(ctx, request.Kind, request.ID)
		if err != nil {
			return RefreshResult{}, err
		}
		if ok {
			request.Title = record.Title
			request.Year = record.Year
		}
	}
	if strings.TrimSpace(request.Title) == "" {
		return RefreshResult{}, errors.New("title is required")
	}

	result := RefreshResult{
		Kind:       request.Kind,
		ID:         request.ID,
		Configured: map[string]bool{"omdb": s.cfg.OMDbAPIKey != "", "tmdb": s.cfg.TMDBAPIKey != ""},
	}
	if s.cfg.OMDbAPIKey == "" {
		result.Warnings = append(result.Warnings, "OMDb is not configured. Set VYRDEN_OMDB_API_KEY to fetch IMDb, Rotten Tomatoes, and Metacritic ratings.")
	} else if err := s.refreshOMDb(ctx, request, &result); err != nil {
		result.Warnings = append(result.Warnings, "OMDb refresh failed: "+err.Error())
	}
	if s.cfg.TMDBAPIKey == "" {
		result.Warnings = append(result.Warnings, "TMDB is not configured. Set VYRDEN_TMDB_API_KEY to fetch TMDB ratings and external IDs.")
	} else if err := s.refreshTMDB(ctx, request, &result); err != nil {
		result.Warnings = append(result.Warnings, "TMDB refresh failed: "+err.Error())
	}

	records, err := s.catalog.ListMetadataRecords(ctx, request.Kind, request.ID)
	if err != nil {
		return RefreshResult{}, err
	}
	ratings, err := s.catalog.ListRatings(ctx, request.Kind, request.ID)
	if err != nil {
		return RefreshResult{}, err
	}
	externalIDs, err := s.catalog.ListExternalIDs(ctx, request.Kind, request.ID)
	if err != nil {
		return RefreshResult{}, err
	}
	result.Records = records
	result.Ratings = ratings
	result.ExternalIDs = externalIDs
	if s.events != nil {
		s.events.Publish("metadata.ratings.updated", result)
	}
	return result, nil
}

func (s *Service) refreshOMDb(ctx context.Context, request RefreshRequest, result *RefreshResult) error {
	endpoint := "https://www.omdbapi.com/?" + url.Values{
		"apikey": {s.cfg.OMDbAPIKey},
		"t":      {request.Title},
		"type":   {omdbType(request.Kind)},
		"plot":   {"full"},
	}.Encode()
	if request.Year > 0 {
		endpoint += "&y=" + strconv.Itoa(request.Year)
	}
	var payload omdbResponse
	if err := s.getJSON(ctx, endpoint, &payload); err != nil {
		return err
	}
	if strings.EqualFold(payload.Response, "false") {
		return errors.New(firstNonEmpty(payload.Error, "no OMDb match"))
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	record := catalog.MetadataRecord{
		Kind:       request.Kind,
		ItemID:     request.ID,
		Provider:   "omdb",
		ExternalID: payload.ImdbID,
		Title:      firstNonEmpty(payload.Title, request.Title),
		Year:       parseYear(payload.Year, request.Year),
		Overview:   payload.Plot,
		PosterURL:  emptyNA(payload.Poster),
		Confidence: 0.88,
		RawJSON:    mustJSON(payload),
		FetchedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.catalog.UpsertMetadataRecord(ctx, record); err != nil {
		return err
	}
	result.Records = append(result.Records, record)
	if payload.ImdbID != "" {
		_ = s.catalog.UpsertExternalID(ctx, catalog.ExternalID{
			Kind:       request.Kind,
			ItemID:     request.ID,
			Provider:   "imdb",
			ExternalID: payload.ImdbID,
		})
	}
	if payload.ImdbID != "" {
		result.ExternalIDs = append(result.ExternalIDs, catalog.ExternalID{
			Kind:       request.Kind,
			ItemID:     request.ID,
			Provider:   "imdb",
			ExternalID: payload.ImdbID,
			UpdatedAt:  now,
		})
	}
	ratings := omdbRatings(request, payload, now)
	if err := s.catalog.UpsertRatings(ctx, ratings); err != nil {
		return err
	}
	result.Ratings = append(result.Ratings, ratings...)
	return nil
}

func (s *Service) refreshTMDB(ctx context.Context, request RefreshRequest, result *RefreshResult) error {
	if request.Kind != "movie" && request.Kind != "series" {
		return nil
	}
	path := "movie"
	if request.Kind == "series" {
		path = "tv"
	}
	searchURL := fmt.Sprintf("https://api.themoviedb.org/3/search/%s?", path) + url.Values{
		"api_key": {s.cfg.TMDBAPIKey},
		"query":   {request.Title},
	}.Encode()
	if request.Year > 0 && request.Kind == "movie" {
		searchURL += "&year=" + strconv.Itoa(request.Year)
	}
	var search tmdbSearch
	if err := s.getJSON(ctx, searchURL, &search); err != nil {
		return err
	}
	if len(search.Results) == 0 {
		return errors.New("no TMDB match")
	}
	match := search.Results[0]
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_ = s.catalog.UpsertExternalID(ctx, catalog.ExternalID{Kind: request.Kind, ItemID: request.ID, Provider: "tmdb", ExternalID: strconv.Itoa(match.ID)})
	rating := catalog.Rating{
		Kind:         request.Kind,
		ItemID:       request.ID,
		Provider:     "tmdb",
		RatingType:   "tmdb",
		Value:        match.VoteAverage,
		DisplayValue: fmt.Sprintf("%.1f/10", match.VoteAverage),
		Scale:        10,
		Votes:        match.VoteCount,
		SourceURL:    fmt.Sprintf("https://www.themoviedb.org/%s/%d", path, match.ID),
		FetchedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.catalog.UpsertRatings(ctx, []catalog.Rating{rating}); err != nil {
		return err
	}
	result.Ratings = append(result.Ratings, rating)
	return nil
}

func (s *Service) getJSON(ctx context.Context, endpoint string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("provider returned %s", response.Status)
	}
	return json.NewDecoder(response.Body).Decode(target)
}

type omdbResponse struct {
	Title      string       `json:"Title"`
	Year       string       `json:"Year"`
	ImdbID     string       `json:"imdbID"`
	ImdbRating string       `json:"imdbRating"`
	ImdbVotes  string       `json:"imdbVotes"`
	Metascore  string       `json:"Metascore"`
	Plot       string       `json:"Plot"`
	Poster     string       `json:"Poster"`
	Ratings    []omdbRating `json:"Ratings"`
	Response   string       `json:"Response"`
	Error      string       `json:"Error"`
}

type omdbRating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

type tmdbSearch struct {
	Results []tmdbResult `json:"results"`
}

type tmdbResult struct {
	ID          int     `json:"id"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
}

func omdbRatings(request RefreshRequest, payload omdbResponse, now string) []catalog.Rating {
	output := []catalog.Rating{}
	if value, ok := parseFloat(payload.ImdbRating); ok {
		output = append(output, catalog.Rating{Kind: request.Kind, ItemID: request.ID, Provider: "omdb", RatingType: "imdb", Value: value, DisplayValue: payload.ImdbRating + "/10", Scale: 10, Votes: parseVotes(payload.ImdbVotes), SourceURL: imdbURL(payload.ImdbID), FetchedAt: now, UpdatedAt: now})
	}
	if value, ok := parseFloat(payload.Metascore); ok {
		output = append(output, catalog.Rating{Kind: request.Kind, ItemID: request.ID, Provider: "omdb", RatingType: "metacritic", Value: value, DisplayValue: payload.Metascore, Scale: 100, SourceURL: imdbURL(payload.ImdbID), FetchedAt: now, UpdatedAt: now})
	}
	for _, rating := range payload.Ratings {
		switch strings.ToLower(rating.Source) {
		case "rotten tomatoes":
			if value, ok := parsePercent(rating.Value); ok {
				output = append(output, catalog.Rating{Kind: request.Kind, ItemID: request.ID, Provider: "omdb", RatingType: "rottenTomatoesCritics", Value: value, DisplayValue: rating.Value, Scale: 100, SourceURL: imdbURL(payload.ImdbID), FetchedAt: now, UpdatedAt: now})
			}
		}
	}
	return output
}

func omdbType(kind string) string {
	if kind == "series" || kind == "episode" {
		return "series"
	}
	return "movie"
}

func parseYear(value string, fallback int) int {
	if len(value) >= 4 {
		if parsed, err := strconv.Atoi(value[:4]); err == nil {
			return parsed
		}
	}
	return fallback
}

func parseVotes(value string) int {
	value = strings.ReplaceAll(value, ",", "")
	parsed, _ := strconv.Atoi(value)
	return parsed
}

func parsePercent(value string) (float64, bool) {
	value = strings.TrimSuffix(strings.TrimSpace(value), "%")
	return parseFloat(value)
}

func parseFloat(value string) (float64, bool) {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "N/A") {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	return parsed, err == nil
}

func imdbURL(id string) string {
	if id == "" {
		return ""
	}
	return "https://www.imdb.com/title/" + id + "/"
}

func emptyNA(value string) string {
	if strings.EqualFold(value, "N/A") {
		return ""
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func mustJSON(value any) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}
