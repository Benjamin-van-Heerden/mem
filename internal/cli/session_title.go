package cli

// Session titles name the work, so sessions in the desktop app's list can be told apart. The instructions name
// the desktop app because that is where an agent can rename its own session.

// nameSession asks for the session's first real title, once the user has said what to work on.
func nameSession() string {
	return `If you can rename this session (as in the Claude desktop app, whose rename tool takes the session id "self" for this session), give it a short title for the work, a few words such as "Tugela charger startup issue". Name only the work, not the project: the app already shows which folder a session belongs to. Do it as soon as the user has said what to work on: now if they already have, otherwise as the first thing you do after their next message. If the rename fails, correct the call and retry.`
}

// sessionTitleReminder ends a context long enough to be written to a file: the agent reads that file after the
// instruction, so without it the naming step sits far behind what it read last when the user's next message comes.
func sessionTitleReminder() string {
	return "Remember to name this session. " + nameSession()
}

// retitleSession asks the agent to check the title against where the session has gone since it was named,
// at points where enough work has happened for it to have moved on.
func retitleSession() string {
	return `If you can rename this session (as in the Claude desktop app, whose session tools take the session id "self" for this session), look up its current title and compare it with what the session is working on now. Sessions drift: if the title no longer describes the current work, rename it to a short title for the current work now. Leave it only if it still fits. If the rename fails, correct the call and retry.`
}
