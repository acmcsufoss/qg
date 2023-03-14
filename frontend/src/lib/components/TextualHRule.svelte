<!-- Code borrowed from https://codepen.io/scottzirkel/pen/yNxNME -->
<script lang="ts">
  export let text: string;
</script>

<hr class="textual-h-rule" data-content={text} />

<style>
  .textual-h-rule {
    position: relative;
    outline: 0;
    border: 0;
    text-align: center;
    height: 1.5em;
    width: 100%;
  }

  .textual-h-rule:before {
    content: "";
    background: var(--muted-color);
    position: absolute;
    left: 0;
    top: 50%;
    width: 100%;
    height: 1px;
  }

  .textual-h-rule:after {
    content: attr(data-content);
    position: relative;
    display: inline-block;
    padding: 0 0.5em;
    line-height: 1.5em;
    background-color: var(--background-color);
  }
</style>
