package error

// State represents the error state when something has gone wrong
type State struct {
	Err error
}

// New creates a new error State.
func New(err error) State {
	return State{Err: err}
}
