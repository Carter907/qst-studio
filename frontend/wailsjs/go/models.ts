export namespace main {
	
	export class AgentConfigDTO {
	    endpoint: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpoint = source["endpoint"];
	        this.model = source["model"];
	    }
	}
	export class AgentStatusDTO {
	    available: boolean;
	    endpoint: string;
	    active_model: string;
	    models: string[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentStatusDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.endpoint = source["endpoint"];
	        this.active_model = source["active_model"];
	        this.models = source["models"];
	        this.error = source["error"];
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
	
	export class LearnerMasteryDTO {
	    guide_id: string;
	    segment: string;
	    p_mastery: number;
	    stability_days: number;
	    // Go type: time
	    last_reviewed_at: any;
	    due_for_review: boolean;
	    attempts_count: number;
	    correct_count: number;
	
	    static createFrom(source: any = {}) {
	        return new LearnerMasteryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.guide_id = source["guide_id"];
	        this.segment = source["segment"];
	        this.p_mastery = source["p_mastery"];
	        this.stability_days = source["stability_days"];
	        this.last_reviewed_at = this.convertValues(source["last_reviewed_at"], null);
	        this.due_for_review = source["due_for_review"];
	        this.attempts_count = source["attempts_count"];
	        this.correct_count = source["correct_count"];
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
	    prompt: string;
	    choices: QuizChoiceDTO[];
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
	        this.prompt = source["prompt"];
	        this.choices = this.convertValues(source["choices"], QuizChoiceDTO);
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
	export class PracticeTurnDTO {
	    item: QuizItemDTO;
	    source: string;
	    reason: string;
	    linked_sub_guide: string;
	    unmastered_prereq_id: string;
	    target_mastery: number;
	
	    static createFrom(source: any = {}) {
	        return new PracticeTurnDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.item = this.convertValues(source["item"], QuizItemDTO);
	        this.source = source["source"];
	        this.reason = source["reason"];
	        this.linked_sub_guide = source["linked_sub_guide"];
	        this.unmastered_prereq_id = source["unmastered_prereq_id"];
	        this.target_mastery = source["target_mastery"];
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
	export class PracticeTurnRequest {
	    dir_path: string;
	    mode: string;
	    focus_guide_id: string;
	    focus_segment: string;
	    seen_item_ids: string[];
	    exclude_guides: string[];
	
	    static createFrom(source: any = {}) {
	        return new PracticeTurnRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir_path = source["dir_path"];
	        this.mode = source["mode"];
	        this.focus_guide_id = source["focus_guide_id"];
	        this.focus_segment = source["focus_segment"];
	        this.seen_item_ids = source["seen_item_ids"];
	        this.exclude_guides = source["exclude_guides"];
	    }
	}
	
	
	

}

