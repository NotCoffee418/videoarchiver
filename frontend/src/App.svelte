<script>
  import './style.css';
  import Navbar from './components/Navbar.svelte';
  import Router from 'svelte-spa-router';
  import { onMount } from 'svelte';
  
  // Import your route components
  import PlaylistsPage from './routes/PlaylistsPage.svelte';
  import LogsPage from './routes/LogsPage.svelte';
  import SettingsPage from './routes/SettingsPage.svelte';
  import LoadingSpinner from './components/LoadingSpinner.svelte';

  const routes = {
    '/': PlaylistsPage,
    '/logs': LogsPage,
    '/settings': SettingsPage
  };

  let isRuntimeReady = $state(false);
  let hasError = $state(false);

  // Listen for wails ready event
  onMount(() => {
    // Check if we're in a Wails context
    if (typeof window !== 'undefined') {
      // First check if runtime is already available
      if (window.runtime) {
        isRuntimeReady = true;
        return;
      }

      // If not, wait for the 'wails:ready' event
      const readyHandler = () => {
        console.log("Wails runtime is ready");
        isRuntimeReady = true;
        // Clean up the event listener
        document.removeEventListener('wails:ready', readyHandler);
      };

      document.addEventListener('wails:ready', readyHandler);

      // Add a timeout of 10 seconds
      setTimeout(() => {
        if (!isRuntimeReady) {
          console.error("Wails runtime failed to initialize within timeout");
          hasError = true;
        }
      }, 10000);
    }
  });
</script>

{#if hasError}
  <div class="error-container">
    <div class="error-icon">⚠️</div>
    <h2>Runtime Initialization Error</h2>
    <p>Something went wrong starting the application. Please restart the application.</p>
  </div>
{:else if !isRuntimeReady}
  <LoadingSpinner />
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
</style>
