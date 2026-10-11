package cli

import "fmt"

// renameSession asks the agent to title its session after the project and the work, so sessions listed
// together across projects can be told apart. It names the desktop app because that is where an agent can
// rename its own session.
func renameSession(name, when string) string {
	return fmt.Sprintf("If you can rename this session (as in the Claude desktop app), title it \"%s · <the work>\" %s.", name, when)
}
