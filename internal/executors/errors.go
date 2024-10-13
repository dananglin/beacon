package executors

type UnrecognisedCommandError struct {
	command string
}

func (e UnrecognisedCommandError) Error() string {
	return "unrecognised command: " + e.command
}
