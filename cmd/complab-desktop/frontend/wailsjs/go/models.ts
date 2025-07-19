export namespace main {
	
	export class FileData {
	    headers: string[];
	    rowNames: string[];
	    data: number[][];
	
	    static createFrom(source: any = {}) {
	        return new FileData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.headers = source["headers"];
	        this.rowNames = source["rowNames"];
	        this.data = source["data"];
	    }
	}
	export class PCARequest {
	    data: number[][];
	    headers: string[];
	    rowNames: string[];
	    components: number;
	    meanCenter: boolean;
	    standardScale: boolean;
	    robustScale: boolean;
	    method: string;
	
	    static createFrom(source: any = {}) {
	        return new PCARequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = source["data"];
	        this.headers = source["headers"];
	        this.rowNames = source["rowNames"];
	        this.components = source["components"];
	        this.meanCenter = source["meanCenter"];
	        this.standardScale = source["standardScale"];
	        this.robustScale = source["robustScale"];
	        this.method = source["method"];
	    }
	}
	export class PCAResponse {
	    success: boolean;
	    error?: string;
	    result?: types.PCAResult;
	
	    static createFrom(source: any = {}) {
	        return new PCAResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.result = this.convertValues(source["result"], types.PCAResult);
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

}

export namespace types {
	
	export class PCAResult {
	    scores: number[][];
	    loadings: number[][];
	    explained_variance: number[];
	    cumulative_variance: number[];
	    component_labels: string[];
	
	    static createFrom(source: any = {}) {
	        return new PCAResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.scores = source["scores"];
	        this.loadings = source["loadings"];
	        this.explained_variance = source["explained_variance"];
	        this.cumulative_variance = source["cumulative_variance"];
	        this.component_labels = source["component_labels"];
	    }
	}

}

