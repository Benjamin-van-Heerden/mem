// Package gittest gives a test binary a home of its own, so that its tests can run in parallel against real Git.
package gittest

import (
	"os"
	"path/filepath"
	"testing"
)

// Main runs the tests with one HOME shared by the whole binary: tests that call t.Setenv cannot run in parallel, and
// every template library lives at a path of its own, so nothing in the user cache collides. The Git settings turn off
// automatic maintenance. Newer Git runs it detached after pushes and fetches, and a repack still writing into a test
// repository makes t.TempDir's cleanup, or a clone running at the same time, fail after the test itself has passed.
// They go in the global config rather than GIT_CONFIG_COUNT, because Git clears GIT_CONFIG_* for the receiving side of
// a push to a local path, which is where the test remotes' maintenance runs, and Git for Windows keeps the line-ending
// defaults of its system config, which tests should see.
func Main(m *testing.M) {
	home, err := os.MkdirTemp("", "mem-test-home")
	if err != nil {
		panic(err)
	}
	config := "[maintenance]\n\tauto = false\n[gc]\n\tauto = 0\n[receive]\n\tautogc = false\n"
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte(config), 0o644); err != nil {
		panic(err)
	}
	os.Setenv("HOME", home)
	os.Setenv("XDG_CACHE_HOME", "")
	os.Setenv("XDG_CONFIG_HOME", "")
	os.Setenv("LocalAppData", filepath.Join(home, "cache"))
	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}
