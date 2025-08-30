<script>
    import { onMount, onDestroy } from "svelte";
    import LoadingSpinner from "../components/LoadingSpinner.svelte";

    let downloads = [];
    let offset = 0;
    let limit = 2; // Changed from 10 to 2 for testing
    let loading = false;
    let error = "";

    // ids currently retrying
    let retrying = new Set();

    let showFailed = true;
    let showSuccessful = true;

    function statusLabel(s) {
        switch (s) {
            case 1: return "Success";
            case 2: return "Failed (retry)";
            case 3: return "Failed (give up)";
            default: return "Unknown";
        }
    }

    async function copyToClipboard(text) {
        if (!text) return;
        try { await navigator.clipboard.writeText(text); }
        catch (e) { console.error("copy failed", e); }
    }

    function onRetry(d) {
        if (retrying.has(d.id)) return;
        retrying = new Set([...retrying, d.id]);
        // placeholder UI-only retry simulation
        setTimeout(() => {
            const s = new Set(retrying);
            s.delete(d.id);
            retrying = s;
        }, 1500);
    }

    async function fetchPage(showLoading = true) {
        if (showLoading) loading = true;
        error = "";
        try {
            if (window?.go?.main?.App?.GetDownloadHistoryPage) {
                const res = await window.go.main.App.GetDownloadHistoryPage(offset, limit, showSuccessful, showFailed);
                downloads = Array.isArray(res) ? res : [];
            } else {
                downloads = [];
            }
        } catch (err) {
            error = String(err ?? "Unknown error");
            downloads = [];
        } finally {
            if (showLoading) loading = false;
        }
    }

    function onFilterChange() {
        offset = 0; // Reset to first page
        fetchPage(true);
    }

    function prevPage() {
        if (offset === 0) return;
        offset = Math.max(0, offset - limit);
        fetchPage(true); // Show loading on manual navigation
    }

    function nextPage() {
        offset += limit;
        fetchPage(true); // Show loading on manual navigation
    }

    let refreshInterval;
    const REFRESH_RATE = 5000; // 5 seconds

    onMount(() => {
        fetchPage(true); // Show loading on initial load
        refreshInterval = setInterval(() => fetchPage(false), REFRESH_RATE); // Hide loading on auto-refresh
    });

    onDestroy(() => {
        if (refreshInterval) {
            clearInterval(refreshInterval);
        }
    });

    $: nextDisabled = downloads.length < limit;
    $: pageNumber = Math.floor(offset / limit) + 1;

    $: if (showFailed !== undefined && showSuccessful !== undefined) {
        onFilterChange();
    }

    function displayTitle(d) {
        return d.output_filename?.String ?? (d.url?.split?.("/").pop() ?? "Untitled");
    }

    function formatTimestamp(ts) {
        if (!ts) return "";
        const n = Number(ts);
        if (isNaN(n)) return "";
        return new Date(n * 1000).toLocaleString(); // multiply by 1000 to convert seconds to milliseconds
    }
</script>

