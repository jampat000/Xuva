package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jampat000/Xuva/server/internal/buildinfo"
	"github.com/jampat000/Xuva/server/internal/config"
)

const defaultUpdateReleaseAPI = "https://api.github.com/repos/jampat000/Xuva/releases/latest"
const maxInstallerDownloadBytes = 700 * 1024 * 1024

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

type updateApplyResponse struct {
	Status          string `json:"status"`
	Version         string `json:"version"`
	InstallerPath   string `json:"installerPath"`
	InstallerSHA256 string `json:"installerSha256"`
	PendingPath     string `json:"pendingPath"`
	RequiresRestart bool   `json:"requiresRestart"`
	Message         string `json:"message"`
}

type stagedUpdateRequest struct {
	Version         string `json:"version"`
	InstallerPath   string `json:"installerPath"`
	InstallerSHA256 string `json:"installerSha256"`
	InstallerName   string `json:"installerName"`
	ReleaseURL      string `json:"releaseUrl,omitempty"`
	RequestedAt     string `json:"requestedAt"`
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
		status, err := checkForUpdates(r.Context(), deps.Config)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, status)
	}
}

func updatesApplyHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := stageUpdateApply(r.Context(), deps.Config)
		if err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, result)
	}
}

func checkForUpdates(ctx context.Context, cfg config.Config) (updateStatus, error) {
	current := buildinfo.Current().Version
	checkedAt := time.Now().UTC().Format(time.RFC3339)
	status := updateStatus{
		CurrentVersion:         current,
		Assets:                 []updateAsset{},
		InstallMode:            "download-and-run-installer",
		ApplySupported:         canApplyUpdates(cfg),
		ApplyUnsupportedReason: updateApplyUnsupportedReason(cfg),
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

func stageUpdateApply(ctx context.Context, cfg config.Config) (updateApplyResponse, error) {
	status, err := checkForUpdates(ctx, cfg)
	if err != nil {
		return updateApplyResponse{}, err
	}
	if !status.UpdateAvailable {
		return updateApplyResponse{}, fmt.Errorf("no newer Xuva release is available")
	}
	if !status.ApplySupported {
		return updateApplyResponse{}, errors.New(status.ApplyUnsupportedReason)
	}
	installer, ok := findUpdateAsset(status.Assets, "windows-installer")
	if !ok {
		return updateApplyResponse{}, fmt.Errorf("latest release does not include a Windows installer")
	}
	checksum, _ := findUpdateAsset(status.Assets, "windows-installer-checksum")

	root := updateRoot(cfg)
	updateDir := filepath.Join(root, "updates", safePathName(status.LatestVersion))
	if err := os.MkdirAll(updateDir, 0o755); err != nil {
		return updateApplyResponse{}, fmt.Errorf("create update directory: %w", err)
	}
	installerPath := filepath.Join(updateDir, installer.Name)
	if err := downloadURLToFile(ctx, installer.URL, installerPath); err != nil {
		return updateApplyResponse{}, err
	}
	expectedSHA := strings.TrimPrefix(strings.TrimSpace(installer.Digest), "sha256:")
	if expectedSHA == "" && checksum.URL != "" {
		checksumPath := filepath.Join(updateDir, checksum.Name)
		if err := downloadURLToFile(ctx, checksum.URL, checksumPath); err != nil {
			return updateApplyResponse{}, err
		}
		expectedSHA = parseSHA256File(checksumPath)
	}
	// An installer we cannot verify must never be staged. If neither the GitHub
	// asset digest nor a published .exe.sha256 sidecar gave us an expected hash,
	// fail closed rather than handing an unverified binary to the elevated
	// launcher.
	if expectedSHA == "" {
		_ = os.Remove(installerPath)
		return updateApplyResponse{}, fmt.Errorf("release is missing a checksum for the installer; refusing to stage an unverified update")
	}
	actualSHA, err := sha256File(installerPath)
	if err != nil {
		return updateApplyResponse{}, err
	}
	if !strings.EqualFold(expectedSHA, actualSHA) {
		_ = os.Remove(installerPath)
		return updateApplyResponse{}, fmt.Errorf("installer checksum mismatch")
	}

	request := stagedUpdateRequest{
		Version:         status.LatestVersion,
		InstallerPath:   installerPath,
		InstallerSHA256: actualSHA,
		InstallerName:   installer.Name,
		ReleaseURL:      status.ReleaseURL,
		RequestedAt:     time.Now().UTC().Format(time.RFC3339),
	}
	pendingPath := filepath.Join(root, "updates", "pending-update.json")
	if err := writeJSONFileAtomic(pendingPath, request); err != nil {
		return updateApplyResponse{}, fmt.Errorf("stage update request: %w", err)
	}
	return updateApplyResponse{
		Status:          "staged",
		Version:         request.Version,
		InstallerPath:   request.InstallerPath,
		InstallerSHA256: request.InstallerSHA256,
		PendingPath:     pendingPath,
		RequiresRestart: true,
		Message:         "Update staged. The Windows launcher will stop Xuva, run the installer, and restart the web app.",
	}, nil
}

func canApplyUpdates(cfg config.Config) bool {
	if os.Getenv("XUVA_UPDATE_APPLY_SUPPORTED") == "1" {
		return strings.TrimSpace(updateRoot(cfg)) != ""
	}
	return runtime.GOOS == "windows" && strings.TrimSpace(cfg.RuntimeHome) != ""
}

func updateApplyUnsupportedReason(cfg config.Config) string {
	if canApplyUpdates(cfg) {
		return ""
	}
	if runtime.GOOS != "windows" && os.Getenv("XUVA_UPDATE_APPLY_SUPPORTED") != "1" {
		return "Automatic apply is only supported by the Windows tray launcher. Docker and Linux installs should upgrade by replacing the image or package."
	}
	return "Automatic apply requires a writable Xuva runtime home managed by the Windows launcher."
}

func updateRoot(cfg config.Config) string {
	if root := strings.TrimSpace(cfg.RuntimeHome); root != "" {
		return root
	}
	if dataDir := strings.TrimSpace(cfg.DataDir); dataDir != "" {
		return filepath.Dir(dataDir)
	}
	return ""
}

func findUpdateAsset(assets []updateAsset, packageType string) (updateAsset, bool) {
	for _, asset := range assets {
		if asset.PackageType == packageType {
			return asset, true
		}
	}
	return updateAsset{}, false
}

func safePathName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, string(filepath.Separator), "-")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, "\\", "-")
	if value == "" {
		return "unknown"
	}
	return value
}

