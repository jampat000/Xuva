package probe

import "testing"

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
