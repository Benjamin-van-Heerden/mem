package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// releases serves a latest redirect, one release asset and its checksums.
func releases(t *testing.T, tag string, binary []byte, sum string) {
	t.Helper()
	if sum == "" {
		digest := sha256.Sum256(binary)
		sum = hex.EncodeToString(digest[:])
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/tag/"+tag, http.StatusFound)
	})
	mux.HandleFunc("/download/"+tag+"/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s  %s\n", sum, Asset(tag))
	})
	mux.HandleFunc("/download/"+tag+"/"+Asset(tag), func(w http.ResponseWriter, r *http.Request) {
		w.Write(binary)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	t.Setenv("MEM_RELEASES_URL", server.URL)
}

func TestNewerOnlyUpdatesReleasesToLaterReleases(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"v0.2.0", "v0.3.0", true},
		{"v0.9.0", "v0.10.0", true},
		{"v1.0.0", "v0.9.9", false},
		{"v0.3.0", "v0.3.0", false},
		{"dev", "v0.3.0", false},
		{"v0.3.0", "v2026.09.28.1", false},
	}
	for _, c := range cases {
		if got := Newer(c.current, c.latest); got != c.want {
			t.Errorf("Newer(%s, %s) = %v", c.current, c.latest, got)
		}
	}
}

func TestInstallReplacesTheExecutableOnlyWithAVerifiedBinary(t *testing.T) {
	target := filepath.Join(t.TempDir(), "mem")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	releases(t, "v0.3.1", []byte("new"), "")
	latest, err := Latest(context.Background())
	if err != nil || latest != "v0.3.1" {
		t.Fatalf("latest = %q, %v", latest, err)
	}
	if err := Install(context.Background(), latest, target); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(target); string(data) != "new" {
		t.Fatalf("target = %q after install", data)
	}
	if info, _ := os.Stat(target); info.Mode().Perm()&0o100 == 0 {
		t.Fatal("the installed binary is not executable")
	}

	releases(t, "v0.3.2", []byte("tampered"), "0000")
	if err := Install(context.Background(), "v0.3.2", target); err == nil {
		t.Fatal("a binary with a wrong checksum was installed")
	}
	if data, _ := os.ReadFile(target); string(data) != "new" {
		t.Fatalf("a failed install changed the target to %q", data)
	}
}
