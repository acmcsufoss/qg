import * as store from "svelte/store";

export enum Urgency {
  Info,
  Warning,
  Error,
}

export type Toast = {
  urgency: Urgency;
  message: string;
  timeout: number;
};

export const toasts = store.writable<Toast[]>([]);

export function addToast(toast: Toast) {
  toasts.update((toasts) => {
    toasts.push(toast);
    return toasts;
  });

  setTimeout(() => {
    toasts.update((toasts) => {
      toasts.shift();
      return toasts;
    });
  }, toast.timeout);
}
