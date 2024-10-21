package actions

type UnrecognisedResouceError struct {
	resource string
}

func (e UnrecognisedResouceError) Error() string {
	return "unrecognised resource: " + e.resource
}

type Executor interface {
	Execute(args []string) error
}
