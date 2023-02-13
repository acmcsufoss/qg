local schema = import '../lib/schema.jsonnet';
{
  EventJoinedGame: schema.description(
    |||
      EventJoinedGame is emitted when the current player joins a game. It is a
      reply to CommandJoinGame and is only for the current player. Not to be
      confused with EventPlayerJoinedGame, which is emitted when any player
      joins the current game.
    |||,
    schema.description({
      game: schema.typeUnion({
        jeopardy: 'JeopardyGameInfo',
        // kahoot: 'KahootGameInfo',
      }),
      isModerator: schema.boolean,
      gameData: schema.nullable(schema.ref('Game')),
    }),
  ),

  EventPlayerJoined: schema.description(
    |||
      EventPlayerJoined is emitted when a player joins the current game.
    |||,
    schema.properties({
      playerName: schema.ref('PlayerName'),
    }),
  ),

  EventGameEnded: schema.description(
    |||
      EventGameEnded is emitted when the current game ends.
    |||,
    schema.properties({
      leaderboard: schema.ref('Leaderboard'),
    }),
  ),

  CommandJoinGame: schema.description(
    |||
      CommandJoinGame is sent by a client to join a game. The client (or the
      user) supplies a game ID and a player name. The server will respond with
      an EventJoinedGame.
    |||,
    schema.properties({
      gameID: schema.description(
        'gameID is the ID of the game to join.',
        schema.ref('GameID')
      ),
      playerName: schema.description(
        'playerName is the wanted name of the user.',
        schema.ref('PlayerName')
      ),
      moderatorPassword: schema.description(
        'moderatorPassword is the password of the moderator of the game.',
        schema.nullable(schema.string)
      ),
    })
  ),

  CommandEndGame: schema.description(
    |||
      CommandEndGame is sent by a client to end the current game. The server
      will respond with an EventGameEnded. Only game moderators (including the
      host) can end the game.
    |||,
    schema.properties({
      declareWinner: schema.description(
        |||
          declareWinner determines whether the game should be ended with a
          winner or not. If true, the game will be ended with a winner. If
          false, the game will be ended abruptly.
        |||,
        schema.boolean
      ),
    })
  ),
}