func downloadURLToFile(ctx context.Context, rawURL string, destination string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("build download request: %w", err)
	}
	req.Header.Set("User-Agent", "Xuva/"+buildinfo.Current().Version)
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download update asset: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download update asset failed with HTTP %d", resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("create update asset directory: %w", err)
	}
	tmp := destination + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create update asset: %w", err)
	}
	_, copyErr := io.Copy(file, io.LimitReader(resp.Body, maxInstallerDownloadBytes+1))
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("write update asset: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close update asset: %w", closeErr)
	}
	if info, err := os.Stat(tmp); err == nil && info.Size() > maxInstallerDownloadBytes {
		_ = os.Remove(tmp)
		return fmt.Errorf("update asset exceeds maximum allowed size")
	}
	if err := os.Rename(tmp, destination); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("move update asset into place: %w", err)
	}
	return nil
}

func parseSHA256File(path string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return ""
	}
	value := strings.TrimPrefix(strings.TrimSpace(fields[0]), "sha256:")
	if len(value) != sha256.Size*2 {
		return ""
	}
	return value
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	sum := sha256.New()
	if _, err := io.Copy(sum, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}

func writeJSONFileAtomic(path string, payload any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encodeErr := encoder.Encode(payload)
	closeErr := file.Close()
	if encodeErr != nil {
		_ = os.Remove(tmp)
		return encodeErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func classifyUpdateAsset(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.HasSuffix(lower, ".msi.sha256"):
		return "windows-msi-checksum"
	case strings.HasSuffix(lower, ".exe.sha256"):
		return "windows-installer-checksum"
	case strings.HasSuffix(lower, ".zip.sha256"):
		return "windows-portable-checksum"
	case strings.HasSuffix(lower, ".msi"):
		return "windows-msi"
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
