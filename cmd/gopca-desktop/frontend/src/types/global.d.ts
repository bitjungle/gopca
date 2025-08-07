// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

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