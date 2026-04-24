package movies

import "github.com/vyrdenhq/vyrden/server/internal/media"

type Movie struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Year      int    `json:"year"`
	SortTitle string `json:"sortTitle"`
}

type Version struct {
	MovieID       string         `json:"movieId"`
	MediaSourceID media.SourceID `json:"mediaSourceId"`
	Edition       string         `json:"edition,omitempty"`
	QualityLabel  string         `json:"qualityLabel,omitempty"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}
