<script>
    import DaemonManagement from '../components/DaemonManagement.svelte';
    import { onMount } from 'svelte';

    let daemonLogs = $state([]);
    let uiLogs = $state([]);
    let activeTab = $state('daemon');
    let loading = $state(false);
    let error = $state('');
    let minLogLevel = $state('info'); // Default to INFO level

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
            
            // Load both daemon and UI logs with level filtering
            const [daemonResult, uiResult] = await Promise.all([
                window.go.main.App.GetDaemonLogLinesWithLevel(100, minLogLevel),
                window.go.main.App.GetUILogLinesWithLevel(100, minLogLevel)
            ]);
            
            daemonLogs = daemonResult || [];
            uiLogs = uiResult || [];
        } catch (err) {
            console.error('Failed to load logs:', err);
            error = `Failed to load logs: ${err.message || 'Unknown error'}`;
        } finally {
            loading = false;
        }
    }

    function selectTab(tab) {
        activeTab = tab;
    }

    function onLogLevelChange(newLevel) {
        minLogLevel = newLevel;
        loadLogs(); // Reload logs with new level
    }

    function formatLogLine(line) {
        if (!line) return '';
        
        // Try to parse JSON formatted logs
        try {
            const parsed = JSON.parse(line);
            const timestamp = parsed.time || parsed.timestamp || '';
            const level = parsed.level || 'info';
            const message = parsed.msg || parsed.message || line;
            const time = timestamp ? new Date(timestamp).toLocaleString() : '';
            
            return {
                timestamp: time,
                level: level.toUpperCase(),
                message: message,
                raw: line
            };
        } catch {
            // If not JSON, treat as plain text
            return {
                timestamp: '',
                level: 'INFO',
                message: line,
                raw: line
            };
        }
    }

    function getLogLevelClass(level) {
        switch (level?.toLowerCase()) {
            case 'debug': return 'log-debug';
            case 'info': return 'log-info';
            case 'warn': case 'warning': return 'log-warning';
            case 'error': return 'log-error';
            case 'fatal': return 'log-fatal';
            default: return 'log-info';
        }
    }
</script>

