package media

type Kind string

const (
	KindMovie   Kind = "movie"
	KindEpisode Kind = "episode"
	KindExtra   Kind = "extra"
	KindUnknown Kind = "unknown"
)

type SourceID string

type Source struct {
	ID        SourceID `json:"id"`
	LibraryID string   `json:"libraryId"`
	Kind      Kind     `json:"kind"`
	Path      string   `json:"path"`
	Container string   `json:"container"`
	Duration  float64  `json:"duration"`
	Bitrate   int64    `json:"bitrate"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}
