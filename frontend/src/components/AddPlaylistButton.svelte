<script>  
    import LoadingSpinner from './LoadingSpinner.svelte';
    import SelectDirectoryButton from './SelectDirectoryButton.svelte';

    let {
        onPlaylistAdded = async () => {
            console.warn("onPlaylistAdded has no handler.");
        }
    } = $props();

    let showModal = $state(false);
    let modalProcessing = $state(false);
    /** @type {string | null} */
    let modalError = $state(null);
    let playlistUrl = $state("");
    let saveDirectory = $state("");
    let format = $state("mp3");
  
    function openModal() {
      showModal = true;
      /** @type {HTMLDialogElement | null} */
      const dialog = document.querySelector('dialog#add-playlist-dialog');
      if (dialog) dialog.showModal();
    }
  
    function closeModal() {
      showModal = false;
      /** @type {HTMLDialogElement | null} */
      const dialog = document.querySelector('dialog#add-playlist-dialog');
      if (dialog) dialog.close();

      // Reset modal inputs
      playlistUrl = "";
      saveDirectory = "";
      format = "mp3";
      modalError = null;
      modalProcessing = false;
    }
  
    async function handleAddPlaylist() {
      modalProcessing = true;
      try {
        // Validate and add playlist
        await window.go.main.App.ValidateAndAddPlaylist(playlistUrl, saveDirectory, format);

        // Notify caller and cleanup
        if (onPlaylistAdded) {
          await onPlaylistAdded();
        }
        closeModal();
      } catch (error) {
        modalError = error;
      } finally {
        modalProcessing = false;
      }
    }
  

    async function pasteUrl() {
        try {
            const text = await window.go.main.App.GetClipboard();
            if (text) {
                playlistUrl = text;
            }
        } catch (error) {
            console.error('Failed to paste:', error);
        }
    }

    async function setDirectory(newPath) {
        saveDirectory = newPath;
    }

  </script>
  
  <!-- Button to Open Modal -->
  <button class="add-playlist-btn" onclick={openModal}>+ Add Playlist</button>
  
  <!-- Modal -->
  <dialog id="add-playlist-dialog" class="modal">
    {#if modalProcessing}
        <LoadingSpinner />
    {:else if modalError}
        <button class="close-btn" onclick={closeModal}>✕</button>
        <p class="error-message">Error: {modalError}</p>
    {:else}
        <button class="close-btn" onclick={closeModal}>✕</button>
    
        <h1>Add Playlist</h1>
    
        <div class="form-group">
            <label for="playlist-url">Playlist URL</label>
            <div class="input-group">
                <input id="playlist-url" type="text" bind:value={playlistUrl} />
                <button class="btn-add-playlist-modal-button" onclick={pasteUrl}>Paste</button>
            </div>
        </div>
    
        <div class="form-group">
            <label for="save-directory">Directory</label>
            <div class="input-group">
                <input id="save-directory" type="text" bind:value={saveDirectory} />        
                <SelectDirectoryButton
                    text="Change"
                    clickHandlerAsync={setDirectory}
                    class="btn-add-playlist-modal-button" />
            </div>
        </div>
    
        <div class="form-group">
            <label for="format">Format</label>
            <div class="input-group">
                <select id="format" bind:value={format}>
                    <option value="mp3">MP3</option>
                    <option value="mp4">MP4</option>
                </select>
            </div>
        </div>
    
        <button class="add-btn" onclick={handleAddPlaylist}>Add Playlist</button>
    {/if}
  </dialog>
  
  <style>
    h1 {
        margin-bottom: 1rem;
    }

    .add-playlist-btn {
        background-color: transparent;
        color: #4caf50;
        text-decoration: none;
        padding: 0.4rem 0.8rem;
        font-size: 0.9rem;
        border: 1px solid #4caf50;
        cursor: pointer;
        transition: background-color 0.2s ease, color 0.2s ease;
        display: inline-block;
    }

    .add-playlist-btn:hover {
        background-color: #4caf50;
        color: #fff;
    }

    .add-btn {
        background-color: #4caf50;
        margin-top: 1rem;
        width: 100%;

    }

    .add-btn:hover {
        background-color: #393;
        color: #fff;
        border-color: #393;
    }

    .modal {
        width: 50rem;
        height: 27rem;
        position: fixed;
        inset: 0;
        margin: auto;
        background-color: #222;
        color: #fff;
        padding: 2rem;
        border-radius: 12px;
        border: none;
    }

    .modal::backdrop {
        background: rgba(0, 0, 0, 0.5);
    }

    .close-btn {
        position: absolute;
        top: 0.5rem;
        right: 0.5rem;
        background: none;
        color: #fff;
        font-size: 1.5rem;
        border: none;
        cursor: pointer;
    }

    .form-group {
        margin-bottom: 1rem;
    }

    .input-group {
        display: flex;
        gap: 0.5rem;
        align-items: center;
    }
    input {
        flex-grow: 1;
    }

    :global(.btn-add-playlist-modal-button) {
        width: 6rem;
        flex-shrink: 0;
    }

    .error-message {
        color: #f00;
    }


  </style>