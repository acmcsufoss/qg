<script lang="ts">
  import * as svelte from "svelte";
  import * as toasts from "#/lib/stores/toasts.js";

  import { loading, game, name } from "#/lib/stores/state.js";
  import { event } from "#/lib/stores/session.js";
  import { fade } from "svelte/transition";

  import Join from "./index/Join.svelte";
  import Game from "./index/Game.svelte";
  import Loading from "#/lib/components/Loading.svelte";

  function capitalizeFirst(str: string): string {
    return str.charAt(0).toUpperCase() + str.slice(1);
  }

  function errorMessage(error: any): string {
    return capitalizeFirst(`${error}`.replace(/^Error: /, ""));
  }

  svelte.onMount(() => {
    event.subscribe((ev) => {
      if (ev && ev.type == "Error") {
        console.error("server error:", ev.error);
        toasts.add({
          urgency: toasts.Urgency.Error,
          message: `Error: ${ev.error.message}`,
          timeout: 10000,
        });
      }
    });
  });
</script>

<div class="loadable">
  {#await $loading.promise}
    <div class="loading-container" transition:fade={{ duration: 250 }}>
      <Loading />
      {#if $loading.message}
        <p class="loading-message">{$loading.message}</p>
      {/if}
    </div>
  {:catch error}
    <div class="error-container">
      <main>
        <h1>Error!</h1>
        <p>{errorMessage(error)}</p>
      </main>
    </div>
  {/await}

  <div class="content-container">
    {#if $game}
      <Game />
    {:else}
      <Join />
    {/if}
  </div>
</div>

<style>
  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--typography-spacing-vertical);

    width: 100vw;
    height: 100vh;

    position: absolute;
    top: 0;
    z-index: 2;

    background-color: var(--background-color);
  }

  .content-container {
    z-index: 1;
  }

  .error-container {
    height: 100vh;
    display: flex;
    flex-direction: row;
    align-items: center;
    margin: auto;
    background-image: var(--del-background-gradient);

    position: absolute;
    top: 0;
    z-index: 2;
  }

  .error-container main {
    --typography-spacing-vertical: 1em;
    margin: auto;
    border: 2px solid var(--del-color);
    border-left: 0;
    border-right: 0;
    padding: 0 1em;
    max-width: 600px;
  }

  .error-container main > * {
    margin: 1rem 0;
  }
</style>
