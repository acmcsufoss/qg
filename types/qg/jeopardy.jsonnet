local schema = import '../lib/schema.jsonnet';
{
  Jeopardy: schema.description(
    |||
      JeopardyGame is the game data for a Jeopardy game.
    |||,
    schema.properties(
      {
        categories: schema.arrayOf(schema.ref('JeopardyCategory')),
      },
      optionalProperties={
        moderators: schema.description(
          |||
            moderators enables moderators being able to join.
          |||,
          schema.boolean,
        ),
        require_name: schema.description(
          |||
            require_name, if true, will require members to input a name before
            we can participate.
          |||,
          schema.boolean,
        ),
        score_multiplier: schema.description(
          |||
            score_multiplier is the score multiplier for each question. The
            default is 100.
          |||,
          schema.float32,
        ),
        // score_to_win: schema.description(
        //   |||
        //     score_to_win is the score required to win the game.
        //   |||,
        //   schema.uint32,
        // ),
      },
    ),
  ),
  JeopardyCategory: schema.description(
    |||
      JeopardyCategory is a category in a Jeopardy game.
    |||,
    schema.properties(
      {
        name: schema.description(
          |||
            name is the name of the category.
          |||,
          schema.string,
        ),
        questions: schema.description(
          |||
            questions are the questions in the category.
          |||,
          schema.arrayOf(schema.ref('JeopardyQuestion')),
        ),
      },
    ),
  ),
  JeopardyQuestion: schema.description(
    |||
      JeopardyQuestion is a question in a Jeopardy game.
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
        correct_answer: schema.description(
          |||
            correct_answer is the correct answer within the list of answers
            above. The index starts at 1.
          |||,
          schema.int32,
        ),
      },
    ),
  ),
}
