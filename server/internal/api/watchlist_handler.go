package api

import (
	"net/http"

	"github.com/jampat000/Xuva/server/internal/catalog"
	"github.com/jampat000/Xuva/server/internal/watchlist"
)

type watchlistAddRequest struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	ItemID string `json:"itemId"`
}

func watchlistListHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Watchlist == nil {
			writeError(w, http.StatusServiceUnavailable, "watchlist service unavailable")
			return
		}
		items, err := deps.Watchlist.List(r.Context(), requestUserID(r))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "watchlist lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": watchlistResponseItems(r, deps, items)})
	}
}

func watchlistAddHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Watchlist == nil {
			writeError(w, http.StatusServiceUnavailable, "watchlist service unavailable")
			return
		}
		var payload watchlistAddRequest
		if !decodeJSON(w, r, &payload) {
			return
		}
		itemID := firstNonEmpty(payload.ItemID, payload.ID)
		kind := watchlist.NormalizeKind(payload.Kind)
		if kind == "" || itemID == "" {
			writeError(w, http.StatusBadRequest, "kind and item id are required")
			return
		}
		if !watchlistCatalogItemExists(r, deps, kind, itemID) {
			writeError(w, http.StatusNotFound, "watchlist item not found")
			return
		}
		item, err := deps.Watchlist.Add(r.Context(), requestUserID(r), kind, itemID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "watchlist update failed")
			return
		}
		responseItems := watchlistResponseItems(r, deps, []watchlist.Item{item})
		if len(responseItems) == 0 {
			writeError(w, http.StatusNotFound, "watchlist item not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": responseItems[0]})
	}
}

func watchlistRemoveHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Watchlist == nil {
			writeError(w, http.StatusServiceUnavailable, "watchlist service unavailable")
			return
		}
		if err := deps.Watchlist.Remove(r.Context(), requestUserID(r), r.PathValue("id")); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func watchlistCatalogItemExists(r *http.Request, deps Deps, kind string, itemID string) bool {
	if deps.Catalog == nil {
		return true
	}
	switch kind {
	case "movie":
		_, ok, err := deps.Catalog.GetMovie(r.Context(), itemID)
		return err == nil && ok
	case "series":
		_, ok, err := deps.Catalog.GetSeries(r.Context(), itemID)
		return err == nil && ok
	default:
		return false
	}
}

func watchlistResponseItems(r *http.Request, deps Deps, items []watchlist.Item) []map[string]any {
	output := make([]map[string]any, 0, len(items))
	for _, item := range items {
		entry, ok := watchlistResponseItem(r, deps, item)
		if ok {
			output = append(output, entry)
		}
	}
	return output
}

func watchlistResponseItem(r *http.Request, deps Deps, item watchlist.Item) (map[string]any, bool) {
	if deps.Catalog != nil {
		switch item.Kind {
		case "movie":
			detail, ok, err := deps.Catalog.GetMovie(r.Context(), item.ItemID)
			if err != nil || !ok {
				return nil, false
			}
			entry := tvMovieItems([]catalog.MovieListItem{detail.MovieListItem})[0]
			entry["id"] = item.ID
			entry["itemId"] = item.ItemID
			entry["addedAt"] = item.AddedAt
			return entry, true
		case "series":
			detail, ok, err := deps.Catalog.GetSeries(r.Context(), item.ItemID)
			if err != nil || !ok {
				return nil, false
			}
			entry := tvSeriesItems([]catalog.SeriesListItem{detail.SeriesListItem})[0]
			entry["id"] = item.ID
			entry["itemId"] = item.ItemID
			entry["addedAt"] = item.AddedAt
			return entry, true
		}
	}
	return map[string]any{
		"id":      item.ID,
		"kind":    item.Kind,
		"itemId":  item.ItemID,
		"addedAt": item.AddedAt,
	}, true
}
