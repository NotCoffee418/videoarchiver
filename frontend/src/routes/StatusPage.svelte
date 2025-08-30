<script>
    import DaemonManagement from '../components/DaemonManagement.svelte';
</script>

<div class="container">
    <h1>Status</h1>

    <section class="daemon-section">
        <DaemonManagement />
    </section>

    <section class="logs-section">
        <h2>Logs</h2>
        <div class="logs-display">
            Logs display not implemented yet
        </div>
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

    .logs-display {
        min-height: 400px;
        padding: 1rem;
        background: #151515;
        border: 1px solid #2a2a2a;
        border-radius: 8px;
        font-family: monospace;
        color: #999;
        display: flex;
        align-items: center;
        justify-content: center;
    }
</style>