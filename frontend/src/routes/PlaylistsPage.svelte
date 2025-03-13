<script>
  import { onMount } from 'svelte';
  import PlaylistItem from '../components/PlaylistItem.svelte';

  let playlists = $state([]);

  async function loadPlaylists() {
    try {
      if (!window.go?.main?.App) {
        console.error("App binding not available");
        return;
      }
      const data = await window.go.main.App.GetPlaylists();
      if (data) {
        playlists = data; // ✅ Use .set() to update the state properly
        console.log("Playlists loaded:", data);
      } else {
        console.error("No data returned from GetPlaylists");
      }
    } catch (error) {
      console.error("Failed to load playlists:", error);
    }
  }


  onMount(() => {
    loadPlaylists();
  });
</script>

<main>
  <h1>Playlists</h1>

  {#if playlists.length > 0}
    <ul>
      {#each playlists as playlist (playlist.id)}
        <PlaylistItem playlist={playlist} />
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
</style>
