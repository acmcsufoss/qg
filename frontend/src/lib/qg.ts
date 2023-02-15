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
}

const sessionEventTarget = EventTarget as {
  new (): SessionEventTarget;
  prototype: SessionEventTarget;
};

// Session is a qg websocket session.
export class Session extends sessionEventTarget {
  ws: WebSocket;

  private openPromise: Promise<void>;

  constructor(url: string) {
    super();

    const parsed = new URL(url);
    parsed.protocol = parsed.protocol.replace("http", "ws"); // preserve HTTPS
    parsed.pathname = `/api/${APIVersion}/ws`;

    console.log("connecting to", parsed.toString());

    this.ws = new WebSocket(parsed.href);
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

  async connect() {
    await this.openPromise;
  }

  async send(event: qg.Command) {
    this.ws.send(JSON.stringify(event));
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
