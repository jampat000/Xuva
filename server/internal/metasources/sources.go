package metasources

import (
	"strings"

	"github.com/jampat000/Xuva/server/internal/config"
)

type SourceDefinition struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Coverage       string   `json:"coverage,omitempty"`
	Note           string   `json:"note,omitempty"`
	Kinds          []string `json:"kinds"`
	Local          bool     `json:"local"`
	Managed        bool     `json:"managed"`
	RequiresConfig bool     `json:"requiresConfig"`
	Available      bool     `json:"available"`
}

func SourceCatalog(cfg config.Config) []SourceDefinition {
	return []SourceDefinition{
		{ID: "filename", Name: "Filename and folders", Description: "Fast local title and year parsing from library paths.", Coverage: "Movies and TV", Note: "Always available", Kinds: []string{"movie", "series"}, Local: true, Available: true},
		{ID: "nfo", Name: "Local NFO", Description: "Reads movie.nfo, tvshow.nfo, and sidecar NFO files in the library.", Coverage: "Movies and TV with sidecars", Note: "Always available", Kinds: []string{"movie", "series"}, Local: true, Available: true},
		{ID: "artwork", Name: "Local artwork", Description: "Uses poster, folder, cover, fanart, and backdrop image sidecars in the library.", Coverage: "Movies and TV with sidecar images", Note: "Always available", Kinds: []string{"movie", "series"}, Local: true, Available: true},
		{ID: "wikipedia", Name: "Wikipedia", Description: "Adds richer summaries and artwork when a matching article is available.", Coverage: "Movies and TV", Note: "No user account required", Kinds: []string{"movie", "series"}, Available: true},
		{ID: "wikidata", Name: "Wikidata", Description: "Adds structured labels, descriptions, poster files, and external IDs from Wikimedia data.", Coverage: "Movies and TV", Note: "No user account required", Kinds: []string{"movie", "series"}, Available: true},
		{ID: "tvmaze", Name: "TVMaze", Description: "Adds series metadata, external IDs, and TV ratings without a user account.", Coverage: "TV libraries", Note: "No user account required", Kinds: []string{"series"}, Available: true},
		{ID: "tmdb", Name: "TMDB", Description: "Adds TMDB IDs, artwork, and community ratings through the server-managed metadata layer.", Coverage: "Movies and TV", Note: "Managed by server credentials", Kinds: []string{"movie", "series"}, Managed: true, RequiresConfig: true, Available: strings.TrimSpace(cfg.TMDBAPIKey) != ""},
		{ID: "tvdb", Name: "TheTVDB", Description: "Adds TV and movie metadata, IDs, and ratings through the server-managed metadata layer.", Coverage: "Movies and TV", Note: "Managed by server credentials", Kinds: []string{"movie", "series"}, Managed: true, RequiresConfig: true, Available: strings.TrimSpace(cfg.TVDBAPIKey) != ""},
		{ID: "omdb", Name: "OMDb", Description: "Adds IMDb, Rotten Tomatoes, and Metacritic ratings through the server-managed metadata layer.", Coverage: "Movies and TV", Note: "Managed by server credentials", Kinds: []string{"movie", "series"}, Managed: true, RequiresConfig: true, Available: strings.TrimSpace(cfg.OMDbAPIKey) != ""},
	}
}

func DefaultSourceOrder(kind string) []string {
	switch NormalizeKind(kind) {
	case "series":
		return []string{"nfo", "artwork", "tvmaze", "tvdb", "tmdb", "wikipedia", "wikidata", "omdb", "filename"}
	default:
		return []string{"nfo", "artwork", "tmdb", "tvdb", "wikipedia", "wikidata", "omdb", "filename"}
	}
}

func NormalizeSourceOrder(kind string, requested []string, cfg config.Config) []string {
	return NormalizeRequestedSourceOrder(kind, requested)
}

func NormalizeRequestedSourceOrder(kind string, requested []string) []string {
	normalizedKind := NormalizeKind(kind)
	known := map[string]SourceDefinition{}
	for _, source := range SourceCatalog(config.Config{}) {
		if SupportsKind(source, normalizedKind) {
			known[source.ID] = source
		}
	}
	output := []string{}
	seen := map[string]struct{}{}
	for _, id := range requested {
		id = strings.ToLower(strings.TrimSpace(id))
		_, ok := known[id]
		if !ok {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		output = append(output, id)
	}
	if len(output) > 0 {
		return output
	}
	for _, id := range DefaultSourceOrder(normalizedKind) {
		_, ok := known[id]
		if !ok {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		output = append(output, id)
	}
	return output
}

func SourceCatalogByKind(cfg config.Config) map[string][]SourceDefinition {
	output := map[string][]SourceDefinition{
		"movie":  {},
		"series": {},
	}
	for _, source := range SourceCatalog(cfg) {
		for _, kind := range source.Kinds {
			output[kind] = append(output[kind], source)
		}
	}
	return output
}

func SupportsKind(source SourceDefinition, kind string) bool {
	for _, candidate := range source.Kinds {
		if candidate == kind {
			return true
		}
	}
	return false
}

func NormalizeKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "tv", "series":
		return "series"
	default:
		return "movie"
	}
}
