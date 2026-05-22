package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jampat000/Xuva/server/internal/auth"
	"github.com/jampat000/Xuva/server/internal/chapters"
)

func chaptersGetHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "media source id required")
			return
		}
		if deps.Chapters == nil {
			writeJSON(w, http.StatusOK, chapters.Chapters{MediaSourceID: id})
			return
		}
		ch, _, err := deps.Chapters.Get(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read chapters")
			return
		}
		writeJSON(w, http.StatusOK, ch)
	}
}

func chaptersAnalyzeHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "media source id required")
			return
		}
		if deps.Chapters == nil || deps.Catalog == nil {
			writeJSON(w, http.StatusOK, map[string]string{"status": "unavailable"})
			return
		}
		src, ok, err := deps.Catalog.GetMediaSource(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read media source")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "media source not found")
			return
		}

		// Fire credits analysis in background.
		go deps.Chapters.AnalyzeCredits(r.Context(), id, src.Path, src.DurationSeconds)

		// Fire season intro analysis if this is a TV episode.
		go func() {
			peers, err := deps.Catalog.GetSeasonPeers(r.Context(), id)
			if err != nil || len(peers) < 2 {
				return
			}
			eps := make([]chapters.EpisodeInput, 0, len(peers))
			for _, p := range peers {
				eps = append(eps, chapters.EpisodeInput{
					MediaSourceID: p.MediaSourceID,
					Path:          p.Path,
					Duration:      p.Duration,
				})
			}
			deps.Chapters.AnalyzeSeason(r.Context(), eps)
		}()

		writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
	}
}

func userPreferencesUpdateHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resolved, ok := auth.ResolvedSessionFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}
		var body auth.UserPreferencesPatch
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if deps.Auth == nil {
			writeJSON(w, http.StatusOK, body)
			return
		}
		preferences, err := deps.Auth.UpdateUserPreferences(r.Context(), resolved.Principal.ID, body)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidPreferences) {
				writeError(w, http.StatusBadRequest, "posterSize must be S, M, or L")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to save preferences")
			return
		}
		writeJSON(w, http.StatusOK, preferences)
	}
}
