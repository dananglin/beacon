package executors

type command struct {
	name string
	args []string
}

func newCommand(args []string) command {
	if len(args) == 0 {
		return command{
			name: "help",
			args: make([]string, 0),
		}
	}

	if len(args) == 1 {
		return command{
			name: args[0],
			args: make([]string, 0),
		}
	}

	return command{name: args[0], args: args[1:]}
}
