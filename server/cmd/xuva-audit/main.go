// xuva-audit — overnight media library health checker.
//
// Usage:
//
//	xuva-audit --data /path/to/data [--ffprobe ffprobe] [--limit N] [--out report.md]
//
// The tool reads Xuva's SQLite database, checks every media source for common
// problems, optionally re-runs ffprobe on files that are missing probe data,
// and writes a Markdown report.
//
// Exit codes: 0 = clean, 1 = issues found, 2 = fatal error.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	dataDir := flag.String("data", "data", "Xuva data directory (contains xuva.db)")
	ffprobePath := flag.String("ffprobe", "ffprobe", "path to ffprobe binary")
	limit := flag.Int("limit", 0, "max files to inspect (0 = all)")
	outPath := flag.String("out", "", "write report to file (default: stdout)")
	reprobeFlag := flag.Bool("reprobe", false, "re-run ffprobe on files missing probe data")
	flag.Parse()

	ctx := context.Background()

	dbPath := filepath.Join(*dataDir, "xuva.db")
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&mode=ro")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	sources, err := loadSources(ctx, db)
	if err != nil {
		log.Fatalf("load sources: %v", err)
	}

	if *limit > 0 && len(sources) > *limit {
		sources = sources[:*limit]
	}

	log.Printf("xuva-audit: inspecting %d media sources", len(sources))

	var issues []issue
	var stats summary

	stats.Total = len(sources)
	for i, src := range sources {
		if (i+1)%100 == 0 {
			log.Printf("  progress %d/%d …", i+1, len(sources))
		}
		fileIssues := auditSource(ctx, src, *ffprobePath, *reprobeFlag)
		issues = append(issues, fileIssues...)
		tally(&stats, src, fileIssues)
	}

	var w io.Writer = os.Stdout
	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			log.Fatalf("create output file: %v", err)
		}
		defer f.Close()
		w = f
	}

	writeReport(w, stats, issues)

	if *outPath != "" {
		log.Printf("xuva-audit: report written to %s", *outPath)
	}

	if len(issues) > 0 {
		os.Exit(1)
	}
}

// ─── Data model ──────────────────────────────────────────────────────────────

type source struct {
	ID          string
	Path        string
	RelPath     string
	Name        string
	Extension   string
	SizeBytes   int64
	Kind        string
	// Probe fields (nullable from LEFT JOIN)
	Container       string
	DurationSeconds float64
	Bitrate         int64
	VideoCodec      string
	VideoProfile    string
	VideoLevel      string
	VideoBitDepth   int
	VideoFrameRate  float64
	Width           int
	Height          int
	AudioStreams     int
	SubtitleStreams  int
	HDRFormat       string
	HasProbe        bool
}

type issueKind string

const (
	issueMissingFile     issueKind = "MISSING_FILE"
	issueUnprobed        issueKind = "UNPROBED"
	issueZeroDuration    issueKind = "ZERO_DURATION"
	issueNullBitrate     issueKind = "NULL_BITRATE"
	issueHighBitrate     issueKind = "HIGH_BITRATE"
	issueLowBitrate      issueKind = "LOW_BITRATE"
	issueNoAudio         issueKind = "NO_AUDIO"
	issueTranscodeCodec  issueKind = "TRANSCODE_NEEDED_CODEC"
	issueUnusualContainer issueKind = "UNUSUAL_CONTAINER"
	issueSmallFile       issueKind = "SMALL_FILE"
	issueHDR             issueKind = "HDR_CONTENT"
	issueNullResolution  issueKind = "NULL_RESOLUTION"
)

type issue struct {
	SourceID string
	RelPath  string
	Kind     issueKind
	Detail   string
}

type summary struct {
	Total           int
	Missing         int
	Unprobed        int
	HighBitrate     int
	TranscodeNeeded int
	UnusualContainer int
	NoAudio         int
	HDR             int
	Other           int
	Clean           int
}

// ─── Loader ───────────────────────────────────────────────────────────────────

