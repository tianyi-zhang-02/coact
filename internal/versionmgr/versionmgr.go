// Package versionmgr manages coact binaries installed under ~/.coact.
package versionmgr

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Manifest describes one published coact release.
type Manifest struct {
	Version     string            `json:"version"`
	Channel     string            `json:"channel"`
	Stability   string            `json:"stability"`
	Recommended bool              `json:"recommended"`
	Summary     string            `json:"summary"`
	Supports    Supports          `json:"supports"`
	Notes       []string          `json:"notes"`
	Assets      []Asset           `json:"assets"`
	Checksums   map[string]string `json:"checksums"`
}

// Supports captures feature compatibility for the UI versions panel.
type Supports struct {
	CLI       bool     `json:"cli"`
	UI        bool     `json:"ui"`
	Realtime  string   `json:"realtime,omitempty"`
	Autopilot string   `json:"autopilot,omitempty"`
	Agents    []string `json:"agents,omitempty"`
	Platforms []string `json:"platforms,omitempty"`
}

// Asset is a downloadable release asset for one OS/architecture.
type Asset struct {
	Name   string `json:"name"`
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// LocalInfo is one locally managed binary.
type LocalInfo struct {
	Version string `json:"version"`
	Path    string `json:"path"`
	Active  bool   `json:"active,omitempty"`
}

// UpdateOptions controls GitHub release installation.
type UpdateOptions struct {
	Repo    string
	Channel string
	Home    string
	GOOS    string
	GOARCH  string
	Client  *http.Client
}

const bundledManifestJSON = `{
  "version": "dev",
  "channel": "beta",
  "stability": "experimental",
  "recommended": false,
  "summary": "Initial release with local control center, shared brief, UI task/message controls, and managed versions.",
  "supports": {
    "cli": true,
    "ui": true,
    "realtime": "experimental",
    "autopilot": "not included",
    "agents": ["claude", "codex", "gemini"],
    "platforms": ["darwin/amd64", "darwin/arm64", "linux/amd64", "linux/arm64", "windows/amd64", "windows/arm64"]
  },
  "notes": [
    "coact with no arguments opens the local-only control center.",
    "Existing CLI subcommands remain available through coact help.",
    "coact update installs into ~/.coact and never overwrites system binaries.",
    "Real-time push remains experimental; the default UI uses polling.",
    "The control center is local-only: it binds 127.0.0.1, checks the Host header, and requires a per-run token.",
    "coact update verifies SHA-256 checksums over HTTPS; releases are not yet cryptographically signed."
  ],
  "assets": [],
  "checksums": {}
}`

// BundledManifest returns the version metadata embedded into this binary for
// display in the local UI.
func BundledManifest() *Manifest {
	manifest, err := ParseManifest([]byte(bundledManifestJSON))
	if err != nil {
		return nil
	}
	return manifest
}

// ParseManifest parses and validates a release manifest.
func ParseManifest(data []byte) (*Manifest, error) {
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	manifest.Version = strings.TrimSpace(manifest.Version)
	if manifest.Version == "" {
		return nil, errors.New("manifest missing version")
	}
	if manifest.Channel == "" {
		manifest.Channel = "stable"
	}
	if manifest.Stability == "" {
		manifest.Stability = manifest.Channel
	}
	if manifest.Checksums == nil {
		manifest.Checksums = map[string]string{}
	}
	return &manifest, nil
}

// SelectAsset chooses the asset matching the target OS/arch.
func SelectAsset(manifest *Manifest, goos, goarch string) (Asset, bool) {
	if manifest == nil {
		return Asset{}, false
	}
	for _, asset := range manifest.Assets {
		if asset.OS == goos && asset.Arch == goarch {
			if asset.SHA256 == "" {
				asset.SHA256 = manifest.Checksums[asset.Name]
			}
			return asset, true
		}
	}
	for _, asset := range manifest.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, goos) && strings.Contains(name, goarch) {
			if asset.SHA256 == "" {
				asset.SHA256 = manifest.Checksums[asset.Name]
			}
			return asset, true
		}
	}
	return Asset{}, false
}

// VerifySHA256 verifies a file against a lowercase hex SHA-256 checksum.
func VerifySHA256(path, expected string) error {
	expected = strings.TrimSpace(strings.TrimPrefix(expected, "sha256:"))
	if expected == "" {
		return errors.New("missing sha256 checksum")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sum := sha256.New()
	if _, err := io.Copy(sum, f); err != nil {
		return err
	}
	got := hex.EncodeToString(sum.Sum(nil))
	if !strings.EqualFold(got, expected) {
		return fmt.Errorf("checksum mismatch: got %s, want %s", got, expected)
	}
	return nil
}

// DefaultHome returns the managed coact home.
func DefaultHome() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".coact")
	}
	return ".coact"
}

// BinDir returns the directory containing managed binaries.
func BinDir(home string) string {
	return filepath.Join(home, "bin")
}

