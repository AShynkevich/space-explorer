export namespace session {
	
	export class Entry {
	    name: string;
	    path: string;
	    kind: string;
	    size: number;
	    percent: number;
	    items: number;
	    hasChildren: boolean;
	    unreadable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.kind = source["kind"];
	        this.size = source["size"];
	        this.percent = source["percent"];
	        this.items = source["items"];
	        this.hasChildren = source["hasChildren"];
	        this.unreadable = source["unreadable"];
	    }
	}
	export class ErrorInfo {
	    path: string;
	    err: string;
	
	    static createFrom(source: any = {}) {
	        return new ErrorInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.err = source["err"];
	    }
	}
	export class Listing {
	    path: string;
	    size: number;
	    total: number;
	    entries: Entry[];
	
	    static createFrom(source: any = {}) {
	        return new Listing(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.size = source["size"];
	        this.total = source["total"];
	        this.entries = this.convertValues(source["entries"], Entry);
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

export namespace volumes {
	
	export class Volume {
	    name: string;
	    path: string;
	    total: number;
	    free: number;
	
	    static createFrom(source: any = {}) {
	        return new Volume(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.total = source["total"];
	        this.free = source["free"];
	    }
	}

}

