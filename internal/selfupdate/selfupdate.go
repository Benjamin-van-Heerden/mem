// Package selfupdate replaces the running mem executable with the latest
// GitHub release when that release is newer.
package selfupdate

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// BaseURL is the release page of the mem repository; MEM_RELEASES_URL overrides it for tests.
const BaseURL = "https://github.com/Benjamin-van-Heerden/mem/releases"

// SkipEnv disables automatic updates when set to 1. mem sets it for the
// process it re-runs after updating.
const SkipEnv = "MEM_NO_UPDATE"

const (
	checkTimeout    = 5 * time.Second
	downloadTimeout = 2 * time.Minute
)

var semverTag = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

func baseURL() string {
	if url := os.Getenv("MEM_RELEASES_URL"); url != "" {
		return strings.TrimRight(url, "/")
	}
	return BaseURL
}

// Newer reports whether latest is a later release than current. Builds that are
// not releases (such as "dev") are never updated.
func Newer(current, latest string) bool {
	c, okc := parse(current)
	l, okl := parse(latest)
	if !okc || !okl {
		return false
	}
	for i := range c {
		if c[i] != l[i] {
			return l[i] > c[i]
		}
	}
	return false
}

// IsRelease reports whether version is a release build that can update itself.
func IsRelease(version string) bool { return semverTag.MatchString(version) }

func parse(v string) ([3]int, bool) {
	var parts [3]int
	m := semverTag.FindStringSubmatch(v)
	if m == nil {
		return parts, false
	}
	for i := range parts {
		parts[i], _ = strconv.Atoi(m[i+1])
	}
	return parts, true
}

// Latest returns the tag of the latest release, read from the redirect of
// <releases>/latest so no API token or rate limit is involved.
func Latest(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, baseURL()+"/latest", nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	location := resp.Header.Get("Location")
	tag := path.Base(location)
	if location == "" || !semverTag.MatchString(tag) {
		return "", fmt.Errorf("could not read the latest release from %s/latest (status %s)", baseURL(), resp.Status)
	}
	return tag, nil
}

// Asset is the release file name for this platform, as GoReleaser names it.
func Asset(tag string) string {
	name := fmt.Sprintf("mem_%s_%s_%s", tag, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

// Install downloads tag's binary for this platform, verifies it against the
// release checksums and replaces the executable at target.
func Install(ctx context.Context, tag, target string) error {
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()
	asset := Asset(tag)
	sums, err := fetch(ctx, fmt.Sprintf("%s/download/%s/checksums.txt", baseURL(), tag))
	if err != nil {
		return err
	}
	want, err := checksum(sums, asset)
	if err != nil {
		return err
	}
	binary, err := fetch(ctx, fmt.Sprintf("%s/download/%s/%s", baseURL(), tag, asset))
	if err != nil {
		return err
	}
	sum := sha256.Sum256(binary)
	if got := hex.EncodeToString(sum[:]); got != want {
		return fmt.Errorf("the downloaded %s does not match its checksum; nothing was replaced", asset)
	}
	return replace(target, binary)
}

func fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func checksum(sums []byte, asset string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(sums)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[1] == asset {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("the release has no checksum for %s", asset)
}

// replace writes the new binary next to target and renames it into place.
// Windows cannot overwrite a running executable, so the old one is moved aside first.
func replace(target string, binary []byte) error {
	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, ".mem-update-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(binary); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o755); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		old := target + ".old"
		os.Remove(old)
		if err := os.Rename(target, old); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return os.Rename(tmp.Name(), target)
}

// Executable is the resolved path of the running mem.
func Executable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}
