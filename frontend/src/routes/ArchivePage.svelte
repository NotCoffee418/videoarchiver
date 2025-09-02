<script>
  import { onMount, tick } from 'svelte';
  import PlaylistItem from '../components/PlaylistItem.svelte';
  import AddPlaylistButton from '../components/AddPlaylistButton.svelte';
  import FileRegistrationProgressModal from '../components/FileRegistrationProgressModal.svelte';

  let playlists = $state([]);
  let showProgressModal = $state(false);

  async function reloadPlaylists() {
    try {
      if (!window.go?.main?.App) {
        console.error("App binding not available");
        return;
      }
      const data = await window.go.main.App.GetActivePlaylists();
      if (data) {
        playlists = data;
        await tick();
      } else {
        console.error("No data returned from GetActivePlaylists");
      }
    } catch (error) {
      console.error("Failed to load playlists:", error);
    }
  }

  async function handleRegisterFiles() {
    try {
      // Select directory first
      const directory = await window.go.main.App.SelectDirectory();
      if (!directory) {
        return; // User cancelled directory selection
      }

      // Show progress modal
      showProgressModal = true;

      // Start file registration with progress
      await window.go.main.App.RegisterFilesWithProgress(directory);
    } catch (error) {
      console.error("Failed to register files:", error);
      showProgressModal = false;
    }
  }

  function onRegistrationComplete() {
    console.log("File registration completed successfully");
    // Could reload playlists or refresh data here if needed
  }

  onMount(() => {
    reloadPlaylists();
  });
</script>

<main>
  <div class="header">
    <h1>Archive</h1>
    <div class="header-buttons">
      <button class="register-files-btn" onclick={handleRegisterFiles}>
        📁 Register Files
      </button>
      <AddPlaylistButton onPlaylistAdded={reloadPlaylists} />
    </div>
  </div>

  {#if playlists.length > 0}
    <ul>
      {#each playlists as playlist (playlist.id)}
        <PlaylistItem playlist={playlist} refreshFunction={reloadPlaylists} />
      {/each}
    </ul>
  {:else}
    <p>No playlists found.</p>
  {/if}
</main>

<!-- Progress Modal -->
<FileRegistrationProgressModal 
  bind:isOpen={showProgressModal}
  onComplete={onRegistrationComplete}
/>

<style>
  main {
    padding: 1rem;
    font-family: inherit;
  }

  ul {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .header-buttons {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .register-files-btn {
    background-color: #2196F3;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    transition: background-color 0.2s;
  }

  .register-files-btn:hover {
    background-color: #1976D2;
  }
</style>
