export namespace main {
	
	export class FileData {
	    headers: string[];
	    data: string[][];
	    rows: number;
	    columns: number;
	
	    static createFrom(source: any = {}) {
	        return new FileData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.headers = source["headers"];
	        this.data = source["data"];
	        this.rows = source["rows"];
	        this.columns = source["columns"];
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

