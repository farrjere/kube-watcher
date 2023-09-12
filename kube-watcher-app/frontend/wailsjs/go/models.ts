export namespace app {
	
	export class PodLogMessage {
	    message: string;
	    pod: string;
	
	    static createFrom(source: any = {}) {
	        return new PodLogMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.pod = source["pod"];
	    }
	}

}

export namespace kube {
	
	export class SearchResult {
	    pod_name: string;
	    matches: string[];
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pod_name = source["pod_name"];
	        this.matches = source["matches"];
	    }
	}

}

