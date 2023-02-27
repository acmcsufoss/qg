package jeopardy

import (
	"context"
	"sort"
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

// Storer is a storage interface for the server.
type Storer interface {
	qg.GameStorer
	// JeopardyGameData returns the game data for the given game.
	JeopardyGameData(ctx context.Context, id qg.GameID) (qg.JeopardyGameData, error)
}

// Game is in charge of creating and managing new Jeopardy games.
type Game struct {
	store Storer
}

// New creates a new Game instance.
func New(store Storer) Game {
	return Game{store}
}

// CreateGame implements the games.GameCreator.
func (g Game) CreateGame(ctx context.Context, id qg.GameID, data qg.GameData) (qg.CommandHandlerFactory, error) {
	jeopardyData, ok := data.Value.(qg.GameDataJeopardy)
	if !ok {
		return nil, errors.Errorf("invalid game data type: %T", data.Value)
	}

	managedData := managedGameData{
		pubsub: pubsub.NewPublisher(),
		storer: g.store,
		data:   jeopardyData.Data,
		id:     id,
	}

	machineData := newMachineData(managedData)
	return &managedGame{
		machine:         cando.NewMachine(machineData),
		managedGameData: managedData,
	}, nil
}

type managedGame struct {
	machine *cando.Machine
	managedGameData
}

type managedGameData struct {
	pubsub *pubsub.Publisher
	storer Storer
	data   qg.JeopardyGameData
	id     qg.GameID
}

func (m *managedGame) NewCommandHandler(ctx context.Context, evs chan<- qg.Event) (qg.CommandHandler, error) {
	m.pubsub.Subscribe(ctx, evs)
	return &playerHandler{m: m, evs: evs}, nil
}

type playerHandler struct {
	evs    chan<- qg.Event
	m      *managedGame
	player *playerState
}

// HandleCommand handles the given command.
func (m *playerHandler) HandleCommand(ctx context.Context, cmd qg.Command) error {
	ctx = context.WithValue(ctx, playerHandlerKey, m)
	return m.m.machine.Change(ctx, cmd)
}

func (m *playerHandler) Close() error {
	m.m.pubsub.Unsubscribe(context.Background(), m.evs)
	return nil
}

func playerHandlerFromContext(ctx context.Context) *playerHandler {
	return ctx.Value(playerHandlerKey).(*playerHandler)
}

func playerFromContext(ctx context.Context) *playerState {
	return playerHandlerFromContext(ctx).player
}

// GameState is the current state of a Jeopardy game that's used within
// the Jeopardy state machine.
type GameState struct {
	Players           map[qg.PlayerName]*PlayerState
	AnsweredQuestions map[[2]int32]qg.PlayerName
	ChoosingPlayer    qg.PlayerName
	AnsweringPlayer   qg.PlayerName
	CurrentCategory   int32
	CurrentQuestion   int32
}

func (s GameState) canContinueQuestion() bool {
	return len(s.alreadyAnsweredPlayers()) < len(s.Players) &&
		s.CurrentCategory > 0 &&
		s.CurrentQuestion > 0
}

func (s GameState) alreadyAnsweredPlayers() []qg.PlayerName {
	players := make([]qg.PlayerName, 0, len(s.Players))
	for name, player := range s.Players {
		if !player.AlreadyPressed && !player.IsAdmin {
			players = append(players, name)
		}
	}
	return players
}

func (s GameState) buildLeaderboard() qg.Leaderboard {
	leaderboard := make(qg.Leaderboard, 0, len(s.Players))
	for name, player := range s.Players {
		leaderboard = append(leaderboard, qg.LeaderboardEntry{
			PlayerName: name,
			Score:      player.Score,
		})
	}

	sort.Slice(leaderboard, func(i, j int) bool {
		return leaderboard[i].Score > leaderboard[j].Score
	})

	return leaderboard
}

// PlayerState is the state of a Jeopardy player.
type PlayerState struct {
	Score          float32
	AlreadyPressed bool
	IsAdmin        bool
}

type playerState struct {
	*PlayerState
	Name qg.PlayerName
}

func newMachineData(m managedGameData) cando.MachineData {
	var mutex sync.Mutex
	state := GameState{
		Players:           make(map[qg.PlayerName]*PlayerState),
		AnsweredQuestions: make(map[[2]int32]qg.PlayerName),
		CurrentCategory:   -1,
		CurrentQuestion:   -1,
	}

	// moveToNextTurn returns the states for a new turn.
	moveToNextTurn := func(ctx context.Context, stillAnswering bool) (cando.NextStates, error) {
		if stillAnswering && state.canContinueQuestion() {
			return cando.NextStates{
				cando.Next[qg.CommandJeopardyPressButton](),
			}, nil
		}

		// Check that we still have questions that aren't yet answered.
		if len(state.AnsweredQuestions) < m.data.TotalQuestions() {
			return cando.NextStates{
				cando.Next[qg.CommandJeopardyChooseQuestion](),
			}, nil
		}

		// End the game since we don't have any questions left.
		// Return no next states, indicating that the game is over.
		return nil, nil
	}

	return cando.MachineData{
		EnterMachine: func(context.Context) error {
			mutex.Lock()
			return nil
		},
		LeaveMachine: func(context.Context) error {
			mutex.Unlock()
			return nil
		},
		Reactors: []cando.AnyReactor{
			cando.Reactor(func(ctx context.Context, _ any, next cando.EndReaction) error {
				return m.pubsub.Publish(ctx, qg.Event{
					Value: qg.EventGameEnded{
						Leaderboard: state.buildLeaderboard(),
					},
				})
			}),
			cando.Reactor(func(ctx context.Context, _ any, next qg.CommandJeopardyChooseQuestion) error {
				return m.pubsub.Publish(ctx, qg.Event{
					Value: qg.EventJeopardyTurnEnded{
						Chooser:     state.ChoosingPlayer,
						Leaderboard: state.buildLeaderboard(),
					},
				})
			}),
			cando.Reactor(func(ctx context.Context, _ any, next qg.CommandJeopardyPressButton) error {
				// We can still accept answers, so don't end the turn yet.
				return m.pubsub.Publish(ctx, qg.Event{
					Value: qg.EventJeopardyResumeButton{
						AlreadyAnsweredPlayers: state.alreadyAnsweredPlayers(),
					},
				})
			}),
			cando.Reactor(func(ctx context.Context, prev qg.CommandJoinGame, _ any) error {
				currentPlayer := playerFromContext(ctx)
				return m.pubsub.Publish(ctx, qg.Event{
					Value: qg.EventJoinedGame{
						GameID: m.id,
						GameInfo: qg.GameInfo{
							Value: qg.GameInfoJeopardy{Data: qg.ConvertJeopardyGameData(m.data)},
						},
						IsAdmin: currentPlayer.IsAdmin,
					},
				})
			}),
			cando.Reactor(func(ctx context.Context, prev qg.CommandJoinGame, _ any) error {
				return m.pubsub.Publish(ctx, qg.Event{
					Value: qg.EventPlayerJoined{
						PlayerName: prev.PlayerName,
					},
				})
			}),
			cando.Reactor(func(ctx context.Context, _ any, next qg.CommandBeginGame) error {
				return m.pubsub.Publish(ctx, qg.Event{
					Value: qg.EventGameStarted{},
				})
			}),
			cando.Reactor(func(ctx context.Context, prev qg.CommandJeopardyChooseQuestion, _ any) error {
				return m.pubsub.Publish(ctx, qg.Event{
					Value: qg.EventJeopardyBeginQuestion{
						Chooser:  state.ChoosingPlayer,
						Category: prev.Category,
						Question: prev.Question,
						Points:   m.data.QuestionPoints(prev.Question),
					},
				})
			}),
			cando.Reactor(func(ctx context.Context, prev qg.CommandJeopardyPressButton, _ any) error {
				currentPlayer := playerFromContext(ctx)
				return m.pubsub.Publish(ctx, qg.Event{
					Value: qg.EventJeopardyButtonPressed{
						PlayerName: currentPlayer.Name,
					},
				})
			}),
		},
		States: []cando.AnyState{
			cando.InitState(func(ctx context.Context) cando.NextStates {
				return cando.NextStates{
					cando.Next[qg.CommandJoinGame](),
				}
			}),
			cando.State(func(ctx context.Context, cmd qg.CommandJoinGame) (cando.NextStates, error) {
				var isAdmin bool
				if cmd.AdminPassword != nil {
					ok, err := m.storer.CompareGamePassword(ctx, m.id, *cmd.AdminPassword)
					if err != nil {
						return nil, errors.Wrap(err, "failed to compare game password")
					}
					if !ok {
						return nil, errors.New("invalid admin password")
					}
					isAdmin = true
				}

				player, ok := state.Players[cmd.PlayerName]
				if !ok {
					player = &PlayerState{IsAdmin: isAdmin}
					state.Players[cmd.PlayerName] = player

					handler := playerHandlerFromContext(ctx)
					handler.player = &playerState{
						PlayerState: player,
						Name:        string(cmd.PlayerName),
					}
				}

				return cando.NextStates{
					cando.Next[qg.CommandJeopardyChooseQuestion](),
					cando.Next[qg.CommandBeginGame](),
				}, nil
			}),
			cando.State(func(ctx context.Context, cmd qg.CommandBeginGame) (cando.NextStates, error) {
				currentPlayer := playerFromContext(ctx)
				if !currentPlayer.IsAdmin {
					return nil, errors.New("only admins can begin the game")
				}

				// Pick a random player to start.
				for name := range state.Players {
					state.ChoosingPlayer = name
					break
				}

				return moveToNextTurn(ctx, false)
			}),
			cando.State(func(ctx context.Context, cmd qg.CommandJeopardyChooseQuestion) (cando.NextStates, error) {
				currentPlayer := playerFromContext(ctx)
				if currentPlayer.Name != state.ChoosingPlayer {
					return nil, errors.New("not your turn")
				}

				_, _, err := m.data.QuestionAt(cmd.Category, cmd.Question)
				if err != nil {
					return nil, errors.Wrap(err, "invalid question")
				}

				state.CurrentCategory = cmd.Category
				state.CurrentQuestion = cmd.Question

				return cando.NextStates{
					cando.Next[qg.CommandJeopardyPressButton](),
				}, nil
			}),
			cando.State(func(ctx context.Context, cmd qg.CommandJeopardyPressButton) (cando.NextStates, error) {
				currentPlayer := playerFromContext(ctx)
				if currentPlayer.AlreadyPressed {
					return nil, errors.New("you already pressed your button")
				}

				currentPlayer.AlreadyPressed = true
				state.AnsweringPlayer = currentPlayer.Name

				return cando.NextStates{
					cando.Next[qg.CommandJeopardyPlayerJudgment](),
				}, nil
			}),
			cando.State(func(ctx context.Context, cmd qg.CommandJeopardyPlayerJudgment) (cando.NextStates, error) {
				currentPlayer := playerFromContext(ctx)
				if !currentPlayer.IsAdmin {
					return nil, errors.New("only admins can judge players")
				}

				questionPos := [2]int32{state.CurrentCategory, state.CurrentQuestion}

				if cmd.Correct {
					// Correct, so reward points and move on to the next
					// question.
					winningPlayer := state.Players[state.AnsweringPlayer]
					winningPlayer.Score += m.data.QuestionPoints(state.CurrentQuestion)

					// Mark the question as answered.
					state.AnsweredQuestions[questionPos] = state.AnsweringPlayer
					state.ChoosingPlayer = state.AnsweringPlayer
				} else {
					// No one answered correctly, so we don't mark the name.
					// Keep the same answering player, but move on to the next
					// question.
					state.AnsweredQuestions[questionPos] = ""
				}

				return moveToNextTurn(ctx, false)
			}),
		},
	}
}