func loadSources(ctx context.Context, db *sql.DB) ([]source, error) {
	const q = `
		SELECT
			ms.id, ms.path, ms.rel_path, ms.name, ms.extension,
			COALESCE(ms.size_bytes, 0), ms.kind,
			COALESCE(mp.container, ''),
			COALESCE(mp.duration_seconds, 0),
			COALESCE(mp.bitrate, 0),
			COALESCE(mp.video_codec, ''),
			COALESCE(mp.video_profile, ''),
			COALESCE(mp.video_level, ''),
			COALESCE(mp.video_bit_depth, 0),
			COALESCE(mp.video_frame_rate, 0),
			COALESCE(mp.width, 0),
			COALESCE(mp.height, 0),
			COALESCE(mp.audio_streams, 0),
			COALESCE(mp.subtitle_streams, 0),
			COALESCE(mp.hdr_format, ''),
			CASE WHEN mp.media_source_id IS NOT NULL THEN 1 ELSE 0 END AS has_probe
		FROM media_sources ms
		LEFT JOIN media_probes mp ON mp.media_source_id = ms.id
		ORDER BY ms.rel_path`

	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []source
	for rows.Next() {
		var s source
		var hasProbe int
		if err := rows.Scan(
			&s.ID, &s.Path, &s.RelPath, &s.Name, &s.Extension,
			&s.SizeBytes, &s.Kind,
			&s.Container, &s.DurationSeconds, &s.Bitrate,
			&s.VideoCodec, &s.VideoProfile, &s.VideoLevel,
			&s.VideoBitDepth, &s.VideoFrameRate,
			&s.Width, &s.Height,
			&s.AudioStreams, &s.SubtitleStreams,
			&s.HDRFormat, &hasProbe,
		); err != nil {
			return nil, err
		}
		s.HasProbe = hasProbe == 1
		out = append(out, s)
	}
	return out, rows.Err()
}

// ─── Auditor ─────────────────────────────────────────────────────────────────

// directPlayCodecs are video codecs Xuva will attempt to direct-play on Apple TV.
var directPlayCodecs = map[string]bool{
	"h264": true, "hevc": true, "av1": true, "vp9": true,
	"h265": true, "avc":  true,
}

// directPlayContainers are containers that usually direct-play on Apple TV.
var directPlayContainers = map[string]bool{
	"mov": true, "mp4": true, "m4v": true, "mkv": true, "m2ts": true, "ts": true,
}

const (
	highBitrateBps  = 40_000_000  // 40 Mbps — may cause buffering on LAN
	lowBitrateBps   = 100_000     // 100 kbps — probably corrupt/audio-only
	smallFileSizeKB = 1024        // < 1 MB — probably a stub or sample
)

func auditSource(ctx context.Context, s source, ffprobePath string, reprobe bool) []issue {
	var issues []issue

	add := func(k issueKind, detail string) {
		issues = append(issues, issue{SourceID: s.ID, RelPath: s.RelPath, Kind: k, Detail: detail})
	}

	// 1. Physical file existence
	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		add(issueMissingFile, "file not found on disk: "+s.Path)
		return issues // no point checking further
	}

	// 2. Suspiciously small file
	if s.SizeBytes > 0 && s.SizeBytes < smallFileSizeKB*1024 {
		add(issueSmallFile, fmt.Sprintf("only %d bytes — may be a stub", s.SizeBytes))
	}

	// 3. No probe data
	if !s.HasProbe {
		if reprobe && ffprobePath != "" {
			// Best-effort reprobe to check readability (we don't write back to DB).
			if err := quickProbe(ctx, ffprobePath, s.Path); err != nil {
				add(issueUnprobed, "no probe data and ffprobe failed: "+err.Error())
			} else {
				add(issueUnprobed, "no probe data in DB (ffprobe reads OK — run server probe job to fix)")
			}
		} else {
			add(issueUnprobed, "no probe data — run a probe job from the admin panel")
		}
		return issues
	}

	// 4. Zero duration
	if s.DurationSeconds <= 0 {
		add(issueZeroDuration, "probe reports 0 duration")
	}

	// 5. Null/zero bitrate (distinct from intentional 0 — means probe didn't extract it)
	if s.Bitrate <= 0 {
		add(issueNullBitrate, "probe has no bitrate — file may be incomplete or unusual format")
	} else if s.Bitrate > highBitrateBps {
		add(issueHighBitrate, fmt.Sprintf("%.0f Mbps — may cause buffering; remux may help", float64(s.Bitrate)/1_000_000))
	} else if s.Bitrate < lowBitrateBps && s.DurationSeconds > 10 {
		add(issueLowBitrate, fmt.Sprintf("%d kbps and >10s duration — file may be corrupt or audio-only", s.Bitrate/1000))
	}

	// 6. Null resolution
	if s.Width == 0 || s.Height == 0 {
		add(issueNullResolution, "probe has no resolution — may be audio-only or corrupt video track")
	}

	// 7. Codec compatibility
	codec := strings.ToLower(s.VideoCodec)
	if codec != "" && !directPlayCodecs[codec] {
		add(issueTranscodeCodec, fmt.Sprintf("codec %q is not in the direct-play list — playback will transcode", s.VideoCodec))
	}

	// 8. Container compatibility
	ext := strings.ToLower(strings.TrimPrefix(s.Extension, "."))
	container := strings.ToLower(s.Container)
	check := ext
	if container != "" {
		check = container
	}
	if check != "" && !directPlayContainers[check] {
		add(issueUnusualContainer, fmt.Sprintf("container/ext %q — verify compatibility", check))
	}

	// 9. No audio streams
	if s.AudioStreams == 0 && s.DurationSeconds > 0 {
		add(issueNoAudio, "no audio streams detected")
	}

	// 10. HDR content — not a bug, but worth flagging for review
	if s.HDRFormat != "" {
		add(issueHDR, fmt.Sprintf("HDR: %s — verify tone mapping on your client", s.HDRFormat))
	}

	return issues
}

