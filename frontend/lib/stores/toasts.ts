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

export const list = store.writable<Toast[]>([]);

export function add(toast: Toast) {
  list.update((list) => {
    list.push(toast);
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

export function remove(toast: Toast) {
  list.update((list) => {
    list = list.filter((t) => t !== toast);
    return list;
  });
}
