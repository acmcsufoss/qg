local schema = import '../lib/schema.jsonnet';
{
  EventJoinedGame: schema.description(
    |||
      EventJoinedGame is emitted when the current player joins a game. It is a
      reply to CommandJoinGame and is only for the current player. Not to be
      confused with EventPlayerJoinedGame, which is emitted when any player
      joins the current game.
    |||,
    schema.properties({
      gameInfo: schema.ref('GameInfo'),
      gameData: schema.nullable(schema.ref('GameData')),
      isAdmin: schema.boolean,
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
      adminPassword: schema.description(
        'adminPassword is the password of the admin of the game.',
        schema.nullable(schema.string)
      ),
    })
  ),

  CommandBeginGame: schema.description(
    |||
      CommandBeginGame is sent by a client to begin a game.
    |||,
    schema.empty
  ),

  CommandEndGame: schema.description(
    |||
      CommandEndGame is sent by a client to end the current game. The server
      will respond with an EventGameEnded. Only game admins (including the
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
