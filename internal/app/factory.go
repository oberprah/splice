package app

import (
	"fmt"
)

// stateFactories holds the registered state factory functions
var stateFactories = map[Screen]func(any) State{}

// RegisterStateFactory registers a factory function for a given screen type.
// This should be called in init() functions of state packages to register themselves.
func RegisterStateFactory(screen Screen, factory func(any) State) {
	stateFactories[screen] = factory
}

// CreateState creates a new state for the given screen using the registered factory.
// Panics if no factory is registered for the screen.
func CreateState(screen Screen, data any) State {
	factory, ok := stateFactories[screen]
	if !ok {
		panic(fmt.Sprintf("no factory registered for screen %v", screen))
	}
	return factory(data)
}
