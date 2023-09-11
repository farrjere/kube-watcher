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

