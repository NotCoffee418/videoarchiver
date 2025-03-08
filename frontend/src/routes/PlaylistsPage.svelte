<script>
    import { onMount } from 'svelte';

    let playlists = [];

    async function loadPlaylists() {
    try {
        if (!window.go?.main?.App) {
        console.error("App binding not available");
        return;
        }

        playlists = await window.go.main.App.GetPlaylists(); // ✅ FIXED PATH
        console.log("Playlists loaded:", playlists);
    } catch (error) {
        console.error("Failed to load playlists:", error);
    }
    }

    onMount(() => {
    loadPlaylists();
    });

    async function openDirectory(path) {
    console.log(`Opening directory: ${path}`);

    try {
        if (!window.go?.main?.App) {
        console.error("App binding not available");
        return;
        }
        await window.go.main.App.OpenDirectory(path); // ✅ FIXED PATH
    } catch (error) {
        console.error("Failed to open directory:", error);
    }
    }

    async function changeDirectory() {
        console.log("Changing directory...");

        try {
            if (!window.go?.main?.App) {
            console.error("App binding not available");
            return;
            }
            const path = await window.go.main.App.SelectDirectory(); // ✅ FIXED PATH
            if (path) {
            console.log(`Selected directory: ${path}`);
            // TODO: Add code to update playlist's save directory
            }
        } catch (error) {
            console.error("Failed to select directory:", error);
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
        {#each playlists as playlist}
          <li>
            <!-- Thumbnail -->
            <div class="thumbnail">
              {#if playlist.thumbnail_base64}
                <img src={`data:image/jpeg;base64,${playlist.thumbnail_base64}`} alt="Thumbnail" />
              {:else}
                <div class="thumbnail-placeholder">No Image</div>
              {/if}
            </div>
  
            <!-- Playlist Info -->
            <div class="content">
              <!-- Title + Hover for Added At -->
              <div class="title-container">
                <h2 title={`Added At: ${playlist.added_at ? new Date(playlist.added_at).toLocaleString() : 'N/A'}`}>
                  {playlist.name}
                </h2>
                <a href={playlist.url} target="_blank" class="open-link">Open Playlist</a>
              </div>
  
              <!-- Path + Buttons -->
              <div class="path-container">
                <input type="text" value={playlist.save_directory} readonly class="path" />
                <button on:click={() => changeDirectory(playlist.save_directory)} class="btn">Change</button>
                <button on:click={() => openDirectory(playlist.save_directory)} class="btn">Open Directory</button>
              </div>
  
              <!-- Format Combobox -->
              <div class="format-container">
                <select disabled>
                  <option value="mp3" selected={playlist.output_format === 'mp3'}>MP3</option>
                  <option value="mp4" selected={playlist.output_format === 'mp4'}>MP4</option>
                </select>
              </div>
            </div>
          </li>
        {/each}
      </ul>
    {:else}
      <p>No playlists found.</p>
    {/if}
  </main>
  
  <!-- ✅ Styles -->
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
  
    /* ✅ Thumbnail */
    .thumbnail {
      width: 100px;
      height: 100px;
      border-radius: 8px;
      overflow: hidden;
      background-color: #333;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  
    .thumbnail img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  
    .thumbnail-placeholder {
      color: #999;
      font-size: 0.8rem;
      text-align: center;
    }
  
    /* ✅ Content Area */
    .content {
      flex: 1;
      display: flex;
      flex-direction: column;
    }
  
    /* ✅ Title and Open Link */
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
  
    h2[title] {
      cursor: help;
    }
  
    .open-link {
      font-size: 0.9rem;
      color: #7abaff;
      text-decoration: none;
      transition: color 0.2s;
    }
  
    .open-link:hover {
      color: #4ea1d3;
    }
  
    /* ✅ Path Container */
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
  
    .btn {
      background-color: #444;
      color: #ddd;
      border: 1px solid #555;
      padding: 0.4rem 0.7rem;
      border-radius: 4px;
      cursor: pointer;
      font-size: 0.9rem;
      transition: background-color 0.2s ease;
    }
  
    .btn:hover {
      background-color: #555;
    }
  
    /* ✅ Format Container */
    .format-container {
      margin-top: 0.5rem;
    }
  
    select {
      background-color: #222;
      color: inherit;
      border: 1px solid var(--border-color, #555);
      padding: 0.4rem;
      border-radius: 4px;
      font-size: 0.9rem;
      appearance: none;
    }
  
    /* ✅ Dark mode support */
    @media (prefers-color-scheme: dark) {
      :root {
        --border-color: #444;
      }
    }
  
    @media (prefers-color-scheme: light) {
      :root {
        --border-color: #ccc;
      }
  
      .path {
        background-color: #f9f9f9;
        color: #222;
      }
  
      .btn {
        background-color: #eee;
        color: #222;
        border: 1px solid #ddd;
      }
  
      select {
        background-color: #f9f9f9;
        color: #222;
      }
    }
  </style>
  