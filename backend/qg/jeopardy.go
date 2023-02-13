package qg

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

// ValidateJeopardyGameData performs additional validation on the given Jeopardy
// game data. The function will make changes to the data.
func ValidateJeopardyGameData(data *JeopardyGameData) error {
	if len(data.Categories) == 0 {
		return fmt.Errorf("no categories found, must have at least one")
	}

	nQuestions := len(data.Categories[0].Questions)
	for i, c := range data.Categories {
		if len(c.Questions) != nQuestions {
			return fmt.Errorf("category %d has %d questions, expected %d", i+1, len(c.Questions), nQuestions)
		}
	}

	if data.ScoreMultiplier == nil {
		*data.ScoreMultiplier = 100
	}

	return nil
}

// JeopardyGameManager is a game manager for Jeopardy games.
type JeopardyGameManager struct {
	gamesMu sync.RWMutex
	games   map[GameID]*ManagedJeopardyGame
	store   JeopardyGameStorer
	pubsub  Broadcaster
}

// NewJeopardyGameManager creates a new JeopardyGameManager.
func NewJeopardyGameManager(store JeopardyGameStorer, pubsub Broadcaster) JeopardyGameManager {
	return JeopardyGameManager{
		games:  make(map[GameID]*ManagedJeopardyGame),
		store:  store,
		pubsub: pubsub,
	}
}

// AddGame adds the given game into the manager. A managed game is returned.
func (m *JeopardyGameManager) AddGame(ctx context.Context, id GameID) (*ManagedJeopardyGame, error) {
	m.gamesMu.RLock()
	if game, ok := m.games[id]; ok {
		m.gamesMu.RUnlock()
		return game, nil
	}
	m.gamesMu.RUnlock()

	data, err := m.store.Data(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get game data")
	}

	if err := ValidateJeopardyGameData(data); err != nil {
		return nil, errors.Wrap(err, "invalid game data")
	}

	game := &ManagedJeopardyGame{
		players: make(map[PlayerName]*ManagedJeopardyGamePlayer),
		id:      id,
		data:    *data,
		info:    jeopardyGameInfo(*data),
		store:   m.store,
		pubsub:  m.pubsub,
	}

	m.gamesMu.Unlock()
	defer m.gamesMu.Unlock()

	// Check in case AddGame is called concurrently.
	if game, ok := m.games[id]; ok {
		return game, nil
	}

	m.games[id] = game
	return game, nil
}

// ManagedJeopardyGame is a managed Jeopardy game.
type ManagedJeopardyGame struct {
	playersMu sync.RWMutex
	players   map[PlayerName]*ManagedJeopardyGamePlayer

	id     GameID
	data   JeopardyGameData
	info   JeopardyGameInfo
	store  JeopardyGameStorer
	pubsub Broadcaster
}

// ID returns the game ID.
func (g *ManagedJeopardyGame) ID() GameID {
	return g.id
}

// Info gets the current Jeopardy game's info.
func (g *ManagedJeopardyGame) Info() JeopardyGameInfo {
	return g.info
}

// AddPlayer adds the given player to the game.
func (g *ManagedJeopardyGame) AddPlayer(ctx context.Context, name PlayerName) (*ManagedJeopardyGamePlayer, error) {
	if err := g.store.AddPlayer(ctx, g.id, name); err != nil {
		return nil, errors.Wrap(err, "failed to add player to game")
	}

	g.mutex.Lock()
	player, ok := g.players[name]
	if !ok {
		player = &ManagedJeopardyGamePlayer{
			name: name,
			game: g,
		}
		g.players[name] = player
	}
	g.mutex.Unlock()

	if err := TopicEventForGame(g.id).Broadcast(ctx, g.pubsub, Event{
		Value: EventPlayerJoined{
			PlayerName: name,
		},
	}); err != nil {
		return player, errors.Wrap(err, "failed to broadcast player joined event")
	}

	return player, nil
}

