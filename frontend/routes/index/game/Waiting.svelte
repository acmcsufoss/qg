<script lang="ts">
  import { session } from "#/lib/stores/session.js";
  import { loading, game } from "#/lib/stores/state.js";

  function startGame() {
    $loading = {
      promise: (async () => {
        await $session.send({ type: "BeginGame" });
        await $session.waitForEvent();
      })(),
      message: "Starting game...",
      style: "pulsating",
    };
  }
</script>

{#if $game}
  <main class="container">
    <section id="game-info">
      <h2>
        You're in: Jeopardy!
        <br />
        <small>Game ID: <code>{$game.id}</code></small>
      </h2>
    </section>

    <section id="player-list">
      <h3>You're in the room with:</h3>
      <ul>
        {#each $game.players as player}
          <li>
            {player.playerName}
            {#if $game.isAdmin}
              <!-- TODO: add a button to click the player -->
            {/if}
          </li>
        {/each}
      </ul>
    </section>

    {#if $game.isAdmin}
      <section id="admin-panel">
        <button on:click={() => startGame()}>Start Game</button>
      </section>
    {/if}
  </main>
{/if}
