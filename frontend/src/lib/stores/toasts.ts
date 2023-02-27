import * as store from "svelte/store";

export enum Urgency {
  Info,
  Warning,
  Error,
}

export type Toast = {
  urgency: Urgency;
  message: string;
};

export const toasts = store.writable<Toast[]>([]);
