<script lang="ts">
  import { slide } from "svelte/transition";
  import { event } from "#/lib/stores/session.js";
  import { loading, game, name } from "#/lib/stores/state.js";
  import * as session from "#/lib/stores/session.js";

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
    if (!ev || !$game) return;

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
        await session.send({ type: "JeopardyPressButton" });
        alreadyPressed = false;
      })(),
    };
  }

  function chooseQuestion(category: number, question: number) {
    $loading = {
      promise: (async () => {
        await session.send({
          type: "JeopardyChooseQuestion",
          category,
          question,
        });
      })(),
    };
  }
</script>

{#if $game && $game.jeopardy}
  {#if state == State.ChoosingQuestion}
    <section id="choosing-question" class="container" transition:slide>
      {#if chooser == $name}
        <h3>
          It's your turn,
          <br />
          <small>Choose a question!</small>
        </h3>
        <!-- TODO: questions table -->
        <!-- TODO: mobile layout, probably a 2-column table and a drop-down -->
        <ul>
          {#each $game.jeopardy?.categories ?? [] as category, c}
            <li>
              <p class="category">{category}</p>
              <ol>
                {#each { length: $game.jeopardy.numQuestions ?? 0 } as _, q}
                  <li>
                    <a
                      href="#answer"
                      role="button"
                      on:click={() => chooseQuestion(c, q)}
                    >
                      {(q + 1) * ($game.jeopardy?.scoreMultiplier ?? 100)}
                    </a>
                  </li>
                {/each}
              </ol>
            </li>
          {/each}
        </ul>
      {:else}
        <h3>
          Waiting for <span class="user">{chooser}</span> to choose a question...
        </h3>
      {/if}
    </section>
  {/if}

  {#if state == State.RacingForAnswer}
    <section id="racing-for-answer" class="container" transition:slide>
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
    <section id="answering" class="container">
      {#if alreadyPressed}
        <h3>You're answering!</h3>
      {:else}
        <h3>Someone's answering...</h3>
      {/if}
      <h4 class="category">{category}</h4>
      <p class="question">{question}</p>
    </section>
  {/if}
{/if}

<style>
  #choosing-question ul,
  #choosing-question ol,
  #choosing-question li {
    padding: 0;
    margin: 0;
  }

  #choosing-question li {
    list-style: none;
  }

  #choosing-question ul > li {
    overflow: hidden;
    border-radius: calc(2 * var(--border-radius));
    margin-bottom: var(--typography-spacing-vertical);
    background-color: var(--form-element-background-color);
  }

  #choosing-question ul ol {
    display: flex;
    justify-content: space-around;
  }

  #choosing-question p.category {
    padding: 0 1rem;
    font-size: 2rem;
  }

  #choosing-question ul ol li {
    flex: 1;
  }

  #choosing-question ul ol li:not(:last-child) {
    border-right: 1px solid var(--form-element-background-color);
  }

  #choosing-question ul ol li a {
    width: 100%;
    padding-left: 0;
    padding-right: 0;
    border-radius: 0;
  }

  #racing-for-answer {
    display: flex;
    flex-direction: column;
  }

  #racing-for-answer .category,
  #racing-for-answer .question {
    font-size: 2rem;
    margin-top: 0;
    margin-bottom: 2rem;
  }

  #racing-for-answer button {
    height: 100%;
    font-size: 2.5rem;
    font-weight: bold;
    margin-top: 1rem;
    border-radius: calc(2 * var(--border-radius));
  }
</style>
