<script lang="ts">
  import { event, session } from "#lib/stores/session";
  import { game } from "#lib/stores/state";
  import { fade } from "svelte/transition";

  import Join from "./index/Join.svelte";
  import Waiting from "./index/Waiting.svelte";
  import Jeopardy from "./index/Jeopardy.svelte";

  enum Screen {
    JoinGame,
    Waiting,
    JeopardyGame,
  }

  let screen = Screen.JoinGame;

  $session.addEventListener("close", () => {
    // Kick the user back to the home page if the session closes.
    screen = Screen.JoinGame;
  });

  event.subscribe((ev) => {
    switch (ev.type) {
      case "Error": {
        // TODO: show a toast or something
        console.error("error from server:", ev.error.message);
        break;
      }
      case "JoinedGame": {
        screen = Screen.Waiting;
        break;
      }
      case "GameStarted": {
        if ($game.jeopardy) {
          screen = Screen.JeopardyGame;
        }
      }
    }
  });
</script>

{#if screen == Screen.JoinGame}
  <div transition:fade>
    <Join />
  </div>
{:else if screen == Screen.Waiting}
  <div transition:fade>
    <Waiting />
  </div>
{:else if screen == Screen.JeopardyGame}
  <div transition:fade>
    <Jeopardy />
  </div>
{/if}
