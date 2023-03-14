<script lang="ts">
  import * as qg from "#lib/qg.js";

  import YAML from "yaml";
  import Brand from "#lib/components/Brand.svelte";

  type SupportedFormat = "yaml" | "json";

  let loading: Promise<any>;
  let busy = false;

  let createdGame: qg.ResponseNewGame | undefined;

  let uploadType: "file" | "paste" = "file";
  let files: FileList;
  let paste = {
    data: "",
    format: "",
  };

  function fileFormat(name: string): SupportedFormat {
    switch (name.split(".").pop()) {
      case "yaml":
      case "yml":
        return "yaml";
      case "json":
        return "json";
      default:
        throw new Error("Unsupported file format");
    }
  }

  function parseGameData(raw: string, format: SupportedFormat): qg.GameData {
    let gameData: qg.GameData;
    switch (format) {
      case "yaml":
        gameData = YAML.parse(raw);
        break;
      case "json":
        gameData = JSON.parse(raw);
        break;
      default:
        throw new Error(`Unknown format: ${format}`);
    }

    qg.Assert("GameData", gameData);
    return gameData;
  }

  async function submitAsync() {
    let data: qg.GameData;
    if (files.length > 0) {
      let file = files[0];
      let text = await file.text();
      data = parseGameData(text, fileFormat(file.name));
    } else {
      data = parseGameData(paste.data, paste.format as "yaml" | "json");
    }

    console.log("will be submitting data", data);
  }

  function submit() {
    loading = (async () => {
      busy = true;
      try {
        await submitAsync();
        busy = false;
      } catch (err) {
        console.error(err);
        throw err;
      }
    })();
  }
</script>

<main>
  <Brand />

  {#if createdGame}
    <h2>Game created!</h2>
    <p>
      Your game code is <code>{createdGame.gameID}</code>. Give it to your
      friends so they can join!
    </p>
  {:else}
    <hgroup>
      <h2>Create Game</h2>

      <formset id="type-form" class="radio">
        <input
          type="radio"
          bind:group={uploadType}
          value="file"
          id="upload-type-file"
        />
        <label for="upload-type-file">Upload a file</label>

        <input
          type="radio"
          bind:group={uploadType}
          value="paste"
          id="upload-type-paste"
        />
        <label for="upload-type-paste">Paste data</label>
      </formset>
    </hgroup>

    <form on:submit|preventDefault={submit}>
      {#await loading catch err}
        <section class="error">
          <h3>Error</h3>
          <p>{err}</p>
        </section>
      {/await}

      {#if uploadType == "file"}
        <formset id="upload-form">
          <input type="file" bind:files id="upload" accept=".yaml,.yml,.json" />
        </formset>
      {/if}

      {#if uploadType == "paste"}
        <formset id="paste-form">
          <label for="paste-format-form">Paste Format</label>
          <formset id="paste-format-form" class="radio">
            <input
              type="radio"
              name="paste-type"
              id="paste-yaml"
              value="yaml"
              bind:group={paste.format}
              required
            />
            <label for="paste-yaml">YAML</label>
            <input
              type="radio"
              name="paste-type"
              id="paste-json"
              value="json"
              bind:group={paste.format}
              required
            />
            <label for="paste-json">JSON</label>
          </formset>

          <textarea
            id="paste-data"
            placeholder="Paste the game file here"
            bind:value={paste.data}
          />
        </formset>
      {/if}
      <button
        type="submit"
        disabled={!busy && !files && !paste.data}
        aria-busy={busy}
      >
        Create
      </button>
    </form>
  {/if}
</main>

<style>
  main {
    margin: 0 auto;
    width: max-content;
  }

  hgroup > * {
    margin: calc(var(--typography-spacing-vertical) / 2) 0;
  }

  form {
    display: flex;
    flex-direction: column;
    max-width: 500px;
    width: 100%;
  }

  formset {
    margin-bottom: var(--typography-spacing-vertical);
  }

  formset:not(.radio) {
    display: flex;
    flex-direction: column;
  }

  formset.radio {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.25em;
    align-items: center;
  }

  formset > *:last-child {
    margin-bottom: 0;
  }

  #paste-form textarea {
    margin: var(--typography-spacing-vertical) 0;
  }
</style>
