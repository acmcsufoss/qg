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
  <main>
    <header>
      <div class="container">
        <p class="id">
          <span class="user">{$name}</span>@<code class="game">{$game.id}</code>
        </p>
        <p class="game">Jeopardy</p>
      </div>
    </header>

    {#if screen == Screen.Waiting}
      <Waiting />
    {:else if screen == Screen.JeopardyGame}
      <Jeopardy />
    {/if}
  </main>
{/if}

<style>
  main {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  main > :global(:not(header)) {
    flex: 1;
  }

  header {
    background-color: black;
    padding: 0.25rem 0;
  }

  header div.container {
    display: flex;
    justify-content: space-between;
  }

  @media (max-width: 300px) {
    header div.container {
      flex-direction: column;
    }
  }

  header p {
    color: var(--muted-color);
    margin: 0;
    font-size: 0.9rem;
  }

  header .id .user,
  header .id .game {
    color: white;
  }

  code.game {
    padding: 0;
    font-size: inherit;
    background-color: transparent;
  }
</style>
