package cando

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// State is a state in the FSM.
type State struct {
	Name string

	Enter  func(ctx context.Context, prev State) error
	Update func(ctx context.Context) error
	Exit   func(ctx context.Context) error
}

// Transitions is a list of allowed state transitions.
type Transitions map[string][]string

// MachineData is a struct that holds the data for creating a finite state
// machine.
type MachineData struct {
	Transitions Transitions
	States      []State

	EnterMachine func(ctx context.Context) error
	LeaveMachine func(ctx context.Context) error
}

// Machine represents a Finite State Machine, which can have one State active at
// a time.
type Machine struct {
	current     State
	state       map[string]State
	transitions Transitions
	data        MachineData
}

// NewMachine creates a new FSM and returns it.
func NewMachine(data MachineData) *Machine {
	mac := &Machine{
		state:       make(map[string]State, len(data.States)),
		transitions: data.Transitions,
		data:        data,
	}

	for _, state := range data.States {
		mac.state[state.Name] = state
	}

	return mac
}

// Update triggers an update on the active State.
func (f *Machine) Update(ctx context.Context) {
	if f.current.Update != nil {
		f.current.Update(ctx)
	}
}

// Change allows you to change the current, "main" State assigned to the FSM. If
// you run Change(), it will call Exit() on the previous State and Enter() on
// the next State.
func (f *Machine) Change(ctx context.Context, stateName string) (err error) {
	if err := f.data.EnterMachine(ctx); err != nil {
		return err
	}

	if f.data.LeaveMachine != nil {
		defer func() {
			if innerErr := f.data.LeaveMachine(ctx); innerErr != nil && err == nil {
				err = innerErr
			}
		}()
	}

	prev := f.current

	// Validate that the state exists and that we can enter it before we do so.
	next, ok := f.state[stateName]
	if !ok {
		return fmt.Errorf("cannot change to state %q: state does not exist", stateName)
	}

	transition := f.transitions[prev.Name]
	if transition != nil {
		for _, t := range transition {
			if t == stateName {
				goto allowed
			}
		}
		return fmt.Errorf("cannot change to state %q: transition not allowed", stateName)
	allowed:
	}

	if prev.Exit != nil {
		if err := prev.Exit(ctx); err != nil {
			return errors.Wrapf(err, "error exiting state %q", prev.Name)
		}
	}

	if next.Enter != nil {
		if err := prev.Enter(ctx, prev); err != nil {
			return errors.Wrapf(err, "error entering state %q", next.Name)
		}
	}

	f.current = next
	return nil
}
