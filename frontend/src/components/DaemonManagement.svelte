<script>
    import { onMount, onDestroy } from "svelte";

    let isDaemonRunning = false;
    let startCooldown = false;
    let stopCooldown = false;
    let checkInterval;
    let error = '';

    async function checkDaemonStatus() {
        try {
            const status = await window.go.main.App.IsDaemonRunning();
            // Only update if status actually changed to avoid unnecessary rerenders
            if (status !== isDaemonRunning) {
                isDaemonRunning = status;
                error = ''; // Clear any previous errors on successful status change
            }
        } catch (err) {
            console.error("Failed to check daemon status:", err);
            error = "Failed to check daemon status";
        }
    }

    async function onStart() {
        if (startCooldown) return;
        startCooldown = true;
        error = '';
        
        try {
            await window.go.main.App.StartDaemon();
            // Poll status with longer intervals for service startup
            for (let i = 0; i < 6; i++) {
                await checkDaemonStatus();
                if (isDaemonRunning) break;
                await new Promise(r => setTimeout(r, 5000)); // 5 second intervals
            }
            
            if (!isDaemonRunning) {
                error = "Failed to start daemon - service did not start within 30 seconds";
            }
        } catch (err) {
            console.error("Failed to start daemon:", err);
            error = `Failed to start daemon: ${err.message || 'Unknown error'}`;
        } finally {
            setTimeout(() => {
                startCooldown = false;
            }, 10000);
        }
    }

    async function onStop() {
        if (stopCooldown) return;
        stopCooldown = true;
        error = '';
        
        try {
            await window.go.main.App.StopDaemon();
            // Poll status with longer intervals for service shutdown
            for (let i = 0; i < 6; i++) {
                await checkDaemonStatus();
                if (!isDaemonRunning) break;
                await new Promise(r => setTimeout(r, 5000)); // 5 second intervals
            }
            
            if (isDaemonRunning) {
                error = "Failed to stop daemon - service did not stop within 30 seconds";
            }
        } catch (err) {
            console.error("Failed to stop daemon:", err);
            error = `Failed to stop daemon: ${err.message || 'Unknown error'}`;
        } finally {
            setTimeout(() => {
                stopCooldown = false;
            }, 10000);
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
        on:click={onStart} 
        disabled={startCooldown || isDaemonRunning}>
        {#if startCooldown}Starting...{:else}Start Daemon{/if}
    </button>

    <button 
        class="control-btn" 
        on:click={onStop} 
        disabled={stopCooldown || !isDaemonRunning}>
        {#if stopCooldown}Stopping...{:else}Stop Daemon{/if}
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