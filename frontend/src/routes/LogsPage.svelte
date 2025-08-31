<script>
    import LoadingSpinner from '../components/LoadingSpinner.svelte';
    import { onMount } from 'svelte';

    let logs = $state([]);
    let loading = $state(true);
    let error = $state('');

    onMount(() => {
        loadLogs();
        // Refresh logs every 10 seconds
        const interval = setInterval(loadLogs, 10000);
        return () => clearInterval(interval);
    });

    async function loadLogs() {
        try {
            loading = true;
            error = '';
            logs = await window.go.main.App.GetRecentLogs();
        } catch (err) {
            console.error('Failed to load logs:', err);
            error = `Failed to load logs: ${err.message || 'Unknown error'}`;
        } finally {
            loading = false;
        }
    }

    function formatTimestamp(timestamp) {
        if (!timestamp) return '';
        return new Date(timestamp * 1000).toLocaleString();
    }

    function getLogLevelClass(verbosity) {
        switch (verbosity) {
            case 0: return 'log-debug';
            case 1: return 'log-info';
            case 2: return 'log-warning';
            case 3: return 'log-error';
            default: return 'log-info';
        }
    }

    function getLogLevelName(verbosity) {
        switch (verbosity) {
            case 0: return 'DEBUG';
            case 1: return 'INFO';
            case 2: return 'WARN';
            case 3: return 'ERROR';
            default: return 'INFO';
        }
    }
</script>

<div class="container">
    <h1>Application Logs</h1>
    
    <div class="controls">
        <button on:click={loadLogs} disabled={loading}>
            {#if loading}
                Refreshing...
            {:else}
                Refresh Logs
            {/if}
        </button>
        <div class="info">
            Showing most recent 250 entries • Auto-refreshes every 10 seconds
        </div>
    </div>

    {#if loading && logs.length === 0}
        <div class="center"><LoadingSpinner size="3rem" /></div>
    {:else if error}
        <p class="error">Error: {error}</p>
    {:else}
        <div class="logs-container">
            {#if logs.length === 0}
                <div class="empty-state">No log entries found</div>
            {:else}
                <div class="logs-list">
                    {#each logs as log (log.id)}
                        <div class="log-entry {getLogLevelClass(log.verbosity)}">
                            <div class="log-header">
                                <span class="log-level">{getLogLevelName(log.verbosity)}</span>
                                <span class="log-timestamp">{formatTimestamp(log.timestamp)}</span>
                            </div>
                            <div class="log-message">{log.message}</div>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
    {/if}
</div>

<style>
    .container {
        max-width: 1200px;
        margin: 1.5rem auto;
        padding: 0 1rem;
    }

    h1 {
        margin-bottom: 1rem;
    }

    .controls {
        display: flex;
        align-items: center;
        gap: 1rem;
        margin-bottom: 1rem;
        padding: 1rem;
        background: #f5f5f5;
        border-radius: 8px;
    }

    .info {
        color: #666;
        font-size: 0.9rem;
    }

    button {
        padding: 0.5rem 1rem;
        border: 1px solid #ddd;
        background: white;
        border-radius: 4px;
        cursor: pointer;
        transition: background-color 0.2s;
    }

    button:hover:not(:disabled) {
        background: #f0f0f0;
    }

    button:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .center {
        display: flex;
        justify-content: center;
        padding: 3rem 0;
    }

    .error {
        color: #ff6b6b;
        background: #ffe0e0;
        padding: 1rem;
        border-radius: 4px;
        border: 1px solid #ffcccc;
    }

    .empty-state {
        text-align: center;
        color: #666;
        padding: 3rem;
        background: #f9f9f9;
        border-radius: 8px;
    }

    .logs-container {
        background: white;
        border: 1px solid #ddd;
        border-radius: 8px;
        overflow: hidden;
    }

    .logs-list {
        max-height: 70vh;
        overflow-y: auto;
    }

    .log-entry {
        padding: 0.75rem 1rem;
        border-bottom: 1px solid #eee;
        font-family: 'Courier New', monospace;
        font-size: 0.85rem;
    }

    .log-entry:last-child {
        border-bottom: none;
    }

    .log-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 0.25rem;
    }

    .log-level {
        font-weight: bold;
        padding: 0.2rem 0.5rem;
        border-radius: 3px;
        font-size: 0.75rem;
    }

    .log-timestamp {
        color: #666;
        font-size: 0.75rem;
    }

    .log-message {
        word-break: break-word;
        margin-left: 0.5rem;
    }

    .log-debug {
        background: #f8f9fa;
        border-left: 3px solid #6c757d;
    }

    .log-debug .log-level {
        background: #6c757d;
        color: white;
    }

    .log-info {
        background: #f8f9ff;
        border-left: 3px solid #0066cc;
    }

    .log-info .log-level {
        background: #0066cc;
        color: white;
    }

    .log-warning {
        background: #fff8e1;
        border-left: 3px solid #ff9800;
    }

    .log-warning .log-level {
        background: #ff9800;
        color: white;
    }

    .log-error {
        background: #ffebee;
        border-left: 3px solid #f44336;
    }

    .log-error .log-level {
        background: #f44336;
        color: white;
    }
</style>