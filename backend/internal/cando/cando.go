package cando

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

// StatefulContext is a type that holds a value of type T in any
// context.Context.
type StatefulContext[T any] struct {
	k reflect.Type
}

type statefulBox[T any] struct {
	v *T
}

// NewStatefulContext creates a new StatefulContext.
func NewStatefulContext[T any]() StatefulContext[T] {
	var z T
	return StatefulContext[T]{
		k: reflect.TypeOf(z),
	}
}

// WithContext adds a stateful placeholder value to the given context.
func (c StatefulContext[T]) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, reflect.TypeOf(c.k), &statefulBox[T]{})
}

// Set sets the stateful value in the given context. The context must
// have been created with WithContext; a panic occurs otherwise.
func (c StatefulContext[T]) Set(ctx context.Context, v *T) {
	ctx.Value(reflect.TypeOf(c.k)).(*statefulBox[T]).v = v
}

// Get retrieves the stateful value from the given context. The context must
// have been created with WithContext; a panic occurs otherwise.
func (c StatefulContext[T]) Get(ctx context.Context) *T {
	return ctx.Value(reflect.TypeOf(c.k)).(*statefulBox[T]).v
}

// NextStates is a list of NextState.
type NextStates []NextState

// NextState describes the next state to transition to.
type NextState struct {
	nextType reflect.Type
}

// Next constructs a NextState that can be used to change the current State to
// the given State.
func Next[T any]() NextState {
	var z T
	return NextState{
		nextType: reflect.TypeOf(z),
	}
}

// EndReaction is a special type that indicates to the machine that the function
// is meant to react to the end of the state machine.
type EndReaction struct{}

// React describes a function that reacts to a change in state. The previous
// and next types will describe the state that was left and the state that was
// entered, respectively.
func React[PrevT, NextT any](f func(ctx context.Context, prev PrevT) error) AnyReactor {
	var prevZ PrevT
	var nextZ NextT
	return reactor[PrevT, NextT]{
		f: f,
		t: [2]reflect.Type{
			reflect.TypeOf(prevZ),
			reflect.TypeOf(nextZ),
		},
	}
}

type reactor[PrevT, NextT any] struct {
	f func(ctx context.Context, prev PrevT) error
	t [2]reflect.Type
}

// AnyReactor is an interface that any reactor will implement.
type AnyReactor interface {
	dataTypes() [2]reflect.Type
	react(ctx context.Context, prev any, machine *Machine) error
}

// JoinReactors flattens the given reactors list (or reactors) into a single
// list.
func JoinReactors(reactors ...any) []AnyReactor {
	var out []AnyReactor
	for _, r := range reactors {
		switch r := r.(type) {
		case []AnyReactor:
			out = append(out, r...)
		case AnyReactor:
			out = append(out, r)
		default:
			panic(fmt.Sprintf("unexpected reactor type: %T", r))
		}
	}
	return out
}

var _ AnyReactor = reactor[any, any]{}

func (r reactor[PrevT, NextT]) dataTypes() [2]reflect.Type {
	return r.t
}

func (r reactor[PrevT, NextT]) react(ctx context.Context, prev any, _ *Machine) error {
	return r.f(ctx, prev.(PrevT))
}

// State creates a state in the FSM.
func State[T any](f func(ctx context.Context, value T) (NextStates, error)) AnyState {
	return stateFunc[T](f)
}

// InitState describes the initial state of the FSM. Its sole job should be to
// describe the next possible states when the FSM is started.
type InitState func(ctx context.Context) NextStates

type stateFunc[T any] func(ctx context.Context, value T) (NextStates, error)

// AnyState is an interface that any State will implement.
type AnyState interface {
	dataType() reflect.Type
	enter(ctx context.Context, data any) (NextStates, error)
}

var (
	_ AnyState = (*stateFunc[any])(nil)
	_ AnyState = (*InitState)(nil)
)

func (s stateFunc[T]) dataType() reflect.Type {
	var z T
	return reflect.TypeOf(z)
}

