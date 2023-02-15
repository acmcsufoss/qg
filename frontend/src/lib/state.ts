import * as store from "svelte/store";
import type * as qg from "$lib/qg";

export const get = store.get;

export const session = store.writable<qg.Session>();

export type GameState = {
  info: qg.JeopardyGameInfo | null;
};

export const game = store.writable<GameState>({
  info: null,
});
