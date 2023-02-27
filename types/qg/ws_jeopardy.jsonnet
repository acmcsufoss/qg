local schema = import '../lib/schema.jsonnet';
{
  EventJeopardyTurnEnded: schema.description(
    |||
      EventJeopardyTurnEnded is emitted when a turn ends or when the game first
      starts.
    |||,
    schema.properties({
      leaderboard: schema.ref('Leaderboard'),
      chooser: schema.ref('PlayerName'),
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
      chooser: schema.ref('PlayerName'),
      category: schema.int32,
      question: schema.string,
      points: schema.float,
    }),
  ),

  EventJeopardyButtonPressed: schema.description(
    |||
      EventJeopardyButtonPressed is emitted when any player had pressed a button
      on their device, voiding other players' buttons. This event is only
      emitted when the game is in the "question" state.
    |||,
    schema.properties({
      playerName: schema.ref('PlayerName'),
    }),
  ),

  EventJeopardyResumeButton: schema.description(
    |||
      EventJeopardyResumeButton is emitted when the player can now continue to
      press the button whenever they are ready to answer the question. This
      could happen if the other player who pressed the button first got the
      question wrong.

      Note that if alreadyPressed is true, then the player has already pressed
      the button, so they cannot press it again.
    |||,
    schema.properties({
      alreadyAnsweredPlayers: schema.arrayOf(schema.ref('PlayerName')),
    }),
  ),

  CommandJeopardyChooseQuestion: schema.description(
    |||
      CommandJeopardyChooseQuestion is sent by a player to choose a question.
      The server must do validation to ensure that the player is allowed to
      choose the question.
    |||,
    schema.properties({
      category: schema.int,
      question: schema.int,
    }),
  ),

  CommandJeopardyPressButton: schema.description(
    |||
      CommandJeopardyPressButton is emitted when a player presses the button
      during a question. It is only valid to emit this command when the game is
      in the question state.
    |||,
    schema.empty,
  ),

  CommandJeopardyPlayerJudgment: schema.description(
    |||
      CommandJeopardyPlayerJudgment is emitted by a game admin to indicate
      whether a player has answered a question correctly. The winning player is
      whoever the last EventJeopardyButtonPressed event indicated. That player
      will instantly receive the points for the question, and the game will let
      them choose the next category and question. If the player answered wrong,
      then the game will let others press the button.
    |||,
    schema.properties({
      correct: schema.boolean,
    }),
  ),
}
