import * as jtd from "jtd";
import * as qg from "./qg-jtd/index";
import schema from "./qg-jtd/schema.js";

export * from "./qg-jtd/index";
export { schema };

// Assert asserts that the given value is a valid type within the qg schema.
export function Assert(type: string, value: any) {
  const defn = schema.definitions[type];
  if (!defn) {
    throw new Error(`unknown type ${type}`);
  }
  defn.definitions = schema.definitions;

  const errors = jtd.validate(defn, value, {
    maxErrors: 1,
    maxDepth: 200,
  });
  if (errors) {
    const error = errors[0];
    let path = "$";
    if (error.instancePath.length > 0) {
      path += "." + error.instancePath.join(".");
    }
    throw new Error(`error at ${path}`);
  }
}
