/// <reference types="svelte" />
/// <reference types="vite/client" />

// Wails type declarations
declare global {
  interface Window {
    go: {
      main: {
        App: {
          DeletePlaylist: (arg1: number) => Promise<void>;
          GetClipboard: () => Promise<string>;
          GetPlaylists: () => Promise<Array<any>>;
          GetSettingString: (arg1: string) => Promise<string>;
          HandleFatalError: (arg1: string) => Promise<void>;
          IsStartupComplete: () => Promise<boolean>;
          OpenDirectory: (arg1: string) => Promise<void>;
          SelectDirectory: () => Promise<string>;
          SetSettingPreparsed: (arg1: string, arg2: string) => Promise<void>;
          UpdatePlaylistDirectory: (arg1: number, arg2: string) => Promise<void>;
          ValidateAndAddPlaylist: (arg1: string, arg2: string, arg3: string) => Promise<void>;
          DirectDownload: (arg1: string, arg2: string, arg3: string) => Promise<void>;
        };
      };
    };
  }
}

export {};
