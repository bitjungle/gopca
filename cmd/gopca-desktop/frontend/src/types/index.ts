export interface FileData {
  headers: string[];
  rowNames: string[];
  data: (number | null)[][];  // null represents NaN values
  categoricalColumns?: {
    [columnName: string]: string[];  // Column name -> array of values for each row
  };
}

export interface PCARequest {
  data: number[][];
  headers: string[];
  rowNames: string[];
  components: number;
  meanCenter: boolean;
  standardScale: boolean;
  robustScale: boolean;
  method: string;
  excludedRows?: number[];
  excludedColumns?: number[];
  // Kernel PCA parameters
  kernelType?: string;
  kernelGamma?: number;
  kernelDegree?: number;
  kernelCoef0?: number;
}

export interface PCAResult {
  scores: number[][];
  loadings: number[][];
  explained_variance: number[];
  cumulative_variance: number[];
  component_labels: string[];
  variable_labels?: string[];
}

export interface PCAMetrics {
  mahalanobisDistances: number[];
  hotellingT2: number[];
  qResiduals: number[];
  contributions: number[][];
  outliersMahalanobis: boolean[];
  outliersT2: boolean[];
  outliersQResiduals: boolean[];
  t2Threshold: number;
  qThreshold: number;
  ellipseParams?: {
    center: [number, number];
    width: number;
    height: number;
    angle: number;
  };
}

export interface PCAResponse {
  success: boolean;
  error?: string;
  result?: PCAResult;
}