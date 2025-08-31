<script>  
  // Props from parent
  let { currentPage = '/', navigateToPage } = $props();
  
  let links = [
    { name: 'Archive', path: '/' },
    { name: 'Direct', path: '/direct' },
    { name: 'History', path: '/history' },
    { name: 'Status', path: '/status' },
    { name: 'Settings', path: '/settings' }
  ];
  
  function handleNavigation(path) {
    navigateToPage(path);
  }
</script>
  
  <nav class="navbar">
    {#each links as link}
    <button
        class="nav-button"
        class:active={currentPage === link.path}
        onclick={() => handleNavigation(link.path)}
    >
        {link.name}
    </button>
    {/each}
  </nav>
  
  <style>
    .navbar {
      display: flex;
      gap: 1rem;
      background-color: #333;
      padding: 1rem;
    }

    .nav-button {
        color: white;
        background: none;
        border: none;
        padding: 0.5rem 1rem;
        transition: background-color 0.2s ease;
        border-radius: 4px;
        cursor: pointer;
        font-size: inherit;
        font-family: inherit;
    }

    .nav-button:hover {
        background-color: rgba(255, 255, 255, 0.1);
    }

    .active {
        background-color: rgba(255, 255, 255, 0.2);
    }
  </style>
  