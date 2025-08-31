import { writable } from 'svelte/store';

// Store for HistoryPage state
export const historyState = writable({
    downloads: [],
    offset: 0,
    limit: 10,
    loading: false,
    error: "",
    showFailed: true,
    showSuccessful: true,
    retrying: new Set(),
    retryAllCooldown: false
});

// Helper function to update specific properties
export function updateHistoryState(updates) {
    historyState.update(state => ({
        ...state,
        ...updates
    }));
}

// Reset function for clearing state when needed
export function resetHistoryState() {
    historyState.set({
        downloads: [],
        offset: 0,
        limit: 10,
        loading: false,
        error: "",
        showFailed: true,
        showSuccessful: true,
        retrying: new Set(),
        retryAllCooldown: false
    });
}