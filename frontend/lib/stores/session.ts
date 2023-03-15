import * as store from "svelte/store";
import * as qg from "#/lib/qg.js";

export const event = store.writable<qg.Event | null>();

let internalSession: qg.Session | undefined;

export const session = internalSession;

export async function open() {
  if (!internalSession) {
    internalSession = qg.Session.newLocal();
    internalSession.addEventListener("event", (ev) => event.set(ev.detail));
    // internalSession.addEventListener("close", () => connect());
    return internalSession.open();
  }
}

export async function send(cmd: qg.Command) {
  await open();
  return internalSession!.send(cmd);
}

export async function waitForEvent(): Promise<qg.Event> {
  await open();
  return internalSession!.waitForEvent();
}
