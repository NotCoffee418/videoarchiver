<!-- Expander.svelte -->
<script>
  let { 
    /** @type {boolean} */
    expanded = false,
    /** @type {string} */
    title = "Settings",
    children
  } = $props();

  function toggle() {
    expanded = !expanded;
  }
</script>

<div class="expander">
  <button class="expander-header" onclick={toggle}>
    <span class="expander-icon" class:expanded>{expanded ? '▼' : '▶'}</span>
    <span>{title}</span>
  </button>
  
  {#if expanded}
    <div class="expander-content">
      {@render children()}
    </div>
  {/if}
</div>

<style>
  .expander {
    width: 100%;
  }

  .expander-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
    background: none;
    border: none;
    color: inherit;
    font-size: 1rem;
    font-weight: bold;
    cursor: pointer;
    padding: 0.5rem 0;
    text-align: left;
    transition: color 0.2s ease;
  }

  .expander-header:hover {
    color: #ccc;
  }

  .expander-icon {
    font-size: 0.8rem;
    transition: transform 0.2s ease;
    display: inline-block;
    width: 1rem;
  }

  .expander-content {
    margin-top: 1rem;
    padding-left: 1.5rem;
  }
</style>