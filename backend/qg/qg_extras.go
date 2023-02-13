package qg

import (
	"context"
	"fmt"
)

// CommandHandler describes a command handler. Usually, a game manager will
// implement this interface to manage a game.
type CommandHandler interface {
	// HandleCommand handles a command.
	HandleCommand(ctx context.Context, cmd Command) error
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