<div class="container">
    <h1>Status</h1>

    <section class="daemon-section">
        <DaemonManagement />
    </section>

    <section class="logs-section">
        <h2>Logs</h2>
        
        <div class="tabs">
            <button 
                class="tab {activeTab === 'daemon' ? 'active' : ''}"
                onclick={() => selectTab('daemon')}
            >
                Daemon Logs
            </button>
            <button 
                class="tab {activeTab === 'ui' ? 'active' : ''}"
                onclick={() => selectTab('ui')}
            >
                UI Logs
            </button>
        </div>

        <div class="controls">
            <button onclick={loadLogs} disabled={loading}>
                {#if loading}
                    Refreshing...
                {:else}
                    Refresh Logs
                {/if}
            </button>
            
            <div class="log-level-controls">
                <span class="log-level-label">Minimum Log Level:</span>
                <div class="log-level-radios">
                    <label class="radio-label">
                        <input 
                            type="radio" 
                            name="minLogLevel" 
                            value="debug" 
                            checked={minLogLevel === 'debug'}
                            onchange={() => onLogLevelChange('debug')}
                        />
                        DEBUG
                    </label>
                    <label class="radio-label">
                        <input 
                            type="radio" 
                            name="minLogLevel" 
                            value="info" 
                            checked={minLogLevel === 'info'}
                            onchange={() => onLogLevelChange('info')}
                        />
                        INFO
                    </label>
                    <label class="radio-label">
                        <input 
                            type="radio" 
                            name="minLogLevel" 
                            value="warn" 
                            checked={minLogLevel === 'warn'}
                            onchange={() => onLogLevelChange('warn')}
                        />
                        WARN
                    </label>
                    <label class="radio-label">
                        <input 
                            type="radio" 
                            name="minLogLevel" 
                            value="error" 
                            checked={minLogLevel === 'error'}
                            onchange={() => onLogLevelChange('error')}
                        />
                        ERROR
                    </label>
                    <label class="radio-label">
                        <input 
                            type="radio" 
                            name="minLogLevel" 
                            value="fatal" 
                            checked={minLogLevel === 'fatal'}
                            onchange={() => onLogLevelChange('fatal')}
                        />
                        FATAL
                    </label>
                </div>
            </div>
            
            <div class="info">
                Showing last 100 lines • Auto-refreshes every 10 seconds
            </div>
        </div>

        {#if error}
            <p class="error">Error: {error}</p>
        {:else}
            <div class="logs-display">
                {#if activeTab === 'daemon'}
                    {#if daemonLogs.length === 0}
                        <div class="empty-state">No daemon log entries found</div>
                    {:else}
                        <div class="logs-list">
                            {#each [...daemonLogs].reverse() as logLine}
                                {@const formatted = formatLogLine(logLine)}
                                <div class="log-entry {getLogLevelClass(formatted.level)}">
                                    <div class="log-header">
                                        <span class="log-level">{formatted.level}</span>
                                        {#if formatted.timestamp}
                                            <span class="log-timestamp">{formatted.timestamp}</span>
                                        {/if}
                                    </div>
                                    <div class="log-message">{formatted.message}</div>
                                </div>
                            {/each}
                        </div>
                    {/if}
                {:else}
                    {#if uiLogs.length === 0}
                        <div class="empty-state">No UI log entries found</div>
                    {:else}
                        <div class="logs-list">
                            {#each [...uiLogs].reverse() as logLine}
                                {@const formatted = formatLogLine(logLine)}
                                <div class="log-entry {getLogLevelClass(formatted.level)}">
                                    <div class="log-header">
                                        <span class="log-level">{formatted.level}</span>
                                        {#if formatted.timestamp}
                                            <span class="log-timestamp">{formatted.timestamp}</span>
                                        {/if}
                                    </div>
                                    <div class="log-message">{formatted.message}</div>
                                </div>
                            {/each}
                        </div>
                    {/if}
                {/if}
            </div>
        {/if}
    </section>
</div>

<style>
    .container {
        max-width: 1100px;
        margin: 1.5rem auto;
        padding: 0 1rem;
    }

    h1 {
        margin-bottom: 1.5rem;
    }

    .daemon-section {
        margin-bottom: 2rem;
        padding: 1rem;
        background: #151515;
        border: 1px solid #2a2a2a;
        border-radius: 8px;
    }

    h2 {
        margin-bottom: 1rem;
        font-size: 1.25rem;
    }

    .tabs {
        display: flex;
        margin-bottom: 1rem;
        border-bottom: 1px solid #2a2a2a;
    }

    .tab {
        background: none;
        border: none;
        color: #999;
        padding: 0.75rem 1.5rem;
        cursor: pointer;
        transition: color 0.2s, border-color 0.2s;
        border-bottom: 2px solid transparent;
        font-size: 1rem;
    }

    .tab:hover {
        color: #fff;
    }

    .tab.active {
        color: #fff;
        border-bottom-color: #4CAF50;
    }

    .controls {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        margin-bottom: 1rem;
        padding: 0.75rem;
        background: #1a1a1a;
        border: 1px solid #2a2a2a;
        border-radius: 4px;
    }

    .controls button {
        padding: 0.5rem 1rem;
        border: 1px solid #2a2a2a;
        background: #151515;
        color: #fff;
        border-radius: 4px;
        cursor: pointer;
        transition: background-color 0.2s;
        align-self: flex-start;
    }

    .controls button:hover:not(:disabled) {
        background: #2a2a2a;
    }

    .controls button:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .log-level-controls {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .log-level-label {
        color: #fff;
        font-weight: bold;
        font-size: 0.9rem;
    }

    .log-level-radios {
        display: flex;
        gap: 1rem;
    }

    .radio-label {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        color: #ccc;
        font-size: 0.85rem;
        font-weight: normal;
        cursor: pointer;
    }

    .radio-label input[type="radio"] {
        margin: 0;
        cursor: pointer;
    }

    .info {
        color: #999;
        font-size: 0.9rem;
    }

    .error {
        color: #ff6b6b;
        background: #2a1a1a;
        padding: 1rem;
        border-radius: 4px;
        border: 1px solid #443333;
        margin-bottom: 1rem;
    }

    .logs-display {
        min-height: 400px;
        background: #151515;
        border: 1px solid #2a2a2a;
        border-radius: 8px;
        overflow: hidden;
    }

    .empty-state {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 400px;
        color: #999;
        font-family: monospace;
    }

    .logs-list {
        max-height: 60vh;
        overflow-y: auto;
    }

    .log-entry {
        padding: 0.75rem 1rem;
        border-bottom: 1px solid #2a2a2a;
        font-family: 'Courier New', monospace;
        font-size: 0.85rem;
        background: #1a1a1a;
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
        color: #999;
        font-size: 0.75rem;
    }

    .log-message {
        word-break: break-word;
        margin-left: 0.5rem;
        color: #ddd;
    }

    .log-debug {
        border-left: 3px solid #6c757d;
    }

    .log-debug .log-level {
        background: #6c757d;
        color: white;
    }

    .log-info {
        border-left: 3px solid #0066cc;
    }

    .log-info .log-level {
        background: #0066cc;
        color: white;
    }

    .log-warning {
        border-left: 3px solid #ff9800;
    }

    .log-warning .log-level {
        background: #ff9800;
        color: white;
    }

    .log-error {
        border-left: 3px solid #f44336;
    }

    .log-error .log-level {
        background: #f44336;
        color: white;
    }

    .log-fatal {
        border-left: 3px solid #d32f2f;
    }

    .log-fatal .log-level {
        background: #d32f2f;
        color: white;
    }
</style>