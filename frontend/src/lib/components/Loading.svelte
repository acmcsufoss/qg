<script lang="ts">
  type LoadingStyle = "pulsating" | "spinning";
  export let style: LoadingStyle;
  export let animationSpeed: "fast" | "slow" = "fast";

  const animationDurations = {
    fast: "1.00s",
    slow: "1.75s",
  };

  const animationDuration = animationDurations[animationSpeed];
</script>

<div
  class="loading-circle-box"
  style="--animation-duration: {animationDuration}"
>
  {#if style == "pulsating"}
    <div class="pulsating-circle-box">
      <div class="pulsating-circle" />
    </div>
  {/if}
  {#if style == "spinning"}
    <div class="spinning-circle" />
  {/if}
</div>

<style>
  .loading-circle-box {
    position: relative;
  }

  /* Pulsating circle is taken from https://codepen.io/peeke/pen/BjxXZa */
  .pulsating-circle-box {
    --glow-size: 128px;
    --circle-size: 48px;
    width: var(--glow-size);
    height: var(--glow-size);
  }

  .pulsating-circle {
    position: absolute;
    width: var(--glow-size);
    height: var(--glow-size);
  }

  .pulsating-circle:before,
  .pulsating-circle:after {
    content: "";
    display: block;
    top: 50%;
    left: 50%;
  }

  .pulsating-circle:before {
    position: relative;
    width: var(--glow-size);
    height: var(--glow-size);
    margin-top: calc(var(--glow-size) / -2);
    margin-left: calc(var(--glow-size) / -2);
    box-sizing: border-box;
    border-radius: 100%;
    background-color: var(--primary);
    animation: pulse-ring var(--animation-duration)
      cubic-bezier(0.215, 0.61, 0.355, 1) infinite;
  }

  .pulsating-circle:after {
    position: absolute;
    width: var(--circle-size);
    height: var(--circle-size);
    margin-top: calc(var(--circle-size) / -2);
    margin-left: calc(var(--circle-size) / -2);
    background-color: var(--color);
    border-radius: 100%;
    box-shadow: 0 0 8px rgba(0, 0, 0, 0.3);
    animation: pulse-dot var(--animation-duration)
      cubic-bezier(0.455, 0.03, 0.515, 0.955) -0.4s infinite;
    animation-delay: calc(var(--animation-duration) / -4);
  }

  @keyframes pulse-ring {
    0% {
      transform: scale(0.35);
      opacity: 1;
    }
    30% {
      opacity: 0.5;
    }
    90%,
    100% {
      opacity: 0;
    }
  }

  @keyframes pulse-dot {
    0% {
      transform: scale(0.65);
    }
    50% {
      transform: scale(1);
    }
    100% {
      transform: scale(0.65);
    }
  }
</style>
