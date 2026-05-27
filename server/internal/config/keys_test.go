package config

import "testing"

func TestApplyEmbeddedProviderDefaultsUsesEnvWhenUnset(t *testing.T) {
	origTMDB := DefaultTMDBAPIKey
	origFanart := DefaultFanartTVAPIKey
	origOMDb := DefaultOMDbAPIKey
	t.Cleanup(func() {
		DefaultTMDBAPIKey = origTMDB
		DefaultFanartTVAPIKey = origFanart
		DefaultOMDbAPIKey = origOMDb
	})

	t.Setenv("XUVA_DEFAULT_TMDB_API_KEY", "tmdb-env-key")
	t.Setenv("XUVA_DEFAULT_FANARTTV_API_KEY", "fanart-env-key")
	t.Setenv("XUVA_DEFAULT_OMDB_API_KEY", "omdb-env-key")

	DefaultTMDBAPIKey = ""
	DefaultFanartTVAPIKey = ""
	DefaultOMDbAPIKey = ""

	applyEmbeddedProviderDefaults()

	if DefaultTMDBAPIKey != "tmdb-env-key" {
		t.Fatalf("expected TMDB env default, got %q", DefaultTMDBAPIKey)
	}
	if DefaultFanartTVAPIKey != "fanart-env-key" {
		t.Fatalf("expected Fanart env default, got %q", DefaultFanartTVAPIKey)
	}
	if DefaultOMDbAPIKey != "omdb-env-key" {
		t.Fatalf("expected OMDb env default, got %q", DefaultOMDbAPIKey)
	}
}

func TestApplyEmbeddedProviderDefaultsPreservesEmbeddedValues(t *testing.T) {
	origTMDB := DefaultTMDBAPIKey
	origFanart := DefaultFanartTVAPIKey
	origOMDb := DefaultOMDbAPIKey
	t.Cleanup(func() {
		DefaultTMDBAPIKey = origTMDB
		DefaultFanartTVAPIKey = origFanart
		DefaultOMDbAPIKey = origOMDb
	})

	t.Setenv("XUVA_DEFAULT_TMDB_API_KEY", "tmdb-env-key")
	t.Setenv("XUVA_DEFAULT_FANARTTV_API_KEY", "fanart-env-key")
	t.Setenv("XUVA_DEFAULT_OMDB_API_KEY", "omdb-env-key")

	DefaultTMDBAPIKey = "tmdb-embedded-key"
	DefaultFanartTVAPIKey = "fanart-embedded-key"
	DefaultOMDbAPIKey = "omdb-embedded-key"

	applyEmbeddedProviderDefaults()

	if DefaultTMDBAPIKey != "tmdb-embedded-key" {
		t.Fatalf("expected TMDB embedded default to remain, got %q", DefaultTMDBAPIKey)
	}
	if DefaultFanartTVAPIKey != "fanart-embedded-key" {
		t.Fatalf("expected Fanart embedded default to remain, got %q", DefaultFanartTVAPIKey)
	}
	if DefaultOMDbAPIKey != "omdb-embedded-key" {
		t.Fatalf("expected OMDb embedded default to remain, got %q", DefaultOMDbAPIKey)
	}
}
