<script>
    import { onMount, onDestroy } from "svelte";

    let isDaemonRunning = false;
    let isProcessing = false;
    let error = '';
    let checkInterval;

    async function checkDaemonStatus() {
        try {
            const status = await window.go.main.App.IsDaemonRunning();
            if (status !== isDaemonRunning) {
                isDaemonRunning = status;
                isProcessing = false; // Clear processing state when status changes
                error = '';
            }
        } catch (err) {
            console.error("Failed to check daemon status:", err);
            error = "Failed to check daemon status";
        }
    }

    async function onStart() {
        if (isProcessing || isDaemonRunning) return;
        isProcessing = true;
        error = '';
        
        try {
            await window.go.main.App.StartDaemon();
            // Only poll if start command succeeded
            for (let i = 0; i < 6 && !isDaemonRunning; i++) {
                await checkDaemonStatus();
                if (isDaemonRunning) break;
                await new Promise(r => setTimeout(r, 5000));
            }
            
            if (!isDaemonRunning) {
                error = "Failed to start daemon - service did not start within 30 seconds";
                isProcessing = false;
            }
        } catch (err) {
            console.error("Failed to start daemon:", err);
            error = `Failed to start daemon: ${err.message || 'Unknown error'}`;
            isProcessing = false;
        }
    }

    async function onStop() {
        if (isProcessing || !isDaemonRunning) return;
        isProcessing = true;
        error = '';
        
        try {
            await window.go.main.App.StopDaemon();
            // Only poll if stop command succeeded
            for (let i = 0; i < 6 && isDaemonRunning; i++) {
                await checkDaemonStatus();
                if (!isDaemonRunning) break;
                await new Promise(r => setTimeout(r, 5000));
            }
            
            if (isDaemonRunning) {
                error = "Failed to stop daemon - service did not stop within 30 seconds";
                isProcessing = false;
            }
        } catch (err) {
            console.error("Failed to stop daemon:", err);
            error = `Failed to stop daemon: ${err.message || 'Unknown error'}`;
            isProcessing = false;
        }
    }

    onMount(() => {
        checkDaemonStatus();
        checkInterval = setInterval(checkDaemonStatus, 5000);
    });

    onDestroy(() => {
        if (checkInterval) {
            clearInterval(checkInterval);
        }
    });
</script>

<div class="daemon-row">
    <button 
        class="control-btn" 
        onclick={onStart} 
        disabled={isProcessing || isDaemonRunning}>
        {#if isProcessing && !isDaemonRunning}Starting...{:else}Start Daemon{/if}
    </button>

    <button 
        class="control-btn" 
        onclick={onStop} 
        disabled={isProcessing || !isDaemonRunning}>
        {#if isProcessing && isDaemonRunning}Stopping...{:else}Stop Daemon{/if}
    </button>

    <div class="status">
        <span class="status-dot" class:active={isDaemonRunning}></span>
        Daemon {isDaemonRunning ? 'Running' : 'Stopped'}
    </div>

    {#if error}
        <div class="error">{error}</div>
    {/if}
</div>

<style>
    .daemon-row {
        display: flex;
        align-items: center;
        gap: 1rem;
        padding: 1rem 0;
    }

    .control-btn {
        min-width: 120px;
        padding: 0.5rem 1rem;
    }

    .status {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        font-size: 0.95rem;
    }

    .status-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #ff4444;
        transition: background-color 0.2s ease;
    }

    .status-dot.active {
        background: #44ff44;
    }

    .error {
        color: #ff6666;
        font-size: 0.9rem;
        margin-left: 1rem;
    }
</style>