// Package update implements self-update from GitHub release archives.
package update

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
	"runtime"
	"strings"
	"time"
)

const (
	// Repository is the GitHub repository that publishes release archives.
	Repository = "Kandru/cs2-demo-analyzer"
	// BinaryName is the installed executable name.
	BinaryName = "cs2-demo-analyzer"
	// DefaultAPIBase is the unauthenticated GitHub API endpoint.
	DefaultAPIBase = "https://api.github.com"

	checksumsAsset = "SHA256SUMS"
	maxAssetSize   = 256 << 20
)

// Asset is a downloadable release asset.
type Asset struct {
	Name string
	URL  string
}

// Status describes the result of checking the latest GitHub release.
type Status struct {
	Current   string
	Latest    string
	Newer     bool
	Asset     Asset
	Checksums Asset
}

// Client checks and applies GitHub releases. Zero-value fields fall back to
// production defaults; tests can override APIBase, HTTP, GOOS and GOARCH.
type Client struct {
	HTTP    *http.Client
	APIBase string
	GOOS    string
	GOARCH  string
	Binary  string
}

// NewClient returns a client configured for the current platform.
func NewClient() *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 5 * time.Minute},
		APIBase: DefaultAPIBase,
		GOOS:    runtime.GOOS,
		GOARCH:  runtime.GOARCH,
		Binary:  BinaryName,
	}
}

type releaseResponse struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// Check fetches the latest release and selects the archive for this platform.
func (c *Client) Check(ctx context.Context, current string) (Status, error) {
	st := Status{Current: strings.TrimSpace(current)}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase()+"/repos/"+Repository+"/releases/latest", nil)
	if err != nil {
		return st, fmt.Errorf("build release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", c.binaryName()+"/"+st.Current)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return st, fmt.Errorf("query latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return st, fmt.Errorf("no GitHub releases found for %s", Repository)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return st, fmt.Errorf("query latest release: GitHub returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var release releaseResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&release); err != nil {
		return st, fmt.Errorf("decode latest release: %w", err)
	}
	latest, err := ParseVersion(release.TagName)
	if err != nil {
		return st, fmt.Errorf("latest release tag: %w", err)
	}
	st.Latest = latest.String()
	if cur, err := ParseVersion(st.Current); err == nil {
		st.Newer = Compare(latest, cur) > 0
	}

	wantArchive := fmt.Sprintf("%s-%s-%s-%s.tar.gz", c.binaryName(), st.Latest, c.targetOS(), c.targetArch())
	var available []string
	for _, asset := range release.Assets {
		switch asset.Name {
		case wantArchive:
			st.Asset = Asset{Name: asset.Name, URL: asset.URL}
		case checksumsAsset:
			st.Checksums = Asset{Name: asset.Name, URL: asset.URL}
		default:
			available = append(available, asset.Name)
		}
	}
	if st.Asset.URL == "" {
		return st, fmt.Errorf("release %s has no archive for %s/%s (expected %s); available assets: %s",
			st.Latest, c.targetOS(), c.targetArch(), wantArchive, strings.Join(available, ", "))
	}
	if st.Checksums.URL == "" {
		return st, fmt.Errorf("release %s has no %s asset", st.Latest, checksumsAsset)
	}
	return st, nil
}

// Apply downloads, verifies and atomically replaces exePath with the release
// selected by Check. The caller should pass the path returned by os.Executable.
func (c *Client) Apply(ctx context.Context, st Status, exePath string) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("self-update is not supported on Windows; download %s from https://github.com/%s/releases", st.Asset.Name, Repository)
	}
	if st.Asset.URL == "" || st.Checksums.URL == "" {
		return fmt.Errorf("update status is incomplete; run Check first")
	}

	sumsData, err := c.download(ctx, st.Checksums.URL, 4<<20, "checksums")
	if err != nil {
		return err
	}
	sums, err := ParseChecksums(sumsData)
	if err != nil {
		return err
	}
	want, ok := sums[st.Asset.Name]
	if !ok {
		return fmt.Errorf("%s does not list %s", checksumsAsset, st.Asset.Name)
	}

	archive, err := c.download(ctx, st.Asset.URL, maxAssetSize, "release archive")
	if err != nil {
		return err
	}
	sum := sha256.Sum256(archive)
	got := hex.EncodeToString(sum[:])
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("checksum mismatch for %s: got %s, want %s", st.Asset.Name, got, want)
	}

	binary, err := extractBinary(archive, c.binaryName())
	if err != nil {
		return err
	}
	target, err := resolveExecutable(exePath)
	if err != nil {
		return err
	}
	if err := replaceFile(target, binary); err != nil {
		return err
	}
	return nil
}

func (c *Client) download(ctx context.Context, url string, limit int64, what string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build %s request: %w", what, err)
	}
	req.Header.Set("User-Agent", c.binaryName())
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", what, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: server returned %s", what, resp.Status)
	}
	if resp.ContentLength > limit {
		return nil, fmt.Errorf("download %s: response is too large (%d bytes)", what, resp.ContentLength)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", what, err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("download %s: response is too large", what)
	}
	return data, nil
}

func (c *Client) httpClient() *http.Client {
	if c != nil && c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 5 * time.Minute}
}

func (c *Client) apiBase() string {
	if c != nil && c.APIBase != "" {
		return strings.TrimRight(c.APIBase, "/")
	}
	return DefaultAPIBase
}

func (c *Client) binaryName() string {
	if c != nil && c.Binary != "" {
		return c.Binary
	}
	return BinaryName
}

func (c *Client) targetOS() string {
	if c != nil && c.GOOS != "" {
		return c.GOOS
	}
	return runtime.GOOS
}

func (c *Client) targetArch() string {
	if c != nil && c.GOARCH != "" {
		return c.GOARCH
	}
	return runtime.GOARCH
}

func resolveExecutable(path string) (string, error) {
	if path == "" {
		var err error
		path, err = os.Executable()
		if err != nil {
			return "", fmt.Errorf("locate executable: %w", err)
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err == nil {
		return resolved, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return abs, nil
	}
	return "", fmt.Errorf("resolve executable symlinks: %w", err)
}

func replaceFile(target string, data []byte) error {
	dir := filepath.Dir(target)
	mode := os.FileMode(0o755)
	if info, err := os.Stat(target); err == nil {
		mode = info.Mode().Perm()
		if mode&0o111 == 0 {
			mode |= 0o111
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat executable: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(target)+".update-*")
	if err != nil {
		return fmt.Errorf("create update file in %s: %w (run with sudo or reinstall with install.sh)", dir, err)
	}
	tmpName := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}
	if err := tmp.Chmod(mode); err != nil {
		cleanup()
		return fmt.Errorf("set update file mode: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write update file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync update file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close update file: %w", err)
	}
	if err := os.Rename(tmpName, target); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("replace %s: %w (run with sudo or reinstall with install.sh)", target, err)
	}
	return nil
}
