package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jampat000/Xuva/server/internal/buildinfo"
)

const defaultUpdateReleaseAPI = "https://api.github.com/repos/jampat000/Xuva/releases/latest"

type updateAsset struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Size        int64  `json:"size"`
	Digest      string `json:"digest,omitempty"`
	PackageType string `json:"packageType"`
}

type updateStatus struct {
	CurrentVersion         string        `json:"currentVersion"`
	LatestVersion          string        `json:"latestVersion,omitempty"`
	UpdateAvailable        bool          `json:"updateAvailable"`
	ReleaseURL             string        `json:"releaseUrl,omitempty"`
	PublishedAt            string        `json:"publishedAt,omitempty"`
	Assets                 []updateAsset `json:"assets"`
	DockerImage            string        `json:"dockerImage,omitempty"`
	InstallMode            string        `json:"installMode"`
	ApplySupported         bool          `json:"applySupported"`
	ApplyUnsupportedReason string        `json:"applyUnsupportedReason,omitempty"`
	CheckedAt              string        `json:"checkedAt"`
}

type githubRelease struct {
	TagName     string `json:"tag_name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
		Digest             string `json:"digest"`
	} `json:"assets"`
}

func updatesCheckHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := checkForUpdates(r.Context())
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, status)
	}
}

func checkForUpdates(ctx context.Context) (updateStatus, error) {
	current := buildinfo.Current().Version
	checkedAt := time.Now().UTC().Format(time.RFC3339)
	status := updateStatus{
		CurrentVersion:         current,
		Assets:                 []updateAsset{},
		InstallMode:            "download-and-run-installer",
		ApplySupported:         false,
		ApplyUnsupportedReason: "Automatic self-apply requires a separate supervised updater process so Xuva can stop the running server, verify the installer, apply it, and restart safely.",
		CheckedAt:              checkedAt,
	}

	apiURL := strings.TrimSpace(os.Getenv("XUVA_UPDATE_RELEASE_API"))
	if apiURL == "" {
		apiURL = defaultUpdateReleaseAPI
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return status, fmt.Errorf("build update request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Xuva/"+current)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return status, fmt.Errorf("update check failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return status, fmt.Errorf("update check failed with HTTP %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return status, fmt.Errorf("decode update response: %w", err)
	}
	status.LatestVersion = strings.TrimSpace(release.TagName)
	status.ReleaseURL = strings.TrimSpace(release.HTMLURL)
	status.PublishedAt = strings.TrimSpace(release.PublishedAt)
	status.UpdateAvailable = compareReleaseVersions(status.LatestVersion, current) > 0
	status.DockerImage = "ghcr.io/jampat000/xuva:" + status.LatestVersion
	for _, asset := range release.Assets {
		name := strings.TrimSpace(asset.Name)
		url := strings.TrimSpace(asset.BrowserDownloadURL)
		if name == "" || url == "" {
			continue
		}
		status.Assets = append(status.Assets, updateAsset{
			Name:        name,
			URL:         url,
			Size:        asset.Size,
			Digest:      strings.TrimSpace(asset.Digest),
			PackageType: classifyUpdateAsset(name),
		})
	}
	return status, nil
}

func classifyUpdateAsset(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.HasSuffix(lower, ".exe.sha256"):
		return "windows-installer-checksum"
	case strings.HasSuffix(lower, ".zip.sha256"):
		return "windows-portable-checksum"
	case strings.HasSuffix(lower, ".exe"):
		return "windows-installer"
	case strings.HasSuffix(lower, ".zip"):
		return "windows-portable"
	default:
		return "asset"
	}
}

var releaseVersionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)

func compareReleaseVersions(a, b string) int {
	av, aok := parseReleaseVersion(a)
	bv, bok := parseReleaseVersion(b)
	if !aok || !bok {
		return 0
	}
	for i := 0; i < len(av); i++ {
		if av[i] > bv[i] {
			return 1
		}
		if av[i] < bv[i] {
			return -1
		}
	}
	return 0
}

func parseReleaseVersion(value string) ([3]int, bool) {
	var output [3]int
	matches := releaseVersionPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(matches) != 4 {
		return output, false
	}
	for i := 1; i < len(matches); i++ {
		n, err := strconv.Atoi(matches[i])
		if err != nil {
			return output, false
		}
		output[i-1] = n
	}
	return output, true
}
