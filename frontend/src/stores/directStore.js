import { writable } from 'svelte/store';

// Store for DirectPage state
export const directState = writable({
    format: "mp3",
    url: "",
    directory: "",
    isDownloading: false,
    error: ""
});

// Helper function to update specific properties
export function updateDirectState(updates) {
    directState.update(state => ({
        ...state,
        ...updates
    }));
}

// Reset function for clearing state when needed
export function resetDirectState() {
    directState.set({
        format: "mp3",
        url: "",
        directory: "",
        isDownloading: false,
        error: ""
    });
}