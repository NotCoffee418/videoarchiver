<script module>
    export const SettingType = Object.freeze({
      BOOL: "bool",
      INT: "int",
      FLOAT: "float",
      STRING: "string",
      SELECT: "select",
      MULTISELECT: "multiselect",
    });

    /**
     * @typedef {Object} Option
     * @property {string} label - Display label
     * @property {string} value - Internal value
     */

  </script>
  
  <script>
    let { 
        /** @type {string} */
        key, 
        /** @type {string} */
        label, 
        /** @type {string} */
        description = "",
        /** @type {SettingType} */
        type = SettingType.STRING, 
        /** @type {Option[]} */
        options = [],
        /** @type {(any) => boolean} */
        validationFunction = (parsedInput) => {
            // Input is parsedInput
            // Optional function for special validation
            return true;
        }
    } = $props();
  
    let value = $state(null);
    let initialValueSet = $state(false); // CHANGE ME

    // Reload setting on page load
    $effect(() => {
        initialValueSet = false; // for reload

        // Run in background, no rush
        (async () => {
            await loadSetting(key);
            initialValueSet = true;
        })();
    });
    


    // Regex for number validation
    const intRegex = /^[0-9]*$/;
    const floatRegex = /^[-+]?\d+(\.\d+)?$/;

    function onChangeBoolean(e) {
        let value = e.target.checked;
        saveSetting(key, value.toString());
    }

    function onChangeString(e) {
        let value = e.target.value;
        let isValid = true;
        isValid = validationFunction(value);
        if (!isValid) {
            document.querySelector(`#string-input-${key}`).classList.add("invalid-input");
            return;
        } else {
            document.querySelector(`#string-input-${key}`).classList.remove("invalid-input");
            saveSetting(key, value);
        }
    }

    function onChangeNumber(e) {
        let value = e.target.value;
        let isValid = false;

        if (type === SettingType.INT) {
            isValid = intRegex.test(value);
        } else if (type === SettingType.FLOAT) {
            isValid = floatRegex.test(value);
        }

        // parse and do special validation if still valid
        let parsedValue = 0;
        if (isValid) {
            if (type === SettingType.INT) {
                parsedValue = parseInt(value);
            } else if (type === SettingType.FLOAT) {
                parsedValue = parseFloat(value);
            }

            if (isValid) {
                isValid = validationFunction(parsedValue);
            }
        }
            
        // Update UI
        if (!isValid) {
            document.querySelector(`#number-input-${key}`).classList.add("invalid-input");
            return;
        } else {
            document.querySelector(`#number-input-${key}`).classList.remove("invalid-input");
            saveSetting(key, parsedValue.toString());
        }
    }

    function onChangeSelect(e) {
        let value = e.target.value;
        let isValid = options.some(option => option.value === value);
        if (isValid) {
            isValid = validationFunction(value);
        }
        if (!isValid) {
            document.querySelectorAll(`#select-input-${key}`).forEach(el => el.classList.add("invalid-input"));
            return;
        } else {
            document.querySelectorAll(`#select-input-${key}`).forEach(el => el.classList.remove("invalid-input"));
            saveSetting(key, value);
        }
    }

    function onChangeMultiSelect(e) {
        // Parse all options
        /** @type {NodeListOf<HTMLInputElement>} */
        const elements = document.querySelectorAll(`[id^="multiselect-checkbox-${key}"]`);
        if (elements.length === 0) {
            return;
        }

        let resultMap = new Map();
        let isValid = true;
        elements.forEach(/** @type {HTMLInputElement} */ el => {
            const cbKey = el.id.split("-")[3];
            const cbValue = el.checked;
            resultMap.set(cbKey, cbValue);
            if (!options.some(option => option.value === cbKey)) {
                isValid = false;
            }
        });

        // Additional validation on map
        if (isValid) {
            isValid = validationFunction(resultMap);
        }

        if (!isValid) {
            elements[0].parentElement.parentElement.classList.add("invalid-input");
            return;
        } else {
            elements[0].parentElement.parentElement.classList.remove("invalid-input");
            
            // Parse to comma seperated string of enabled items
            let enabledItems = [];
            resultMap.forEach((value, key) => {
                if (value) {
                    enabledItems.push(key);
                }
            });
            saveSetting(key, enabledItems.join(","));
        }
    }

    /**
     * Save the setting to the database
     * @param {string} key - The key of the setting
     * @param {string} parsedValue - The value of the setting as it is placed in db
     */
    function saveSetting(key, parsedValue) {
        // Don't save on page load
        if (!initialValueSet) {
            return;
        }

        // Save to backend
        window.go.main.App.SetSettingPreparsed(key, parsedValue);
    }

    async function loadSetting(key) {
        console.log(key);
        const loadedValue = await window.go.main.App.GetSettingString(key);
        console.log("loadedValue", loadedValue);
        const parsedValue = JSON.parse(loadedValue);
        console.log("parsedValue", parsedValue);
        
        // Handle loaded value per type
        if (type === SettingType.BOOL) { 
            console.log(loadedValue);
            value = parsedValue == true;
        } else if (type === SettingType.MULTISELECT) {
            const selectedValues = loadedValue.split(",");
            // Find options that match selectedValues
            options.forEach(option => {
                /** @type {HTMLInputElement | null} */
                const checkbox = document.querySelector(`#multiselect-checkbox-${key}-${option.value}`);
                if (selectedValues.includes(option.value) && checkbox) {
                    checkbox.checked = true;
                }
            });

        } else {
            // Everything else acts like a string
            value = parsedValue;
        }
    }
    
  </script>
  
  <div class="setting-row">
    <div class="key-slot">
      <span class="setting-title">{label}</span>
      <span class="setting-description">{description}</span>
    </div>
    <div class="value-slot">
      {#if type === SettingType.BOOL}
        <input 
            id={`checkbox-input-${key}`}
            class="checkbox-input" 
            type="checkbox" 
            bind:checked={value} 
            oninput={onChangeBoolean} 
        />
      
      {:else if type === SettingType.INT || type === SettingType.FLOAT}
        <input 
            id={`number-input-${key}`}
            class="number-input" 
            type="text" 
            bind:value={value} 
            oninput={onChangeNumber} 
        />
      
      {:else if type === SettingType.STRING}
        <input 
            id={`string-input-${key}`}
            type="text" 
            bind:value={value} 
            oninput={onChangeString} 
        />
      
      {:else if type === SettingType.SELECT}
        <select 
            id={`select-input-${key}`}
            bind:value={value} 
            oninput={onChangeSelect}
        >
          {#each options as option}
            <option 
                value={option.value}
                id={`select-option-${key}-${option.value}`}>
                    {option.label}
            </option>
          {/each}
        </select>
      
      {:else if type === SettingType.MULTISELECT}
        <div class="value-slot multiselect-slot">
        {#each options as option}
          <label class="multiselect-slot">
            <input 
              id={`multiselect-checkbox-${key}-${option.value}`}
              class="checkbox-input"
              type="checkbox"
              oninput={onChangeMultiSelect} 
            />
            {option.label}
          </label>
        {/each}
        </div>
      {/if}      
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

    .multiselect-slot {
        display: block;
    }

    label {
        display: flex; /* ✅ Flexbox for alignment */
        align-items: center; /* ✅ Vertical alignment */
        gap: 0.5rem; /* ✅ Clean spacing between checkbox and label */
        cursor: pointer;
    }

    input {
        padding: 0.5rem 0 0.5rem 0.5rem;
        width: 100%;
    }

    select {
        padding: 0.25rem 0.5rem;
        width: 100%;
    }

    .number-input {
        padding-left: 0.5rem;
    }

    .checkbox-input {
        margin-right: 0.5rem;
        height: 0.9rem;
        width: 0.9rem;
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
</style>