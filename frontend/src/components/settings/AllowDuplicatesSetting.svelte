<script>
  import { SettingType } from './SettingView.svelte';

  let {
    /** @type {string} */
    key,
    /** @type {string} */
    label,
    /** @type {string} */
    description = "",
  } = $props();

  let value = $state(false);
  let initialValueSet = $state(false);
  let showConfirmModal = $state(false);

  // Reload setting on page load
  $effect(() => {
    initialValueSet = false;

    // Run in background, no rush
    (async () => {
      await loadSetting(key);
      initialValueSet = true;
    })();
  });

  function onChangeBoolean(e) {
    let newValue = e.target.checked;
    
    if (newValue && !value) {
      // User is trying to enable duplicates - show confirmation modal
      e.target.checked = false; // Reset checkbox until confirmed
      showConfirmModal = true;
      /** @type {HTMLDialogElement | null} */
      const dialog = document.querySelector('#allow-duplicates-confirm-modal');
      if (dialog) dialog.showModal();
    } else {
      // User is disabling - allow immediately
      value = newValue;
      saveSetting(key, value.toString());
    }
  }

  function confirmEnableDuplicates() {
    value = true;
    /** @type {HTMLInputElement | null} */
    const checkbox = document.querySelector(`#checkbox-input-${key}`);
    if (checkbox) checkbox.checked = true;
    saveSetting(key, value.toString());
    closeConfirmModal();
  }

  function closeConfirmModal() {
    showConfirmModal = false;
    /** @type {HTMLDialogElement | null} */
    const dialog = document.querySelector('#allow-duplicates-confirm-modal');
    if (dialog) dialog.close();
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
    
    // Handle loaded value for boolean
    console.log(loadedValue);
    value = JSON.parse(loadedValue) == true;
  }
</script>

<div class="setting-row">
  <div class="key-slot">
    <span class="setting-title">{label}</span>
    <span class="setting-description">{description}</span>
  </div>
  <div class="value-slot">
    <input 
        id={`checkbox-input-${key}`}
        class="checkbox-input" 
        type="checkbox" 
        bind:checked={value} 
        oninput={onChangeBoolean} 
    />
  </div>
</div>

<!-- Confirmation Modal -->
<dialog id="allow-duplicates-confirm-modal">
  <button class="dialog-close-btn" onclick={closeConfirmModal}>✕</button>
  <h1>⚠️ Enable Duplicate Downloads?</h1>
    <p><strong>Warning:</strong> This may fill your playlist directories with duplicate files which already exist in other directories depending on your setup.</p>
    <p>Enabling this setting will also skip checking the registered files.</p>
    <p>Are you sure you want to proceed?</p>
  <div class="modal-buttons">
    <button class="danger-btn" onclick={confirmEnableDuplicates}>Yes, Allow Duplicates</button>
    <button onclick={closeConfirmModal}>Cancel</button>
  </div>
</dialog>

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
  }

  .value-slot {
      text-align: right;
  }

  .setting-title {
      color: #fff;
      display: block;
      font-weight: bold;
      margin-bottom: 0.25rem;
  }

  .setting-description {
      font-size: 0.9rem;
      color: #666;
      display: block;
      overflow: hidden;
      text-overflow: ellipsis;
      overflow-wrap: break-word;
      word-break: break-word;
      white-space: normal;
  }

  .modal-buttons {
    display: flex;
    gap: 1rem;
    justify-content: center;
    margin-top: 1.5rem;
  }

  .danger-btn {
    background-color: #dc3545;
    border-color: #dc3545;
  }

  .danger-btn:hover {
    background-color: #c82333;
    border-color: #bd2130;
  }
</style>