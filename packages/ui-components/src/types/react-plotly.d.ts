declare module 'react-plotly.js' {
  import { Component } from 'react';
  import { Layout, Config, Data } from 'plotly.js';

  interface PlotParams {
    data: Data[];
    layout?: Partial<Layout>;
    config?: Partial<Config>;
    onInitialized?: (figure: any, graphDiv: HTMLElement) => void;
    onUpdate?: (figure: any, graphDiv: HTMLElement) => void;
    onPurge?: (figure: any, graphDiv: HTMLElement) => void;
    onError?: (error: Error) => void;
    useResizeHandler?: boolean;
    style?: React.CSSProperties;
    className?: string;
    divId?: string;
  }

  export default class Plot extends Component<PlotParams> {}
}