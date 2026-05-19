// Package chapters provides intro/credits detection for media files.
// Credits are detected via FFmpeg silencedetect (always available).
// Intro detection uses Chromaprint via fpcalc (optional, disabled when path is empty).
package chapters

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/bits"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	// FingerprintLengthSecs is how many seconds of audio to fingerprint.
	FingerprintLengthSecs = 300.0

	// MaxIntroSecs caps the detected intro length.
	MaxIntroSecs = 120.0

	// MinIntroSecs is the minimum intro length to consider valid.
	MinIntroSecs = 10.0

	// IntroSearchSecs is how far into each episode to search for an intro.
	IntroSearchSecs = 300.0

	// berThreshold is the maximum bit error rate to consider two segments matching.
	berThreshold = 0.15

	// frameRate is approximately how many fpcalc uint32 values are emitted per second.
	frameRate = 2.0
)

// Segment is a [Start, End] time range in seconds.
type Segment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// FingerprintResult holds raw Chromaprint output for one file.
type FingerprintResult struct {
	Duration    float64  `json:"duration"`
	Fingerprint []uint32 `json:"fingerprint"`
}

// EpisodeFingerprint pairs a media-source ID with its fingerprint data.
type EpisodeFingerprint struct {
	MediaSourceID string
	Data          *FingerprintResult
}

// Analyzer wraps FFmpeg and fpcalc subprocess calls.
type Analyzer struct {
	FFmpegPath string
	FpcalcPath string // empty = intro detection disabled
}

// New returns an Analyzer. FpcalcPath may be empty to disable intro detection.
func New(ffmpegPath, fpcalcPath string) *Analyzer {
	return &Analyzer{FFmpegPath: ffmpegPath, FpcalcPath: fpcalcPath}
}

// IntroEnabled reports whether fpcalc is configured.
func (a *Analyzer) IntroEnabled() bool {
	return strings.TrimSpace(a.FpcalcPath) != ""
}

// Fingerprint extracts a Chromaprint fingerprint for the first maxSecs of the file.
func (a *Analyzer) Fingerprint(ctx context.Context, path string, maxSecs float64) (*FingerprintResult, error) {
	if !a.IntroEnabled() {
		return nil, fmt.Errorf("fpcalc not configured")
	}
	args := []string{
		"-json", "-raw",
		"-length", strconv.FormatFloat(maxSecs, 'f', 0, 64),
		path,
	}
	cmd := exec.CommandContext(ctx, a.FpcalcPath, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("fpcalc: %w", err)
	}
	var raw struct {
		Duration    float64 `json:"duration"`
		Fingerprint []int   `json:"fingerprint"` // fpcalc -raw emits signed ints
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse fpcalc output: %w", err)
	}
	fp := make([]uint32, len(raw.Fingerprint))
	for i, v := range raw.Fingerprint {
		fp[i] = uint32(v)
	}
	return &FingerprintResult{Duration: raw.Duration, Fingerprint: fp}, nil
}

// DetectCredits uses FFmpeg silencedetect to find where credits begin.
// It looks for the last silence gap after 60% of the file's duration.
// Returns nil when no silence is found or the file is too short.
func (a *Analyzer) DetectCredits(ctx context.Context, path string, durationSecs float64) (*Segment, error) {
	if durationSecs < 60 {
		return nil, nil
	}
	startOffset := durationSecs - 600
	if startOffset < 0 {
		startOffset = 0
	}
	args := []string{
		"-ss", strconv.FormatFloat(startOffset, 'f', 2, 64),
		"-i", path,
		"-vn",
		"-af", "silencedetect=noise=-50dB:duration=2",
		"-f", "null",
		"-",
	}
	cmd := exec.CommandContext(ctx, a.FFmpegPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	_ = cmd.Run() // -f null always exits non-zero
	return parseLastSilence(stderr.String(), startOffset, durationSecs), nil
}

var silenceStartRe = regexp.MustCompile(`silence_start: ([\d.]+)`)

func parseLastSilence(output string, timeOffset, totalDuration float64) *Segment {
	scanner := bufio.NewScanner(strings.NewReader(output))
	minStart := totalDuration * 0.6
	var lastStart float64 = -1
	for scanner.Scan() {
		if m := silenceStartRe.FindStringSubmatch(scanner.Text()); m != nil {
			t, _ := strconv.ParseFloat(m[1], 64)
			abs := t + timeOffset
			if abs > minStart {
				lastStart = abs
			}
		}
	}
	if lastStart < 0 {
		return nil
	}
	return &Segment{Start: lastStart, End: totalDuration}
}

// FindIntros compares fingerprints across episodes and returns a map from
// mediaSourceID to its detected intro segment. Requires at least 2 episodes.
func FindIntros(episodes []EpisodeFingerprint) map[string]Segment {
	if len(episodes) < 2 {
		return nil
	}
	type vote struct{ start, end float64 }
	votes := make(map[string][]vote)

	for i := 0; i < len(episodes); i++ {
		for j := i + 1; j < len(episodes); j++ {
			a, b := episodes[i], episodes[j]
			segA, segB, ok := findBestMatch(a.Data.Fingerprint, b.Data.Fingerprint)
			if !ok {
				continue
			}
			votes[a.MediaSourceID] = append(votes[a.MediaSourceID], vote{segA.Start, segA.End})
			votes[b.MediaSourceID] = append(votes[b.MediaSourceID], vote{segB.Start, segB.End})
		}
	}

	result := make(map[string]Segment)
	for id, vs := range votes {
		var sumS, sumE float64
		for _, v := range vs {
			sumS += v.start
			sumE += v.end
		}
		n := float64(len(vs))
		seg := Segment{Start: sumS / n, End: sumE / n}
		if seg.End-seg.Start >= MinIntroSecs {
			result[id] = seg
		}
	}
	return result
}

// findBestMatch finds the best-matching window between two fingerprints using
// a coarse sliding-window search limited to the first IntroSearchSecs of each.
func findBestMatch(a, b []uint32) (segA, segB Segment, ok bool) {
	maxA := int(IntroSearchSecs * frameRate)
	if maxA > len(a) {
		maxA = len(a)
	}
	maxB := int(IntroSearchSecs * frameRate)
	if maxB > len(b) {
		maxB = len(b)
	}

	win := int(MaxIntroSecs * frameRate)
	if win > maxA || win > maxB {
		win = min(maxA, maxB) / 2
	}
	if win < int(MinIntroSecs*frameRate) {
		return Segment{}, Segment{}, false
	}

	const step = 4 // ~2 second increments
	bestBER := 1.0
	var bestOA, bestOB int

	for oA := 0; oA+win <= maxA; oA += step {
		for oB := 0; oB+win <= maxB; oB += step {
			ber := hammingBER(a[oA:oA+win], b[oB:oB+win])
			if ber < bestBER {
				bestBER = ber
				bestOA, bestOB = oA, oB
			}
		}
	}
	if bestBER > berThreshold {
		return Segment{}, Segment{}, false
	}
	return Segment{
		Start: float64(bestOA) / frameRate,
		End:   float64(bestOA+win) / frameRate,
	}, Segment{
		Start: float64(bestOB) / frameRate,
		End:   float64(bestOB+win) / frameRate,
	}, true
}

func hammingBER(a, b []uint32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 1.0
	}
	var diff int
	for i := range a {
		diff += bits.OnesCount32(a[i] ^ b[i])
	}
	return float64(diff) / float64(len(a)*32)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
