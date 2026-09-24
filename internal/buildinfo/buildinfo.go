package buildinfo

import (
	"regexp"
	"runtime/debug"
)

// Set by the release build.
var Version = "dev"
var Commit = "unknown"

var release = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// A `go install module@version` build carries its release version in the
// module info instead. Local builds carry pseudo-versions, which stay "dev".
func init() {
	if Version != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && release.MatchString(info.Main.Version) {
		Version = info.Main.Version
	}
}
