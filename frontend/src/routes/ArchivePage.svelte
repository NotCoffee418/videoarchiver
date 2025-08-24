<script>
  import { onMount, tick } from 'svelte';
  import PlaylistItem from '../components/PlaylistItem.svelte';
  import AddPlaylistButton from '../components/AddPlaylistButton.svelte';

  let playlists = $state([]);

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

  onMount(() => {
    reloadPlaylists();
  });
</script>

<main>
  <div class="header">
    <h1>Archive</h1>
    <AddPlaylistButton onPlaylistAdded={reloadPlaylists} />
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
</style>
