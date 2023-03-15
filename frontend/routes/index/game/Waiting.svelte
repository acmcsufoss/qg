<script lang="ts">
  import * as session from "#/lib/stores/session.js";
  import { loading, game } from "#/lib/stores/state.js";

  function startGame() {
    $loading = {
      promise: (async () => {
        await session.send({ type: "BeginGame" });
        await session.waitForEvent();
      })(),
      message: "Starting game...",
    };
  }
</script>

{#if $game}
  <div class="container">
    <div class="top">
      <section id="game-info">
        <h1>
          {#if $game.jeopardy}
            Jeopardy!
          {:else}
            ???
          {/if}
        </h1>
        <h2>
          Game code
          <br />
          <code class="game">{$game.id}</code>
        </h2>
      </section>

      <section id="player-list">
        <h3>You're in the room with {$game.players.length} player(s)</h3>
        <ul>
          {#each $game.players as player}
            <li>
              <span class="player">
                {player.playerName}
              </span>
              {#if $game.isAdmin}
                <!-- TODO: add a button to click the player -->
                <a class="kick" href="#kick" role="button">✖</a>
              {/if}
            </li>
          {/each}
        </ul>
      </section>
    </div>

    {#if $game.isAdmin}
      <section id="admin-panel">
        <button on:click={() => startGame()}>Start Game</button>
      </section>
    {/if}
  </div>
{/if}

<style>
  div.container {
    display: flex;
    flex-direction: column;
  }

  div.container > div.top {
    flex: 1;
  }

  code.game {
    font-size: 3rem;
    color: var(--color);
  }

  #game-info {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
  }

  #game-info > * {
    margin-bottom: 0;
  }

  #game-info h2 {
    text-align: right;
  }

  #player-list * {
    font-size: 1.2rem;
  }

  #player-list h3 {
    margin-bottom: calc(var(--typography-spacing-vertical) / 2);
  }

  #player-list ul {
    padding: 0 1rem;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(10rem, 1fr));
    column-gap: 1rem;
    overflow-y: scroll;
  }

  #player-list ul li {
    list-style: none;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  #player-list ul li span.player {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  a.kick {
    --primary: var(--del-color);
    --primary-focus: var(--del-color-focus);

    color: var(--primary);
    padding: 0;
    line-height: 1.2rem;
    width: 1.2rem;
    height: 1.2rem;
    border-color: transparent;
    background-color: transparent;
  }

  a.kick:hover {
    color: var(--color);
    background-color: var(--primary);
  }

  #admin-panel button {
    margin: 0;
  }

  @media (max-width: 600px) {
    #game-info {
      flex-direction: column;
    }

    #game-info h2 {
      text-align: left;
    }
  }
</style>
