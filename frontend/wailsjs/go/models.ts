export namespace download {
	
	export class Download {
	    id: number;
	    playlist_id: number;
	    url: string;
	    status: number;
	    format_downloaded: string;
	    md5?: sql.NullString;
	    output_filename?: sql.NullString;
	    last_attempt: number;
	    fail_message?: sql.NullString;
	    attempt_count: number;
	
	    static createFrom(source: any = {}) {
	        return new Download(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.playlist_id = source["playlist_id"];
	        this.url = source["url"];
	        this.status = source["status"];
	        this.format_downloaded = source["format_downloaded"];
	        this.md5 = this.convertValues(source["md5"], sql.NullString);
	        this.output_filename = this.convertValues(source["output_filename"], sql.NullString);
	        this.last_attempt = source["last_attempt"];
	        this.fail_message = this.convertValues(source["fail_message"], sql.NullString);
	        this.attempt_count = source["attempt_count"];
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

export namespace main {
	
	export class RegisteredFile {
	    id: number;
	    filename: string;
	    file_path: string;
	    md5_hash: string;
	    registered_at: number;
	
	    static createFrom(source: any = {}) {
	        return new RegisteredFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.filename = source["filename"];
	        this.file_path = source["file_path"];
	        this.md5_hash = source["md5_hash"];
	        this.registered_at = source["registered_at"];
	    }
	}

}

export namespace playlist {
	
	export class Playlist {
	    id: number;
	    name: string;
	    url: string;
	    output_format: string;
	    save_directory: string;
	    thumbnail_base64?: sql.NullString;
	    is_enabled: boolean;
	    added_at: number;
	
	    static createFrom(source: any = {}) {
	        return new Playlist(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.output_format = source["output_format"];
	        this.save_directory = source["save_directory"];
	        this.thumbnail_base64 = this.convertValues(source["thumbnail_base64"], sql.NullString);
	        this.is_enabled = source["is_enabled"];
	        this.added_at = source["added_at"];
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

export namespace sql {
	
	export class NullString {
	    String: string;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullString(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.String = source["String"];
	        this.Valid = source["Valid"];
	    }
	}

}

