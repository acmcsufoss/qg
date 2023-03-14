<script lang="ts">
  import * as svelte from "svelte";
  import * as store from "svelte/store";

  import { slide } from "svelte/transition";
  import { name } from "#/lib/stores/state.js";
  import { session } from "#/lib/stores/session.js";

  import Brand from "#/lib/components/Brand.svelte";
  import Loadable from "#/lib/components/Loadable.svelte";
  import TextualHRule from "#/lib/components/TextualHRule.svelte";

  const gamecodeRegex = `^[a-z0-9]*$`;
  const usernameRegex = `^[a-zA-Z0-9_]{1,20}$`;

  let promise: Promise<any>;
  let gameID = "";
  let isAdmin = false;
  let adminPassword = "";

  svelte.onMount(async () => {
    promise = $session.open();
  });

  function submit() {
    promise = (async () => {
      await store.get(session).send({
        type: "JoinGame",
        gameID: gameID,
        playerName: $name,
        adminPassword: isAdmin ? adminPassword : null,
      });

      await store.get(session).waitForEvent(["JoinedGame", "Error"]);
    })();
  }
</script>

<Loadable
  {promise}
  message="Connecting to the game server..."
  style="pulsating"
>
  <main>
    <Brand />

    <form on:submit|preventDefault={submit} class={isAdmin ? "isAdmin" : ""}>
      <formset id="gamecode-form">
        <label for="gamecode">
          Enter your
          <span title="The game code that's displayed on the big screen.">
            game code</span
          >:
        </label>
        <input
          type="text"
          id="gamecode"
          placeholder="XXXX"
          pattern={gamecodeRegex}
          minlength="4"
          maxlength="4"
          required
          autocomplete="off"
          bind:value={gameID}
        />
      </formset>

      <formset id="username-form">
        <label for="gamecode">
          Enter your <span title="Your display name.">name</span>:
        </label>
        <input
          type="text"
          id="username"
          placeholder="Player 1"
          pattern={usernameRegex}
          minlength="1"
          maxlength="20"
          required
          bind:value={$name}
        />
      </formset>

      {#if isAdmin}
        <formset id="admin-password-form" transition:slide={{ duration: 200 }}>
          <label for="admin-password">Enter your admin password:</label>
          <input
            type="password"
            id="admin-password"
            required={isAdmin}
            bind:value={adminPassword}
          />
        </formset>
      {/if}

      <formset id="submit-form" class="last-in-form">
        <input type="submit" id="submit" value="Join" />
      </formset>

      {#if isAdmin}
        <formset
          id="create-new-form"
          class="last-in-form"
          transition:slide={{ duration: 200 }}
        >
          <TextualHRule text="or" />
          <a role="button" class="secondary" id="create-new" href="/create">
            Create New
          </a>
        </formset>
      {/if}

      <formset id="is-admin-form">
        <input type="checkbox" id="is-admin" bind:checked={isAdmin} />
        <label for="is-admin">I'm an admin/organizer</label>
      </formset>
    </form>

    <footer>
      <p>
        Source code:
        <a
          href="https://github.com/acmcsufoss/qg"
          target="_blank"
          rel="noreferrer"
        >
          acmcsufoss/qg
        </a>
        <br />
        Licensed under the
        <a
          href="https://opensource.org/license/mit/"
          target="_blank"
          rel="noreferrer"
        >
          MIT
        </a>
        license.
      </p>
    </footer>
  </main>
</Loadable>

<style>
  main {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    height: 100vh;
    position: relative;
  }

  main :global(#brand) {
    position: absolute;
  }

  footer {
    width: 100%;
    max-width: 600px;
    margin: auto;
    padding: 0 var(--spacing);
    border-top: 1px solid var(--muted-color);
  }

  form {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: stretch;
    margin: 0 auto;
    flex: 1;
    position: relative;
  }

  form.isAdmin {
    /* Hack: the form with isAdmin is too long. */
    min-height: 650px;
  }

  form input[type="submit"],
  form a[role="button"] {
    width: 100%;
  }

  form #is-admin-form {
    position: absolute;
    bottom: 0;
  }

  .last-in-form * {
    margin: 0;
  }

  #is-admin-form {
    display: inline-flex;
    align-items: center;
    bottom: 0;
    margin: var(--typography-spacing-vertical) 0;
  }

  #gamecode {
    font-family: var(--monospace-font-family);
    font-size: 1.2rem;
    font-weight: 500;
    flex: 1;
    min-width: 8ch;
  }

  #username {
    font-size: 1.2rem;
  }
</style>
