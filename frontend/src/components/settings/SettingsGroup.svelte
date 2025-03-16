<!-- SettingsGroup.svelte -->
<script>
    export let title;
  </script>
  
  <div class="settings-table">
    {#if title}
      <h2>{title}</h2>
    {/if}
    <slot />
  </div>
  
  <style>
    .settings-table {
        display: grid;
        gap: 1rem;
        max-width: 50rem;
        margin-top: 1rem;
        margin-left: auto;
        margin-right: auto;
    }

  
    
  </style>
  