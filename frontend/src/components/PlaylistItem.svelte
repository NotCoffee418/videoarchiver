<script>  
    import SelectDirectoryButton from './SelectDirectoryButton.svelte'; 
    export let playlist;
    /** @type {() => Promise<void>} */
    export let refreshFunction = async () => {};
  
    async function changeDirectory(newPath) {  
      try {
        if (newPath) {
          playlist.save_directory = newPath;
          await window.go?.main?.App?.UpdatePlaylistDirectory(playlist.id, newPath);
          await refreshFunction();
        }
      } catch (error) {
        console.error("Failed to select directory:", error);
      }
    }
  
    async function openDirectory() {
      try {
        if (!window.go?.main?.App) {
          console.error("App binding not available");
          return;
        }
        await window.go?.main?.App?.OpenDirectory(playlist.save_directory);
      } catch (error) {
        console.error("Failed to open directory:", error);
      }
    }

    async function copyPlaylistUrl() {
      await navigator.clipboard.writeText(playlist.url);
    }

    function openDeletePlaylistItemModal() {
      /** @type {HTMLDialogElement} */
      const deleteModal = document.querySelector(`#delete-playlist-item-modal-${playlist.id}`)
      deleteModal.showModal();
    }

    function closeDeletePlaylistItemModal() {
      /** @type {HTMLDialogElement} */
      const deleteModal = document.querySelector(`#delete-playlist-item-modal-${playlist.id}`)
      deleteModal.close();
    }

    async function confirmDeletePlaylistItem() {

      closeDeletePlaylistItemModal();
      await window.go?.main?.App?.DeletePlaylist(playlist.id);
      await refreshFunction();      
    }
    
  </script>
  
  <li>
    <div class="thumbnail">
      {#if playlist.thumbnail_base64?.Valid && playlist.thumbnail_base64?.String.trim() !== ""}
        <img src={`data:image/jpg;base64,${playlist.thumbnail_base64.String}`} alt="Thumbnail" />
      {:else}
        <div class="thumbnail-placeholder">No Image</div>
      {/if}
    </div>
  
    <div class="content">
      <div class="title-container">
        <h2 title={`Added At: ${playlist.added_at ? new Date(playlist.added_at).toLocaleString() : 'N/A'}`}>
          {playlist.name}
        </h2>
        <button onclick={copyPlaylistUrl} class="button-link">Copy Playlist URL</button>
      </div>
  
      <div class="path-container">
        <input type="text" bind:value={playlist.save_directory} class="path" readonly />
        <SelectDirectoryButton text="Change" clickHandlerAsync={changeDirectory} />
        <button onclick={openDirectory} class="btn">Open</button>
      </div>
  
      <div class="format-container">
        {playlist.output_format.toUpperCase()}
        <button onclick={openDeletePlaylistItemModal} class="delete-btn" aria-label="Delete playlist">

            <svg class="delete-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="3 6 5 6 21 6"></polyline>
              <path d="M19 6L18 20H6L5 6"></path>
              <path d="M10 11V17"></path>
              <path d="M14 11V17"></path>
              <path d="M9 6L10 3H14L15 6"></path>
            </svg>         
        </button>
      </div>
    </div>
  </li>


  <dialog id="delete-playlist-item-modal-{playlist.id}">
    <button class="dialog-close-btn" onclick={closeDeletePlaylistItemModal}>✕</button>
    <h1>Delete Playlist</h1>
    <p>Are you sure you want to delete playlist <span class="playlist-name">'{playlist.name}'</span>?</p>
    <button class="danger-btn" onclick={confirmDeletePlaylistItem}>Delete</button>
    <button class="" onclick={closeDeletePlaylistItemModal}>Cancel</button>
  </dialog>
  

  <style>
    li {
      display: flex;
      align-items: flex-start;
      gap: 1rem;
      border: 1px solid var(--border-color, #555);
      padding: 1rem;
      margin-bottom: 1rem;
      border-radius: 12px;
      transition: background-color 0.2s ease;
    }
  
    li:hover {
      background-color: rgba(255, 255, 255, 0.05);
    }
  
    .thumbnail {
      width: 100px;
      height: 100px;
      border-radius: 4px;
      overflow: hidden;
      display: flex;
      align-items: center;
      justify-content: center;
      border: 1px solid var(--border-color, #555);
    }
  
    .thumbnail img {
      width: auto;
      height: auto;
      object-fit: contain;
    }
  
    .thumbnail-placeholder {
      color: #999;
      font-size: 0.8rem;
      text-align: center;
    }
  
    .content {
      flex: 1;
      display: flex;
      flex-direction: column;
    }
  
    .title-container {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
    }
  
    h2 {
      font-size: 1.2rem;
      font-weight: 500;
      margin: 0;
      color: inherit;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  
    .path-container {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      margin-top: 0.5rem;
    }
  
    .path {
      flex: 1;
      background-color: #222;
      color: inherit;
      border: 1px solid var(--border-color, #555);
      padding: 0.5rem;
      border-radius: 4px;
      font-size: 0.9rem;
    }
  
    .format-container {
      margin-top: 0.5rem;
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
    }

    .button-link {
      all: unset; /* Resets all button styles */
      cursor: pointer;
      color: #4dabf7; /* Subtle blue color for better visibility */
      transition: color 0.2s ease;
    }

    .button-link:hover {
      background-color: transparent !important; /* Override global button hover background */
      color: #74c0fc; /* Lighter blue on hover */
      text-decoration: underline;
    }

    .btn {
      cursor: pointer;
    }

    .delete-btn {
      all: unset;
      cursor: pointer;
      border: 1px solid rgba(255, 0, 0, 0.664);
      border-radius: 4px;
      width: 1rem;
      height: 1rem;
      padding-bottom: 0.1rem;
    }

    .delete-icon {
      height: 1rem;
      width: 1rem;
      color: rgba(255, 0, 0, 0.664);

      border-radius: 4px;
    }

    .danger-btn {
      background-color: rgba(255, 0, 0, 0.664);
      color: #fff;
      border: none;
      padding: 0.5rem 1rem;
      border-radius: 4px;
      margin-top: 1rem;
    }

  </style>
  