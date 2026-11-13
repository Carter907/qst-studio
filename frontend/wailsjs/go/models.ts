export namespace main {
	
	export class BlankDTO {
	    type: string;
	    answers?: string[];
	    options?: string[];
	    correct?: string;
	
	    static createFrom(source: any = {}) {
	        return new BlankDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.answers = source["answers"];
	        this.options = source["options"];
	        this.correct = source["correct"];
	    }
	}
	export class TourDTO {
	    name: string;
	    guides: string[];
	
	    static createFrom(source: any = {}) {
	        return new TourDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.guides = source["guides"];
	    }
	}
	export class SubGuideDTO {
	    guide: string;
	    adherence: string;
	    segment: string;
	
	    static createFrom(source: any = {}) {
	        return new SubGuideDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.guide = source["guide"];
	        this.adherence = source["adherence"];
	        this.segment = source["segment"];
	    }
	}
	export class GuideSummaryDTO {
	    id: string;
	    scope: string;
	    line_count: number;
	    prerequisites: string[];
	    sub_guides: SubGuideDTO[];
	
	    static createFrom(source: any = {}) {
	        return new GuideSummaryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.scope = source["scope"];
	        this.line_count = source["line_count"];
	        this.prerequisites = source["prerequisites"];
	        this.sub_guides = this.convertValues(source["sub_guides"], SubGuideDTO);
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
	export class GraphDTO {
	    title: string;
	    description: string;
	    scopes: string[];
	    adherences: string[];
	    relaxed_subguides: boolean;
	    strict_coverage: boolean;
	    require_subguides: boolean;
	    guides: GuideSummaryDTO[];
	    tours: TourDTO[];
	
	    static createFrom(source: any = {}) {
	        return new GraphDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.description = source["description"];
	        this.scopes = source["scopes"];
	        this.adherences = source["adherences"];
	        this.relaxed_subguides = source["relaxed_subguides"];
	        this.strict_coverage = source["strict_coverage"];
	        this.require_subguides = source["require_subguides"];
	        this.guides = this.convertValues(source["guides"], GuideSummaryDTO);
	        this.tours = this.convertValues(source["tours"], TourDTO);
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
	export class GuideDetailDTO {
	    id: string;
	    scope: string;
	    path: string;
	    body: string;
	    raw_content: string;
	    line_count: number;
	    prerequisites: string[];
	    sub_guides: SubGuideDTO[];
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new GuideDetailDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.scope = source["scope"];
	        this.path = source["path"];
	        this.body = source["body"];
	        this.raw_content = source["raw_content"];
	        this.line_count = source["line_count"];
	        this.prerequisites = source["prerequisites"];
	        this.sub_guides = this.convertValues(source["sub_guides"], SubGuideDTO);
	        this.tags = source["tags"];
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
	
	export class MatchingPairDTO {
	    prompt: string;
	    target: string;
	
	    static createFrom(source: any = {}) {
	        return new MatchingPairDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prompt = source["prompt"];
	        this.target = source["target"];
	    }
	}
	export class PracticeRecordDTO {
	    item_id: string;
	    guide_id: string;
	    attempts: number;
	    correct: number;
	    last_correct: boolean;
	    last_practiced_at: string;
	
	    static createFrom(source: any = {}) {
	        return new PracticeRecordDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.item_id = source["item_id"];
	        this.guide_id = source["guide_id"];
	        this.attempts = source["attempts"];
	        this.correct = source["correct"];
	        this.last_correct = source["last_correct"];
	        this.last_practiced_at = source["last_practiced_at"];
	    }
	}
	export class QuizChoiceDTO {
	    id: string;
	    text: string;
	    correct: boolean;
	    feedback?: string;
	
	    static createFrom(source: any = {}) {
	        return new QuizChoiceDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.text = source["text"];
	        this.correct = source["correct"];
	        this.feedback = source["feedback"];
	    }
	}
	export class QuizItemDTO {
	    id: string;
	    title?: string;
	    guide_id: string;
	    segment: string;
	    scope: string;
	    type: string;
	    points?: number;
	    prompt: string;
	    feedback?: string;
	    choices?: QuizChoiceDTO[];
	    blanks?: Record<string, BlankDTO>;
	    pairs?: MatchingPairDTO[];
	    distractors?: string[];
	    items?: string[];
	    top_label?: string;
	    bottom_label?: string;
	    solution?: string;
	    created_at?: string;
	
	    static createFrom(source: any = {}) {
	        return new QuizItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.guide_id = source["guide_id"];
	        this.segment = source["segment"];
	        this.scope = source["scope"];
	        this.type = source["type"];
	        this.points = source["points"];
	        this.prompt = source["prompt"];
	        this.feedback = source["feedback"];
	        this.choices = this.convertValues(source["choices"], QuizChoiceDTO);
	        this.blanks = this.convertValues(source["blanks"], BlankDTO, true);
	        this.pairs = this.convertValues(source["pairs"], MatchingPairDTO);
	        this.distractors = source["distractors"];
	        this.items = source["items"];
	        this.top_label = source["top_label"];
	        this.bottom_label = source["bottom_label"];
	        this.solution = source["solution"];
	        this.created_at = source["created_at"];
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

