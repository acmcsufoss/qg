-- name: AddGame :exec
INSERT INTO games (id, typ, data) VALUES (?, ?, ?);

-- name: SetGameModeratorPassword :exec
UPDATE games SET mod_password = ? WHERE id = ?;

-- name: GetGameModeratorPassword :one
SELECT mod_password FROM games WHERE id = ?;

-- name: GetGameType :one
SELECT typ FROM games WHERE id = ?;

-- name: GetGameData :one
SELECT data FROM games WHERE id = ? AND typ = ?;

-- name: ListGames :many
SELECT id FROM games;
