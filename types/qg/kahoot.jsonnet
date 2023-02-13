local schema = import '../lib/schema.jsonnet';
{
  KahootGameData: schema.description(
    |||
      KahootGameData is the game data for a Kahoot game.
    |||,
    schema.properties(
      {
        time_limit: schema.description(
          |||
            time_limit is the time limit for each question. The format is in
            Go's time.Duration, e.g. 10s for 10 seconds.
          |||,
          schema.string,
        ),
        questions: schema.description(
          |||
            questions are the questions in the game.
          |||,
          schema.arrayOf(schema.ref('KahootQuestion')),
        ),
      },
    ),
  ),

  KahootQuestion: schema.description(
    |||
      KahootQuestion is a question in a Kahoot game.
    |||,
    schema.properties(
      {
        question: schema.description(
          |||
            question is the question.
          |||,
          schema.string,
        ),
        answers: schema.description(
          |||
            answers are the possible answers.
          |||,
          schema.arrayOf(schema.string),
        ),
      },
    ),
  ),
}
