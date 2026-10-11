package cli

import (
	"os"
	"strconv"
	"testing"
)

// TestMain turns off Git's automatic maintenance for every git process the tests start. Newer Git runs it
// detached after pushes and fetches, and a repack still writing into a test repository makes t.TempDir's
// cleanup fail after the test itself has passed.
func TestMain(m *testing.M) {
	settings := [][2]string{{"maintenance.auto", "false"}, {"gc.auto", "0"}, {"receive.autogc", "false"}}
	os.Setenv("GIT_CONFIG_COUNT", strconv.Itoa(len(settings)))
	for i, s := range settings {
		os.Setenv("GIT_CONFIG_KEY_"+strconv.Itoa(i), s[0])
		os.Setenv("GIT_CONFIG_VALUE_"+strconv.Itoa(i), s[1])
	}
	os.Exit(m.Run())
}
