<script>  
    let showModal = $state(false);
  
    function openModal() {
      showModal = true;
      const dialog = document.getElementById('add-playlist-dialog');
      if (dialog) dialog.showModal();
    }
  
    function closeModal() {
      showModal = false;
      const dialog = document.getElementById('add-playlist-dialog');
      if (dialog) dialog.close();
    }
  
    function handleAddPlaylist() {
      console.log("Adding playlist:", {
        playlistUrl,
        saveDirectory,
        format
      });
      // TODO: Implement actual add logic
    }
  
    let playlistUrl = "";
    let saveDirectory = "";
    let format = "mp3";
  </script>
  
  <!-- Button to Open Modal -->
  <button class="add-playlist-btn" onclick={openModal}>+ Add Playlist</button>
  
  <!-- Modal -->
  <dialog id="add-playlist-dialog" class="modal">
    <button class="close-btn" onclick={closeModal}>✕</button>
  
    <h1>Add Playlist</h1>
  
    <div class="form-group">
      <label for="playlist-url">Playlist URL</label>
      <div class="input-group">
        <input id="playlist-url" type="text" bind:value={playlistUrl} />
        <button class="btn" onclick={pasteUrl}>Paste</button>
      </div>
    </div>
  
    <div class="form-group">
      <label for="save-directory">Directory</label>
      <div class="input-group">
        <input id="save-directory" type="text" bind:value={saveDirectory} readonly />
        <button class="btn">Browse</button>
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
  </dialog>
  
  <style>
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

    .modal {
        position: fixed;
        inset: 0;
        margin: auto;
        background-color: #222;
        color: #fff;
        padding: 2rem;
        border-radius: 12px;
        width: 50rem;
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

    .btn {
        width: 6rem;
        flex-shrink: 0;
    }


  </style>