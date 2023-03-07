import * as jtd from "jtd";
import type * as qg from "./qg-jtd/index.js";
import schema from "./qg-jtd/schema.js";

export * from "./qg-jtd/index.js";
export { schema };

// Assert asserts that the given value is a valid type within the qg schema.
export function Assert(type: string, value: any) {
  const defn = schema.definitions?.[type];
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

export const APIVersion = "v0";

// SessionEvents contains events emitted by Session.
export interface SessionEvents {
  open: CustomEvent<void>;
  close: CustomEvent<void>;
  event: CustomEvent<qg.Event>;
}

interface SessionEventTarget extends EventTarget {
  addEventListener<K extends keyof SessionEvents>(
    type: K,
    listener: (ev: SessionEvents[K]) => void,
    options?: boolean | AddEventListenerOptions
  ): void;
  addEventListener(
    type: string,
    callback: EventListenerOrEventListenerObject | null,
    options?: EventListenerOptions | boolean
  ): void;
  removeEventListener<K extends keyof SessionEvents>(
    type: K,
    listener: (ev: SessionEvents[K]) => void,
    options?: boolean | EventListenerOptions
  ): void;
  removeEventListener(
    type: string,
    callback: EventListenerOrEventListenerObject | null,
    options?: EventListenerOptions | boolean
  ): void;
}

const sessionEventTarget = EventTarget as {
  new (): SessionEventTarget;
  prototype: SessionEventTarget;
};

// Session is a qg websocket session.
export class Session extends sessionEventTarget {
  // newLocal uses window.location to construct a new Session.
  static newLocal(): Session {
    if (!window) {
      throw new Error("window is undefined");
    }

    const host = window.location.host;
    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    return new Session(`${protocol}://${host}/api/${APIVersion}/ws`);
  }

  ws: WebSocket;

  private openPromise: Promise<void>;

  constructor(url: string) {
    super();
    console.log("connecting to", url);

    this.ws = new WebSocket(url);
    this.ws.addEventListener("open", this.onOpen.bind(this));
    this.ws.addEventListener("message", this.onMessage.bind(this));

    this.openPromise = new Promise((resolve, reject) => {
      this.ws.addEventListener("open", () => {
        resolve();
      });
      this.ws.addEventListener("close", (ev) =>
        reject(new Error(`closed (code ${ev.code}): ${ev.reason}`))
      );
      this.ws.addEventListener("error", (ev) => {
        reject(new Error(`websocket connection error: server unreachable`));
      });
    });
  }

  async open() {
    await this.openPromise;
  }

  async close(graceful: boolean = true) {
    this.ws.close(graceful ? 1000 : 1001);
  }

  async send(event: qg.Command) {
    this.ws.send(JSON.stringify(event));
  }

  // waitForEvent waits for any event of the given types.
  async waitForEvent<
    // Ungodly TypeScript magic. This type is the union of all Event type
    // values.
    T1 extends qg.Event["type"],
    // This one uses the above type value to get the actual qg.Event type.
    T2 extends Extract<qg.Event, { type: T1 }>
  >(types?: T1[], timeout = 0): Promise<T2> {
    return new Promise((resolve, reject) => {
      let timeoutID: ReturnType<typeof setTimeout> | undefined;
      if (timeout > 0) {
        timeoutID = setTimeout(() => {
          this.removeEventListener("event", f);
          reject(new Error("timeout waiting for event"));
        }, timeout);
      }

      const f = (ev: CustomEvent<qg.Event>) => {
        if (!types || types.includes(ev.detail.type as T1)) {
          if (timeoutID) {
            clearInterval(timeoutID);
          }

          this.removeEventListener("event", f);
          resolve(ev.detail as T2);
        }
      };
    });
  }

  private async onOpen() {
    this.dispatchEvent(new CustomEvent("open"));
  }

  private async onClose() {
    this.dispatchEvent(new CustomEvent("close"));
  }

  private async onMessage(event: MessageEvent) {
    const data = JSON.parse(event.data) as qg.Event;
    Assert("Event", data);

    this.dispatchEvent(new CustomEvent("event", { detail: data }));
  }
}
