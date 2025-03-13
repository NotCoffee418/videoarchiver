<script>
    /**
     * @typedef {Object} Props
     * @property {string} [name]
     */

    /** @type {Props} */
    let { name = "World" } = $props();
</script>

<h1>Hello {name}!</h1>
<p>This is the settings page.</p>
