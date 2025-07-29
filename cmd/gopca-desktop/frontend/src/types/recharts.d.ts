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