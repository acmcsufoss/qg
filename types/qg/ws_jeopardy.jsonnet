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
      chooser: schema.ref('PlayerName'),
      category: schema.string,
      question: schema.string,
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
      alreadyPressed: schema.boolean,
    }),
  ),

  CommandJeopardyChooseQuestion: schema.description(
    |||
      CommandJeopardyChooseQuestion is sent by a player to choose a question.
      The server must do validation to ensure that the player is allowed to
      choose the question.
    |||,
    schema.properties({
      category: schema.string,
      question: schema.string,
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

  CommandJeopardyPlayerIsCorrect: schema.description(
    |||
      CommandJeopardyPlayerIsCorrect is emitted by a game moderator to indicate
      that a player has answered a question correctly. The winning player is
      whoever the last EventJeopardyButtonPressed event indicated. That player
      will instantly receive the points for the question, and the game will let
      them choose the next category and question.
    |||,
    schema.empty,
  ),
}
