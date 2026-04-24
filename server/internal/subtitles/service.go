package subtitles

type Sidecar struct {
	Path              string `json:"path"`
	Format            string `json:"format"`
	Language          string `json:"language,omitempty"`
	Forced            bool   `json:"forced"`
	HearingImpaired   bool   `json:"hearingImpaired"`
	RequiresVideoBurn bool   `json:"requiresVideoBurn"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}
