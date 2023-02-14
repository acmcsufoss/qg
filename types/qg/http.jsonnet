local schema = import '../lib/schema.jsonnet';
{
  RequestNewGame: schema.properties({
    data: schema.ref('GameData'),
    moderator_password: schema.string,
  }),
  ResponseNewGame: schema.properties({
    gameID: schema.string,
    gameType: schema.ref('GameType'),
  }),

  RequestGetGame: schema.properties({
    gameID: schema.string,
  }),
  ResponseGetGame: schema.properties({
    gameType: schema.ref('GameType'),
  }),

  RequestGetJeopardyGame: schema.properties({
    gameID: schema.string,
  }),
  ResponseGetJeopardyGame: schema.properties({
    info: schema.ref('JeopardyGameInfo'),
  }),
}
