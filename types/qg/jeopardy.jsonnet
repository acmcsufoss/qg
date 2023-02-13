local schema = import '../lib/schema.jsonnet';
{
  JeopardyGameData: schema.description(
    |||
      JeopardyGameData is the game data for a Jeopardy game.
    |||,
    schema.properties(
      {
        categories: schema.arrayOf(schema.ref('JeopardyCategory')),
        moderator_password: schema.string,
      },
      optionalProperties={
        score_multiplier: schema.description(
          |||
            score_multiplier is the score multiplier for each question. The
            default is 100.
          |||,
          schema.float,
        ),
        // moderators: schema.description(
        //   |||
        //     moderators enables moderators being able to join.
        //   |||,
        //   schema.boolean,
        // ),
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
      },
    ),
  ),

  JeopardyGameInfo: schema.description(
    |||
      JeopardyGameInfo is the initial information for a Jeopardy game. This type
      contains no useful information about the entire game data, so it's used to
      send to players the first time they join.
    |||,
    schema.properties({
      categories: schema.arrayOf(schema.string),
      numQuestions: schema.integer,
      scoreMultiplier: schema.float,
    }),
  ),
}
