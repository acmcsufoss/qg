<script lang="ts">
  import { event } from "#/lib/stores/session.js";
  import { game, name } from "#/lib/stores/state.js";

  import Waiting from "./game/Waiting.svelte";
  import Jeopardy from "./game/Jeopardy.svelte";

  enum Screen {
    Waiting,
    JeopardyGame,
  }

  let screen = Screen.Waiting;

  event.subscribe((ev) => {
    if (!ev || !$game) return;

    switch (ev.type) {
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

{#if $game}
  <header>
    <p>
      <span class="user">{$name}</span>@<span class="game">{$game.id}</span>
    </p>
    <p>Jeopardy</p>
  </header>

  {#if screen == Screen.Waiting}
    <main>
      <Waiting />
    </main>
  {:else if screen == Screen.JeopardyGame}
    <main>
      <Jeopardy />
    </main>
  {/if}
{/if}
