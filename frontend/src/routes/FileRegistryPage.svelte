<script>
    import { onMount, onDestroy } from "svelte";
    import LoadingSpinner from "../components/LoadingSpinner.svelte";
    import SelectDirectoryButton from "../components/SelectDirectoryButton.svelte";
    import FileRegistrationProgressModal from "../components/FileRegistrationProgressModal.svelte";

    let registeredFiles = $state([]);
    let offset = $state(0);
    let limit = $state(10);
    let loading = $state(false);
    let error = $state("");

    // Modal states
    let showRegisterModal = $state(false);
    let showClearModal = $state(false);
    let showProgressModal = $state(false);
    let modalProcessing = $state(false);
    let modalError = $state(null);
    let selectedDirectory = $state("");

    // Derived values
    let nextDisabled = $derived(registeredFiles.length < limit);
    let pageNumber = $derived(Math.floor(offset / limit) + 1);

    async function fetchPage(showLoading = true) {
        if (showLoading) loading = true;
        error = "";
        try {
            if (window?.go?.main?.App?.GetRegisteredFiles) {
                const res = await window.go.main.App.GetRegisteredFiles(offset, limit);
                registeredFiles = Array.isArray(res) ? res : [];
            } else {
                registeredFiles = [];
            }
        } catch (err) {
            error = String(err ?? "Unknown error");
            registeredFiles = [];
        } finally {
            if (showLoading) loading = false;
        }
    }

    function prevPage() {
        if (offset === 0) return;
        offset = Math.max(0, offset - limit);
        fetchPage(true);
    }

    function nextPage() {
        offset += limit;
        fetchPage(true);
    }

    let refreshInterval;
    const REFRESH_RATE = 10000; // 10 seconds

    onMount(() => {
        fetchPage(true);
        refreshInterval = setInterval(() => fetchPage(false), REFRESH_RATE);
    });

    onDestroy(() => {
        if (refreshInterval) {
            clearInterval(refreshInterval);
        }
    });

    function formatTimestamp(ts) {
        if (!ts) return "";
        const n = Number(ts);
        if (isNaN(n)) return "";
        return new Date(n * 1000).toLocaleString();
    }

    // Modal functions
    function openRegisterModal() {
        showRegisterModal = true;
        modalError = null;
        selectedDirectory = "";
        const dialog = document.querySelector('dialog#register-directory-dialog');
        if (dialog) dialog.showModal();
    }

    function closeRegisterModal() {
        showRegisterModal = false;
        const dialog = document.querySelector('dialog#register-directory-dialog');
        if (dialog) dialog.close();
        selectedDirectory = "";
        modalError = null;
        modalProcessing = false;
    }

    function openClearModal() {
        showClearModal = true;
        modalError = null;
        const dialog = document.querySelector('dialog#clear-all-dialog');
        if (dialog) dialog.showModal();
    }

    function closeClearModal() {
        showClearModal = false;
        const dialog = document.querySelector('dialog#clear-all-dialog');
        if (dialog) dialog.close();
        modalError = null;
        modalProcessing = false;
    }

    async function handleRegisterDirectory() {
        const dir = selectedDirectory + "";
        if (!dir || dir.trim() === "") {
            modalError = "Please select a directory first";
            return;
        }

        // Prevent double-clicks
        if (modalProcessing) {
            return;
        }
        modalProcessing = true;

        // Close the directory selection modal
        closeRegisterModal();
        
        // Show progress modal
        showProgressModal = true;
        
        try {
            await window.go.main.App.RegisterDirectory(dir);
            // Don't handle success here - let the progress modal handle completion events
        } catch (error) {
            console.error("RegisterDirectory function threw error:", error);
            // Instead of closing progress modal, emit error event manually
            // This handles cases where the function throws before starting the goroutine
            if (window.runtime && window.runtime.EventsEmit) {
                window.runtime.EventsEmit('file-registration-error', {
                    error: String(error)
                });
            } else {
                // Fallback: close progress modal and show error in selection modal
                showProgressModal = false;
                modalError = String(error);
                showRegisterModal = true;
            }
        } finally {
            modalProcessing = false;
        }
    }

    function onRegistrationComplete() {
        console.log("Directory registration completed successfully");
        // Refresh the list after completion
        fetchPage(false);
    }

    async function handleClearAll() {
        modalProcessing = true;
        try {
            await window.go.main.App.ClearAllRegisteredFiles();
            closeClearModal();
            // Refresh the list
            await fetchPage(false);
        } catch (error) {
            modalError = error;
        } finally {
            modalProcessing = false;
        }
    }

    async function setDirectory(newPath) {
        console.log("setDirectory called with:", newPath);
        
        if (!newPath || newPath.trim() === "") {
            modalError = "Invalid directory path selected";
            return;
        }
        selectedDirectory = String(newPath).trim();
        console.log("selectedDirectory set to:", selectedDirectory);
    }
