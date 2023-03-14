<script lang="ts">
  import { game } from "#/lib/stores/state";
  import { session } from "#/lib/stores/session";

  import Loadable from "#/lib/components/Loadable.svelte";

  let promise: Promise<any>;

  function startGame() {
    promise = (async () => {
      await $session.send({ type: "BeginGame" });
      await $session.waitForEvent(["GameStarted"]);
    })();
  }
</script>

<Loadable {promise}>
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
</Loadable>
