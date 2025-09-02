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
          GetActivePlaylists: () => Promise<Array<any>>;
          GetLegalDisclaimerAccepted: () => Promise<boolean>;
          GetConfirmCloseEnabled: () => Promise<boolean>;
          GetSettingString: (arg1: string) => Promise<string>;
          HandleFatalError: (arg1: string) => Promise<void>;
          IsStartupComplete: () => Promise<boolean>;
          OpenDirectory: (arg1: string) => Promise<void>;
          SelectDirectory: () => Promise<string>;
          SetLegalDisclaimerAccepted: (arg1: boolean) => Promise<void>;
          SetConfirmCloseEnabled: (arg1: boolean) => Promise<void>;
          SetSettingPreparsed: (arg1: string, arg2: string) => Promise<void>;
          UpdatePlaylistDirectory: (
            arg1: number,
            arg2: string
          ) => Promise<void>;
          ValidateAndAddPlaylist: (
            arg1: string,
            arg2: string,
            arg3: string
          ) => Promise<void>;
          DirectDownload: (
            arg1: string,
            arg2: string,
            arg3: string
          ) => Promise<void>;
          GetDownloadHistoryPage: (
            arg1: number,
            arg2: number,
            arg3: boolean,
            arg4: boolean
          ) => Promise<Array<any>>;
          GetRecentLogs: () => Promise<Array<any>>;
          GetDaemonLogLines: (arg1: number) => Promise<Array<string>>;
          GetUILogLines: (arg1: number) => Promise<Array<string>>;
          GetDaemonLogLinesWithLevel: (arg1: number, arg2: string) => Promise<Array<string>>;
          GetUILogLinesWithLevel: (arg1: number, arg2: string) => Promise<Array<string>>;
          SetManualRetry: (downloadId: number) => Promise<void>;
          RegisterAllFailedForRetryManual: () => Promise<void>;
          StartDaemon: () => Promise<void>;
          StopDaemon: () => Promise<void>;
          IsDaemonRunning: () => Promise<boolean>;
          CloseApplication: () => Promise<void>;
          GetRegisteredFiles: (arg1: number, arg2: number) => Promise<Array<any>>;
          RegisterDirectory: (arg1: string) => Promise<void>;
          ClearAllRegisteredFiles: () => Promise<void>;
          TestModalProgress: () => Promise<void>;
        };
      };
    };
  }
}

export {};
