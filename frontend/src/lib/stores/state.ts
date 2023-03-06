import type * as qg from "$lib/qg";
import * as store from "svelte/store";
import { event } from "$lib/stores/session";

export type GameState = {
  id: string;
  isAdmin: boolean;
  players: qg.LeaderboardEntry[];
  username: string;
  jeopardy?: qg.JeopardyGameInfo;
};

export const name = store.writable<string>("");
export const game = store.writable<GameState>();

event.subscribe((ev) => {
  console.log("event", ev);
  game.update((game) => {
    switch (ev.type) {
      case "JoinedGame": {
        switch (ev.gameInfo.type) {
          case "jeopardy": {
            game.id = ev.gameID;
            game.isAdmin = ev.isAdmin;
            game.players = [];
            game.jeopardy = ev.gameInfo.data;
            break;
          }
          default: {
            throw new Error(`unknown game type ${ev.gameInfo.type}`);
          }
        }
        break;
      }
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

    return game;
  });
});
