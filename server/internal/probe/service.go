package probe

import (
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

type Result struct {
	Container       string  `json:"container"`
	DurationSeconds float64 `json:"durationSeconds"`
	Bitrate         int64   `json:"bitrate"`
	VideoCodec      string  `json:"videoCodec"`
	VideoProfile    string  `json:"videoProfile,omitempty"`
	VideoLevel      string  `json:"videoLevel,omitempty"`
	VideoBitDepth   int     `json:"videoBitDepth,omitempty"`
	VideoFrameRate  float64 `json:"videoFrameRate,omitempty"`
	PixelFormat     string  `json:"pixelFormat,omitempty"`
	ColorPrimaries  string  `json:"colorPrimaries,omitempty"`
	ColorTransfer   string  `json:"colorTransfer,omitempty"`
	ColorSpace      string  `json:"colorSpace,omitempty"`
	HDRFormat       string  `json:"hdrFormat,omitempty"` // derived: "" | "HDR10" | "HLG" | "BT2020" | "DolbyVision" | "DolbyVision+HDR10"
	DoviProfile     int     `json:"doviProfile,omitempty"` // Dolby Vision profile number (5, 7, 8, 9...), 0 = not DV
	MaxCLL          int     `json:"maxCll,omitempty"`      // Maximum Content Light Level in nits
	MaxFALL         int     `json:"maxFall,omitempty"`     // Maximum Average Frame Average Light Level in nits
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	AudioStreams    int     `json:"audioStreams"`
	SubtitleStreams int     `json:"subtitleStreams"`
	AudioTracks     []Track `json:"audioTracks"`
	SubtitleTracks  []Track `json:"subtitleTracks"`
	RawJSON         string  `json:"rawJson"`
}

type Track struct {
	Index    int    `json:"index"`
	Codec    string `json:"codec"`
	Language string `json:"language,omitempty"`
	Title    string `json:"title,omitempty"`
	Channels int    `json:"channels,omitempty"`
	Forced   bool   `json:"forced,omitempty"`
	Default  bool   `json:"default,omitempty"`
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
				result.VideoProfile = stream.Profile
				result.VideoLevel = levelString(stream.Level)
				result.PixelFormat = stream.PixFmt
				result.VideoBitDepth = bitDepthFromPixFmt(stream.PixFmt)
				result.ColorPrimaries = stream.ColorPrimaries
				result.ColorTransfer = stream.ColorTransfer
				result.ColorSpace = stream.ColorSpace
				result.VideoFrameRate = parseFrameRate(stream.AvgFrameRate, stream.RFrameRate)
				// Scan side data for Dolby Vision config and HDR metadata
				for _, sd := range stream.SideDataList {
					sdType := strings.ToLower(sd.SideDataType)
					if strings.Contains(sdType, "dovi") || strings.Contains(sdType, "dolby vision") {
						if sd.DVProfile > 0 {
							result.DoviProfile = sd.DVProfile
						} else if sd.DVVersionMajor > 0 {
							// DV detected but profile not parseable; record generic DV
							result.DoviProfile = -1
						}
					}
					if strings.Contains(sdType, "content light level") {
						if sd.MaxContent > 0 {
							result.MaxCLL = sd.MaxContent
						}
						if sd.MaxAverage > 0 {
							result.MaxFALL = sd.MaxAverage
						}
					}
				}
				result.HDRFormat = deriveHDRFormat(stream.ColorPrimaries, stream.ColorTransfer, result.VideoBitDepth, result.DoviProfile)
			}
		case "audio":
			result.AudioStreams++
			result.AudioTracks = append(result.AudioTracks, Track{Index: stream.Index, Codec: stream.CodecName, Language: stream.Tags.Language, Title: stream.Tags.Title, Channels: stream.Channels, Default: stream.Disposition.Default == 1})
		case "subtitle":
			result.SubtitleStreams++
			result.SubtitleTracks = append(result.SubtitleTracks, Track{Index: stream.Index, Codec: stream.CodecName, Language: stream.Tags.Language, Title: stream.Tags.Title, Forced: stream.Disposition.Forced == 1, Default: stream.Disposition.Default == 1})
		}
	}
	return result, nil
}

