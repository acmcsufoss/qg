package jeopardy

import (
	"context"
	"sort"
	"sync"

	"etok.codes/qg/backend/qg"
	"etok.codes/qg/backend/qg/internal/cando"
	"github.com/pkg/errors"
)

var playerContext = cando.NewStatefulContext[playerState]()

// GameManager manages a game of Jeopardy.
type GameManager struct {
	machine *cando.Machine

	pubber qg.Broadcaster
	data   qg.JeopardyGameData
	id     qg.GameID
}

// NewGame creates a new managed game of Jeopardy.
func NewGame(id qg.GameID, data qg.JeopardyGameData, pubber qg.Broadcaster) (*GameManager, error) {
	machineData := NewMachineData(id, data, pubber)
	return &GameManager{
		machine: cando.NewMachine(machineData),
		pubber:  pubber,
		data:    data,
		id:      id,
	}, nil
}

// HandleCommand handles the given command.
func (m *GameManager) HandleCommand(ctx context.Context, cmd qg.Command) error {
	return m.machine.Change(ctx, cmd)
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

// NewMachineData creates a new JeopardyMachineData.
func NewMachineData(id qg.GameID, data qg.JeopardyGameData, pubber qg.Broadcaster) cando.MachineData {
	var mutex sync.Mutex
	state := GameState{
		Players:           make(map[qg.PlayerName]*PlayerState),
		AnsweredQuestions: make(map[[2]int32]qg.PlayerName),
		CurrentCategory:   -1,
		CurrentQuestion:   -1,
	}

	gameTopic := qg.TopicEventForGame(id)

	// moveToNextTurn returns the states for a new turn.
	moveToNextTurn := func(ctx context.Context, stillAnswering bool) (cando.NextStates, error) {
		if stillAnswering && state.canContinueQuestion() {
			return cando.NextStates{
				cando.Next[qg.CommandJeopardyPressButton](),
			}, nil
		}

		// Check that we still have questions that aren't yet answered.
		if len(state.AnsweredQuestions) < data.TotalQuestions() {
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
				return gameTopic.Broadcast(ctx, pubber, qg.Event{
					Value: qg.EventGameEnded{
						Leaderboard: state.buildLeaderboard(),
					},
				})
			}),
			cando.Reactor(func(ctx context.Context, _ any, next qg.CommandJeopardyChooseQuestion) error {
				return gameTopic.Broadcast(ctx, pubber, qg.Event{
					Value: qg.EventJeopardyTurnEnded{
						Chooser:     state.ChoosingPlayer,
						Leaderboard: state.buildLeaderboard(),
					},
				})
			}),
			cando.Reactor(func(ctx context.Context, _ any, next qg.CommandJeopardyPressButton) error {
				// We can still accept answers, so don't end the turn yet.
				return gameTopic.Broadcast(ctx, pubber, qg.Event{
					Value: qg.EventJeopardyResumeButton{
						AlreadyAnsweredPlayers: state.alreadyAnsweredPlayers(),
					},
				})
			}),
			cando.Reactor(func(ctx context.Context, prev qg.CommandJoinGame, _ any) error {
				return gameTopic.Broadcast(ctx, pubber, qg.Event{
					Value: qg.EventPlayerJoined{
						PlayerName: prev.PlayerName,
					},
				})
			}),
			cando.Reactor(func(ctx context.Context, prev qg.CommandJeopardyChooseQuestion, _ any) error {
				return gameTopic.Broadcast(ctx, pubber, qg.Event{
					Value: qg.EventJeopardyBeginQuestion{
						Chooser:  state.ChoosingPlayer,
						Category: prev.Category,
						Question: prev.Question,
						Points:   data.QuestionPoints(prev.Question),
					},
				})
			}),
			cando.Reactor(func(ctx context.Context, prev qg.CommandJeopardyPressButton, _ any) error {
				currentPlayer := playerContext.Get(ctx)

				return gameTopic.Broadcast(ctx, pubber, qg.Event{
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
				if cmd.ModeratorPassword != nil && *cmd.ModeratorPassword == data.ModeratorPassword {
					isAdmin = true
				}

				player, ok := state.Players[cmd.PlayerName]
				if !ok {
					player = &PlayerState{IsAdmin: isAdmin}
					playerContext.Set(ctx, &playerState{
						PlayerState: player,
						Name:        string(cmd.PlayerName),
					})
					state.Players[cmd.PlayerName] = player
				}

				return cando.NextStates{
					cando.Next[qg.CommandJoinGame](),
					cando.Next[qg.CommandBeginGame](),
				}, nil
			}),
			cando.State(func(ctx context.Context, cmd qg.CommandBeginGame) (cando.NextStates, error) {
				if !playerContext.Get(ctx).IsAdmin {
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
				currentPlayer := playerContext.Get(ctx)
				if currentPlayer.Name != state.ChoosingPlayer {
					return nil, errors.New("not your turn")
				}

				_, _, err := data.QuestionAt(cmd.Category, cmd.Question)
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
				currentPlayer := playerContext.Get(ctx)
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
				currentPlayer := playerContext.Get(ctx)
				if !currentPlayer.IsAdmin {
					return nil, errors.New("only admins can judge players")
				}

				questionPos := [2]int32{state.CurrentCategory, state.CurrentQuestion}

				if cmd.Correct {
					// Correct, so reward points and move on to the next
					// question.
					winningPlayer := state.Players[state.AnsweringPlayer]
					winningPlayer.Score += data.QuestionPoints(state.CurrentQuestion)

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
