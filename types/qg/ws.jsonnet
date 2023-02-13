local dxsonnet = import '../lib/dxsonnet/dxsonnet.jsonnet';
local schema = import '../lib/schema.jsonnet';
local stdx = import '../lib/stdx.jsonnet';

local types =
  {}
  + (import './ws_types.jsonnet')
  + (import './ws_jeopardy.jsonnet')
  + (import './ws_kahoot.jsonnet');

local events = std.filter(
  function(k)
    std.startsWith(k, 'Event'),
  std.objectFields(types),
);

{
  Event: schema.discriminator('type', dxsonnet.obj.map(
    function(k, v)
      if std.startsWith(k, 'Event')
      then { [stdx.trimPrefix(k, 'Event')]: v },
    types,
  )),
  Command: schema.discriminator('type', dxsonnet.obj.map(
    function(k, v)
      if std.startsWith(k, 'Command')
      then { [stdx.trimPrefix(k, 'Command')]: v },
    types,
  )),
}