<div class="container">
    <h1>Download History</h1>

    <div class="filters">
        <label>
            <input type="checkbox" bind:checked={showSuccessful}>
            Show Successful
        </label>
        <label>
            <input type="checkbox" bind:checked={showFailed}>
            Show Failed
        </label>
    </div>

    {#if loading}
        <div class="center"><LoadingSpinner size="3rem" /></div>
    {:else if error}
        <p class="error">Error: {error}</p>
    {:else}
        <div class="history-list">
            {#if downloads.length === 0}
                <div class="empty-state">No history</div>
            {:else}
                {#each downloads as d}
                    <div class="history-item {d.status === 1 ? 'success' : 'failed'}">
                        <div class="status-ico" title={statusLabel(d.status)}>
                            {#if d.status === 1}✅{:else}❌{/if}
                        </div>

                        {#if d.status === 1}
                            <!-- Success layout -->
                            <div class="content">
                                <div class="title">{displayTitle(d)}</div>
                                <div class="actions">
                                    <a href="/" on:click|preventDefault={() => copyToClipboard(d.url)}>Copy URL</a>
                                    <span class="separator">|</span>
                                    <a href="/" on:click|preventDefault={() => copyToClipboard(d.output_filename.String)}>Copy File Path</a>
                                </div>
                                <div class="timestamp">{formatTimestamp(d.last_attempt)}</div>
                            </div>
                        {:else}
                            <!-- Failed layout -->
                            <div class="content">
                                <div class="url-preview">{d.url}</div>
                                <div class="retry-row">
                                    <button class="retry-btn" on:click={() => onRetry(d)} disabled={retrying.has(d.id)}>
                                        {#if retrying.has(d.id)}Retrying...{:else}Retry Download{/if}
                                    </button>
                                    <span class="attempts">({d.attempt_count} attempts)</span>
                                </div>
                                <div class="actions">
                                    <a href="/" on:click|preventDefault={() => copyToClipboard(d.url)}>Copy URL</a>
                                    {#if d.output_filename?.Valid}
                                        <span class="separator">|</span>
                                        <a href="/" on:click|preventDefault={() => copyToClipboard(d.output_filename.String)}>Copy File Path</a>
                                    {/if}
                                </div>
                                {#if d.fail_message?.String}
                                    <div class="error-message">{d.fail_message.String}</div>
                                {/if}
                                <div class="timestamp">{formatTimestamp(d.last_attempt)}</div>
                            </div>
                        {/if}
                    </div>
                {/each}
            {/if}
        </div>

        <div class="pagination">
            <button on:click={prevPage} disabled={offset === 0}>Previous</button>
            <div class="page-info">Page {pageNumber}</div>
            <button on:click={nextPage} disabled={nextDisabled}>Next</button>
        </div>
    {/if}
</div>

<style>
    .container { max-width: 900px; margin: 1.5rem auto; padding: 0 1rem; }
    h1 { margin-bottom: 1rem; }
    .center { display: flex; justify-content: center; padding: 2rem 0; }
    
    .history-list {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        border: 1px solid #2a2a2a;
        border-radius: 8px;
        overflow: hidden;
    }

    .history-item {
        display: flex;
        padding: 0.75rem;
        gap: 1rem;
        border-bottom: 1px solid #2a2a2a;
        background: #151515;
    }

    .history-item:last-child {
        border-bottom: none;
    }

    .status-ico {
        flex-shrink: 0;
        width: 2rem;
        font-size: 1.25rem;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    .content {
        flex: 1;
        min-width: 0; /* Enables text truncation */
    }

    .success .content {
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
    }

    .failed .content {
        display: grid;
        gap: 0.75rem;
    }

    .title {
        font-weight: 600;
        word-break: break-word;
    }

    .actions {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        font-size: 0.9rem;
    }

    .actions a {
        color: #9ad1ff;
        text-decoration: underline;
    }

    .separator {
        color: #666;
    }

    .timestamp {
        color: #666;
        font-size: 0.85rem;
        margin-top: 0.25rem;
    }

    .url-preview {
        color: #999;
        font-size: 0.9rem;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        margin-bottom: 0.5rem;
    }

    .retry-btn {
        padding: 0.3rem 0.6rem;
        font-size: 0.9rem;
        margin-right: 0.5rem;
    }

    .attempts {
        color: #ccc;
        font-size: 0.9rem;
    }

    .error-message {
        color: #ff6666;
        font-family: monospace;
        font-size: 0.9rem;
        white-space: pre-wrap;
        word-break: break-word;
        padding: 0.5rem;
        background: rgba(255, 0, 0, 0.1);
        border-radius: 4px;
    }

    .empty-state {
        text-align: center;
        padding: 1rem;
        color: #999;
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

    .filters {
        display: flex;
        gap: 1rem;
        margin-bottom: 1rem;
    }

    .filters label {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        cursor: pointer;
    }

    .filters input[type="checkbox"] {
        width: 1rem;
        height: 1rem;
        cursor: pointer;
    }

    @media (max-width: 600px) {
        .history-item {
            font-size: 0.85rem;
        }
    }
</style>