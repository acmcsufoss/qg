<script lang="ts">
  import Loading from "#/lib/components/Loading.svelte";
  import { fade } from "svelte/transition";

  export let style: "pulsating" | "spinning" = "pulsating";

  export let promise: Promise<any> | undefined;
  export let message = "";
  export let errorTitle = "Error!";

  $: promise_ = promise !== undefined ? promise : Promise.resolve();

  function capitalizeFirst(str: string): string {
    return str.charAt(0).toUpperCase() + str.slice(1);
  }

  function errorMessage(error: any): string {
    return capitalizeFirst(`${error}`.replace(/^Error: /, ""));
  }
</script>

{#await promise_}
  <div class="loading-container" transition:fade>
    <Loading {style} />
    {#if message}
      <p class="loading-message">{message}</p>
    {/if}
  </div>
{:then}
  <div class="content-container" transition:fade>
    <slot />
  </div>
{:catch error}
  <div class="error-container">
    <main>
      {#if errorTitle}
        <h1>{errorTitle}</h1>
      {/if}
      <p>{errorMessage(error)}</p>
    </main>
  </div>
{/await}

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
  }

  .error-container {
    height: 100vh;
    display: flex;
    flex-direction: row;
    align-items: center;
    margin: auto;
    background-image: var(--del-background-gradient);
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
