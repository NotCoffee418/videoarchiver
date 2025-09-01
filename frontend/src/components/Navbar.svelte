<script>  
    // Props from parent
    let { currentRoute } = $props();
    
    // Also update router in App.svelte
    let links = [
      { name: 'Archive', path: '/' },
      { name: 'Direct', path: '/direct' },
      { name: 'History', path: '/history' },
      { name: 'Duplicates', path: '/duplicates' },
      { name: 'Status', path: '/status' },
      { name: 'Settings', path: '/settings' }
    ];
  </script>
  
  <nav class="navbar">
    {#each links as link}
    <a
        href={`#${link.path}`}
        class:active={currentRoute === link.path}
    >
        {link.name}
    </a>
    {/each}
  </nav>
  
  <style>
    .navbar {
      display: flex;
      gap: 1rem;
      background-color: #333;
      padding: 1rem;
    }

    a {
        color: white;
        text-decoration: none;
        padding: 0.5rem 1rem;
        transition: background-color 0.2s ease;
        border-radius: 4px;
    }

    a:hover {
        background-color: rgba(255, 255, 255, 0.1);
    }

    .active {
        background-color: rgba(255, 255, 255, 0.2);
    }
  </style>
  