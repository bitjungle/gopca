export namespace main {
	
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

