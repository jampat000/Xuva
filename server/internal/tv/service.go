package tv

import "github.com/vyrdenhq/vyrden/server/internal/media"

type Series struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Year  int    `json:"year"`
}

type Season struct {
	ID           string `json:"id"`
	SeriesID     string `json:"seriesId"`
	SeasonNumber int    `json:"seasonNumber"`
}

type Episode struct {
	ID            string `json:"id"`
	SeriesID      string `json:"seriesId"`
	SeasonID      string `json:"seasonId"`
	EpisodeNumber int    `json:"episodeNumber"`
	Title         string `json:"title"`
}

type EpisodeVersion struct {
	EpisodeID     string         `json:"episodeId"`
	MediaSourceID media.SourceID `json:"mediaSourceId"`
	QualityLabel  string         `json:"qualityLabel,omitempty"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}
