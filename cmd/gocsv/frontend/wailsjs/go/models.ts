export namespace main {
	
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