type ffprobePayload struct {
	Streams []struct {
		CodecType      string `json:"codec_type"`
		CodecName      string `json:"codec_name"`
		Index          int    `json:"index"`
		Width          int    `json:"width"`
		Height         int    `json:"height"`
		Channels       int    `json:"channels"`
		Profile        string `json:"profile"`
		Level          int    `json:"level"`
		PixFmt         string `json:"pix_fmt"`
		ColorPrimaries string `json:"color_primaries"`
		ColorTransfer  string `json:"color_transfer"`
		ColorSpace     string `json:"color_space"`
		AvgFrameRate   string `json:"avg_frame_rate"`
		RFrameRate     string `json:"r_frame_rate"`
		Tags           struct {
			Language string `json:"language"`
			Title    string `json:"title"`
		} `json:"tags"`
		Disposition struct {
			Default int `json:"default"`
			Forced  int `json:"forced"`
		} `json:"disposition"`
		SideDataList []struct {
			SideDataType string `json:"side_data_type"`
			// Dolby Vision configuration record
			DVVersionMajor int `json:"dv_version_major,omitempty"`
			DVProfile      int `json:"dv_profile,omitempty"`
			// MaxCLL / MaxFALL (HDR10+, HDR10 metadata blocks)
			MaxContent    int `json:"max_content,omitempty"`
			MaxAverage    int `json:"max_average,omitempty"`
		} `json:"side_data_list,omitempty"`
	} `json:"streams"`
	Format struct {
		FormatName string `json:"format_name"`
		Duration   string `json:"duration"`
		BitRate    string `json:"bit_rate"`
	} `json:"format"`
}

// bitDepthFromPixFmt maps an ffmpeg pix_fmt string to luma bit depth.
// Covers the common formats used in real-world media; returns 0 for unknown
// so callers can treat the field as "unspecified" rather than assume 8-bit.
func bitDepthFromPixFmt(pixFmt string) int {
	pixFmt = strings.ToLower(strings.TrimSpace(pixFmt))
	switch {
	case pixFmt == "":
		return 0
	case strings.Contains(pixFmt, "p016") || strings.Contains(pixFmt, "16le") || strings.Contains(pixFmt, "16be"):
		return 16
	case strings.Contains(pixFmt, "p012") || strings.Contains(pixFmt, "12le") || strings.Contains(pixFmt, "12be"):
		return 12
	case strings.Contains(pixFmt, "p010") || strings.Contains(pixFmt, "10le") || strings.Contains(pixFmt, "10be"):
		return 10
	default:
		return 8
	}
}

// deriveHDRFormat infers an HDR classification from color metadata and Dolby
// Vision profile number. doviProfile 0 = no DV, -1 = DV detected without
// a parseable profile number.
func deriveHDRFormat(primaries string, transfer string, bitDepth int, doviProfile int) string {
	primaries = strings.ToLower(strings.TrimSpace(primaries))
	transfer = strings.ToLower(strings.TrimSpace(transfer))

	baseHDR := ""
	if bitDepth == 0 || bitDepth >= 10 {
		switch transfer {
		case "smpte2084":
			baseHDR = "HDR10"
		case "arib-std-b67":
			baseHDR = "HLG"
		default:
			if primaries == "bt2020" {
				baseHDR = "BT2020"
			}
		}
	}

	if doviProfile != 0 {
		// Profile 5: DV only (no HDR10 compat layer)
		// Profile 7: DV + HDR10 metadata layer
		// Profile 8: DV + HDR10 compat layer (most common for disc/streaming)
		// Profile 9: DV SDR compat
		if baseHDR == "HDR10" || doviProfile == 7 || doviProfile == 8 {
			return "DolbyVision+HDR10"
		}
		return "DolbyVision"
	}
	return baseHDR
}

// parseFrameRate parses ffprobe rational-number frame-rate fields. It prefers
// avg_frame_rate (the actual playback rate); r_frame_rate is the codec base
// rate and can lie for VFR content.
func parseFrameRate(avg string, base string) float64 {
	if v := parseRational(avg); v > 0 {
		return v
	}
	return parseRational(base)
}

func parseRational(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" || value == "0/0" {
		return 0
	}
	parts := strings.SplitN(value, "/", 2)
	if len(parts) != 2 {
		return parseFloat(value)
	}
	num, _ := strconv.ParseFloat(parts[0], 64)
	den, _ := strconv.ParseFloat(parts[1], 64)
	if den == 0 {
		return 0
	}
	return num / den
}

// levelString renders an ffmpeg integer level (e.g. 51 → "5.1") for the
// codecs that use the convention (H.264, HEVC). For codecs that don't, it
// returns the raw value as a string.
func levelString(level int) string {
	if level <= 0 {
		return ""
	}
	if level >= 10 && level <= 99 {
		major := level / 10
		minor := level % 10
		return strconv.Itoa(major) + "." + strconv.Itoa(minor)
	}
	return strconv.Itoa(level)
}

func parseFloat(value string) float64 {
	parsed, _ := strconv.ParseFloat(value, 64)
	return parsed
}

func parseInt64(value string) int64 {
	parsed, _ := strconv.ParseInt(value, 10, 64)
	return parsed
}
