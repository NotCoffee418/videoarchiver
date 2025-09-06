<script>
  let { 
      /** @type {string} */
      key, 
      /** @type {string} */
      label, 
      /** @type {string} */
      description = "",
      /** @type {(any) => boolean} */
      validationFunction = (value) => {
          // Optional function for special validation
          return true;
      }
  } = $props();

  let value = $state("");
  let initialValueSet = $state(false);

  // Reload setting on page load
  $effect(() => {
      initialValueSet = false; // for reload

      // Run in background, no rush
      (async () => {
          await loadSetting(key);
          initialValueSet = true;
      })();
  });

  function onChangeString(e) {
      let value = e.target.value;
      let isValid = true;
      isValid = validationFunction(value);
      if (!isValid) {
          document.querySelector(`#json-string-input-${key}`).classList.add("invalid-input");
          return;
      } else {
          document.querySelector(`#json-string-input-${key}`).classList.remove("invalid-input");
          saveSetting(key, value);
      }
  }

  /**
   * Save the setting to the config file
   * @param {string} key - The key of the setting
   * @param {string} value - The value of the setting
   */
  function saveSetting(key, value) {
      // Don't save on page load
      if (!initialValueSet) {
          return;
      }

      // Save to backend
      window.go.main.App.SetConfigString(key, value);
  }

  async function loadSetting(key) {
      console.log("Loading config setting:", key);
      const loadedValue = await window.go.main.App.GetConfigString(key);
      
      // Config values are stored as strings directly
      value = loadedValue;
  }
  
</script>

<div class="setting-row">
  <div class="key-slot">
    <span class="setting-title">{label}</span>
    <span class="setting-description">{description}</span>
  </div>
  <div class="value-slot">
    <input 
        id={`json-string-input-${key}`}
        type="text" 
        bind:value={value} 
        oninput={onChangeString} 
    />
  </div>
</div>

<style>
   .setting-row {
        display: grid;
        grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
        align-items: center;
        gap: 1rem;
        padding-bottom: 0.75rem;
        border-bottom: 1px solid #333;
    }

    .key-slot {
        text-align: left;
        font-weight: bold;
        white-space: nowrap;
        max-width: 100%;
        overflow: hidden;
    }

    .value-slot {
        width: 100%;
    }

    input {
        padding: 0.5rem 0 0.5rem 0.5rem;
        width: 100%;
    }

    .setting-title {
        font-size: 1.2rem;
        font-weight: bold;
        display: block;
        white-space: nowrap;
    }
    
    .setting-description {
        font-size: 0.8rem;
        color: #666;
        display: block;
        overflow: hidden;
        text-overflow: ellipsis;
        overflow-wrap: break-word;
        word-break: break-word;
        white-space: normal;
    }

    :global(input.invalid-input) {
        border: 2px solid #ff6b6b !important;
        background-color: rgba(255, 107, 107, 0.1) !important;
    }
</style>