// Begin begins the game.
func (g *ManagedJeopardyGame) Begin(ctx context.Context) error {
	g.mutex.Lock()

	g.currentCategory = -1
	g.currentQuestion = -1
	g.mutex.Unlock()

	if err := TopicEventForGame(g.id).Broadcast(ctx, g.pubsub, Event{
		Value: EventJeopardyTurnEnded{
			Chooser: startingPlayer,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to broadcast turn ended event")
	}

	return nil
}

func (g *ManagedJeopardyGame) PlayerJudgment(ctx context.Context, correct bool) error {
	g.mutex.Lock()
	if g.answering == "" {
		g.mutex.Unlock()
		return errors.New("no player is answering")
	}

	answering := g.answering

	g.answering = ""
	g.chooser = ""

	if correct {
		g.chooser = answering
	}
	g.mutex.Unlock()

	var ev Event
	if correct {
		ev.Value = EventJeopardyTurnEnded{
			Chooser: answering,
		}
	} else {
		ev.Value = EventJeopardyResumeButton{}
	}

	if err := TopicEventForGame(g.id).Broadcast(ctx, g.pubsub, Event{
		Value: EventJeopardyAnswered{
			Answerer: answering,
			Correct:  correct,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to broadcast answer event")
	}
}

// ManagedJeopardyGamePlayer is a managed Jeopardy game player.
type ManagedJeopardyGamePlayer struct {
	answered bool

	game *ManagedJeopardyGame
	name PlayerName
}

func (g *ManagedJeopardyGamePlayer) newTurn() {
	g.answered = false
}

// Game returns the game this player is in.
func (g *ManagedJeopardyGamePlayer) Game() *ManagedJeopardyGame {
	return g.game
}

// ChooseQuestion chooses the given question for the player.
func (g *ManagedJeopardyGamePlayer) ChooseQuestion(ctx context.Context, categoryIx, questionIx int) error {
	category, question, err := g.game.data.QuestionAt(categoryIx, questionIx)
	if err != nil {
		return errors.Wrap(err, "failed to get question")
	}

	g.game.mutex.Lock()
	if g.game.chooser != g.name {
		g.game.mutex.Unlock()
		return errors.New("not the chooser")
	}

	g.game.chooser = ""
	g.game.currentCategory = categoryIx
	g.game.currentQuestion = questionIx
	g.game.mutex.Unlock()

	if err := g.game.store.MarkQuestionAsAnswered(ctx, g.game.id, categoryIx, questionIx); err != nil {
		return errors.Wrap(err, "failed to mark question as answered")
	}

	if err := TopicEventForGame(g.game.id).Broadcast(ctx, g.game.pubsub, Event{
		Value: EventJeopardyBeginQuestion{
			Chooser:  g.name,
			Points:   g.game.data.QuestionPoints(questionIx),
			Category: category.Name,
			Question: question.Question,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to broadcast begin question event")
	}

	return nil
}

// PressButton presses the button for the player.
func (g *ManagedJeopardyGamePlayer) PressButton(ctx context.Context) error {
	g.game.mutex.Lock()

	if g.game.answering {
		g.game.mutex.Unlock()
		return errors.New("already answering")
	}

	if g.game.currentCategory == -1 || g.game.currentQuestion == -1 {
		g.game.mutex.Unlock()
		return errors.New("no question is currently being answered")
	}

	g.game.answering = true
	g.game.mutex.Unlock()

	if err := TopicEventForGame(g.game.id).Broadcast(ctx, g.game.pubsub, Event{
		Value: EventJeopardyButtonPressed{
			PlayerName: g.name,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to broadcast button pressed event")
	}

	return nil
}

func jeopardyGameInfo(data JeopardyGameData) JeopardyGameInfo {
	categories := make([]string, len(data.Categories))
	for i, c := range data.Categories {
		categories[i] = c.Name
	}

	return JeopardyGameInfo{
		Categories:      categories,
		NumQuestions:    int32(len(data.Categories[0].Questions)),
		ScoreMultiplier: *data.ScoreMultiplier,
	}
}
