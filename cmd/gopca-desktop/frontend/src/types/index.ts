export interface FileData {
  headers: string[];
  rowNames: string[];
  data: number[][];
  missingMask?: boolean[][];
  categoricalColumns?: {
    [columnName: string]: string[];  // Column name -> array of values for each row
  };
  numericTargetColumns?: {
    [columnName: string]: number[];  // Column name -> array of numeric values for each row
  };
}

export interface PCARequest {
  data: number[][];
  missingMask?: boolean[][];
  headers: string[];
  rowNames: string[];
  components: number;
  meanCenter: boolean;
  standardScale: boolean;
  robustScale: boolean;
  scaleOnly: boolean;
  snv: boolean;
  vectorNorm: boolean;
  method: string;
  excludedRows?: number[];
  excludedColumns?: number[];
  missingStrategy?: string;
  calculateMetrics?: boolean;
  // Kernel PCA parameters
  kernelType?: string;
  kernelGamma?: number;
  kernelDegree?: number;
  kernelCoef0?: number;
  // Grouping parameters for confidence ellipses
  groupColumn?: string;
  groupLabels?: string[];
}


export interface PCAResult {
  scores: number[][];
  loadings: number[][];
  explained_variance: number[];
  explained_variance_ratio: number[];
  cumulative_variance: number[];
  component_labels: string[];
  variable_labels?: string[];
  components_computed: number;
  method: string;
  preprocessing_applied: boolean;
  means?: number[];
  stddevs?: number[];
  metrics?: SampleMetrics[];
  t2_limit_95?: number;
  t2_limit_99?: number;
  q_limit_95?: number;
  q_limit_99?: number;
}

export interface SampleMetrics {
  hotelling_t2: number;
  mahalanobis: number;
  rss: number;
  is_outlier: boolean;
}

export interface EllipseParams {
  centerX: number;
  centerY: number;
  majorAxis: number;
  minorAxis: number;
  angle: number;
  confidenceLevel: number;
}

export interface PCAResponse {
  success: boolean;
  error?: string;
  result?: PCAResult;
  info?: string;
  groupEllipses90?: Record<string, EllipseParams>;
  groupEllipses95?: Record<string, EllipseParams>;
  groupEllipses99?: Record<string, EllipseParams>;
}