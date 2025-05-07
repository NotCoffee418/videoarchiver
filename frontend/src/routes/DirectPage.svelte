<div class="container">
    <h1>Direct Download</h1>

    <div class="form-group">
        <label for="url">URL</label>
        <div class="input-group">
            <input type="text" id="url" placeholder="Enter URL" bind:value={url} />
            <button class="paste-button" onclick={pasteUrl}>Paste URL</button>
        </div>
    </div>

    <div class="form-group">
        <label for="directory">Directory</label>
        <div class="input-group">
            <input type="text" id="directory" placeholder="Enter directory" bind:value={directory} />
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

    <button class="download-button">Download</button>
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
        width: 100%;
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
</style>


<script>
    let format = "mp3";
    let url = "";
    let directory = "";

    function pasteUrl() {
        navigator.clipboard.readText().then(text => {
            console.log(text);
        });
    }
</script>