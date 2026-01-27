package update

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var ErrAlreadyLatest = errors.New("already latest version")

type releaseInfo struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func SelfUpdate(ctx context.Context, repo string, currentVersion string) (string, error) {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		return "", errors.New("repo is required")
	}

	rel, err := fetchLatestRelease(ctx, repo)
	if err != nil {
		return "", err
	}

	tag := strings.TrimSpace(rel.TagName)
	if tag == "" {
		return "", errors.New("latest release has no tag")
	}

	if isSameVersion(tag, currentVersion) {
		return tag, ErrAlreadyLatest
	}

	assetName := buildAssetName()
	assetURL, checksumURL := findAsset(rel.Assets, assetName)
	if assetURL == "" {
		return "", fmt.Errorf("asset not found for %s", assetName)
	}

	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}

	targetDir := filepath.Dir(exePath)
	tmpFile, err := os.CreateTemp(targetDir, "vcontext-update-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if err := downloadToFile(ctx, assetURL, tmpFile, checksumURL); err != nil {
		_ = tmpFile.Close()
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("close temp file: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpPath, 0o755); err != nil {
			return "", fmt.Errorf("chmod temp file: %w", err)
		}
	}

	if err := replaceBinary(tmpPath, exePath); err != nil {
		return "", err
	}

	return tag, nil
}

func fetchLatestRelease(ctx context.Context, repo string) (*releaseInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "vcontext")

	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if token := strings.TrimSpace(os.Getenv("GH_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("fetch latest release: %s", strings.TrimSpace(string(body)))
	}

	var rel releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}

	return &rel, nil
}

func buildAssetName() string {
	name := fmt.Sprintf("vcontext_%s_%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

func findAsset(assets []releaseAsset, name string) (string, string) {
	var assetURL string
	var checksumURL string
	checksumName := name + ".sha256"
	for _, asset := range assets {
		switch asset.Name {
		case name:
			assetURL = asset.URL
		case checksumName:
			checksumURL = asset.URL
		}
	}
	return assetURL, checksumURL
}

func downloadToFile(ctx context.Context, url string, dst *os.File, checksumURL string) error {
	if err := dst.Truncate(0); err != nil {
		return fmt.Errorf("truncate temp file: %w", err)
	}
	if _, err := dst.Seek(0, 0); err != nil {
		return fmt.Errorf("seek temp file: %w", err)
	}

	expectedHash := ""
	if checksumURL != "" {
		hash, err := downloadChecksum(ctx, checksumURL)
		if err != nil {
			return err
		}
		expectedHash = hash
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "vcontext")
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download asset: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("download asset: %s", strings.TrimSpace(string(body)))
	}

	hasher := sha256.New()
	writer := io.MultiWriter(dst, hasher)
	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("download asset: %w", err)
	}

	if expectedHash != "" {
		actual := fmt.Sprintf("%x", hasher.Sum(nil))
		if !strings.EqualFold(actual, expectedHash) {
			return fmt.Errorf("checksum mismatch: expected %s got %s", expectedHash, actual)
		}
	}

	return nil
}

func downloadChecksum(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "vcontext")
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download checksum: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("download checksum: %s", strings.TrimSpace(string(body)))
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return "", fmt.Errorf("read checksum: %w", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "", errors.New("checksum file empty")
	}

	return strings.ToLower(fields[0]), nil
}

func replaceBinary(tmpPath string, exePath string) error {
	if runtime.GOOS == "windows" {
		backup := exePath + ".old"
		_ = os.Remove(backup)
		if err := os.Rename(exePath, backup); err != nil {
			return fmt.Errorf("replace binary: %w", err)
		}
		if err := os.Rename(tmpPath, exePath); err != nil {
			_ = os.Rename(backup, exePath)
			return fmt.Errorf("replace binary: %w", err)
		}
		_ = os.Remove(backup)
		return nil
	}

	if err := os.Rename(tmpPath, exePath); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}

	return nil
}

func isSameVersion(latest string, current string) bool {
	current = strings.TrimSpace(current)
	if current == "" || current == "dev" {
		return false
	}
	return normalizeVersion(latest) == normalizeVersion(current)
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	return version
}
