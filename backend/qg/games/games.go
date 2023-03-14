package games

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"oss.acmcsuf.com/qg/backend/qg"
)

// GameCreator describes a game manager that can create a new game.
type GameCreator interface {
	// CreateGame creates a new game.
	CreateGame(ctx context.Context, id qg.GameID, data qg.IGameData) (qg.CommandHandlerFactory, error)
}

// Manager manages games and the creation of new games.
type Manager struct {
	gamesMut     sync.RWMutex
	gameCreators map[qg.GameType]GameCreator
	games        map[qg.GameID]qg.CommandHandlerFactory
	store        qg.GameStorer
}

var _ qg.CommandHandlerFactory = (*Manager)(nil)

// NewManager creates a new GameManager.
func NewManager(store qg.GameStorer) *Manager {
	return &Manager{
		gameCreators: make(map[qg.GameType]GameCreator),
		games:        make(map[qg.GameID]qg.CommandHandlerFactory),
		store:        store,
	}
}

// AddGame adds a game creator.
func (g *Manager) AddGame(t qg.GameType, game GameCreator) {
	g.gamesMut.Lock()
	defer g.gamesMut.Unlock()

	_, ok := g.gameCreators[t]
	if ok {
		panic("game already exists")
	}

	g.gameCreators[t] = game
}

// CreateGame creates a new game. The game will be created in the database and
// the game ID will be returned.
func (g *Manager) CreateGame(ctx context.Context, data qg.IGameData) (qg.GameID, error) {
	gameType := qg.GameTypeFromData(data)

	g.gamesMut.RLock()
	gameCreator, ok := g.gameCreators[gameType]
	g.gamesMut.RUnlock()
	if !ok {
		return "", fmt.Errorf("unknown game type %q", gameType)
	}

	id, err := g.store.CreateGame(ctx, data)
	if err != nil {
		return "", errors.Wrap(err, "cannot create game")
	}

	// TODO: use singleflight instead of write mutex
	g.gamesMut.Lock()
	defer g.gamesMut.Unlock()

	game, err := gameCreator.CreateGame(ctx, id, data)
	if err != nil {
		return "", errors.Wrap(err, "cannot create game")
	}

	g.games[id] = game
	return id, nil
}

// NewCommandHandler creates a new command handler.
func (g *Manager) NewCommandHandler(ctx context.Context, evs chan<- qg.IEvent) (qg.CommandHandler, error) {
	return &gameHandler{g, nil, evs}, nil
}

type gameHandler struct {
	gm  *Manager
	gg  qg.CommandHandler
	evs chan<- qg.IEvent
}

func (h *gameHandler) HandleCommand(ctx context.Context, cmd qg.ICommand) (err error) {
	if h.gg != nil {
		return h.gg.HandleCommand(ctx, cmd)
	}

	switch data := cmd.(type) {
	case qg.CommandJoinGame:
		h.gm.gamesMut.RLock()
		game, ok := h.gm.games[data.GameID]
		h.gm.gamesMut.RUnlock()

		if !ok {
			return fmt.Errorf("unknown game with ID %q", data.GameID)
		}

		h.gg, err = game.NewCommandHandler(ctx, h.evs)
		if err != nil {
			return errors.Wrap(err, "cannot create command handler")
		}

		return h.gg.HandleCommand(ctx, cmd)
	default:
		return errors.New("expect join game command")
	}
}

func (h *gameHandler) Close() error {
	if h.gg != nil {
		return h.gg.Close()
	}
	return nil
}
