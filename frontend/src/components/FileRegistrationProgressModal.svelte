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

    <!-- Reserved space for close button -->
    <div class="button-container">
      {#if isComplete}
        <button class="close-button" onclick={closeModal}>Close</button>
      {:else}
        <!-- Reserve space even when button is not visible -->
        <div class="button-spacer"></div>
      {/if}
    </div>
  </div>
  {/if}
</dialog>

<style>
  dialog#file-registration-progress-dialog {
    border: none;
    border-radius: 12px;
    padding: 2rem;
    width: 50rem;
    max-width: 90vw;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
    background-color: #222;
    color: #fff;
  }

  dialog#file-registration-progress-dialog::backdrop {
    background: rgba(0, 0, 0, 0.5);
  }

  .modal-content {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .modal-header {
    text-align: center;
  }

  .modal-header h2 {
    margin: 0;
    color: #fff;
    font-size: 1.5rem;
    font-weight: 500;
  }

  .progress-container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    flex: 1;
  }

  .progress-text {
    color: #fff;
    font-size: 1rem;
    text-align: center;
    min-height: 24px;
    margin-bottom: 0.5rem;
  }

  .progress-bar-container {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .progress-bar {
    flex: 1;
    height: 20px;
    background-color: #333;
    border-radius: 10px;
    overflow: hidden;
    border: 1px solid #555;
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
    color: #fff;
    min-width: 40px;
    text-align: right;
  }

  .loading-indicator {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 1rem;
  }

  .loading-note {
    margin: 0;
    color: #ccc;
    font-style: italic;
    text-align: center;
  }

  .completion-indicator {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 1rem;
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

  .button-container {
    display: flex;
    justify-content: center;
    margin-top: 1rem;
    min-height: 2.5rem;
    align-items: center;
  }

  .close-button {
    background-color: #333;
    color: #fff;
    border: 1px solid #555;
    padding: 0.5rem 2rem;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
    transition: all 0.2s ease;
    min-width: 120px;
  }

  .close-button:hover {
    background-color: #555;
  }

  .button-spacer {
    height: 2.5rem;
    width: 120px;
  }
</style>