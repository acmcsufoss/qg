local schema = import 'lib/schema.jsonnet';
{
  definitions:
    {}
    + (import './qg/kahoot.jsonnet')
    + (import './qg/error.jsonnet')
    + (import './qg/jeopardy.jsonnet')
    + (import './qg/game.jsonnet')
    + (import './qg/http.jsonnet')
    + (import './qg/ws.jsonnet'),
}
