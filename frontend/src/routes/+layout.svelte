<script>
  import * as navigation from "$app/navigation";

  import { fade, fly } from "svelte/transition";
  import { game } from "$lib/stores/state";
  import { event } from "$lib/stores/session";
  import { toasts } from "$lib/stores/toasts";

  event.subscribe((ev) => {
    switch (ev.type) {
      case "Error": {
        // TODO: show a toast or something
        console.error("error from server:", ev.error.message);
        break;
      }
      case "JoinedGame": {
        navigation.goto("/waiting");
        break;
      }
      case "GameStarted": {
        if ($game.jeopardy) navigation.goto(`/jeopardy/${game.id}`);
      }
    }
  });
</script>

<noscript id="noscript-warning">
  <div>
    <h2>Hey!</h2>
    <p>
      It appears you have JavaScript disabled! Unfortunately, this site requires
      JavaScript to function properly. Please enable JavaScript to continue!
    </p>
  </div>
</noscript>

<div class="transition-wrapper" transition:fade>
  <slot />
</div>

{#if $toasts}
  <div class="toast-box">
    {#each $toasts as toast}
      <p class="toast" transition:fly={{ y: 100 }}>
        {toast.message}
      </p>
    {/each}
  </div>
{/if}

<style global>
  @import "normalize.css";
  @import "@picocss/pico/css/pico.min.css";
  @import url("https://fonts.googleapis.com/css2?family=Nunito:wght@300;400;500;600;700;800;900&display=swap");
  @import url("https://fonts.googleapis.com/css2?family=Inconsolata:wght@300;400;500;600;700&display=swap");

  #noscript-warning {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    padding: 1rem;
    z-index: 1000;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background-color: var(--background-color);
  }

  #noscript-warning p {
    max-width: 400px;
    line-height: var(--line-height);
  }

  :root {
    --line-height: 1.35;
  }

  /* https://colordesigner.io/ */
  /* https://colordesigner.io/gradient-generator */

  :root[data-theme="light"],
  :root:not([data-theme="dark"]) {
    --background-color: #e0f4ff;
    --color: #1a1a1a;
    --primary: #13bafb;
    --primary-hover: #66d2fc;
    --primary-focus: #13bafb55;
    --muted-color: #6b9bad;
    --muted-border-color: #56c7f3;
    --form-element-background-color: #bbebfe;
    --form-element-border-color: #a0d9fe;
    --form-element-active-background-color: #aae6fe;
    --form-element-active-border-color: var(--primary);
    --form-element-focus-color: var(--primary);
    --del-color: #fc557a;
    /* TODO: update */
    --del-background-gradient: linear-gradient(45deg, #df6161, #eda4a4);
  }

  @media only screen and (prefers-color-scheme: dark) {
    :root,
    :root:not([data-theme]) {
      --background-color: #013246;
      --color: #fefefe;
      --primary: #13bafb;
      --primary-hover: #66d2fc;
      --primary-focus: #0378a6;
      --muted-color: #86afbf;
      --muted-border-color: #71929f;
      --form-element-background-color: #024863;
      --form-element-border-color: #0378a6;
      --form-element-active-background-color: #026084;
      --form-element-active-border-color: var(--primary);
      --form-element-focus-color: var(--primary);
      --del-color: #fc557a;
      /* TODO: update */
      --del-background-gradient: linear-gradient(-135deg, #4f1010, #280808);
    }
  }

  html {
    background-color: var(--background-color);
    line-height: var(--line-height);
    min-height: 100vh;
  }

  html,
  body {
    /* let body have full height */
    display: flex;
    flex-direction: column;
  }

  body,
  body > div {
    flex: 1;
  }

  body > div {
    display: block; /* reset */
  }

  body {
    --font-family: "Nunito", "Helvetica", "Source Sans Pro", sans-serif;
    --monospace-font-family: "Inconsolata", "Source Code Pro", "Noto Mono",
      monospace;

    margin: 0;
    padding: 0;
    flex: 1;
    font-family: var(--font-family);
  }

  code,
  pre {
    font-family: var(--monospace-font-family);
  }

  span[title] {
    text-decoration: underline dashed;
    cursor: help;
  }

  h1,
  h2,
  h3,
  h4,
  h5,
  h6 {
    margin: var(--typography-spacing-vertical) 0;
  }

  p {
    --typography-spacing-vertical: 1rem;
    margin: var(--typography-spacing-vertical) 0;
    hyphens: auto;
    word-wrap: break-word;
  }

  blockquote {
    padding-top: 0;
    padding-bottom: 0;
  }
</style>
