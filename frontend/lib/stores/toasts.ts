import * as store from "svelte/store";

export enum Urgency {
  Info = "info",
  Warning = "warning",
  Error = "error",
}

export type Toast = {
  urgency: Urgency;
  message: string;
  timeout?: number;
};

export const list = store.writable<(Toast & { id: string })[]>([]);

export function add(toast: Toast) {
  list.update((list) => {
    list.push({ ...toast, id: `${Date.now()}-${toast.message}` });
    return list;
  });

  if (toast.timeout) {
    setTimeout(() => {
      list.update((list) => {
        list.shift();
        return list;
      });
    }, toast.timeout);
  }
}

export function remove(toast: { id: string }) {
  list.update((list) => {
    list = list.filter((t) => t.id !== toast.id);
    return list;
  });
}
