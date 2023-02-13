local schema = import '../lib/schema.jsonnet';
{
  GameData: schema.description(
    |||
      GameData is the game data. It contains all the information about the game.
    |||,
    schema.unionOf('game', {
      jeopardy: 'JeopardyGameData',
      kahoot: 'KahootGameData',
    })
  ),

  GameType: schema.enum([
    'jeopardy',
    'kahoot',
  ]),

  GameID: schema.description(
    |||
      GameID is the unique identifier for a game. Each player must type this
      code to join the game.
    |||,
    schema.string,
  ),

  PlayerName: schema.description(
    |||
      PlayerName is the name of a player.
    |||,
    schema.string
  ),

  Leaderboard: schema.description(
    |||
      Leaderboard is a list of players and their scores.
    |||,
    schema.arrayOf(schema.ref('LeaderboardEntry'))
  ),

  LeaderboardEntry: schema.properties({
    playerName: schema.string,
    score: schema.int,
  }),
}
