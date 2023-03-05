package jeopardy

import (
	"context"
	"sort"

	"github.com/pkg/errors"
	"oss.acmcsuf.com/qg/backend/internal/cando"
	"oss.acmcsuf.com/qg/backend/internal/pubsub"
	"oss.acmcsuf.com/qg/backend/qg"
	"oss.acmcsuf.com/qg/backend/qg/games"
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

// GameState is the current state of a Jeopardy game that's used within
// the Jeopardy state machine.
type GameState struct {
	PlayerScores         map[qg.PlayerName]float32
	PlayerAlreadyPressed map[qg.PlayerName]bool
	AnsweredQuestions    qg.JeopardyAnsweredQuestions
	ChoosingPlayer       qg.PlayerName
	AnsweringPlayer      qg.PlayerName
	CurrentCategory      int32
	CurrentQuestion      int32
}

// PlayerState is the state of a Jeopardy player.
type PlayerState struct {
	Score          float32
	AlreadyPressed bool
}

func newGameState(data qg.JeopardyGameData) *GameState {
	return &GameState{
		PlayerScores:         make(map[qg.PlayerName]float32),
		PlayerAlreadyPressed: make(map[qg.PlayerName]bool),
		AnsweredQuestions:    qg.JeopardyAnsweredQuestions{},
		CurrentCategory:      -1,
		CurrentQuestion:      -1,
	}
}

type gameManager struct {
	pubsub *pubsub.Publisher
	storer Storer

	state   *GameState
	machine *games.MachineState

	data qg.JeopardyGameData
	id   qg.GameID
}

func newGameManager(store Storer, id qg.GameID, data qg.JeopardyGameData, mstate *games.MachineState) *gameManager {
	return &gameManager{
		pubsub:  pubsub.NewPublisher(),
		storer:  store,
		state:   newGameState(data),
		machine: mstate,
		data:    data,
		id:      id,
	}
}

func (m *gameManager) ID() qg.GameID      { return m.id }
func (m *gameManager) Data() qg.IGameData { return qg.GameDataJeopardy{Data: m.data} }

func (m *gameManager) CompareGamePassword(ctx context.Context, input string) (bool, error) {
	return m.storer.CompareGamePassword(ctx, m.id, input)
}

func (m *gameManager) Leaderboard() qg.Leaderboard {
	leaderboard := make(qg.Leaderboard, 0, len(m.machine.Players))
	for name := range m.machine.Players {
		leaderboard = append(leaderboard, qg.LeaderboardEntry{
			PlayerName: name,
			Score:      m.state.PlayerScores[name],
		})
	}

	sort.Slice(leaderboard, func(i, j int) bool {
		return leaderboard[i].Score > leaderboard[j].Score
	})

	return leaderboard
}

func (m *gameManager) BeginGame(ctx context.Context) (cando.NextStates, error) {
	// Pick a random player to start.
	for _, player := range m.machine.Players {
		if player.IsAdmin {
			continue
		}
		m.state.ChoosingPlayer = player.Name
		break
	}

	return m.moveToNextTurn(ctx, false)
}

func (m *gameManager) alreadyAnsweredPlayers() []qg.PlayerName {
	players := make([]qg.PlayerName, 0, len(m.machine.Players))
	for name, player := range m.machine.Players {
		if !m.state.PlayerAlreadyPressed[name] && !player.IsAdmin {
			players = append(players, name)
		}
	}
	return players
}

func (m *gameManager) canContinueQuestion() bool {
	return len(m.alreadyAnsweredPlayers()) < len(m.machine.Players) &&
		m.state.CurrentCategory > 0 &&
		m.state.CurrentQuestion > 0
}

func (m *gameManager) moveToNextTurn(ctx context.Context, stillAnswering bool) (cando.NextStates, error) {
	if stillAnswering && m.canContinueQuestion() {
		return cando.NextStates{
			cando.Next[qg.CommandJeopardyPressButton](),
		}, nil
	}

	// Check that we still have questions that aren't yet answered.
	if len(m.state.AnsweredQuestions) < m.data.TotalQuestions() {
		return cando.NextStates{
			cando.Next[qg.CommandJeopardyChooseQuestion](),
		}, nil
	}

	// End the game since we don't have any questions left.
	// Return no next states, indicating that the game is over.
	return nil, nil
}

// CreateGame implements the games.GameCreator.
func (g Game) CreateGame(ctx context.Context, id qg.GameID, data qg.IGameData) (qg.CommandHandlerFactory, error) {
	jeopardyData, ok := data.(qg.GameDataJeopardy)
	if !ok {
		return nil, errors.Errorf("invalid game data type: %T", data)
	}

	s := games.NewMachineState(ctx)
	m := newGameManager(g.store, id, jeopardyData.Data, s)

	s.AddReactors(
		cando.React[any, qg.CommandJeopardyChooseQuestion](func(ctx context.Context, _ any) error {
			s.Publish(ctx, qg.EventJeopardyTurnEnded{
				Chooser:     m.state.ChoosingPlayer,
				Answered:    m.state.AnsweredQuestions,
				Leaderboard: m.Leaderboard(),
			})
			return nil
		}),
		cando.React[any, qg.CommandJeopardyPressButton](func(ctx context.Context, _ any) error {
			// We can still accept answers, so don't end the turn yet.
			s.Publish(ctx, qg.EventJeopardyResumeButton{
				AlreadyAnsweredPlayers: m.alreadyAnsweredPlayers(),
			})
			return nil
		}),
		cando.React[qg.CommandJeopardyChooseQuestion, any](func(ctx context.Context, prev qg.CommandJeopardyChooseQuestion) error {
			s.Publish(ctx, qg.EventJeopardyBeginQuestion{
				Chooser:  m.state.ChoosingPlayer,
				Category: prev.Category,
				Question: m.data.Categories[prev.Category].Questions[prev.Question].Question,
				Points:   m.data.QuestionPoints(prev.Question),
			})
			return nil
		}),
		cando.React[qg.CommandJeopardyPressButton, any](func(ctx context.Context, prev qg.CommandJeopardyPressButton) error {
			self := games.PlayerFromContext(ctx)
			s.Publish(ctx, qg.EventJeopardyButtonPressed{
				PlayerName: self.Name,
			})
			return nil
		}),
		cando.React[any, qg.CommandJeopardyChooseQuestion](func(ctx context.Context, _ any) error {
			for name := range m.state.PlayerAlreadyPressed {
				m.state.PlayerAlreadyPressed[name] = false
			}
			return nil
		}),
	)

	s.AddState(
		cando.State(func(ctx context.Context, cmd qg.CommandJeopardyChooseQuestion) (cando.NextStates, error) {
			self := games.PlayerFromContext(ctx)
			if self.Name != m.state.ChoosingPlayer {
				return nil, errors.New("not your turn")
			}

			_, _, err := m.data.QuestionAt(cmd.Category, cmd.Question)
			if err != nil {
				return nil, errors.Wrap(err, "invalid question")
			}

			m.state.CurrentCategory = cmd.Category
			m.state.CurrentQuestion = cmd.Question

			return cando.NextStates{
				cando.Next[qg.CommandJeopardyPressButton](),
			}, nil
		}),
		cando.State(func(ctx context.Context, cmd qg.CommandJeopardyPressButton) (cando.NextStates, error) {
			self := games.PlayerFromContext(ctx)

			if m.state.PlayerAlreadyPressed[self.Name] {
				return nil, errors.New("you already pressed your button")
			}

			m.state.PlayerAlreadyPressed[self.Name] = true
			m.state.AnsweringPlayer = self.Name

			return cando.NextStates{
				cando.Next[qg.CommandJeopardyPlayerJudgment](),
			}, nil
		}),
		cando.State(func(ctx context.Context, cmd qg.CommandJeopardyPlayerJudgment) (cando.NextStates, error) {
			self := games.PlayerFromContext(ctx)
			if !self.IsAdmin {
				return nil, errors.New("only admins can judge players")
			}

			if cmd.Correct {
				// Correct, so reward points and move on to the next
				// question.
				pts := m.data.QuestionPoints(m.state.CurrentQuestion)
				m.state.PlayerScores[m.state.AnsweringPlayer] += pts

				// Mark the question as answered.
				m.state.AnsweredQuestions = append(m.state.AnsweredQuestions, qg.JeopardyAnsweredQuestion{
					Player:   m.state.AnsweringPlayer,
					Question: m.state.CurrentQuestion,
					Category: m.state.CurrentCategory,
				})
			}

			return m.moveToNextTurn(ctx, false)
		}),
	)

	return s.StartMachine(ctx, m)
}
