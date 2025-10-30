package states

// State represents the error state when something has gone wrong
type ErrorState struct {
	Err error
}
