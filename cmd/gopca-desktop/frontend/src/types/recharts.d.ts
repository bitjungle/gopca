// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Type definitions for Recharts tooltip and custom components

export interface TooltipProps {
  active?: boolean;
  payload?: Array<{
    name: string;
    value: number;
    payload: any;
    color?: string;
    dataKey?: string;
  }>;
  label?: string | number;
}

export interface LoadingEndpointProps {
  cx: number;
  cy: number;
  fill: string;
  radius: number;
}