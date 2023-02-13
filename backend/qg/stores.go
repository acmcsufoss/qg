package qg

import "context"

// GameStorer is a store for games.
type GameStorer interface {
	// MakeNewGame creates a new game using the given game data. The new game ID
	// is returned.
	MakeNewGame(context.Context, GameData) (GameID, error)
	// GameType gets the game type for the given game ID.
	GameType(context.Context, GameID) (GameType, error)
	// Games gets the game IDs for all games.
	Games(context.Context) ([]GameID, error)
}
