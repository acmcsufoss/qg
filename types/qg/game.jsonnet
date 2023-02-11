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
}
