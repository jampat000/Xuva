package devices

type Profile struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Containers          []string `json:"containers"`
	VideoCodecs         []string `json:"videoCodecs"`
	AudioCodecs         []string `json:"audioCodecs"`
	SubtitleCodecs      []string `json:"subtitleCodecs"`
	SupportsHDR         bool     `json:"supportsHdr"`
	SupportsToneMapping bool     `json:"supportsToneMapping"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}