// LinkPath returns the managed shim path.
func LinkPath(home string) string {
	name := "coact"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(home, name)
}

// BinaryName returns the managed binary filename for version.
func BinaryName(version string) string {
	name := "coact-" + version
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

// LocalVersions lists binaries already installed in the managed layout.
func LocalVersions(home string) []LocalInfo {
	entries, err := os.ReadDir(BinDir(home))
	if err != nil {
		return nil
	}
	active := activeTarget(home)
	var out []LocalInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if runtime.GOOS == "windows" {
			name = strings.TrimSuffix(name, ".exe")
		}
		if !strings.HasPrefix(name, "coact-") {
			continue
		}
		path := filepath.Join(BinDir(home), entry.Name())
		abs, _ := filepath.Abs(path)
		out = append(out, LocalInfo{
			Version: strings.TrimPrefix(name, "coact-"),
			Path:    path,
			Active:  active != "" && abs == active,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Version < out[j].Version
	})
	return out
}

// Switch changes only the managed shim to point at an already installed version.
func Switch(home, version string) error {
	if err := validateVersion(version); err != nil {
		return err
	}
	target := filepath.Join(BinDir(home), BinaryName(version))
	if _, err := os.Stat(target); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("managed version %q is not installed", version)
		}
		return err
	}
	if err := os.MkdirAll(home, 0o755); err != nil {
		return err
	}

	link := LinkPath(home)
	if runtime.GOOS == "windows" {
		return copyFile(target, link, 0o755)
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	tmp := link + ".tmp"
	_ = os.Remove(tmp)
	if err := os.Symlink(absTarget, tmp); err != nil {
		return err
	}
	return os.Rename(tmp, link)
}

// InstallFromFile installs a raw coact binary into the managed layout.
func InstallFromFile(home, version, src string) (string, error) {
	if err := validateVersion(version); err != nil {
		return "", err
	}
	dest := filepath.Join(BinDir(home), BinaryName(version))
	if err := copyFile(src, dest, 0o755); err != nil {
		return "", err
	}
	return dest, nil
}

// CurrentBinaryManaged reports whether the running binary is under home.
func CurrentBinaryManaged(home string) (bool, string) {
	exe, err := os.Executable()
	if err != nil {
		return false, ""
	}
	exeAbs, _ := filepath.Abs(exe)
	homeAbs, _ := filepath.Abs(home)
	rel, err := filepath.Rel(homeAbs, exeAbs)
	if err != nil {
		return false, exeAbs
	}
	return rel != "." && rel != "" && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..", exeAbs
}

// Update downloads the newest matching GitHub release manifest and installs it
// into the managed layout. It does not overwrite system installations.
func Update(opts UpdateOptions) (*Manifest, string, error) {
	if opts.Repo == "" {
		return nil, "", errors.New("repo is required")
	}
	if opts.Home == "" {
		opts.Home = DefaultHome()
	}
	if opts.Channel == "" {
		opts.Channel = "stable"
	}
	if opts.GOOS == "" {
		opts.GOOS = runtime.GOOS
	}
	if opts.GOARCH == "" {
		opts.GOARCH = runtime.GOARCH
	}
	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	manifest, asset, err := findReleaseAsset(client, opts.Repo, opts.Channel, opts.GOOS, opts.GOARCH)
	if err != nil {
		return nil, "", err
	}
	if asset.URL == "" {
		return nil, "", fmt.Errorf("manifest asset %q has no download URL", asset.Name)
	}
	if asset.SHA256 == "" {
		return nil, "", fmt.Errorf("manifest asset %q has no checksum", asset.Name)
	}

	tmp, err := downloadToTemp(client, asset.URL, asset.Name)
	if err != nil {
		return nil, "", err
	}
	defer os.Remove(tmp)
	if err := VerifySHA256(tmp, asset.SHA256); err != nil {
		return nil, "", err
	}
	installed, err := installDownloadedAsset(opts.Home, manifest.Version, tmp, asset.Name)
	if err != nil {
		return nil, "", err
	}
	return manifest, installed, nil
}

