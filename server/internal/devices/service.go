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
	SupportsHLS         bool     `json:"supportsHls"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Profiles() []Profile {
	return []Profile{
		{
			ID:             "web",
			Name:           "Web Player",
			Containers:     []string{"mp4", "mov", "webm"},
			VideoCodecs:    []string{"h264", "av1", "vp9"},
			AudioCodecs:    []string{"aac", "opus", "mp3"},
			SubtitleCodecs: []string{"webvtt", "srt"},
			SupportsHLS:    true,
		},
		{
			ID:                  "android-tv",
			Name:                "Android TV",
			Containers:          []string{"mp4", "mkv", "webm", "mpegts"},
			VideoCodecs:         []string{"h264", "hevc", "av1", "vp9"},
			AudioCodecs:         []string{"aac", "ac3", "eac3", "opus", "mp3"},
			SubtitleCodecs:      []string{"srt", "ass", "webvtt", "pgs"},
			SupportsHDR:         true,
			SupportsToneMapping: true,
			SupportsHLS:         true,
		},
		{
			ID:                  "apple-tv",
			Name:                "Apple TV",
			Containers:          []string{"mp4", "mov", "m4v"},
			VideoCodecs:         []string{"h264", "hevc"},
			AudioCodecs:         []string{"aac", "ac3", "eac3", "alac"},
			SubtitleCodecs:      []string{"webvtt", "srt"},
			SupportsHDR:         true,
			SupportsToneMapping: true,
			SupportsHLS:         true,
		},
		{
			ID:                  "ios",
			Name:                "iPhone / iPad",
			Containers:          []string{"mp4", "mov", "m4v"},
			VideoCodecs:         []string{"h264", "hevc"},
			AudioCodecs:         []string{"aac", "ac3", "eac3", "alac"},
			SubtitleCodecs:      []string{"webvtt", "srt"},
			SupportsHDR:         true,
			SupportsToneMapping: true,
			SupportsHLS:         true,
		},
		{
			ID:             "chromecast",
			Name:           "Chromecast",
			Containers:     []string{"mp4", "webm"},
			VideoCodecs:    []string{"h264", "vp9", "av1"},
			AudioCodecs:    []string{"aac", "ac3", "eac3", "opus"},
			SubtitleCodecs: []string{"webvtt", "srt"},
			SupportsHDR:    true,
			SupportsHLS:    true,
		},
	}
}

func (s *Service) GetProfile(id string) (Profile, bool) {
	if id == "" {
		id = "web"
	}
	for _, profile := range s.Profiles() {
		if profile.ID == id {
			return profile, true
		}
	}
	return Profile{}, false
}
