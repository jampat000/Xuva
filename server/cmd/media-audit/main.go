// media-audit walks a media library, runs ffprobe on every file, and emits a
// structured per-file JSON report + a human-readable summary text file.
// Designed for overnight runs: resumable (skip files already in
// audit-report.json with matching size+mtime), bounded per-file timeout, no
// memory accumulation beyond the report itself, progress logging.
//
// Unlike server/cmd/xuva-audit (which reads Xuva's DB to know what files
// exist), this tool just walks directories — no Xuva install required. Useful
// for one-off library health checks before/after a scan, or for vetting a
// foreign drive of content before importing.
//
// Usage:
//
//	media-audit --library /mnt/media/Movies --library /mnt/media/TV \
//	            --out-dir ./reports [--ffprobe ffprobe] [--timeout 60s]
//
// Outputs land in <out-dir>:
//   - audit-report.json — full structured FileAudit per file
//   - audit-summary.txt — counts by issue + flagged-file list
//
// Exit codes: 0 = all files probed cleanly, 1 = one or more files flagged,
// 2 = fatal (bad arguments, output dir not writable).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// mediaExtensions is the set of file extensions we consider "media files"
// worth probing. Conservative — non-media files (.nfo, .srt, .jpg) are
// skipped fast without the 30 s timeout per file.
var mediaExtensions = map[string]bool{
	".mkv":  true,
	".mp4":  true,
	".m4v":  true,
	".avi":  true,
	".mov":  true,
	".wmv":  true,
	".flv":  true,
	".webm": true,
	".ts":   true,
	".mpg":  true,
	".mpeg": true,
	".vob":  true,
	".m2ts": true,
	".mts":  true,
	".3gp":  true,
	".asf":  true,
}

// FileAudit is one entry in audit-report.json. Captures the ffprobe-derived
// facts we care about for compatibility/health: codecs, container, durations,
// bitrate, HDR metadata, VFR, and any probe error.
type FileAudit struct {
	Path           string   `json:"path"`
	SizeBytes      int64    `json:"sizeBytes"`
	ModTimeUnix    int64    `json:"modTimeUnix"`
	ProbedAt       string   `json:"probedAt"`
	Container      string   `json:"container,omitempty"`
	DurationSec    float64  `json:"durationSec,omitempty"`
	BitrateBPS     int64    `json:"bitrateBps,omitempty"`
	VideoCodec     string   `json:"videoCodec,omitempty"`
	VideoProfile   string   `json:"videoProfile,omitempty"`
	VideoWidth     int      `json:"videoWidth,omitempty"`
	VideoHeight    int      `json:"videoHeight,omitempty"`
	VideoFPS       float64  `json:"videoFps,omitempty"`
	VariableFPS    bool     `json:"variableFps,omitempty"`
	HDRFormat      string   `json:"hdrFormat,omitempty"`
	ColorSpace     string   `json:"colorSpace,omitempty"`
	ColorTransfer  string   `json:"colorTransfer,omitempty"`
	AudioCodecs    []string `json:"audioCodecs,omitempty"`
	AudioChannels  []int    `json:"audioChannels,omitempty"`
	SubtitleTracks int      `json:"subtitleTracks,omitempty"`
	SubtitleCodecs []string `json:"subtitleCodecs,omitempty"`
	Issues         []string `json:"issues,omitempty"`
	ProbeError     string   `json:"probeError,omitempty"`
}

// ffprobeOutput mirrors the subset of the ffprobe JSON we read.
type ffprobeOutput struct {
	Streams []struct {
		CodecType        string `json:"codec_type"`
		CodecName        string `json:"codec_name"`
		Profile          string `json:"profile"`
		Width            int    `json:"width,omitempty"`
		Height           int    `json:"height,omitempty"`
		AvgFrameRate     string `json:"avg_frame_rate,omitempty"`
		RFrameRate       string `json:"r_frame_rate,omitempty"`
		Channels         int    `json:"channels,omitempty"`
		ColorSpace       string `json:"color_space,omitempty"`
		ColorTransfer    string `json:"color_transfer,omitempty"`
		ColorPrimaries   string `json:"color_primaries,omitempty"`
		SideDataList     []struct {
			SideDataType string `json:"side_data_type"`
		} `json:"side_data_list,omitempty"`
	} `json:"streams"`
	Format struct {
		FormatName string            `json:"format_name"`
		Duration   string            `json:"duration"`
		BitRate    string            `json:"bit_rate"`
		Tags       map[string]string `json:"tags,omitempty"`
	} `json:"format"`
}

