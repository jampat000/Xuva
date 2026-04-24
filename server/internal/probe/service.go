package probe

import (
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
)

type Result struct {
	Container       string  `json:"container"`
	DurationSeconds float64 `json:"durationSeconds"`
	Bitrate         int64   `json:"bitrate"`
	VideoCodec      string  `json:"videoCodec"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	AudioStreams    int     `json:"audioStreams"`
	SubtitleStreams int     `json:"subtitleStreams"`
	RawJSON         string  `json:"rawJson"`
}

type Service struct {
	ffprobePath string
}

func NewService(ffprobePath string) *Service {
	if ffprobePath == "" {
		ffprobePath = "ffprobe"
	}
	return &Service{ffprobePath: ffprobePath}
}

func (s *Service) Probe(ctx context.Context, path string) (Result, error) {
	cmd := exec.CommandContext(ctx, s.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)
	raw, err := cmd.Output()
	if err != nil {
		return Result{}, err
	}
	return Parse(raw)
}

func Parse(raw []byte) (Result, error) {
	var payload ffprobePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return Result{}, err
	}
	result := Result{
		Container:       payload.Format.FormatName,
		DurationSeconds: parseFloat(payload.Format.Duration),
		Bitrate:         parseInt64(payload.Format.BitRate),
		RawJSON:         string(raw),
	}
	for _, stream := range payload.Streams {
		switch stream.CodecType {
		case "video":
			if result.VideoCodec == "" {
				result.VideoCodec = stream.CodecName
				result.Width = stream.Width
				result.Height = stream.Height
			}
		case "audio":
			result.AudioStreams++
		case "subtitle":
			result.SubtitleStreams++
		}
	}
	return result, nil
}

type ffprobePayload struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
	Format struct {
		FormatName string `json:"format_name"`
		Duration   string `json:"duration"`
		BitRate    string `json:"bit_rate"`
	} `json:"format"`
}

func parseFloat(value string) float64 {
	parsed, _ := strconv.ParseFloat(value, 64)
	return parsed
}

func parseInt64(value string) int64 {
	parsed, _ := strconv.ParseInt(value, 10, 64)
	return parsed
}