</script>

<div class="container">
    <h1>File Registry</h1>
    
    <div class="explanation">
        <p>
            <strong>What is File Registry?</strong><br>
            This feature allows you to register files from your existing collection to prevent duplicate downloads. 
            When you register a directory, all files (including those in subfolders) are catalogued by filename and MD5 hash. 
            During automatic downloads, any new file with the same MD5 hash as a registered file will be treated as a duplicate and skipped, 
            helping you avoid downloading content you already have.
        </p>
    </div>

    <div class="actions">
        <button class="register-btn" onclick={openRegisterModal}>Register Directory</button>
        <button class="clear-btn" onclick={openClearModal}>Clear All Registered Files</button>
    </div>

    {#if loading}
        <div class="center"><LoadingSpinner size="3rem" /></div>
    {:else if error}
        <p class="error">Error: {error}</p>
    {:else}
        <div class="files-list">
            {#if registeredFiles.length === 0}
                <div class="empty-state">No registered files</div>
            {:else}
                {#each registeredFiles as file}
                    <div class="file-item">
                        <div class="file-icon">📄</div>
                        <div class="content">
                            <div class="filename">{file.filename}</div>
                            <div class="file-path">{file.file_path}</div>
                            <div class="meta-section">
                                <div class="hash">MD5: {file.md5}</div>
                                <div class="timestamp">{formatTimestamp(file.registered_at)}</div>
                            </div>
                        </div>
                    </div>
                {/each}
            {/if}
        </div>

        <div class="pagination">
            <button onclick={prevPage} disabled={offset === 0}>Previous</button>
            <div class="page-info">Page {pageNumber}</div>
            <button onclick={nextPage} disabled={nextDisabled}>Next</button>
        </div>
    {/if}
</div>

<!-- Register Directory Modal -->
<dialog id="register-directory-dialog">
    {#if modalProcessing}
        <div class="modal-processing">
            <LoadingSpinner />
            <p>Registering directory...</p>
        </div>
    {:else if modalError}
        <button class="dialog-close-btn" onclick={closeRegisterModal}>✕</button>
        <p class="error-message">Error: {modalError}</p>
        <div class="modal-actions">
            <button onclick={closeRegisterModal}>Close</button>
        </div>
    {:else}
        <button class="dialog-close-btn" onclick={closeRegisterModal}>✕</button>
        
        <h2>Register Directory</h2>
        <p>Select a directory to register all files for duplicate detection.</p>
        
        <div class="form-group">
            <label for="directory">Directory</label>
            <div class="input-group">
                <input id="directory" type="text" bind:value={selectedDirectory} />
                <SelectDirectoryButton
                    text="Browse"
                    clickHandlerAsync={setDirectory}
                    style="padding: 0.5rem 1rem; background-color: #555; border: 1px solid #777; color: white; border-radius: 4px;" />
            </div>
        </div>
        
        <div class="modal-actions">
            <button onclick={closeRegisterModal}>Cancel</button>
            <button class="primary" onclick={handleRegisterDirectory} disabled={!selectedDirectory || modalProcessing}>Register Directory</button>
        </div>
    {/if}
</dialog>

<!-- Clear All Modal -->
<dialog id="clear-all-dialog">
    {#if modalProcessing}
        <div class="modal-processing">
            <LoadingSpinner />
            <p>Clearing registered files...</p>
        </div>
    {:else if modalError}
        <button class="dialog-close-btn" onclick={closeClearModal}>✕</button>
        <p class="error-message">Error: {modalError}</p>
        <div class="modal-actions">
            <button onclick={closeClearModal}>Close</button>
        </div>
    {:else}
        <button class="dialog-close-btn" onclick={closeClearModal}>✕</button>
        
        <h2>Clear All Registered Files</h2>
        <p>Are you sure you want to remove all registered files? This action cannot be undone.</p>
        
        <div class="modal-actions">
            <button onclick={closeClearModal}>Cancel</button>
            <button class="danger" onclick={handleClearAll}>Clear All</button>
        </div>
    {/if}
</dialog>

<!-- File Registration Progress Modal -->
<FileRegistrationProgressModal 
  bind:isOpen={showProgressModal}
  onComplete={onRegistrationComplete}
/>

<style>
    .container { 
        max-width: 900px; 
        margin: 1.5rem auto; 
        padding: 0 1rem; 
    }
    
    h1 { 
        margin-bottom: 1rem; 
    }
    
    .explanation {
        background-color: #1e1e1e;
        border-left: 4px solid #4caf50;
        padding: 1rem;
        margin-bottom: 1.5rem;
        border-radius: 4px;
    }
    
    .explanation p {
        margin: 0;
        line-height: 1.6;
    }
    
    .actions {
        display: flex;
        gap: 1rem;
        margin-bottom: 1.5rem;
    }
    
    .register-btn {
        background-color: #4caf50;
        border: none;
        color: white;
        padding: 0.75rem 1.5rem;
        border-radius: 4px;
        font-size: 1rem;
        cursor: pointer;
    }
    
    .register-btn:hover {
        background-color: #45a049;
    }
    
    .clear-btn {
        background-color: #f44336;
        border: none;
        color: white;
        padding: 0.75rem 1.5rem;
        border-radius: 4px;
        font-size: 1rem;
        cursor: pointer;
    }
    
    .clear-btn:hover {
        background-color: #da190b;
    }
    
    .test-btn {
        background-color: #2196F3;
        border: none;
        color: white;
        padding: 0.75rem 1.5rem;
        border-radius: 4px;
        font-size: 1rem;
        cursor: pointer;
    }
    
    .test-btn:hover {
        background-color: #1976D2;
    }
    
    .center { 
        display: flex; 
        justify-content: center; 
        padding: 2rem 0; 
    }
    
    .files-list {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        border: 1px solid #2a2a2a;
        border-radius: 8px;
        overflow: hidden;
    }

    .file-item {
        display: flex;
        padding: 1rem;
        gap: 1rem;
        border-bottom: 1px solid #2a2a2a;
        background: #151515;
        align-items: flex-start;
    }

    .file-item:last-child {
        border-bottom: none;
    }

    .file-icon {
        flex-shrink: 0;
        width: 2rem;
        display: flex;
        align-items: center;
        justify-content: center;
        padding-top: 0.2rem;
        font-size: 1.2rem;
    }

    .content {
        flex: 1;
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .filename {
        font-weight: 600;
        word-break: break-word;
        color: #e0e0e0;
    }
    
    .file-path {
        font-size: 0.9rem;
        color: #9ad1ff;
        word-break: break-all;
        font-family: monospace;
    }

    .meta-section {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 1rem;
        border-top: 1px solid #2a2a2a;
        padding-top: 0.5rem;
        margin-top: 0.5rem;
    }

    .hash {
        font-family: monospace;
        font-size: 0.8rem;
        color: #999;
        word-break: break-all;
        flex: 1;
    }

    .timestamp {
        color: #666;
        font-size: 0.85rem;
        white-space: nowrap;
    }

    .empty-state {
        text-align: center;
        padding: 2rem;
        color: #999;
        font-style: italic;
    }

    .pagination {
        display: flex;
        gap: 1rem;
        align-items: center;
        margin-top: 1rem;
    }

    .pagination button {
        padding: 0.45rem 0.75rem;
        border-radius: 6px;
    }

    .page-info {
        color: #ccc;
        font-size: 0.95rem;
    }

    .error {
        color: #ff6666;
        margin: 1rem 0;
    }
    
    /* Modal styles */
    .modal-processing {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
        padding: 2rem;
    }
    
    .error-message {
        color: #ff6666;
        margin: 1rem 0;
    }
    
    .form-group {
        margin-bottom: 1rem;
    }
    
    .form-group label {
        display: block;
        margin-bottom: 0.5rem;
        font-weight: 600;
    }
    
    .input-group {
        display: flex;
        gap: 0.5rem;
        align-items: center;
    }
    
    .input-group input {
        flex: 1;
    }
    
    .modal-actions {
        display: flex;
        gap: 0.5rem;
        justify-content: flex-end;
        margin-top: 1.5rem;
    }
    
    .modal-actions button {
        padding: 0.5rem 1rem;
        border-radius: 4px;
        cursor: pointer;
    }
    
    .modal-actions button.primary {
        background-color: #4caf50;
        border: none;
        color: white;
    }
    
    .modal-actions button.primary:hover {
        background-color: #45a049;
    }
    
    .modal-actions button.danger {
        background-color: #f44336;
        border: none;
        color: white;
    }
    
    .modal-actions button.danger:hover {
        background-color: #da190b;
    }

    @media (max-width: 600px) {
        .file-item {
            font-size: 0.85rem;
        }
        
        .actions {
            flex-direction: column;
        }
        
        .meta-section {
            flex-direction: column;
            align-items: flex-start;
            gap: 0.5rem;
        }
    }
</style>