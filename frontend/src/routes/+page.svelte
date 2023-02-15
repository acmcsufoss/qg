<script lang="ts">
  import * as qg from "$lib/qg";
  import * as state from "$lib/state";
  import * as svelte from "svelte";
  import { slide } from "svelte/transition";
  import Loadable from "$lib/components/Loadable.svelte";
  import TextualHRule from "$lib/components/TextualHRule.svelte";

  const gamecodeRegex = `^[a-z0-9]*$`;
  const usernameRegex = `^[a-zA-Z0-9_]{1,20}$`;

  let ready = false;
  let error: unknown;

  let gameID = "";
  let username = "";
  let isModerator = false;
  let moderatorPassword = "";

  svelte.onMount(async () => {
    error = undefined;
    try {
      const session = new qg.Session(window.location.href);
      state.session.set(session);
      await session.connect();
      ready = true;
    } catch (err) {
      /* error = err; */
    }
  });

  function submit() {
    const session = state.get(state.session);
    session.send({
      type: "JoinGame",
      gameID: gameID,
      playerName: username,
      moderatorPassword: isModerator ? moderatorPassword : null,
    });
  }
</script>

<Loadable
  loading={!ready}
  loadingMessage="Connecting to the game server..."
  style="pulsating"
  {error}
>
  <main>
    <h1 id="brand"><span>q</span>uiz<span>g</span>ame</h1>

    <form on:submit|preventDefault={submit}>
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
          bind:value={username}
        />
      </formset>

      {#if isModerator}
        <formset
          id="moderator-password-form"
          transition:slide={{ duration: 200 }}
        >
          <label for="moderator-password">Enter your moderator password:</label>
          <input
            type="password"
            id="moderator-password"
            required={isModerator}
            bind:value={moderatorPassword}
          />
        </formset>
      {/if}

      <formset id="submit-form" class="last-in-form">
        <input type="submit" id="submit" value="Join" />
      </formset>

      {#if isModerator}
        <formset
          id="create-new-form"
          class="last-in-form"
          transition:slide={{ duration: 200 }}
        >
          <TextualHRule text="or" />
          <input
            type="submit"
            class="secondary"
            id="create-new"
            value="Create New"
          />
        </formset>
      {/if}

      <formset id="is-moderator-form">
        <input type="checkbox" id="is-moderator" bind:checked={isModerator} />
        <label for="is-moderator">I'm a moderator/organizer</label>
      </formset>
    </form>
  </main>
</Loadable>

<style>
  main {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    height: 100vh;
    min-height: 600px;
    position: relative;
  }

  form {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: stretch;
    height: 100%;
    max-width: 600px;
    margin: 0 auto;
    position: relative;
  }

  .last-in-form input {
    margin: 0;
  }

  #brand {
    position: absolute;
    align-self: center;
    top: 0;
    text-align: center;
    font-size: 3rem;
    font-weight: lighter;
  }

  #brand span {
    font-weight: bold;
  }

  #is-moderator-form {
    display: inline-flex;
    align-items: center;
    position: absolute;
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

  #gamecode + input {
    flex: 0.15;
  }

  #username {
    font-size: 1.2rem;
  }
</style>
