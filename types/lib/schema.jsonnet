local stdx = import 'stdx.jsonnet';
// https://jsontypedef.com/docs/jtd-in-5-minutes/#discriminator-schemas
{
  // All jtd-defined types.
  empty: $.properties({}),
  all: {},
  boolean: { type: 'boolean' },
  string: { type: 'string' },
  timestamp: { type: 'timestamp' },
  float32: { type: 'float32' },
  float64: { type: 'float64' },
  int8: { type: 'int8' },
  int16: { type: 'int16' },
  int32: { type: 'int32' },
  uint8: { type: 'uint8' },
  uint16: { type: 'uint16' },
  uint32: { type: 'uint32' },
  integer: $.int32,
  int: $.integer,
  float: $.float32,
  number: $.float,
  enum: function(values) { enum: values },
  arrayOf: $.elements,
  elements: function(type) { elements: type },
  properties: function(properties, optionalProperties={}, additionalProperties=false)
    { properties: properties }
    + (if optionalProperties != {}
       then { optionalProperties: optionalProperties }
       else {})
    + (if additionalProperties != false
       then { additionalProperties: true }
       else {}),
  values: function(type) { values: type },
  unionOf(discriminator, types): {
    discriminator: discriminator,
    mapping: std.mapWithKey(
      function(k, v) {
        properties: {
          data: $.ref(v),
        },
      },
      types
    ),
  },
  typeUnion(types): $.unionOf('type', types),
  discriminator: function(discriminator, mapping) {
    discriminator: discriminator,
    mapping: mapping,
  },
  ref(ref): { ref: ref },
  // Extra overrides.
  nullable: function(any) any { nullable: true },
  metadata: function(metadata, any=$.empty)
    any { metadata: metadata },
  description: function(description, any=$.empty)
    self.metadata({ description: description }, any),
}
