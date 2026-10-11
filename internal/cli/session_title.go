package cli

import "fmt"

// Session titles name the project and the work, so sessions listed together across projects can be told
// apart. The instructions name the desktop app because that is where an agent can rename its own session.

// nameSession asks for the session's first real title, once the user has said what to work on.
func nameSession(name string) string {
	return fmt.Sprintf("If you can rename this session (as in the Claude desktop app), title it \"%s · <the work>\" as soon as the user has said what to work on: now if they already have, otherwise after their next message.", name)
}

// retitleSession asks the agent to check the title against where the session has gone since it was named,
// at points where enough work has happened for it to have moved on.
func retitleSession(name string) string {
	return fmt.Sprintf("If you can rename this session (as in the Claude desktop app), look up its current title and compare it with what the session is working on now. Sessions drift: if the title no longer describes the current work, rename it \"%s · <the current work>\" now. Leave it only if it still fits.", name)
}