// stringSliceFlag lets --library be passed multiple times.
type stringSliceFlag []string

func (s *stringSliceFlag) String() string     { return strings.Join(*s, ",") }
func (s *stringSliceFlag) Set(v string) error { *s = append(*s, v); return nil }

func main() {
	var libraries stringSliceFlag
	flag.Var(&libraries, "library", "media library root (repeatable; e.g. --library /mnt/movies --library /mnt/tv)")
	ffprobePath := flag.String("ffprobe", "ffprobe", "path to ffprobe binary")
	outDir := flag.String("out-dir", ".", "directory to write audit-report.json + audit-summary.txt")
	perFileTimeout := flag.Duration("timeout", 30*time.Second, "ffprobe timeout per file")
	maxFiles := flag.Int("limit", 0, "stop after this many files (0 = all)")
	noResume := flag.Bool("no-resume", false, "ignore an existing audit-report.json and reprobe every file")
	flag.Parse()

	if len(libraries) == 0 {
		fmt.Fprintln(os.Stderr, "media-audit: at least one --library path is required")
		flag.Usage()
		os.Exit(2)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("media-audit: out-dir not writable: %v", err)
	}
	if _, err := exec.LookPath(*ffprobePath); err != nil {
		log.Fatalf("media-audit: ffprobe not found at %q: %v (pass --ffprobe to override)", *ffprobePath, err)
	}

	// Graceful shutdown: SIGINT/SIGTERM cancels the run, then we still write
	// whatever we've collected so an overnight run isn't lost.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	reportPath := filepath.Join(*outDir, "audit-report.json")
	summaryPath := filepath.Join(*outDir, "audit-summary.txt")

	// Resume from a prior run unless told otherwise. Match key = path; reuse
	// if size+mtime are unchanged (the file is bit-identical).
	prior := map[string]FileAudit{}
	if !*noResume {
		if data, err := os.ReadFile(reportPath); err == nil {
			var existing []FileAudit
			if err := json.Unmarshal(data, &existing); err == nil {
				for _, e := range existing {
					prior[e.Path] = e
				}
				log.Printf("media-audit: resuming from %s (%d prior entries)", reportPath, len(prior))
			}
		}
	}

	results := make([]FileAudit, 0, 4096)
	total := 0
	probed := 0
	skipped := 0
	failed := 0
	start := time.Now()

walking:
	for _, root := range libraries {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil {
				log.Printf("media-audit: walk error at %s: %v", path, err)
				return nil
			}
			if d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if !mediaExtensions[ext] {
				return nil
			}
			total++
			info, statErr := d.Info()
			if statErr != nil {
				log.Printf("media-audit: stat %s: %v", path, statErr)
				return nil
			}

			// Resume: if we already audited this path and size+mtime match,
			// reuse the prior entry instead of reprobing.
			if existing, ok := prior[path]; ok && existing.SizeBytes == info.Size() && existing.ModTimeUnix == info.ModTime().Unix() {
				results = append(results, existing)
				skipped++
				return nil
			}

			audit := probeFile(ctx, *ffprobePath, path, *perFileTimeout)
			audit.SizeBytes = info.Size()
			audit.ModTimeUnix = info.ModTime().Unix()
			results = append(results, audit)
			probed++
			if audit.ProbeError != "" {
				failed++
			}
			if probed%100 == 0 {
				log.Printf("media-audit: probed %d new / skipped %d resumed / %d failed (%.0fs elapsed)",
					probed, skipped, failed, time.Since(start).Seconds())
			}
			if *maxFiles > 0 && probed >= *maxFiles {
				return errStopLimit
			}
			return nil
		})
		if errors.Is(err, errStopLimit) {
			log.Printf("media-audit: hit --limit %d; stopping", *maxFiles)
			break walking
		}
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("media-audit: walk of %s failed: %v", root, err)
		}
	}

	// Always write the report — even on Ctrl-C — so a 6-hour run isn't lost.
	if err := writeJSON(reportPath, results); err != nil {
		log.Fatalf("media-audit: write report: %v", err)
	}
	if err := writeSummary(summaryPath, results, total, probed, skipped, failed, time.Since(start)); err != nil {
		log.Fatalf("media-audit: write summary: %v", err)
	}
	log.Printf("media-audit: %d files (%d probed, %d resumed, %d failed) → %s + %s",
		len(results), probed, skipped, failed, reportPath, summaryPath)

	if anyIssues(results) {
		os.Exit(1)
	}
}

