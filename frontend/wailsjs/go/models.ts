export namespace backend {
	
	export class ActionItem {
	    text: string;
	    gene: string;
	
	    static createFrom(source: any = {}) {
	        return new ActionItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.gene = source["gene"];
	    }
	}
	export class ActionPlan {
	    diet: ActionItem[];
	    supplements: ActionItem[];
	    exercise: ActionItem[];
	    sleep: ActionItem[];
	    monitoring: ActionItem[];
	
	    static createFrom(source: any = {}) {
	        return new ActionPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.diet = this.convertValues(source["diet"], ActionItem);
	        this.supplements = this.convertValues(source["supplements"], ActionItem);
	        this.exercise = this.convertValues(source["exercise"], ActionItem);
	        this.sleep = this.convertValues(source["sleep"], ActionItem);
	        this.monitoring = this.convertValues(source["monitoring"], ActionItem);
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
	export class Finding {
	    rsid: string;
	    gene: string;
	    variant: string;
	    chrom: string;
	    cat: string;
	    subcat: string;
	    riskAllele: string;
	    riskLevel: string;
	    trait: string;
	    desc: string;
	    rec: string;
	    conf: string;
	    pmid: string;
	    a1: string;
	    a2: string;
	    status: string;
	    color: string;
	    badge: string;
	    effect: string;
	
	    static createFrom(source: any = {}) {
	        return new Finding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rsid = source["rsid"];
	        this.gene = source["gene"];
	        this.variant = source["variant"];
	        this.chrom = source["chrom"];
	        this.cat = source["cat"];
	        this.subcat = source["subcat"];
	        this.riskAllele = source["riskAllele"];
	        this.riskLevel = source["riskLevel"];
	        this.trait = source["trait"];
	        this.desc = source["desc"];
	        this.rec = source["rec"];
	        this.conf = source["conf"];
	        this.pmid = source["pmid"];
	        this.a1 = source["a1"];
	        this.a2 = source["a2"];
	        this.status = source["status"];
	        this.color = source["color"];
	        this.badge = source["badge"];
	        this.effect = source["effect"];
	    }
	}
	export class Summary {
	    high: Finding[];
	    moderate: Finding[];
	    protective: Finding[];
	    drugs: Finding[];
	
	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.high = this.convertValues(source["high"], Finding);
	        this.moderate = this.convertValues(source["moderate"], Finding);
	        this.protective = this.convertValues(source["protective"], Finding);
	        this.drugs = this.convertValues(source["drugs"], Finding);
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
	export class ParseResult {
	    provider: string;
	    totalSNPs: number;
	    snps: Record<string, Array<string>>;
	
	    static createFrom(source: any = {}) {
	        return new ParseResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.totalSNPs = source["totalSNPs"];
	        this.snps = source["snps"];
	    }
	}
	export class AnalysisResult {
	    parsed: ParseResult;
	    matched: number;
	    categories: Record<string, Array<Finding>>;
	    summary: Summary;
	    actionPlan: ActionPlan;
	    doctorText: string;
	    generatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new AnalysisResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parsed = this.convertValues(source["parsed"], ParseResult);
	        this.matched = source["matched"];
	        this.categories = this.convertValues(source["categories"], Array<Finding>, true);
	        this.summary = this.convertValues(source["summary"], Summary);
	        this.actionPlan = this.convertValues(source["actionPlan"], ActionPlan);
	        this.doctorText = source["doctorText"];
	        this.generatedAt = source["generatedAt"];
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
	export class ComparisonRow {
	    rsid: string;
	    gene: string;
	    trait: string;
	    cat: string;
	    a1Geno: string;
	    a2Geno: string;
	    a1Status: string;
	    a2Status: string;
	    riskAllele: string;
	    notes: string[];
	
	    static createFrom(source: any = {}) {
	        return new ComparisonRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rsid = source["rsid"];
	        this.gene = source["gene"];
	        this.trait = source["trait"];
	        this.cat = source["cat"];
	        this.a1Geno = source["a1Geno"];
	        this.a2Geno = source["a2Geno"];
	        this.a1Status = source["a1Status"];
	        this.a2Status = source["a2Status"];
	        this.riskAllele = source["riskAllele"];
	        this.notes = source["notes"];
	    }
	}
	export class ComparisonResult {
	    a: ParseResult;
	    b: ParseResult;
	    commonSNPs: number;
	    identical: number;
	    differ: number;
	    onlyA: number;
	    onlyB: number;
	    rows: ComparisonRow[];
	
	    static createFrom(source: any = {}) {
	        return new ComparisonResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.a = this.convertValues(source["a"], ParseResult);
	        this.b = this.convertValues(source["b"], ParseResult);
	        this.commonSNPs = source["commonSNPs"];
	        this.identical = source["identical"];
	        this.differ = source["differ"];
	        this.onlyA = source["onlyA"];
	        this.onlyB = source["onlyB"];
	        this.rows = this.convertValues(source["rows"], ComparisonRow);
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
	
	export class DBStats {
	    totalSNPs: number;
	    totalRecords: number;
	    byCategory: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new DBStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalSNPs = source["totalSNPs"];
	        this.totalRecords = source["totalRecords"];
	        this.byCategory = source["byCategory"];
	    }
	}
	
	
	export class SNPRecord {
	    rsid: string;
	    gene: string;
	    variant: string;
	    chrom: string;
	    cat: string;
	    subcat: string;
	    riskAllele: string;
	    riskLevel: string;
	    trait: string;
	    desc: string;
	    rec: string;
	    conf: string;
	    pmid: string;
	
	    static createFrom(source: any = {}) {
	        return new SNPRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rsid = source["rsid"];
	        this.gene = source["gene"];
	        this.variant = source["variant"];
	        this.chrom = source["chrom"];
	        this.cat = source["cat"];
	        this.subcat = source["subcat"];
	        this.riskAllele = source["riskAllele"];
	        this.riskLevel = source["riskLevel"];
	        this.trait = source["trait"];
	        this.desc = source["desc"];
	        this.rec = source["rec"];
	        this.conf = source["conf"];
	        this.pmid = source["pmid"];
	    }
	}
	export class SessionMeta {
	    id: string;
	    filename: string;
	    provider: string;
	    matched: number;
	    highCount: number;
	    modCount: number;
	    savedAt: string;
	    generatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.filename = source["filename"];
	        this.provider = source["provider"];
	        this.matched = source["matched"];
	        this.highCount = source["highCount"];
	        this.modCount = source["modCount"];
	        this.savedAt = source["savedAt"];
	        this.generatedAt = source["generatedAt"];
	    }
	}

}

export namespace main {
	
	export class DatabaseSource {
	    id: string;
	    name: string;
	    description: string;
	    coverage: string;
	    fileSize: string;
	    variants: string;
	    url: string;
	    fileExt: string;
	    warning: string;
	    installed: boolean;
	    rowCount: number;
	    downloadedAt: string;
	    downloading: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseSource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.coverage = source["coverage"];
	        this.fileSize = source["fileSize"];
	        this.variants = source["variants"];
	        this.url = source["url"];
	        this.fileExt = source["fileExt"];
	        this.warning = source["warning"];
	        this.installed = source["installed"];
	        this.rowCount = source["rowCount"];
	        this.downloadedAt = source["downloadedAt"];
	        this.downloading = source["downloading"];
	    }
	}

}

