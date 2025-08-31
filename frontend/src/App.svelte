<script>
  import './style.css';
  import Navbar from './components/Navbar.svelte';
  import { onMount } from 'svelte';
  
  // Import your route components
  import ArchivePage from './routes/ArchivePage.svelte';
  import DirectPage from './routes/DirectPage.svelte';
  import StatusPage from './routes/StatusPage.svelte';
  import SettingsPage from './routes/SettingsPage.svelte';
  import LoadingSpinner from './components/LoadingSpinner.svelte';
  import HistoryPage from './routes/HistoryPage.svelte';
  import LegalDisclaimer from './components/LegalDisclaimer.svelte';

  // Simple hash-based routing that maintains component state
  let currentRoute = $state('/');

  // Components are rendered once and kept alive
  const components = {
    '/': { component: ArchivePage, instance: null },
    '/direct': { component: DirectPage, instance: null },
    '/history': { component: HistoryPage, instance: null },
    '/status': { component: StatusPage, instance: null },
    '/settings': { component: SettingsPage, instance: null }
  };

  let isRuntimeReady = $state(false);
  let hasError = $state(false);
  let startupComplete = $state(false);
  let loadingText = $state("Initializing Application...");
  let disclaimerAccepted = $state(null); // null = checking, false = not accepted, true = accepted

  // Check disclaimer acceptance status
  async function checkDisclaimerAcceptance() {
    try {
      const accepted = await window.go?.main?.App?.GetLegalDisclaimerAccepted();
      disclaimerAccepted = accepted;
    } catch (error) {
      console.error('Failed to check disclaimer acceptance:', error);
      // Default to false if we can't check
      disclaimerAccepted = false;
    }
  }

  // Handle disclaimer acceptance
  async function handleDisclaimerAccept() {
    try {
      await window.go?.main?.App?.SetLegalDisclaimerAccepted(true);
      disclaimerAccepted = true;
    } catch (error) {
      console.error('Failed to save disclaimer acceptance:', error);
    }
  }

  // Hash-based routing
  function updateRoute() {
    const hash = window.location.hash.replace('#', '') || '/';
    if (components[hash]) {
      currentRoute = hash;
    }
  }

  // Listen for wails ready event
  onMount(async () => {
    // Check if we're in a Wails context
    if (typeof window == 'undefined') {      
      console.error("Not in a Wails context");
      hasError = true;
      return;
    }

    // Setup hash routing
    updateRoute();
    window.addEventListener('hashchange', updateRoute);

    if (window.runtime) {
      // On refresh
      isRuntimeReady = true;
      startupComplete = await window.go?.main?.App?.IsStartupComplete();
      // Check disclaimer after runtime is ready
      await checkDisclaimerAcceptance();
    } else {
      // On startup
      // If not, wait for the 'wails:ready' event
      const wailsReadyHandler = async () => {
        isRuntimeReady = true;
        document.removeEventListener('wails:ready', wailsReadyHandler);
        // Check disclaimer after wails is ready
        await checkDisclaimerAcceptance();
      };
      document.addEventListener('wails:ready', wailsReadyHandler);

    }

    // Report startup progress
    window.runtime.EventsOn('startup-progress', (progress) => {
      loadingText = progress;
      console.log("Startup progress:", progress);
    });

    // Startup complete event
    window.runtime.EventsOn('startup-complete', () => {
      startupComplete = true;
      window.runtime.EventsEmit('startup-complete-confirmed');
      console.log("Startup complete");
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
    }, 120000);

  });
</script>

{#if hasError}
  <div class="error-container">
    <div class="error-icon">⚠️</div>
    <h2>Runtime Initialization Error</h2>
    <p>Something went wrong starting the application. Please restart the application.</p>
  </div>
{:else if !isRuntimeReady || disclaimerAccepted === null}
  <div class="app">
    <div class="loader">
      <LoadingSpinner />
      <p class="loading-text">Initializing Application...</p>
    </div>
  </div>
{:else if disclaimerAccepted === false}
  <LegalDisclaimer onAccept={handleDisclaimerAccept} />
{:else if !startupComplete}
  <div class="app">
    <div class="loader">
      <LoadingSpinner />
      <p class="loading-text">{loadingText}</p>
    </div>
  </div>
{:else}
  <div class="app">
    <Navbar {currentRoute} />
    <main>
      {#each Object.entries(components) as [path, { component: Component }]}
        <div class="route-container" class:active={currentRoute === path}>
          <Component />
        </div>
      {/each}
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
    position: relative;
  }

  .route-container {
    display: none;
  }

  .route-container.active {
    display: block;
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
