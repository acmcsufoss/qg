package qg

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sync"

	cryptorand "crypto/rand"
	mathrand "math/rand"
)

var (
	idRand     *mathrand.Rand
	idRandOnce sync.Once
)

func needIDRand() {
	idRandOnce.Do(func() {
		var seedBuf [8]byte

		_, err := cryptorand.Read(seedBuf[:])
		if err != nil {
			panic("cannot read random bytes: " + err.Error())
		}

		seed := binary.BigEndian.Uint64(seedBuf[:])
		idRand = mathrand.New(mathrand.NewSource(int64(seed)))
	})
}

// GenerateGameID generates a new random game ID.
func GenerateGameID() GameID {
	// generate 4 letters or numbers.
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	needIDRand()

	var id [4]byte

	for i := range id {
		n := idRand.Intn(len(letters))
		id[i] = letters[n]
	}

	return GameID(id[:])
}

// GameTypeFromData returns the game type from the given game data.
func GameTypeFromData(data GameData) GameType {
	var gameType GameType
	// TODO: add body.Data.Type().
	switch data.Value.(type) {
	case GameDataJeopardy:
		gameType = GameTypeJeopardy
	case GameDataKahoot:
		gameType = GameTypeKahoot
	default:
		panic("unreachable")
	}
	return gameType
}

// PlayerNameRegex is the regex used to validate player names.
const PlayerNameRegex = `^[a-zA-Z0-9_]{1,20}$`

var playerNameRe = regexp.MustCompile(PlayerNameRegex)

// ValidatePlayerName validates the given player name.
func ValidatePlayerName(name string) error {
	if !playerNameRe.MatchString(name) {
		return fmt.Errorf("invalid player name %q does not match %s", name, PlayerNameRegex)
	}
	return nil
}

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

// ConvertJeopardyGameData converts a Jeopardy game data to a Jeopardy game
// info.
func ConvertJeopardyGameData(data JeopardyGameData) JeopardyGameInfo {
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

func (data JeopardyGameData) assertQuestionIx(categoryIx, questionIx int32) error {
	if 0 > categoryIx || categoryIx >= int32(len(data.Categories)) {
		return fmt.Errorf("invalid category index: %d", categoryIx)
	}
	if 0 > questionIx || questionIx >= int32(len(data.Categories[categoryIx].Questions)) {
		return fmt.Errorf("invalid question index: %d", questionIx)
	}
	return nil
}

// QuestionAt returns the question at the given index.
func (data JeopardyGameData) QuestionAt(categoryIx, questionIx int32) (*JeopardyCategory, *JeopardyQuestion, error) {
	if err := data.assertQuestionIx(categoryIx, questionIx); err != nil {
		return nil, nil, err
	}
	return &data.Categories[categoryIx], &data.Categories[categoryIx].Questions[questionIx], nil
}

// QuestionPoints returns the points for the given question.
func (data JeopardyGameData) QuestionPoints(questionIx int32) float32 {
	return *data.ScoreMultiplier * float32(questionIx+1)
}

// TotalQuestions returns the total number of questions in the game.
func (data JeopardyGameData) TotalQuestions() int {
	return len(data.Categories) * len(data.Categories[0].Questions)
}

// Define this here because we're lazy.

// WriteHTTPError writes the given error to the given response writer.
func WriteHTTPError(w http.ResponseWriter, code int, err error) {
	msg := Error{Message: err.Error()}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(msg)
}
