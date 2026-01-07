package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// mockState is a simple test state for factory testing
type mockState struct {
	data string
}

func (m mockState) View(ctx Context) ViewRenderer {
	return nil
}

func (m mockState) Update(msg tea.Msg, ctx Context) (State, tea.Cmd) {
	return m, nil
}

func TestRegisterStateFactory(t *testing.T) {
	// Save and restore original factories
	originalFactories := stateFactories
	defer func() {
		stateFactories = originalFactories
	}()

	// Clear factories for test
	stateFactories = make(map[Screen]func(any) State)

	// Register a factory
	RegisterStateFactory(LogScreen, func(data any) State {
		return mockState{data: "log"}
	})

	// Verify factory was registered
	if _, ok := stateFactories[LogScreen]; !ok {
		t.Error("Factory was not registered")
	}

	// Verify factory creates correct state
	state := CreateState(LogScreen, nil)
	if ms, ok := state.(mockState); !ok || ms.data != "log" {
		t.Errorf("Factory created incorrect state: %+v", state)
	}
}

func TestCreateState_Panics_WhenNoFactoryRegistered(t *testing.T) {
	// Save and restore original factories
	originalFactories := stateFactories
	defer func() {
		stateFactories = originalFactories
	}()

	// Clear factories for test
	stateFactories = make(map[Screen]func(any) State)

	// Attempt to create state with no factory registered
	defer func() {
		if r := recover(); r == nil {
			t.Error("CreateState did not panic when no factory was registered")
		}
	}()

	CreateState(LogScreen, nil)
}

func TestCreateState_WithData(t *testing.T) {
	// Save and restore original factories
	originalFactories := stateFactories
	defer func() {
		stateFactories = originalFactories
	}()

	// Clear factories for test
	stateFactories = make(map[Screen]func(any) State)

	// Register factory that uses data
	RegisterStateFactory(FilesScreen, func(data any) State {
		d := data.(string)
		return mockState{data: d}
	})

	// Create state with data
	state := CreateState(FilesScreen, "test-data")
	if ms, ok := state.(mockState); !ok || ms.data != "test-data" {
		t.Errorf("Factory did not use data correctly: %+v", state)
	}
}
