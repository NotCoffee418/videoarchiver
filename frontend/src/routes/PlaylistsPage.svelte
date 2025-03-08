<script>
    import { onMount } from 'svelte';
  
    let playlists = [];
  
    async function loadPlaylists() {
      try {
        console.log("Starting loadPlaylists...");
  
        if (!window.go?.playlist?.PlaylistDB) {
          console.error("PlaylistDB binding not available");
          return;
        }
  
        console.log("Calling GetPlaylists...");
        // ✅ Direct assignment since keys match
        playlists = await window.go.playlist.PlaylistDB.GetPlaylists();
        console.log("Playlists loaded:", playlists);
      } catch (error) {
        console.error("Failed to load playlists:", error);
      }
    }
  
    onMount(() => {
      console.log("Component mounted, calling loadPlaylists...");
      loadPlaylists();
    });
  </script>
  
  <main>
    <h1>Playlists</h1>
  
    {#if playlists.length > 0}
      <ul>
        {#each playlists as playlist}
          <li>
            <h2>{playlist.name}</h2>
            <p>URL: <a href={playlist.url} target="_blank">{playlist.url}</a></p>
            <p>Format: {playlist.output_format}</p>
            <p>Save Directory: {playlist.save_directory}</p>
            <p>Status: {playlist.is_enabled ? 'Enabled' : 'Disabled'}</p>
            <p>Added At: {playlist.added_at ? new Date(playlist.added_at * 1000).toLocaleString() : 'N/A'}</p>
            {#if playlist.thumbnail_base64}
              <img src={`data:image/jpeg;base64,${playlist.thumbnail_base64}`} alt="Thumbnail" />
            {/if}
          </li>
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
  
    li {
      border: 1px solid var(--border-color, #555);
      padding: 1rem;
      margin-bottom: 1rem;
      border-radius: 8px;
      transition: background-color 0.2s ease;
    }
  
    a {
      text-decoration: underline;
      color: inherit;
    }
  
    img {
      max-width: 100px;
      height: auto;
      border-radius: 4px;
      margin-top: 0.5rem;
    }
  
    @media (prefers-color-scheme: dark) {
      :root {
        --border-color: #444;
      }
  
      li {
        background-color: #222;
      }
  
      a {
        color: #7abaff;
      }
    }
  
    @media (prefers-color-scheme: light) {
      :root {
        --border-color: #ccc;
      }
  
      li {
        background-color: #f9f9f9;
      }
  
      a {
        color: #007bff;
      }
    }
  </style>
  