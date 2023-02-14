package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"

	"oss.acmcsuf.com/qg/backend/qg"
	"oss.acmcsuf.com/qg/backend/qg/games/jeopardy"
	"oss.acmcsuf.com/qg/backend/qg/stores/sqlite/sqlitec"
	"github.com/pkg/errors"

	_ "modernc.org/sqlite"
)

//go:generate sqlc generate

// https://cj.rs/blog/sqlite-pragma-cheatsheet-for-performance-and-consistency/

const openPragma = `
PRAGMA journal_mode = wal; -- different implementation of the atomicity properties
PRAGMA synchronous = normal; -- synchronise less often to the filesystem
PRAGMA foreign_keys = on; -- check foreign key reference, slightly worst performance
PRAGMA strict = on;
`

const closePragma = `
PRAGMA analysis_limit=400; -- make sure pragma optimize does not take too long
PRAGMA optimize; -- gather statistics to improve query optimization
`

// Store is a SQLite store. It implements qg.Storer.
type Store struct {
	db *sql.DB
	q  *sqlitec.Queries
}

var (
	_ qg.GameStorer   = (*Store)(nil)
	_ jeopardy.Storer = (*Store)(nil)
)

// New creates a new SQLite store.
func New(url string) (*Store, error) {
	db, err := sql.Open("sqlite", url)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(openPragma); err != nil {
		return nil, errors.Wrap(err, "failed to set pragma")
	}

	return &Store{
		db: db,
		q:  sqlitec.New(db),
	}, nil
}

func (s *Store) Close() error {
	if _, err := s.db.Exec(closePragma); err != nil {
		return errors.Wrap(err, "failed to set pragma")
	}

	return s.db.Close()
}

func (s *Store) CreateGame(ctx context.Context, data qg.GameData) (qg.GameID, error) {
	id := qg.GenerateGameID()

	b, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrap(err, "cannot encode data")
	}

	return id, sqliteErr(s.q.AddGame(ctx, sqlitec.AddGameParams{
		ID:   id,
		Typ:  string(qg.GameTypeFromData(data)),
		Data: b,
	}))
}

func (s *Store) GameType(ctx context.Context, id qg.GameID) (qg.GameType, error) {
	v, err := s.q.GetGameType(ctx, id)
	if err != nil {
		return "", sqliteErr(err)
	}
	return qg.GameType(v), nil
}

func (s *Store) Games(ctx context.Context) ([]qg.GameID, error) {
	gameStrs, err := s.q.ListGames(ctx)
	if err != nil {
		return nil, sqliteErr(err)
	}

	games := make([]qg.GameID, len(gameStrs))
	for i, g := range gameStrs {
		games[i] = qg.GameID(g)
	}

	return games, nil
}

func (s *Store) SetGamePassword(ctx context.Context, id qg.GameID, password string) error {
	return sqliteErr(s.q.SetGameModeratorPassword(ctx, sqlitec.SetGameModeratorPasswordParams{
		ID:          id,
		ModPassword: sql.NullString{String: password, Valid: password != ""},
	}))
}

func (s *Store) CompareGamePassword(ctx context.Context, id qg.GameID, password string) (bool, error) {
	p, err := s.q.GetGameModeratorPassword(ctx, id)
	if err != nil {
		return false, sqliteErr(err)
	}

	if !p.Valid {
		return false, nil
	}

	// TODO: give more fucks about timing attacks
	return p.String == password, nil
}

func (s *Store) JeopardyGameData(ctx context.Context, id qg.GameID) (qg.JeopardyGameData, error) {
	var data qg.JeopardyGameData

	b, err := s.q.GetGameData(ctx, sqlitec.GetGameDataParams{
		ID:  id,
		Typ: string(qg.GameTypeJeopardy),
	})
	if err != nil {
		return data, sqliteErr(err)
	}

	if err := json.Unmarshal(b, &data); err != nil {
		return data, errors.Wrap(err, "cannot decode data")
	}

	return data, nil
}

func sqliteErr(err error) error {
	return err
}
