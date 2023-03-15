import type * as qg from "#/lib/qg.js";
import * as store from "svelte/store";
import { session, event } from "#/lib/stores/session.js";

export type PromiseInfo<T> = {
  promise: Promise<T>;
  message?: string;
};

export const loading = store.writable<PromiseInfo<any>>({
  promise: Promise.resolve(),
});

export type GameState = {
  id: string;
  isAdmin: boolean;
  players: qg.LeaderboardEntry[];
  jeopardy?: qg.JeopardyGameInfo;
};

export const name = store.writable<string>("");
export const game = store.writable<GameState | null>();

event.subscribe((ev) => {
  console.log("event", ev);
  game.update((game) => {
    console.log("WS sent event", ev);

    if (!ev) {
      return null;
    }

    if (ev.type == "JoinedGame") {
      switch (ev.gameInfo.type) {
        case "jeopardy": {
          game = {
            id: ev.gameID,
            isAdmin: ev.isAdmin,
            players: [],
            jeopardy: ev.gameInfo.data,
          };
          break;
        }
        default: {
          throw new Error(`unknown game type ${ev.gameInfo.type}`);
        }
      }
    } else {
      if (!game) {
        console.log("ignoring event since they should be after JoinedGame", ev);
        return null;
      }
      switch (ev.type) {
        case "PlayerJoined": {
          game.players.push({
            playerName: ev.playerName,
            score: 0,
          });
          break;
        }
        case "GameEnded": {
          game.jeopardy = undefined;
          break;
        }
      }
    }

    return game;
  });
});
