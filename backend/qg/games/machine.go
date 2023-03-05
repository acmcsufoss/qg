package games

import (
	"context"
	"log"
	"sync"

	"github.com/pkg/errors"
	"oss.acmcsuf.com/qg/backend/internal/cando"
	"oss.acmcsuf.com/qg/backend/internal/pubsub"
	"oss.acmcsuf.com/qg/backend/qg"
)

type ctxKey int

const (
	playerHandlerKey ctxKey = iota
)

// PlayerHandle is a handle for a player. It is somewhat different from an
// actual player: a handle is bound to a connection, while a player is
// identified by a name.
type PlayerHandle struct {
	*pubsub.Publisher
	*PlayerState
}

// PlayerState is the state of a player.
type PlayerState struct {
	Name    qg.PlayerName
	IsAdmin bool
}

func injectPlayerHandler(ctx context.Context, h *PlayerHandle) context.Context {
	return context.WithValue(ctx, playerHandlerKey, h)
}

// PlayerFromContext returns the player handle from the context. The context
// will contain a player name once the player sends a CommandJoinGame.
func PlayerFromContext(ctx context.Context) *PlayerHandle {
	return ctx.Value(playerHandlerKey).(*PlayerHandle)
}

// GameManager is the initial state data supplied to the machine. The data
// contained here are immutable.
type GameManager interface {
	ID() qg.GameID
	Data() qg.IGameData
	// CompareGamePassword is a function to compare the game password.
	CompareGamePassword(context.Context, string) (bool, error)
	// BeginGame is the entrypoint function for the game-specific states.
	BeginGame(ctx context.Context) (cando.NextStates, error)
	// Leaderboard builds a leaderboard for the game.
	Leaderboard() qg.Leaderboard
}

// MachineState controls a game using a state machine.
type MachineState struct {
	*pubsub.Publisher
	Players map[string]*PlayerState

	states   []cando.AnyState
	reactors []cando.AnyReactor
}

// NewMachineState creates a new default machine controller for a game. It
// handles all the basic commands.
func NewMachineState(ctx context.Context) *MachineState {
	return &MachineState{
		Publisher: pubsub.NewPublisher(),
		Players:   make(map[string]*PlayerState),
	}
}

// AddReactors adds the given reactors to the machine.
func (m *MachineState) AddReactors(reactors ...cando.AnyReactor) {
	m.reactors = append(m.reactors, reactors...)
}

// AddState adds a new state to the machine.
func (m *MachineState) AddState(states ...cando.AnyState) {
	m.states = append(m.states, states...)
}

// StartMachine starts a new machine from the current state. Only one machine
// can use the state at a time.
func (s *MachineState) StartMachine(ctx context.Context, game GameManager) (*Machine, error) {
	var mdata cando.MachineData
	var mutex sync.Mutex

	mdata.EnterMachine = func(context.Context) error {
		mutex.Lock()
		return nil
	}

	mdata.LeaveMachine = func(context.Context) error {
		mutex.Unlock()
		return nil
	}

	mdata.States = []cando.AnyState{
		cando.InitState(func(ctx context.Context) cando.NextStates {
			return cando.NextStates{
				cando.Next[qg.CommandJoinGame](),
			}
		}),
		cando.State(func(ctx context.Context, cmd qg.CommandJoinGame) (cando.NextStates, error) {
			var isAdmin bool
			if cmd.AdminPassword != nil {
				ok, err := game.CompareGamePassword(ctx, *cmd.AdminPassword)
				if err != nil {
					return nil, errors.Wrap(err, "failed to compare game password")
				}
				if !ok {
					return nil, errors.New("invalid admin password")
				}
				isAdmin = true
			}

			_, ok := s.Players[cmd.PlayerName]
			if ok {
				return nil, errors.New("player already exists")
			}

			player := &PlayerState{
				Name:    cmd.PlayerName,
				IsAdmin: isAdmin,
			}

			s.Players[cmd.PlayerName] = player

			self := PlayerFromContext(ctx)
			self.PlayerState = player

			return cando.NextStates{
				cando.Next[qg.CommandJoinGame](),
				cando.Next[qg.CommandBeginGame](),
			}, nil
		}),
		cando.State(func(ctx context.Context, cmd qg.CommandBeginGame) (cando.NextStates, error) {
			self := PlayerFromContext(ctx)
			if !self.IsAdmin {
				return nil, errors.New("only admins can begin the game")
			}

			return game.BeginGame(ctx)
		}),
	}

	mdata.Reactors = cando.JoinReactors(
		cando.React[qg.CommandJoinGame, any](func(ctx context.Context, prev qg.CommandJoinGame) error {
			self := PlayerFromContext(ctx)

			var gameData *qg.GameData
			if self.IsAdmin {
				gameData = &qg.GameData{Value: game.Data()}
			}

			log.Println("publish to self")
			self.Publish(ctx, qg.EventJoinedGame{
				GameID:   game.ID(),
				GameInfo: qg.GameInfo{Value: qg.GameInfoFromData(game.Data())},
				GameData: gameData,
				IsAdmin:  gameData != nil,
			})

			return nil
		}),
		cando.React[qg.CommandJoinGame, any](func(ctx context.Context, prev qg.CommandJoinGame) error {
			// Broadcast to all players that this player has joined.
			s.Publish(ctx, qg.EventPlayerJoined{
				PlayerName: prev.PlayerName,
			})

			// Broadcast to this player the names of all other players.
			for player := range s.Players {
				if player == prev.PlayerName {
					continue
				}

				self := PlayerFromContext(ctx)
				self.Publish(ctx, qg.EventPlayerJoined{PlayerName: player})
			}

			return nil
		}),
		cando.React[qg.CommandBeginGame, any](func(ctx context.Context, _ qg.CommandBeginGame) error {
			s.Publish(ctx, qg.EventGameStarted{})
			return nil
		}),
		cando.React[any, cando.EndReaction](func(ctx context.Context, _ any) error {
			s.Publish(ctx, qg.EventGameEnded{
				Leaderboard: game.Leaderboard(),
			})
			return nil
		}),
	)

	mdata.States = append(mdata.States, s.states...)
	mdata.Reactors = append(mdata.Reactors, s.reactors...)

	m := cando.NewMachine(mdata)
	if err := m.Start(ctx); err != nil {
		return nil, err
	}

	return &Machine{s, m}, nil
}

// Machine is a running game state machine.
type Machine struct {
	s *MachineState
	m *cando.Machine
}

// NewCommandHandler creates a new command handler for a player.
func (m *Machine) NewCommandHandler(ctx context.Context, evs chan<- qg.IEvent) (qg.CommandHandler, error) {
	pubsub := pubsub.NewPublisher()
	pubsub.Subscribe(evs)
	m.s.Publisher.SubscribePublisher(pubsub)

	return &playerCommandHandler{
		handle:  &PlayerHandle{Publisher: pubsub},
		machine: m,
	}, nil
}

type playerCommandHandler struct {
	handle  *PlayerHandle
	machine *Machine
}

func (h *playerCommandHandler) HandleCommand(ctx context.Context, cmd qg.ICommand) error {
	ctx = injectPlayerHandler(ctx, h.handle)
	return h.machine.m.Change(ctx, cmd)
}

func (h *playerCommandHandler) Close() error {
	h.machine.s.Publisher.UnsubscribePublisher(h.handle.Publisher)
	return nil
}