var errStopLimit = errors.New("media-audit: limit reached")

func probeFile(ctx context.Context, ffprobePath, path string, timeout time.Duration) FileAudit {
	probed := FileAudit{
		Path:     path,
		ProbedAt: time.Now().UTC().Format(time.RFC3339),
	}
	pCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	out, err := exec.CommandContext(pCtx, ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		path,
	).Output()
	if err != nil {
		probed.ProbeError = strings.TrimSpace(err.Error())
		probed.Issues = append(probed.Issues, "ffprobe failed (file may be corrupt or unreadable)")
		return probed
	}
	var raw ffprobeOutput
	if err := json.Unmarshal(out, &raw); err != nil {
		probed.ProbeError = "ffprobe output unparseable: " + err.Error()
		probed.Issues = append(probed.Issues, "ffprobe output invalid")
		return probed
	}
	probed.Container = raw.Format.FormatName
	if raw.Format.Duration != "" {
		probed.DurationSec, _ = strconv.ParseFloat(raw.Format.Duration, 64)
	}
	if raw.Format.BitRate != "" {
		probed.BitrateBPS, _ = strconv.ParseInt(raw.Format.BitRate, 10, 64)
	}
	for _, s := range raw.Streams {
		switch s.CodecType {
		case "video":
			if probed.VideoCodec == "" {
				probed.VideoCodec = s.CodecName
				probed.VideoProfile = s.Profile
				probed.VideoWidth = s.Width
				probed.VideoHeight = s.Height
				avg := parseFraction(s.AvgFrameRate)
				rTime := parseFraction(s.RFrameRate)
				probed.VideoFPS = avg
				probed.VariableFPS = avg > 0 && rTime > 0 && math.Abs(avg-rTime)/math.Max(avg, rTime) > 0.02
				probed.ColorSpace = s.ColorSpace
				probed.ColorTransfer = s.ColorTransfer
				probed.HDRFormat = detectHDR(s.ColorTransfer, s.ColorPrimaries, s.SideDataList)
			}
		case "audio":
			probed.AudioCodecs = append(probed.AudioCodecs, s.CodecName)
			probed.AudioChannels = append(probed.AudioChannels, s.Channels)
		case "subtitle":
			probed.SubtitleTracks++
			probed.SubtitleCodecs = append(probed.SubtitleCodecs, s.CodecName)
		}
	}
	probed.Issues = flagIssues(probed)
	return probed
}

// detectHDR maps ffprobe color/transfer/side-data into a friendly HDR tag.
// Dolby Vision side-data lives in "DOVI configuration record"; HDR10 / HDR10+
// live in "Mastering display metadata"/"Content light level metadata".
func detectHDR(transfer, primaries string, sideData []struct {
	SideDataType string `json:"side_data_type"`
}) string {
	if strings.EqualFold(transfer, "smpte2084") {
		for _, sd := range sideData {
			if strings.Contains(sd.SideDataType, "Dolby Vision") || strings.Contains(strings.ToLower(sd.SideDataType), "dovi") {
				return "Dolby Vision (PQ)"
			}
		}
		return "HDR10"
	}
	if strings.EqualFold(transfer, "arib-std-b67") {
		return "HLG"
	}
	for _, sd := range sideData {
		if strings.Contains(sd.SideDataType, "Dolby Vision") {
			return "Dolby Vision"
		}
	}
	_ = primaries
	return ""
}

