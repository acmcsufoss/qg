local dxsonnet = import 'dxsonnet/dxsonnet.jsonnet';
local filterHidden = function(hidden)
  function(obj)
    dxsonnet.obj.map(
      function(k, v)
        if hidden
        then
          if (std.objectHas(obj, k))
          then { [k]::: v }
          else {}
        else
          if (std.objectHasAll(obj, k) && !std.objectHas(obj, k))
          then { [k]::: v }
          else {}
      ,
      obj,
    );
{
  onlyHidden: filterHidden(true),
  onlyVisible: filterHidden(false),
}
