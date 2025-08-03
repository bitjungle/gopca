export namespace main {
	
	export class OutlierInfo {
	    rowIndex: number;
	    value: string;
	    method: string;
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new OutlierInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rowIndex = source["rowIndex"];
	        this.value = source["value"];
	        this.method = source["method"];
	        this.score = source["score"];
	    }
	}
	export class HistogramBin {
	    min: number;
	    max: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new HistogramBin(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.min = source["min"];
	        this.max = source["max"];
	        this.count = source["count"];
	    }
	}
	export class DistributionInfo {
	    histogram?: HistogramBin[];
	    isNormal: boolean;
	    normalityPValue?: number;
	    distType: string;
	
	    static createFrom(source: any = {}) {
	        return new DistributionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.histogram = this.convertValues(source["histogram"], HistogramBin);
	        this.isNormal = source["isNormal"];
	        this.normalityPValue = source["normalityPValue"];
	        this.distType = source["distType"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ColumnStatistics {
	    count: number;
	    missing: number;
	    missingPercent: number;
	    unique: number;
	    mean?: number;
	    median?: number;
	    mode?: string;
	    stdDev?: number;
	    min?: number;
	    max?: number;
	    q1?: number;
	    q3?: number;
	    iqr?: number;
	    skewness?: number;
	    kurtosis?: number;
	    categories?: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new ColumnStatistics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.missing = source["missing"];
	        this.missingPercent = source["missingPercent"];
	        this.unique = source["unique"];
	        this.mean = source["mean"];
	        this.median = source["median"];
	        this.mode = source["mode"];
	        this.stdDev = source["stdDev"];
	        this.min = source["min"];
	        this.max = source["max"];
	        this.q1 = source["q1"];
	        this.q3 = source["q3"];
	        this.iqr = source["iqr"];
	        this.skewness = source["skewness"];
	        this.kurtosis = source["kurtosis"];
	        this.categories = source["categories"];
	    }
	}
	export class ColumnAnalysis {
	    name: string;
	    type: string;
	    stats: ColumnStatistics;
	    distribution: DistributionInfo;
	    outliers: OutlierInfo[];
	    qualityScore: number;
	
	    static createFrom(source: any = {}) {
	        return new ColumnAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.stats = this.convertValues(source["stats"], ColumnStatistics);
	        this.distribution = this.convertValues(source["distribution"], DistributionInfo);
	        this.outliers = this.convertValues(source["outliers"], OutlierInfo);
	        this.qualityScore = source["qualityScore"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ColumnMissing {
	    name: string;
	    totalValues: number;
	    missingValues: number;
	    missingPercent: number;
	    pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new ColumnMissing(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.totalValues = source["totalValues"];
	        this.missingValues = source["missingValues"];
	        this.missingPercent = source["missingPercent"];
	        this.pattern = source["pattern"];
	    }
	}
	
	export class DataProfile {
	    rows: number;
	    columns: number;
	    numericColumns: number;
	    categoricalColumns: number;
	    targetColumns: number;
	    missingPercent: number;
	    duplicateRows: number;
	    memorySize: string;
	
	    static createFrom(source: any = {}) {
	        return new DataProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = source["rows"];
	        this.columns = source["columns"];
	        this.numericColumns = source["numericColumns"];
	        this.categoricalColumns = source["categoricalColumns"];
	        this.targetColumns = source["targetColumns"];
	        this.missingPercent = source["missingPercent"];
	        this.duplicateRows = source["duplicateRows"];
	        this.memorySize = source["memorySize"];
	    }
	}
	export class Recommendation {
	    priority: string;
	    category: string;
	    action: string;
	    description: string;
	    columns?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Recommendation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.priority = source["priority"];
	        this.category = source["category"];
	        this.action = source["action"];
	        this.description = source["description"];
	        this.columns = source["columns"];
	    }
	}
	export class QualityIssue {
	    severity: string;
	    category: string;
	    description: string;
	    affected: string[];
	    impact: string;
	
	    static createFrom(source: any = {}) {
	        return new QualityIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.severity = source["severity"];
	        this.category = source["category"];
	        this.description = source["description"];
	        this.affected = source["affected"];
	        this.impact = source["impact"];
	    }
	}
	export class DataQualityReport {
	    dataProfile: DataProfile;
	    columnAnalysis: ColumnAnalysis[];
	    qualityScore: number;
	    issues: QualityIssue[];
	    recommendations: Recommendation[];
	
	    static createFrom(source: any = {}) {
	        return new DataQualityReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dataProfile = this.convertValues(source["dataProfile"], DataProfile);
	        this.columnAnalysis = this.convertValues(source["columnAnalysis"], ColumnAnalysis);
	        this.qualityScore = source["qualityScore"];
	        this.issues = this.convertValues(source["issues"], QualityIssue);
	        this.recommendations = this.convertValues(source["recommendations"], Recommendation);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class FileData {
	    headers: string[];
	    rowNames?: string[];
	    data: string[][];
	    rows: number;
	    columns: number;
	    categoricalColumns?: Record<string, Array<string>>;
	    numericTargetColumns?: Record<string, Array<number>>;
	    columnTypes?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new FileData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.headers = source["headers"];
	        this.rowNames = source["rowNames"];
	        this.data = source["data"];
	        this.rows = source["rows"];
	        this.columns = source["columns"];
	        this.categoricalColumns = source["categoricalColumns"];
	        this.numericTargetColumns = source["numericTargetColumns"];
	        this.columnTypes = source["columnTypes"];
	    }
	}
	export class FillMissingValuesRequest {
	    strategy: string;
	    column: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new FillMissingValuesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strategy = source["strategy"];
	        this.column = source["column"];
	        this.value = source["value"];
	    }
	}
	
	export class RowMissing {
	    index: number;
	    totalValues: number;
	    missingValues: number;
	    missingPercent: number;
	
	    static createFrom(source: any = {}) {
	        return new RowMissing(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.totalValues = source["totalValues"];
	        this.missingValues = source["missingValues"];
	        this.missingPercent = source["missingPercent"];
	    }
	}
	export class MissingValueStats {
	    totalCells: number;
	    missingCells: number;
	    missingPercent: number;
	    columnStats: Record<string, ColumnMissing>;
	    rowStats: Record<number, RowMissing>;
	
	    static createFrom(source: any = {}) {
	        return new MissingValueStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalCells = source["totalCells"];
	        this.missingCells = source["missingCells"];
	        this.missingPercent = source["missingPercent"];
	        this.columnStats = this.convertValues(source["columnStats"], ColumnMissing, true);
	        this.rowStats = this.convertValues(source["rowStats"], RowMissing, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	export class ValidationResult {
	    isValid: boolean;
	    messages: string[];
	
	    static createFrom(source: any = {}) {
	        return new ValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isValid = source["isValid"];
	        this.messages = source["messages"];
	    }
	}

}

