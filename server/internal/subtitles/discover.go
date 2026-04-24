package subtitles

import (
	"os"
	"path/filepath"
	"strings"
)

func DiscoverSidecars(mediaPath string) []Sidecar {
	dir := filepath.Dir(mediaPath)
	base := strings.TrimSuffix(filepath.Base(mediaPath), filepath.Ext(mediaPath))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var output []Sidecar
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if !isSubtitleExt(ext) || !strings.HasPrefix(strings.TrimSuffix(name, filepath.Ext(name)), base) {
			continue
		}
		path := filepath.Join(dir, name)
		rel, _ := filepath.Rel(dir, path)
		output = append(output, Sidecar{Path: path, RelPath: rel, Format: strings.TrimPrefix(ext, "."), Language: inferLanguage(name, base), Forced: strings.Contains(strings.ToLower(name), ".forced"), HearingImpaired: strings.Contains(strings.ToLower(name), ".sdh") || strings.Contains(strings.ToLower(name), ".hi")})
	}
	return output
}

func isSubtitleExt(ext string) bool {
	switch ext {
	case ".srt", ".ass", ".ssa", ".vtt", ".sub":
		return true
	default:
		return false
	}
}

func inferLanguage(name string, base string) string {
	stem := strings.TrimSuffix(name, filepath.Ext(name))
	stem = strings.TrimPrefix(stem, base)
	stem = strings.Trim(stem, ".-_ ")
	stem = strings.ReplaceAll(stem, "forced", "")
	stem = strings.Trim(stem, ".-_ ")
	parts := strings.FieldsFunc(stem, func(r rune) bool { return r == '.' || r == '-' || r == '_' || r == ' ' })
	if len(parts) > 0 && len(parts[0]) >= 2 && len(parts[0]) <= 3 {
		return strings.ToLower(parts[0])
	}
	return ""
}
