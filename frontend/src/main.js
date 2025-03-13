import './style.css'
import App from './App.svelte'
import { mount } from "svelte";

let app;

// Wait for the DOM to be ready
document.addEventListener('DOMContentLoaded', () => {
  app = mount(App, {
      target: document.getElementById('app')
    })
})

export default app
