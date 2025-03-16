<script>
  import './style.css';
  import Navbar from './components/Navbar.svelte';
  import Router from 'svelte-spa-router';
  import { onMount } from 'svelte';
  
  // Import your route components
  import ArchivePage from './routes/ArchivePage.svelte';
  import DirectPage from './routes/DirectPage.svelte';
  import LogsPage from './routes/LogsPage.svelte';
  import SettingsPage from './routes/SettingsPage.svelte';
  import LoadingSpinner from './components/LoadingSpinner.svelte';

  const routes = {
    '/': ArchivePage,
    '/direct': DirectPage,
    '/logs': LogsPage,
    '/settings': SettingsPage
  };

  let isRuntimeReady = $state(false);
  let hasError = $state(false);
  let startupComplete = $state(false);

  // Listen for wails ready event
  onMount(async () => {
    // Check if we're in a Wails context
    if (typeof window == 'undefined') {      
      console.error("Not in a Wails context");
      hasError = true;
      return;
    }

    if (window.runtime) {
      // On refresh
      isRuntimeReady = true;
      startupComplete = await window.go?.main?.App?.IsStartupComplete();
    } else {
      // On startup
      // If not, wait for the 'wails:ready' event
      const wailsReadyHandler = () => {
        isRuntimeReady = true;
        document.removeEventListener('wails:ready', wailsReadyHandler);
      };
      document.addEventListener('wails:ready', wailsReadyHandler);
    }

    // Startup complete event
    window.runtime.EventsOn('startup-complete', () => {
      startupComplete = true;
      window.runtime.EventsEmit('startup-complete-confirmed');
    });

    // Add a timeout of 10 seconds
    setTimeout(() => {
      if (!isRuntimeReady) {
        console.error("Wails runtime failed to initialize within timeout");
        hasError = true;
      }
    }, 10000);

    // Error if we never get the startup complete event
    setTimeout(() => {
      if (!startupComplete) {
        console.error("Startup failed to complete within timeout");
        hasError = true;
      }
    }, 30000);

  });
</script>

{#if hasError}
  <div class="error-container">
    <div class="error-icon">⚠️</div>
    <h2>Runtime Initialization Error</h2>
    <p>Something went wrong starting the application. Please restart the application.</p>
  </div>
{:else if !isRuntimeReady || !startupComplete}
  <div class="app">
    <div class="loader">
      <LoadingSpinner />
      <p class="loading-text">Loading...</p>
    </div>
  </div>
{:else}
  <div class="app">
    <Navbar />
    <main>
      <Router {routes} />
    </main>
  </div>
{/if}

<style>
  .app {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  main {
    flex: 1;
    padding: 1.5rem;
    background-color: #121212;
    color: #fff;
  }

  .error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100vh;
    background-color: #121212;
    color: #fff;
    text-align: center;
  }

  .error-container {
    padding: 2rem;
  }

  .error-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
  }

  .error-container h2 {
    color: #ff5555;
    margin: 0 0 1rem 0;
  }

  .error-container p {
    margin: 0.5rem 0;
    opacity: 0.8;
  }

  .loader {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    height: 100vh;
    background-color: #121212;
    color: #fff;
  }

  .loading-text {
    margin-top: 2rem;
    font-size: 1.5rem;
    opacity: 0.8;
  }
</style>