func flagIssues(a FileAudit) []string {
	var issues []string
	if a.VideoCodec == "" && a.DurationSec > 0 {
		issues = append(issues, "no video stream")
	}
	if len(a.AudioCodecs) == 0 && a.DurationSec > 0 {
		issues = append(issues, "no audio streams")
	}
	if a.VariableFPS {
		issues = append(issues, "variable frame rate (may cause sync drift on direct play)")
	}
	if a.HDRFormat != "" {
		issues = append(issues, "HDR ("+a.HDRFormat+") — verify tone-mapping on SDR clients")
	}
	// Transcode-prone codecs: clients that lean on web playback often can't
	// direct-play h.265/HEVC, vp9, or av1 without ffmpeg transcoding.
	switch strings.ToLower(a.VideoCodec) {
	case "hevc", "h265", "vp9", "av1":
		issues = append(issues, "codec "+a.VideoCodec+" likely needs transcoding for web clients")
	}
	if a.BitrateBPS > 50_000_000 {
		issues = append(issues, fmt.Sprintf("high bitrate (%.1f Mbps) — may stress slow networks", float64(a.BitrateBPS)/1_000_000))
	}
	return issues
}

func anyIssues(rs []FileAudit) bool {
	for _, r := range rs {
		if r.ProbeError != "" || len(r.Issues) > 0 {
			return true
		}
	}
	return false
}

// parseFraction handles ffprobe's "30000/1001" style. Returns 0 on garbage.
func parseFraction(s string) float64 {
	if s == "" || s == "0/0" {
		return 0
	}
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}
	num, _ := strconv.ParseFloat(parts[0], 64)
	den, _ := strconv.ParseFloat(parts[1], 64)
	if den == 0 {
		return 0
	}
	return num / den
}

func writeJSON(path string, data any) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func writeSummary(path string, rs []FileAudit, total, probed, skipped, failed int, elapsed time.Duration) error {
	codecCounts := map[string]int{}
	hdrCounts := map[string]int{}
	containerCounts := map[string]int{}
	issueCounts := map[string]int{}
	var flagged []FileAudit

	for _, r := range rs {
		if r.VideoCodec != "" {
			codecCounts[r.VideoCodec]++
		}
		if r.HDRFormat != "" {
			hdrCounts[r.HDRFormat]++
		}
		if r.Container != "" {
			containerCounts[r.Container]++
		}
		for _, i := range r.Issues {
			issueCounts[i]++
		}
		if len(r.Issues) > 0 || r.ProbeError != "" {
			flagged = append(flagged, r)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Xuva media-audit summary\n")
	fmt.Fprintf(&b, "Generated: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "Total media files walked : %d\n", total)
	fmt.Fprintf(&b, "  newly probed           : %d\n", probed)
	fmt.Fprintf(&b, "  resumed (size+mtime)   : %d\n", skipped)
	fmt.Fprintf(&b, "  ffprobe failures       : %d\n", failed)
	fmt.Fprintf(&b, "  flagged with issues    : %d\n", len(flagged))
	fmt.Fprintf(&b, "  clean                  : %d\n", len(rs)-len(flagged))
	fmt.Fprintf(&b, "Elapsed                  : %s\n\n", elapsed.Round(time.Second))

	writeCountSection(&b, "Video codecs", codecCounts)
	writeCountSection(&b, "Containers", containerCounts)
	writeCountSection(&b, "HDR formats", hdrCounts)
	writeCountSection(&b, "Issues", issueCounts)

	if len(flagged) > 0 {
		fmt.Fprintf(&b, "Flagged files (%d)\n", len(flagged))
		b.WriteString(strings.Repeat("-", 60) + "\n")
		sort.Slice(flagged, func(i, j int) bool { return flagged[i].Path < flagged[j].Path })
		for _, f := range flagged {
			fmt.Fprintf(&b, "  %s\n", f.Path)
			if f.ProbeError != "" {
				fmt.Fprintf(&b, "      probe-error: %s\n", f.ProbeError)
			}
			for _, i := range f.Issues {
				fmt.Fprintf(&b, "      - %s\n", i)
			}
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeCountSection(b *strings.Builder, title string, counts map[string]int) {
	if len(counts) == 0 {
		return
	}
	type kv struct {
		k string
		v int
	}
	rows := make([]kv, 0, len(counts))
	for k, v := range counts {
		rows = append(rows, kv{k, v})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].v != rows[j].v {
			return rows[i].v > rows[j].v
		}
		return rows[i].k < rows[j].k
	})
	fmt.Fprintf(b, "%s\n%s\n", title, strings.Repeat("-", len(title)))
	for _, r := range rows {
		fmt.Fprintf(b, "  %-40s %d\n", r.k, r.v)
	}
	fmt.Fprintln(b)
}
