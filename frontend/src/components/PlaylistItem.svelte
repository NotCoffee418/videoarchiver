<script>  
    import SelectDirectoryButton from './SelectDirectoryButton.svelte';

    export let playlist;
  
    async function changeDirectory(newPath) {
      console.log(`Changing directory for: ${playlist.name}`);
  
      try {
        if (newPath) {
          playlist.save_directory = newPath;
          console.log(`Updated directory for ${playlist.name}: ${newPath}`);
          await window.go?.main?.App?.UpdatePlaylistDirectory(playlist.id, newPath);
          window.location.reload();
        }
      } catch (error) {
        console.error("Failed to select directory:", error);
      }
    }
  
    async function openDirectory() {
      console.log(`Opening directory for: ${playlist.save_directory}`);
  
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
        <a href={playlist.url} target="_blank" class="open-link">Open Playlist</a>
      </div>
  
      <div class="path-container">
        <input type="text" bind:value={playlist.save_directory} class="path" readonly />
        <SelectDirectoryButton text="Change" clickHandlerAsync={changeDirectory} />
        <button on:click={openDirectory} class="btn">Open</button>
      </div>
  
      <div class="format-container">
        {playlist.output_format.toUpperCase()}
      </div>
    </div>
  </li>
  

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
    }
  </style>
  