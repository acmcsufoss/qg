local schema = import '../lib/schema.jsonnet';
{
  EventJeopardyTurnEnded: schema.description(
    |||
      EventJeopardyTurnEnded is emitted when a turn ends or when the game first
      starts.
    |||,
    schema.properties({
      currentScore: schema.number,
      leaderboard: schema.ref('Leaderboard'),
      isChooser: schema.boolean,
    }),
  ),

  EventJeopardyBeginQuestion: schema.description(
    |||
      EventJeopardyBeginQuestion is emitted when a question begins within this
      Jeopardy game. It is usually emitted once the chooser player has chosen a
      category and value.

      Each category name and question value will map to a category and question
      within the game data. Note that a question may repeat across multiple
      categories.
    |||,
    schema.properties({
      category: schema.string,
      question: schema.string,
    }),
  ),

  CommandJeopardyChooseQuestion: schema.description(
    |||
      CommandJeopardyChooseQuestion is emitted when a player chooses a question
      to answer. The server must do validation to ensure that the player is
      allowed to choose the question.
    |||,
    schema.properties({
      category: schema.string,
      question: schema.string,
    }),
  ),
}
