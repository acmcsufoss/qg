<script lang="ts">
  import { slide } from "svelte/transition";
  import { session, event } from "#/lib/stores/session.js";
  import { loading, game, name } from "#/lib/stores/state.js";

  enum State {
    ChoosingQuestion,
    RacingForAnswer,
    Answering,
    /* RightAnswer, */
    /* WrongAnswer, */
  }

  let state = State.ChoosingQuestion;
  let chooser: string = "<chooser>";
  let category: string = "<category>"; // for testing
  let question: string = "<question>";
  let alreadyPressed = false;

  event.subscribe((ev) => {
    switch (ev.type) {
      case "JeopardyBeginQuestion":
        state = State.RacingForAnswer;
        category = $game.jeopardy?.categories[ev.category] ?? "";
        question = ev.question;
        alreadyPressed = false;
        break;
      case "JeopardyResumeButton":
        // TODO: at this stage, the user got the wrong answer if alreadyPressed
        // is true. We should show that screen.
        state = State.RacingForAnswer;
        break;
      case "JeopardyButtonPressed":
        state = State.Answering;
        break;
      case "JeopardyTurnEnded":
        state = State.ChoosingQuestion;
        chooser = ev.chooser;
        break;
    }
  });

  function answer() {
    alreadyPressed = true;
    $loading = {
      promise: (async () => {
        await $session.send({ type: "JeopardyPressButton" });
        alreadyPressed = false;
      })(),
      message: "",
      style: "pulsating",
    };
  }
</script>

{#if state == State.ChoosingQuestion}
  <section id="choosing-question" transition:slide>
    {#if chooser == $name}
      <h3>
        It's your turn
        <br />
        <small>Choose a question!</small>
      </h3>
      <!-- TODO: questions table -->
      <!-- TODO: mobile layout, probably a 2-column table and a drop-down -->
      <table>
        <thead>
          <tr>
            {#each $game.jeopardy?.categories ?? [] as category}
              <th>{category}</th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each { length: $game.jeopardy?.numQuestions ?? 0 } as _, i}
            <tr>
              {#each $game.jeopardy?.categories ?? [] as _}
                <td>{i * ($game.jeopardy?.scoreMultiplier ?? 0)}</td>
              {/each}
            </tr>
          {/each}
        </tbody>
      </table>
    {:else}
      <h3>
        Waiting for <span class="user">{chooser}</span> to choose a question...
      </h3>
    {/if}
  </section>
{/if}

{#if state == State.RacingForAnswer}
  <section id="racing-for-answer" transition:slide>
    <h3>Question:</h3>
    <h4 class="category">{category}</h4>
    <p class="question">{question}</p>
    <button on:click={() => answer()} disabled={alreadyPressed}>
      {#if alreadyPressed}
        You already answered :(
      {:else}
        Answer!
      {/if}
    </button>
  </section>
{/if}

{#if state == State.Answering}
  <!-- Intentionally use no transition. -->
  <section id="answering">
    {#if alreadyPressed}
      <h3>You're answering!</h3>
    {:else}
      <h3>Someone's answering...</h3>
    {/if}
    <h4 class="category">{category}</h4>
    <p class="question">{question}</p>
  </section>
{/if}