func (s stateFunc[T]) enter(ctx context.Context, data any) (NextStates, error) {
	return s(ctx, data.(T))
}

func (s InitState) dataType() reflect.Type { return nil }

func (s InitState) enter(ctx context.Context, data any) (NextStates, error) {
	return s(ctx), nil
}

// MachineData is a struct that holds the data for creating a finite state
// machine.
type MachineData struct {
	States       []AnyState
	Reactors     []AnyReactor
	EnterMachine func(ctx context.Context) error
	LeaveMachine func(ctx context.Context) error
}

// TODO: find out how to resume a machine from a saved state. We might need to
// store the machine state and the previous input to reproduce the state, or we
// store a list of valid next states. We also want to give the enter and leave
// functions the saved machine.

type savedMachine struct {
	State      stateIdentifier
	NextStates []stateIdentifier
}

type stateIdentifier string

// Machine represents a Finite State Machine, which can have one State active at
// a time.
type Machine struct {
	current     AnyState
	currentData any
	next        NextStates

	state map[reflect.Type]AnyState
	data  MachineData
}

// NewMachine creates a new FSM and returns it.
func NewMachine(data MachineData) *Machine {
	if _, ok := data.States[0].(InitState); !ok {
		panic("first state must be InitState")
	}

	mac := &Machine{
		state: make(map[reflect.Type]AnyState, len(data.States)),
		data:  data,
	}

	for _, state := range data.States {
		mac.state[state.dataType()] = state
	}

	return mac
}

func (f *Machine) enter(ctx context.Context, do func() error) (err error) {
	if err := f.data.EnterMachine(ctx); err != nil {
		return err
	}

	if err = do(); err != nil {
		f.data.LeaveMachine(ctx)
		return err
	}

	if f.data.LeaveMachine != nil {
		if innerErr := f.data.LeaveMachine(ctx); innerErr != nil && err == nil {
			err = innerErr
		}
	}

	prev := f.current.dataType()

	nexts := make([]reflect.Type, len(f.next))
	for i, next := range f.next {
		nexts[i] = next.nextType
	}

reactorMatch:
	for _, reactor := range f.data.Reactors {
		types := reactor.dataTypes()

		if types[0] != nil {
			if prev == nil || !prev.AssignableTo(types[0]) {
				continue reactorMatch
			}
		}

		if types[1] != nil {
			for _, t := range nexts {
				if !t.AssignableTo(types[1]) {
					continue reactorMatch
				}
				break
			}
		}

		if err := reactor.react(ctx, f.currentData, f); err != nil {
			return errors.Wrapf(err, "error reacting to %v", types)
		}
	}

	return
}

// Start starts the FSM. It will call the EnterMachine function, and then
// transition to the first state.
func (f *Machine) Start(ctx context.Context) (err error) {
	return f.enter(ctx, func() error {
		if f.current != nil {
			return errors.New("machine already started")
		}

		f.current = f.data.States[0].(InitState)
		f.next, _ = f.current.enter(ctx, nil)

		return nil
	})
}

// Change allows you to change the current, "main" State assigned to the FSM.
// The caller must have called Start first, otherwise an error is returned.
func (f *Machine) Change(ctx context.Context, data any) (err error) {
	dataType := reflect.TypeOf(data)

	return f.enter(ctx, func() error {
		if f.current == nil {
			return errors.New("machine not started")
		}

		// Validate the transition.
		var next AnyState
		for _, acceptableNext := range f.next {
			if acceptableNext.nextType == dataType {
				next = f.state[acceptableNext.nextType]
				goto allowed
			}
		}

		return fmt.Errorf("cannot change to state of type %T: not allowed", data)
	allowed:

		nextNexts, err := next.enter(ctx, data)
		if err != nil {
			return err
		}

		f.current = next
		f.currentData = data

		f.next = nextNexts
		if f.next == nil {
			f.next = []NextState{{nextType: reflect.TypeOf(EndReaction{})}}
		}

		return err
	})
}
