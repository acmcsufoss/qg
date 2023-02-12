import * as qg from "./qg-jtd/index";
import * as schema from "./qg-jtd/schema.js";
import * as jtd from "jtd";

export * from "./qg-jtd/index";
export const Schema = schema as jtd.Schema;

export function Assert(type: string, value: any) {
  console.log(qg);
}
