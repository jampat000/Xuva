package probe

import (
	"math"
	"testing"
)

func TestParseFFprobeResult(t *testing.T) {
	raw := []byte(`{
		"streams": [
			{"codec_type":"video","codec_name":"hevc","width":3840,"height":2160},
			{"codec_type":"audio","codec_name":"truehd"},
			{"codec_type":"audio","codec_name":"ac3"},
			{"codec_type":"subtitle","codec_name":"subrip"}
		],
		"format": {"format_name":"matroska,webm","duration":"123.456","bit_rate":"42000000"}
	}`)

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if result.Container != "matroska,webm" {
		t.Fatalf("expected container, got %q", result.Container)
	}
	if result.VideoCodec != "hevc" || result.Width != 3840 || result.Height != 2160 {
		t.Fatalf("unexpected video result: %#v", result)
	}
	if result.AudioStreams != 2 || result.SubtitleStreams != 1 {
		t.Fatalf("unexpected stream counts: %#v", result)
	}
	if result.Bitrate != 42000000 {
		t.Fatalf("expected bitrate, got %d", result.Bitrate)
	}
}

func TestParseExtractsProfileLevelAndBitDepth(t *testing.T) {
	// H.264-style level encoding (level_idc, where 4.1 → 41). HEVC encodes
	// differently (level_idc * 30); we store HEVC values as-is and rely on
	// the display layer for pretty rendering. This test exercises the
	// H.264 path of levelString().
	raw := []byte(`{
		"streams": [
			{"codec_type":"video","codec_name":"h264","width":1920,"height":1080,
			 "profile":"High 10","level":41,"pix_fmt":"yuv420p10le",
			 "avg_frame_rate":"24000/1001","r_frame_rate":"24000/1001"}
		],
		"format": {"format_name":"mp4","duration":"100","bit_rate":"40000000"}
	}`)

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if result.VideoProfile != "High 10" {
		t.Fatalf("expected video profile High 10, got %q", result.VideoProfile)
	}
	if result.VideoLevel != "4.1" {
		t.Fatalf("expected level 4.1 (H.264 level 41), got %q", result.VideoLevel)
	}
	if result.VideoBitDepth != 10 {
		t.Fatalf("expected bit depth 10 for yuv420p10le, got %d", result.VideoBitDepth)
	}
	if math.Abs(result.VideoFrameRate-23.976) > 0.01 {
		t.Fatalf("expected ~23.976 fps, got %f", result.VideoFrameRate)
	}
}

func TestParseDerivesHDRFormatFromColorMetadata(t *testing.T) {
	cases := []struct {
		name     string
		primaries string
		transfer string
		pixFmt   string
		want     string
	}{
		{"hdr10", "bt2020", "smpte2084", "yuv420p10le", "HDR10"},
		{"hlg", "bt2020", "arib-std-b67", "yuv420p10le", "HLG"},
		{"bt2020_only", "bt2020", "", "yuv420p10le", "BT2020"},
		{"sdr_bt709", "bt709", "bt709", "yuv420p", ""},
		// HDR transfer at 8-bit is treated as bogus (no real-world file does
		// this); we suppress the HDR classification rather than promising
		// HDR playback we can't actually deliver.
		{"hdr10_at_8bit_ignored", "bt2020", "smpte2084", "yuv420p", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte(`{
				"streams": [{
					"codec_type":"video","codec_name":"hevc","width":3840,"height":2160,
					"pix_fmt":"` + tc.pixFmt + `","color_primaries":"` + tc.primaries + `",
					"color_transfer":"` + tc.transfer + `","color_space":"bt2020nc"
				}],
				"format":{"format_name":"mp4"}
			}`)
			result, err := Parse(raw)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if result.HDRFormat != tc.want {
				t.Fatalf("expected HDR=%q, got %q (pix_fmt=%q transfer=%q)", tc.want, result.HDRFormat, tc.pixFmt, tc.transfer)
			}
		})
	}
}
