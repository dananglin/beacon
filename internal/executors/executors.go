package executors

func Execute(args []string) error {
	command := newCommand(args)

	executorFuncMap := map[string]func(args []string) error{
		"serve":   executeServeCommand,
		"version": executeVersionCommand,
	}

	executeFunc, ok := executorFuncMap[command.name]
	if !ok {
		return UnrecognisedCommandError{command.name}
	}

	return executeFunc(command.args)
}
