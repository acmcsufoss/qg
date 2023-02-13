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

types {
  Event: schema.typeUnion(dxsonnet.obj.map(
    function(k, v)
      if std.startsWith(k, 'Event')
      then { [stdx.trimPrefix(k, 'Event')]: k },
    types,
  )),
  Command: schema.typeUnion(dxsonnet.obj.map(
    function(k, v)
      if std.startsWith(k, 'Command')
      then { [stdx.trimPrefix(k, 'Command')]: k },
    types,
  )),
}
