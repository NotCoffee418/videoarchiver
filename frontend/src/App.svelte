<script>
  import './style.css';
  import Navbar from './components/Navbar.svelte';
  import Router from 'svelte-spa-router';
  
  // Import your route components
  import PlaylistsPage from './routes/PlaylistsPage.svelte';
  import LogsPage from './routes/LogsPage.svelte';
  import SettingsPage from './routes/SettingsPage.svelte';
  
  const routes = {
    '/': PlaylistsPage,
    '/logs': LogsPage,
    '/settings': SettingsPage
  };
</script>

<div class="app">
  <Navbar />
  <main>
    <Router {routes} />
  </main>
</div>

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
</style>
