package qg

import "fmt"

func (data JeopardyGameData) assertQuestionIx(categoryIx, questionIx int) error {
	if 0 > categoryIx || categoryIx >= len(data.Categories) {
		return fmt.Errorf("invalid category index: %d", categoryIx)
	}
	if 0 > questionIx || questionIx >= len(data.Categories[categoryIx].Questions) {
		return fmt.Errorf("invalid question index: %d", questionIx)
	}
	return nil
}

// QuestionAt returns the question at the given index.
func (data JeopardyGameData) QuestionAt(categoryIx, questionIx int) (*JeopardyCategory, *JeopardyQuestion, error) {
	if err := data.assertQuestionIx(categoryIx, questionIx); err != nil {
		return nil, nil, err
	}
	return &data.Categories[categoryIx], &data.Categories[categoryIx].Questions[questionIx], nil
}

// QuestionPoints returns the points for the given question.
func (data JeopardyGameData) QuestionPoints(questionIx int) float64 {
	return *data.ScoreMultiplier * float64(questionIx+1)
}
