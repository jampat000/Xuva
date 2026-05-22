package watchlist

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jampat000/Xuva/server/internal/database"
)

const DefaultUserID = "local"

type Item struct {
	ID      string `json:"id"`
	UserID  string `json:"userId"`
	Kind    string `json:"kind"`
	ItemID  string `json:"itemId"`
	AddedAt string `json:"addedAt"`
}

type Service struct {
	db *sql.DB
}

func NewService(database *database.Service) *Service {
	return &Service{db: database.DB()}
}

func (s *Service) List(ctx context.Context, userID string) ([]Item, error) {
	userID = normalizeUserID(userID)
	rows, err := s.db.QueryContext(ctx, `
		SELECT user_id, kind, item_id, added_at
		FROM watchlist_items
		WHERE user_id = ?
		ORDER BY added_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.UserID, &item.Kind, &item.ItemID, &item.AddedAt); err != nil {
			return nil, err
		}
		item.ID = EntryID(item.Kind, item.ItemID)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Add(ctx context.Context, userID string, kind string, itemID string) (Item, error) {
	userID = normalizeUserID(userID)
	kind = NormalizeKind(kind)
	itemID = strings.TrimSpace(itemID)
	if kind == "" {
		return Item{}, errors.New("kind must be movie or series")
	}
	if itemID == "" {
		return Item{}, errors.New("item id is required")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO watchlist_items(user_id, kind, item_id, added_at, updated_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(user_id, kind, item_id) DO UPDATE SET
			updated_at = excluded.updated_at
	`, userID, kind, itemID, now, now)
	if err != nil {
		return Item{}, err
	}
	return s.Get(ctx, userID, kind, itemID)
}

func (s *Service) Get(ctx context.Context, userID string, kind string, itemID string) (Item, error) {
	userID = normalizeUserID(userID)
	kind = NormalizeKind(kind)
	itemID = strings.TrimSpace(itemID)
	var item Item
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, kind, item_id, added_at
		FROM watchlist_items
		WHERE user_id = ? AND kind = ? AND item_id = ?
	`, userID, kind, itemID).Scan(&item.UserID, &item.Kind, &item.ItemID, &item.AddedAt)
	if err != nil {
		return Item{}, err
	}
	item.ID = EntryID(item.Kind, item.ItemID)
	return item, nil
}

func (s *Service) Remove(ctx context.Context, userID string, entryID string) error {
	userID = normalizeUserID(userID)
	kind, itemID := ParseEntryID(entryID)
	if itemID == "" {
		return errors.New("watchlist item id is required")
	}
	if kind == "" {
		_, err := s.db.ExecContext(ctx, `DELETE FROM watchlist_items WHERE user_id = ? AND item_id = ?`, userID, itemID)
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM watchlist_items WHERE user_id = ? AND kind = ? AND item_id = ?`, userID, kind, itemID)
	return err
}

func NormalizeKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "movie", "movies":
		return "movie"
	case "series", "tv", "show":
		return "series"
	default:
		return ""
	}
}

func EntryID(kind string, itemID string) string {
	return NormalizeKind(kind) + ":" + strings.TrimSpace(itemID)
}

func ParseEntryID(entryID string) (string, string) {
	entryID = strings.TrimSpace(entryID)
	if entryID == "" {
		return "", ""
	}
	kind, itemID, found := strings.Cut(entryID, ":")
	if !found {
		return "", entryID
	}
	return NormalizeKind(kind), strings.TrimSpace(itemID)
}

func normalizeUserID(userID string) string {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return DefaultUserID
	}
	return userID
}
