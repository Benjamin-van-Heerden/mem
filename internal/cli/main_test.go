package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain turns off Git's automatic maintenance for every git process the tests start. Newer Git runs it
// detached after pushes and fetches, and a repack still writing into a test repository makes t.TempDir's
// cleanup, or a clone running at the same time, fail after the test itself has passed. The settings go in a
// system config file rather than GIT_CONFIG_COUNT, because Git clears GIT_CONFIG_* for the receiving side of a
// push to a local path, which is where the test remotes' maintenance runs.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "mem-cli-test-git")
	if err != nil {
		panic(err)
	}
	config := filepath.Join(dir, "gitconfig")
	if err := os.WriteFile(config, []byte("[maintenance]\n\tauto = false\n[gc]\n\tauto = 0\n[receive]\n\tautogc = false\n"), 0o644); err != nil {
		panic(err)
	}
	os.Setenv("GIT_CONFIG_SYSTEM", config)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
