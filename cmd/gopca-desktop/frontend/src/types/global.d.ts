declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          SaveFile?: (fileName: string, dataUrl: string) => Promise<void>;
        };
      };
    };
  }
}

export {};