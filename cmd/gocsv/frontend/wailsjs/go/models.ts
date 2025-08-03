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

}

