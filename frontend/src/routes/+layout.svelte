<script>
  import "normalize.css";
  import "@picocss/pico/css/pico.min.css";
</script>

<main>
  <slot />
</main>

<style global>
  main {
    max-width: 800px;
    margin: 0 auto;
    padding: 0 1rem;
  }
  :global(h1),
  :global(h2),
  :global(h3),
  :global(h4),
  :global(h5),
  :global(h6) {
    margin: var(--typography-spacing-vertical) 0;
  }
</style>