type githubRelease struct {
	Draft      bool `json:"draft"`
	Prerelease bool `json:"prerelease"`
	Assets     []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func findReleaseAsset(client *http.Client, repo, channel, goos, goarch string) (*Manifest, Asset, error) {
	var releases []githubRelease
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=20", repo)
	if err := getJSON(client, url, &releases); err != nil {
		return nil, Asset{}, err
	}

	var lastErr error
	for _, release := range releases {
		if release.Draft {
			continue
		}
		assets := map[string]string{}
		for _, asset := range release.Assets {
			assets[asset.Name] = asset.URL
		}
		manifestURL := assets["coact_manifest.json"]
		if manifestURL == "" {
			continue
		}
		data, err := getBytes(client, manifestURL)
		if err != nil {
			lastErr = err
			continue
		}
		manifest, err := ParseManifest(data)
		if err != nil {
			lastErr = err
			continue
		}
		if !channelAllowed(channel, manifest.Channel) {
			continue
		}
		if checksumURL := assets["checksums.txt"]; len(manifest.Checksums) == 0 && checksumURL != "" {
			if data, err := getBytes(client, checksumURL); err == nil {
				manifest.Checksums = parseChecksumFile(data)
			}
		}
		attachAssetURLs(manifest, assets)
		asset, ok := SelectAsset(manifest, goos, goarch)
		if !ok {
			lastErr = fmt.Errorf("release %s has no asset for %s/%s", manifest.Version, goos, goarch)
			continue
		}
		return manifest, asset, nil
	}
	if lastErr != nil {
		return nil, Asset{}, lastErr
	}
	return nil, Asset{}, fmt.Errorf("no %s coact release manifest found for %s/%s", channel, goos, goarch)
}

func attachAssetURLs(manifest *Manifest, urls map[string]string) {
	for i := range manifest.Assets {
		if manifest.Assets[i].URL == "" {
			manifest.Assets[i].URL = urls[manifest.Assets[i].Name]
		}
		if manifest.Assets[i].URL == "" {
			if name, url := findAssetByPlatform(urls, manifest.Assets[i].OS, manifest.Assets[i].Arch); url != "" {
				manifest.Assets[i].Name = name
				manifest.Assets[i].URL = url
			}
		}
		if manifest.Assets[i].SHA256 == "" {
			manifest.Assets[i].SHA256 = manifest.Checksums[manifest.Assets[i].Name]
		}
	}
}

func findAssetByPlatform(urls map[string]string, goos, goarch string) (string, string) {
	var names []string
	for name := range urls {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		lower := strings.ToLower(name)
		if name == "coact_manifest.json" || name == "checksums.txt" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(goos)) && strings.Contains(lower, strings.ToLower(goarch)) {
			return name, urls[name]
		}
	}
	return "", ""
}

func channelAllowed(requested, got string) bool {
	rank := map[string]int{"stable": 0, "beta": 1, "alpha": 2}
	want, ok := rank[requested]
	if !ok {
		want = 0
	}
	have, ok := rank[got]
	if !ok {
		have = 0
	}
	return have <= want
}

func parseChecksumFile(data []byte) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := filepath.Base(fields[len(fields)-1])
		out[name] = fields[0]
	}
	return out
}

func getJSON(client *http.Client, url string, out any) error {
	data, err := getBytes(client, url)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func getBytes(client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "coact-version-manager")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func downloadToTemp(client *http.Client, url, name string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	tmp, err := os.CreateTemp("", "coact-"+filepath.Base(name)+"-*")
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		_ = os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}

func installDownloadedAsset(home, version, path, name string) (string, error) {
	if err := validateVersion(version); err != nil {
		return "", err
	}
	dest := filepath.Join(BinDir(home), BinaryName(version))
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		if err := installFromTarGz(path, dest); err != nil {
			return "", err
		}
	case strings.HasSuffix(lower, ".zip"):
		if err := installFromZip(path, dest); err != nil {
			return "", err
		}
	default:
		if err := copyFile(path, dest, 0o755); err != nil {
			return "", err
		}
	}
	return dest, nil
}

func installFromTarGz(path, dest string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if h.FileInfo().IsDir() || !isCoactBinaryName(filepath.Base(h.Name)) {
			continue
		}
		return writeBinary(dest, tr)
	}
	return errors.New("archive did not contain coact binary")
}

func installFromZip(path, dest string) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || !isCoactBinaryName(filepath.Base(f.Name)) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		err = writeBinary(dest, rc)
		rc.Close()
		return err
	}
	return errors.New("archive did not contain coact binary")
}

func isCoactBinaryName(name string) bool {
	return name == "coact" || name == "coact.exe"
}

func writeBinary(dest string, src io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	tmp := dest + ".tmp"
	_ = os.Remove(tmp)
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, src); err != nil {
		f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	_ = os.Chmod(tmp, 0o755)
	_ = os.Remove(dest)
	return os.Rename(tmp, dest)
}

func copyFile(src, dest string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	tmp := dest + ".tmp"
	_ = os.Remove(tmp)
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	_ = os.Chmod(tmp, mode)
	_ = os.Remove(dest)
	return os.Rename(tmp, dest)
}

func validateVersion(version string) error {
	if strings.TrimSpace(version) == "" {
		return errors.New("version is required")
	}
	if strings.ContainsAny(version, `/\`) || version == "." || version == ".." {
		return fmt.Errorf("invalid version %q", version)
	}
	return nil
}

func activeTarget(home string) string {
	link := LinkPath(home)
	target, err := os.Readlink(link)
	if err != nil {
		return ""
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(link), target)
	}
	abs, _ := filepath.Abs(target)
	return abs
}
