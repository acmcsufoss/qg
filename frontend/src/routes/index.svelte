<script lang="ts">
  import * as routing from "svelte-routing";

  import { event } from "#lib/stores/session.ts";
  import { game } from "#lib/stores/state.ts";
  import { fade } from "svelte/transition";

  import Join from "#src/routes/index/Join.svelte";
  import Waiting from "#src/routes/index/Waiting.svelte";
  import Jeopardy from "#src/routes/index/Jeopardy.svelte";

  enum Screen {
    JoinGame,
    Waiting,
    JeopardyGame,
  }

  let screen = Screen.JoinGame;

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
        if ($game.jeopardy) screen = Screen.JeopardyGame;
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
