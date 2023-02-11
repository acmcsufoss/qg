{
  // Apply the given function to each field:value pair of the object to form a new object
  map(func, obj, _mfunc=function(o, f) std.get(o, f, {}), _field=[], _result={})::
    if std.length(_field) == 0 && std.length(std.objectFieldsAll(_result)) == 0 then
      self.map(func, obj, _mfunc, std.objectFieldsAll(obj))
    else
      if std.length(_field) > 0 then
        self.map(func,
                 obj,
                 _mfunc,
                 _field[1:std.length(_field)],
                 local f = func(_field[0], _mfunc(obj, _field[0]));
                 if std.isObject(f) then _result + f else _result)
      else
        _result,

  // Alias for map
  forEach(func, obj):: self.map(func, obj),

  // Returns fields in object `obj` that matches entries specified in the array `fields`
  filterFields(obj, fields)::
    self.forEach(function(f, v) if std.member(fields, f) then if std.objectHas(obj, f)
                 then { [f]: v }
                 else { [f]:: v },
                 obj),

  // Remove fields in object `obj` that matches entries specified in the array `fields`
  removeFields(obj, fields)::
    self.filterFields(obj, std.setDiff(std.objectFieldsAll(obj), fields)),

  // Alias for removeFields
  ignoreFields(obj, fields):: self.removeFields(obj, fields),

  // Removes field from `obj`
  remove(obj, field):: self.removeFields(obj, [field]),

  // Alias for removeFields
  ignore(obj, field):: self.remove(obj, field),

  // Alias for remove
  pop(obj, field):: self.remove(obj, field),

  // Return a copy of object `obj` where all top level fields are converted to hidden fields
  hideFields(obj)::
    self.forEach(function(field, value) { [field]:: value }, obj),

  // Return a copy of object `obj` where all top level fields are converted to visible fields
  showFields(obj)::
    self.forEach(function(field, value) { [field]::: value }, obj),

  // Expands object `obj` with array `array`
  getTraverse(obj, arr)::
    std.foldl(function(prev, this) std.get(prev, this, {}), arr, obj),

  // Flattens an array (`arr`) of object and returns it as an object: `[{x: 0}, {y: 1}]` => `{x: 0, y: 1}`
  flattenObjArray(arr)::
    std.foldl(function(prev, this) prev + this, arr, {}),

  // add all top level fields of the object `obj` together
  addFields(obj)::
    self.flattenObjArray([obj[field] for field in std.objectFieldsAll(obj)]),
}
