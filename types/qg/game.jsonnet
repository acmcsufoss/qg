local schema = import '../lib/schema.jsonnet';
{
  Game: schema.description(
    |||
      Game is the main game object. It contains all the information about the
      game.
    |||,
    schema.unionOf('game', {
      jeopardy: 'Jeopardy',
      kahoot: 'Kahoot',
    })
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