// quickProbe runs ffprobe on a file and returns an error if it can't be read.
func quickProbe(ctx context.Context, ffprobePath, path string) error {
	ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx2, ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ─── Tallying ────────────────────────────────────────────────────────────────

func tally(s *summary, src source, issues []issue) {
	if len(issues) == 0 {
		s.Clean++
		return
	}
	for _, iss := range issues {
		switch iss.Kind {
		case issueMissingFile:
			s.Missing++
		case issueUnprobed:
			s.Unprobed++
		case issueHighBitrate:
			s.HighBitrate++
		case issueTranscodeCodec, issueUnusualContainer:
			s.TranscodeNeeded++
		case issueNoAudio:
			s.NoAudio++
		case issueHDR:
			s.HDR++
		default:
			s.Other++
		}
	}
	_ = src
}

// ─── Report writer ───────────────────────────────────────────────────────────

func writeReport(w io.Writer, s summary, issues []issue) {
	now := time.Now().UTC().Format("2006-01-02 15:04 UTC")
	fmt.Fprintf(w, "# Xuva Media Audit Report\n\n")
	fmt.Fprintf(w, "Generated: **%s**\n\n", now)

	// Summary table
	fmt.Fprintf(w, "## Summary\n\n")
	fmt.Fprintf(w, "| Metric | Count |\n|---|---|\n")
	fmt.Fprintf(w, "| Total files inspected | %d |\n", s.Total)
	fmt.Fprintf(w, "| ✅ Clean (no issues) | %d |\n", s.Clean)
	fmt.Fprintf(w, "| ❌ Missing from disk | %d |\n", s.Missing)
	fmt.Fprintf(w, "| ⚠️ Unprobed (no metadata) | %d |\n", s.Unprobed)
	fmt.Fprintf(w, "| 🔴 High bitrate (>40 Mbps) | %d |\n", s.HighBitrate)
	fmt.Fprintf(w, "| 🔄 Will need transcoding | %d |\n", s.TranscodeNeeded)
	fmt.Fprintf(w, "| 🔕 No audio streams | %d |\n", s.NoAudio)
	fmt.Fprintf(w, "| 🌈 HDR content | %d |\n", s.HDR)
	fmt.Fprintf(w, "| ℹ️ Other issues | %d |\n\n", s.Other)

	if len(issues) == 0 {
		fmt.Fprintf(w, "✅ **No issues found.** All files look good.\n")
		return
	}

	// Group by kind for the detail sections
	byKind := map[issueKind][]issue{}
	for _, iss := range issues {
		byKind[iss.Kind] = append(byKind[iss.Kind], iss)
	}

	// Print sections in priority order
	order := []issueKind{
		issueMissingFile,
		issueUnprobed,
		issueSmallFile,
		issueZeroDuration,
		issueNullBitrate,
		issueNullResolution,
		issueHighBitrate,
		issueLowBitrate,
		issueTranscodeCodec,
		issueUnusualContainer,
		issueNoAudio,
		issueHDR,
	}
	headers := map[issueKind]string{
		issueMissingFile:      "❌ Missing files",
		issueUnprobed:         "⚠️ Unprobed files",
		issueSmallFile:        "🔍 Suspiciously small files",
		issueZeroDuration:     "🔍 Zero duration",
		issueNullBitrate:      "🔍 Null/zero bitrate",
		issueNullResolution:   "🔍 No resolution data",
		issueHighBitrate:      "🔴 High bitrate (>40 Mbps)",
		issueLowBitrate:       "🔍 Abnormally low bitrate",
		issueTranscodeCodec:   "🔄 Non-direct-play codec",
		issueUnusualContainer: "🔄 Unusual container",
		issueNoAudio:          "🔕 No audio streams",
		issueHDR:              "🌈 HDR content",
	}

	fmt.Fprintf(w, "## Issues\n\n")

	for _, kind := range order {
		list, ok := byKind[kind]
		if !ok {
			continue
		}
		sort.Slice(list, func(i, j int) bool { return list[i].RelPath < list[j].RelPath })
		fmt.Fprintf(w, "### %s (%d)\n\n", headers[kind], len(list))
		fmt.Fprintf(w, "| File | Detail |\n|---|---|\n")
		for _, iss := range list {
			// Escape pipes in paths/details for markdown tables
			p := strings.ReplaceAll(iss.RelPath, "|", "\\|")
			d := strings.ReplaceAll(iss.Detail, "|", "\\|")
			fmt.Fprintf(w, "| `%s` | %s |\n", p, d)
		}
		fmt.Fprintf(w, "\n")
	}

	fmt.Fprintf(w, "---\n*Generated by xuva-audit*\n")
}
