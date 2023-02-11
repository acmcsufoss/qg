local schema = import '../lib/schema.jsonnet';
{
  Error: schema.description(
    |||
      Error is returned on every API error.
    |||,
    schema.properties({
      message: schema.description(
        'Message is the error message',
        schema.string
      ),
    }),
  ),
}
