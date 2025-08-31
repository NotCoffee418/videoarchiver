<script>
    import LoadingSpinner from "../components/LoadingSpinner.svelte";
    import SelectDirectoryButton from "../components/SelectDirectoryButton.svelte";
    import { directState, updateDirectState } from "../stores/directStore.js";

    // Reactive state from store
    $: ({ format, url, directory, isDownloading, error } = $directState);

    async function selectDirectory(path) {
        updateDirectState({ directory: path });
    }

    function pasteUrl() {
        window.go.main.App.GetClipboard().then(text => {
            updateDirectState({ url: text });
        });
    }

    function directDownload() {
        updateDirectState({ isDownloading: true, error: "" });
        window.go.main.App.DirectDownload(url, directory, format).then(() => {
            updateDirectState({ isDownloading: false });
        }).catch(err => {
            updateDirectState({ isDownloading: false, error: err.toString() });
        }).then(path => {
            //window.go.main.App.OpenDirectory(path);
        });
    }
</script>

<div class="container">
    <h1>Direct Download</h1>

    <div class="form-group">
        <label for="url">URL</label>
        <div class="input-group">
            <input type="text" id="url" placeholder="Enter URL" value={url} oninput={(e) => updateDirectState({ url: e.target.value })} />
            <button class="paste-button" onclick={pasteUrl}>Paste URL</button>
        </div>
    </div>

    <div class="form-group">
        <label for="directory">Directory</label>
        <div class="input-group">
            <input type="text" id="directory" placeholder="Enter directory" value={directory} oninput={(e) => updateDirectState({ directory: e.target.value })} />
            <SelectDirectoryButton text="Select Directory" clickHandlerAsync={selectDirectory} />
        </div>
    </div>

    <div class="form-group">
        <label for="format">Format</label>
        <div class="input-group">
            <select id="format" value={format} onchange={(e) => updateDirectState({ format: e.target.value })}>
                <option value="mp3">MP3</option>
                <option value="mp4">MP4</option>
            </select>
        </div>
    </div>

    {#if isDownloading}
        <LoadingSpinner size="4rem" />
    {:else if error}
        <p class="error">Error: {error}</p>
    {:else}
        <div class="spinner-filler"></div>
    {/if}

    <button class="download-button" id="download-button" onclick={directDownload} disabled={isDownloading}>Download</button>
</div>

<style>
    .container {
        max-width: 800px;
        margin: 2rem auto;
        padding: 0 1rem;
    }

    h1 {
        margin-bottom: 2rem;
        text-align: center;
    }
    
    input {
        width: 100%;
        padding: 0.5rem;
        border: 1px solid #ccc;
        border-radius: 4px;
    }

    .input-group {
        display: flex;
        gap: 1rem;
    }

    .input-group input {
        flex: 1;
    }

    .paste-button {
        width: 10rem;
        padding: 0.5rem;
        border-radius: 4px;
        cursor: pointer;
        white-space: nowrap;
    }

    .form-group {
        margin-bottom: 2rem;
    }

    .form-group label {
        display: block;
        margin-bottom: 0.5rem;
    }

    .input-group select {
        width: 100%;
        padding: 0.5rem;
        border: 1px solid #ccc;
        border-radius: 4px;
    }

    .download-button {
        display: block;
        width: 10rem;
        padding: 0.75rem;
        background-color: #4CAF50;
        color: white;
        border: none;
        border-radius: 4px;
        cursor: pointer;
        font-size: 1rem;
    }

    .download-button:hover {
        background-color: #45a049;
    }

    .download-button:disabled {
        background-color: #ccc;
        cursor: not-allowed;
    }

    .spinner-filler {
        height: 5rem;
    }

    .error {
        color: red;
        margin-bottom: 1rem;
    }
</style>
