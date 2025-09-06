<script>
    /** @type {string} */
    export let text = "Select Directory";
    
    /** @type {(path: string) => Promise<any>} */
    export let clickHandlerAsync = async (path) => {
        console.warn("No handler provided. Path: ", path);
        return null;
    };

    async function selectDirectory() {
      try {
        /** @type {string} */
        const path = await window.go?.main?.App?.SelectDirectory();
        console.log("SelectDirectory returned:", path);
        return await clickHandlerAsync(path);
      } catch (error) {
        console.error("Failed to open directory:", error);
        return null;
      }
    }
</script>

<button onclick={selectDirectory} {...$$props}>{text}</button>

<style>
    button {
        cursor: pointer;
    }
</style>