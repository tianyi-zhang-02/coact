package versionmgr

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseManifestAndSelectAsset(t *testing.T) {
	data := []byte(`{
		"version": "v0.2.0",
		"channel": "beta",
		"stability": "recommended",
		"summary": "local UI release",
		"supports": {"cli": true, "ui": true, "agents": ["claude", "codex"]},
		"assets": [
			{"name": "coact_v0.2.0_darwin_arm64.tar.gz", "os": "darwin", "arch": "arm64"},
			{"name": "coact_v0.2.0_linux_amd64.tar.gz", "os": "linux", "arch": "amd64"}
		],
		"checksums": {
			"coact_v0.2.0_darwin_arm64.tar.gz": "abc123"
		}
	}`)
	manifest, err := ParseManifest(data)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Version != "v0.2.0" || manifest.Channel != "beta" {
		t.Fatalf("unexpected manifest: %#v", manifest)
	}
	asset, ok := SelectAsset(manifest, "darwin", "arm64")
	if !ok {
		t.Fatal("expected darwin/arm64 asset")
	}
	if asset.SHA256 != "abc123" {
		t.Fatalf("checksum was not populated from manifest map: %#v", asset)
	}
}

func TestVerifySHA256RejectsMismatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "coact")
	if err := os.WriteFile(path, []byte("actual"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := VerifySHA256(path, strings.Repeat("0", 64))
	if err == nil {
		t.Fatal("expected checksum mismatch")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseChecksumFile(t *testing.T) {
	checksums := parseChecksumFile([]byte("abc123  coact_v0.2.0_darwin_arm64.tar.gz\nbad line\nfff999  ./dist/coact_v0.2.0_linux_amd64.tar.gz\n"))
	if checksums["coact_v0.2.0_darwin_arm64.tar.gz"] != "abc123" {
		t.Fatalf("missing darwin checksum: %#v", checksums)
	}
	if checksums["coact_v0.2.0_linux_amd64.tar.gz"] != "fff999" {
		t.Fatalf("missing linux checksum: %#v", checksums)
	}
}

func TestAttachAssetURLsFallsBackByPlatform(t *testing.T) {
	manifest := &Manifest{
		Assets: []Asset{{Name: "expected-but-not-actual.tar.gz", OS: "darwin", Arch: "arm64"}},
		Checksums: map[string]string{
			"coact_0.2.0_darwin_arm64.tar.gz": "abc123",
		},
	}
	attachAssetURLs(manifest, map[string]string{
		"coact_manifest.json":             "https://example.invalid/manifest",
		"checksums.txt":                   "https://example.invalid/checksums",
		"coact_0.2.0_darwin_arm64.tar.gz": "https://example.invalid/darwin-arm64",
		"coact_0.2.0_windows_amd64.zip":   "https://example.invalid/windows-amd64",
		"coact_0.2.0_linux_amd64.tar.gz":  "https://example.invalid/linux-amd64",
		"coact_0.2.0_darwin_amd64.tar.gz": "https://example.invalid/darwin-amd64",
		"coact_0.2.0_windows_arm64.zip":   "https://example.invalid/windows-arm64",
		"coact_0.2.0_linux_arm64.tar.gz":  "https://example.invalid/linux-arm64",
	})
	asset := manifest.Assets[0]
	if asset.Name != "coact_0.2.0_darwin_arm64.tar.gz" || asset.URL == "" || asset.SHA256 != "abc123" {
		t.Fatalf("fallback did not populate asset fields: %#v", asset)
	}
}

func TestSwitchOnlyChangesManagedShim(t *testing.T) {
	home := t.TempDir()
	version := "v1.0.0"
	target := filepath.Join(BinDir(home), BinaryName(version))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("first"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Switch(home, version); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("installed binary should remain: %v", err)
	}

	link := LinkPath(home)
	if runtime.GOOS == "windows" {
		data, err := os.ReadFile(link)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "first" {
			t.Fatalf("shim copy mismatch: %q", data)
		}
		return
	}

	got, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.Abs(target)
	if got != want {
		t.Fatalf("shim target mismatch: got %q want %q", got, want)
	}
}
