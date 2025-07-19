export interface FileData {
  headers: string[];
  rowNames: string[];
  data: number[][];
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
}

export interface PCAResult {
  scores: number[][];
  loadings: number[][];
  explained_variance: number[];
  cumulative_variance: number[];
  component_labels: string[];
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