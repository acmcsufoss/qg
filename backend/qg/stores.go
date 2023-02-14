package qg

import "context"

// GameStorer is a store for games.
type GameStorer interface {
	// MakeNewGame creates a new game using the given game data. The new game ID
	// is returned.
	MakeNewGame(context.Context, GameData) (GameID, error)
	// SetGamePassword sets the moderator password for the given game. Users
	// that connect to the game with the given password will be considered
	// moderators.
	SetGamePassword(context.Context, GameID, string) error
	// CompareGamePassword compares the given password to the moderator
	// password for the given game. If the passwords match, true is returned.
	CompareGamePassword(context.Context, GameID, string) (bool, error)
	// GameType gets the game type for the given game ID.
	GameType(context.Context, GameID) (GameType, error)
	// Games gets the game IDs for all games.
	Games(context.Context) ([]GameID, error)
}
