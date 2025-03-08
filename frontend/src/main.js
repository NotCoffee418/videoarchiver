import './style.css'
import App from './App.svelte'

let app;

// Wait for the DOM to be ready
document.addEventListener('DOMContentLoaded', () => {
  app = new App({
    target: document.getElementById('app')
  })
})

export default app
