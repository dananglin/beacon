package utilities

import "errors"

type ParsedArgs struct {
	Name string
	Args []string
}

var ErrNoArgumentsProvided = errors.New("no arguments were provided")

func ParseArgs(args []string) (ParsedArgs, error) {
	if len(args) == 0 {
		return ParsedArgs{}, ErrNoArgumentsProvided
	}

	if len(args) == 1 {
		return ParsedArgs{
			Name: args[0],
			Args: make([]string, 0),
		}, nil
	}

	return ParsedArgs{
		Name: args[0],
		Args: args[1:],
	}, nil
}
