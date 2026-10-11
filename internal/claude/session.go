package claude

import (
	"os"
	"path/filepath"
	"regexp"
)

// SessionEnv names the Claude Code session a command runs in; Claude Code sets it for the commands it runs.
const SessionEnv = "CLAUDE_CODE_SESSION_ID"

var sessionID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

// MarkOnboarded records that the session onboarded, so its next prompt carries the session-title hint. dir is
// a directory Git ignores; an unusable session id records nothing.
func MarkOnboarded(dir, session string) error {
	if !sessionID.MatchString(session) {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, session), nil, 0o644)
}

// TakeTitleHint reports whether the session onboarded since its last prompt, and clears that record so the
// hint is given once.
func TakeTitleHint(dir, session string) bool {
	if !sessionID.MatchString(session) {
		return false
	}
	return os.Remove(filepath.Join(dir, session)) == nil
}
