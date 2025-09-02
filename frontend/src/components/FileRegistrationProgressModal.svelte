<script>
  import LoadingSpinner from './LoadingSpinner.svelte';

  let {
    isOpen = $bindable(false),
    onComplete = () => {}
  } = $props();

  let progress = $state(0);
  let isComplete = $state(false);
  let progressText = $state("Initializing file registration...");
  
  // Listen for file registration progress events
  $effect(() => {
    if (isOpen && !isComplete) {
      // Listen for progress events
      const unsubscribeProgress = window.runtime.EventsOn('file-registration-progress', (data) => {
        progress = data.percent;
        progressText = data.message;
        console.log(`File registration progress: ${progress}% - ${progressText}`);
      });

      // Listen for completion events
      const unsubscribeComplete = window.runtime.EventsOn('file-registration-complete', () => {
        progress = 100;
        progressText = "File registration completed successfully!";
        isComplete = true;
        console.log("File registration completed");
      });

      // Cleanup function
      return () => {
        unsubscribeProgress();
        unsubscribeComplete();
      };
    }
  });

  // Handle modal opening/closing using the same pattern as other modals
  $effect(() => {
    console.log(`Modal isOpen changed to: ${isOpen}`);
    const dialog = document.querySelector('dialog#file-registration-progress-dialog');
    if (dialog) {
      if (isOpen && !dialog.open) {
        console.log("Opening modal dialog");
        dialog.showModal();
      } else if (!isOpen && dialog.open) {
        console.log("Closing modal dialog");
        dialog.close();
      }
    } else {
      console.log("Dialog element not found");
    }
  });

  function closeModal() {
    if (!isComplete) return; // Modal is unclosable until complete
    
    console.log("Closing modal and resetting state");
    isOpen = false;
    progress = 0;
    isComplete = false;
    progressText = "Initializing file registration...";
    onComplete();
  }
</script>

<!-- File Registration Progress Modal -->
<dialog id="file-registration-progress-dialog">
  {#if isOpen}
  <div class="modal-content">
    <div class="modal-header">
      <h2>File Registration Progress</h2>
      <!-- Close button only appears when complete -->
      {#if isComplete}
        <button class="dialog-close-btn" onclick={closeModal}>✕</button>
      {/if}
    </div>

    <div class="progress-container">
      <div class="progress-text">
        {progressText}
      </div>
      
      <div class="progress-bar-container">
        <div class="progress-bar">
          <div class="progress-fill" style="width: {progress}%"></div>
        </div>
        <div class="progress-percentage">
          {Math.round(progress)}%
        </div>
      </div>

      {#if !isComplete}
        <div class="loading-indicator">
          <LoadingSpinner />
          <p class="loading-note">Please wait, this process cannot be cancelled...</p>
        </div>
      {:else}
        <div class="completion-indicator">
          <div class="checkmark">✓</div>
          <p class="completion-note">Registration completed successfully!</p>
        </div>
      {/if}
    </div>
  </div>
  {/if}
</dialog>

<style>
  dialog#file-registration-progress-dialog {
    border: none;
    border-radius: 8px;
    padding: 0;
    width: 500px;
    max-width: 90vw;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
    background: var(--color-background, #ffffff);
  }

  .modal-content {
    padding: 24px;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .modal-header h2 {
    margin: 0;
    color: var(--color-text, #333333);
    font-size: 1.5rem;
  }

  .dialog-close-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    color: var(--color-text-secondary, #666666);
    padding: 4px;
    border-radius: 4px;
    transition: background-color 0.2s;
  }

  .dialog-close-btn:hover {
    background-color: var(--color-hover, rgba(0, 0, 0, 0.1));
  }

  .progress-container {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .progress-text {
    color: var(--color-text, #333333);
    font-size: 1rem;
    text-align: center;
    min-height: 24px;
  }

  .progress-bar-container {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .progress-bar {
    flex: 1;
    height: 20px;
    background-color: var(--color-background-secondary, #f0f0f0);
    border-radius: 10px;
    overflow: hidden;
    border: 1px solid var(--color-border, #dddddd);
  }

  .progress-fill {
    height: 100%;
    background: linear-gradient(90deg, #4CAF50, #45a049);
    border-radius: 10px;
    transition: width 0.3s ease;
    min-width: 2px;
  }

  .progress-percentage {
    font-weight: bold;
    color: var(--color-text, #333333);
    min-width: 40px;
    text-align: right;
  }

  .loading-indicator {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 16px;
  }

  .loading-note {
    margin: 0;
    color: var(--color-text-secondary, #666666);
    font-style: italic;
    text-align: center;
  }

  .completion-indicator {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 16px;
  }

  .checkmark {
    font-size: 3rem;
    color: #4CAF50;
    font-weight: bold;
  }

  .completion-note {
    margin: 0;
    color: #4CAF50;
    font-weight: bold;
    text-align: center;
  }
</style>