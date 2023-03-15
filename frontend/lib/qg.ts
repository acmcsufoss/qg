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
  if (errors.length > 0) {
    const error = errors[0];
    let path = "$";
    if (error.instancePath.length > 0) {
      path += "." + error.instancePath.join(".");
    }
    throw new Error(`error: ${errors}`);
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

  private ws: WebSocket | null;
  private url: string;
  private openPromise: Promise<void> = Promise.resolve();

  constructor(url: string) {
    super();
    console.log("connecting to", url);

    this.ws = null;
    this.url = url;
  }

  async open() {
    this.init();
    await this.openPromise;
  }

  async close(graceful: boolean = true) {
    if (this.ws) this.ws.close(graceful ? 1000 : 1001);
  }

  async send(event: qg.Command) {
    if (!this.ws) throw new Error("websocket is closed");
    this.ws.send(JSON.stringify(event));
  }

  async waitForEvent(): Promise<qg.Event> {
    return new Promise((resolve, reject) => {
      const onEvent = (ev: CustomEvent<qg.Event> | null) => {
        this.removeEventListener("event", onEvent);
        this.removeEventListener("close", onClose);

        if (!ev) {
          reject(new Error("session closed"));
          return;
        }

        resolve(ev.detail);
      };

      const onClose = () => onEvent(null);

      this.addEventListener("event", onEvent);
      this.addEventListener("close", onClose);
    });
  }

  private init() {
    if (this.ws) return;

    console.log("connecting to websocket...");

    const ws = new WebSocket(this.url);
    this.ws = ws;
    this.ws.addEventListener("open", this.onOpen.bind(this));
    this.ws.addEventListener("close", this.onClose.bind(this));
    this.ws.addEventListener("message", this.onMessage.bind(this));

    this.openPromise = new Promise((resolve, reject) => {
      if (!ws) {
        throw "ws should not be null";
      }
      ws.addEventListener("open", () => {
        resolve();
      });
      ws.addEventListener("close", (ev) =>
        reject(new Error(`closed (code ${ev.code}): ${ev.reason}`))
      );
      ws.addEventListener("error", (ev) => {
        reject(new Error(`websocket connection error: server unreachable`));
      });
    });
  }

  private async onOpen() {
    this.dispatchEvent(new CustomEvent("open"));
  }

  private async onClose() {
    this.dispatchEvent(new CustomEvent("close"));
    this.ws = null;
    // this.init(); // immediately reconnect
  }

  private async onMessage(event: MessageEvent) {
    const data = JSON.parse(event.data) as qg.Event;
    Assert("Event", data);

    this.dispatchEvent(new CustomEvent("event", { detail: data }));
  }
}
