import {
  parseMarkdownBlocksWithLines,
  type RenderedGuideBlock
} from './markdown';
import {
  LoadGraph,
  GetGuide,
  GetLearningPath,
  GetQuizItems,
  SaveQuizItem,
  DeleteQuizItem,
  SelectGraphDirectory,
  GetRecentGraphs,
  GetPracticeHistory,
  RecordPracticeAttempt
} from '../wailsjs/go/main/App.js';
import type { main } from '../wailsjs/go/models';
import {
  type RightTab,
  type EditableChoice,
  type EditableBlank,
  type EditablePair,
  type BuilderNotice,
  defaultChoices,
  defaultTrueFalseChoices,
  shuffleArray
} from './stores/types';
import { initPracticeAnswer, gradePracticeAnswer } from './stores/practiceGrading';
import { buildQuizItemPayload } from './stores/builderHelpers';

export type { RightTab, EditableChoice, EditableBlank, EditablePair, BuilderNotice };

class StudioState {
  // Workspace State
  graphDir = $state('');
  graph = $state<main.GraphDTO | null>(null);
  recentGraphs = $state<string[]>([]);
  leftPaneOpen = $state<boolean>(true);
  rightPaneOpen = $state<boolean>(true);
  activeRightTab = $state<RightTab>('inspector');
  statusMessage = $state<string>('Ready');
  showCommandPalette = $state(false);

  // Selected Guide & Navigator State
  selectedGuideId = $state<string>('');
  selectedGuide = $state<main.GuideDetailDTO | null>(null);
  guideBodyLines = $state<string[]>([]);
  guideRenderedBlocks = $state<RenderedGuideBlock[]>([]);
  learningPath = $state<string[]>([]);
  quizItems = $state<main.QuizItemDTO[]>([]);

  // Segment Anchoring
  startLine = $state<number | null>(null);
  endLine = $state<number | null>(null);

  // Question Builder Form
  editingItemId = $state<string | null>(null);
  itemTitle = $state('');
  itemPrompt = $state('');
  itemType = $state('multiple_choice');
  builderNotice = $state<BuilderNotice | null>(null);

  // Type-Specific Builder States
  itemChoices = $state<EditableChoice[]>(defaultChoices());
  itemBlanks = $state<EditableBlank[]>([
    { token: 'b1', type: 'text', answers: [''], options: [], correct: '' }
  ]);
  itemPairs = $state<EditablePair[]>([
    { prompt: '', target: '' },
    { prompt: '', target: '' }
  ]);
  itemDistractors = $state<string[]>([]);
  itemOrderItems = $state<string[]>(['', '']);
  itemTopLabel = $state('');
  itemBottomLabel = $state('');
  itemSolution = $state('');

  // Practice Runner State
  practiceHistory = $state<Record<string, main.PracticeRecordDTO>>({});
  practiceScope = $state<'guide' | 'graph'>('guide');
  practiceDeck = $state<main.QuizItemDTO[]>([]);
  practiceIndex = $state<number>(0);
  practiceSessionScore = $state<{ correct: number; total: number }>({ correct: 0, total: 0 });
  practiceAnswer = $state<any>({});
  practiceChecked = $state<boolean>(false);
  practiceResult = $state<{ isCorrect: boolean; feedback?: string } | null>(null);

  // Zoom & UI Scaling State
  zoomLevel = $state<number>(100);

  async init() {
    this.initZoom();
    await this.loadRecents();
    if (this.recentGraphs.length > 0) {
      await this.handleLoadGraph(this.recentGraphs[0]);
    }
  }

  initZoom() {
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        const saved = window.localStorage.getItem('qst_studio_ui_zoom');
        if (saved) {
          const parsed = parseInt(saved, 10);
          if (!isNaN(parsed) && parsed >= 70 && parsed <= 200) {
            this.zoomLevel = parsed;
          }
        }
      }
    } catch (e) {
      console.warn('Failed to load zoom from localStorage', e);
    }
    this.applyZoom();
  }

  zoomIn() {
    if (this.zoomLevel < 200) {
      this.zoomLevel = Math.min(200, this.zoomLevel + 10);
      this.applyZoom();
    }
  }

  zoomOut() {
    if (this.zoomLevel > 70) {
      this.zoomLevel = Math.max(70, this.zoomLevel - 10);
      this.applyZoom();
    }
  }

  resetZoom() {
    this.zoomLevel = 100;
    this.applyZoom();
  }

  applyZoom() {
    if (typeof document !== 'undefined' && document.documentElement) {
      (document.documentElement.style as any).zoom = `${this.zoomLevel}%`;
    }
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        window.localStorage.setItem('qst_studio_ui_zoom', this.zoomLevel.toString());
      }
    } catch (e) {
      console.warn('Failed to save zoom to localStorage', e);
    }
  }

  async loadRecents() {
    try {
      this.recentGraphs = await GetRecentGraphs();
    } catch {
      this.recentGraphs = [];
    }
  }

  async handleLoadGraph(dir: string) {
    if (!dir) return;
    try {
      this.statusMessage = 'Loading graph...';
      this.graphDir = dir;
      this.graph = await LoadGraph(dir);
      await this.loadRecents();
      await this.loadPracticeHistory();
      if (this.graph && this.graph.guides.length > 0) {
        await this.selectGuide(this.graph.guides[0].id);
      } else {
        this.selectedGuideId = '';
        this.selectedGuide = null;
      }
      this.statusMessage = `Loaded "${this.graph?.title}" (${this.graph?.guides.length} guides)`;
    } catch (err: any) {
      this.statusMessage = `Error: ${err}`;
    }
  }

  async loadPracticeHistory() {
    try {
      this.practiceHistory = await GetPracticeHistory(this.graphDir);
    } catch {
      this.practiceHistory = {};
    }
  }

  async handleNativeFolderPick() {
    try {
      const selected = await SelectGraphDirectory();
      if (selected) {
        await this.handleLoadGraph(selected);
      }
    } catch (err: any) {
      this.statusMessage = `Folder picker error: ${err}`;
    }
  }

  private parseSegmentBounds(segmentStr: string): [number, number] | null {
    const trimmed = (segmentStr || '').trim();
    if (!trimmed || trimmed === 'ALL') {
      return null;
    }
    const maxLine = Math.max(1, this.guideBodyLines.length);
    if (trimmed === '1-1' && maxLine > 1) {
      return null;
    }
    const parts = trimmed.split('-');
    const s = parseInt(parts[0], 10);
    const e = parts.length > 1 ? parseInt(parts[1], 10) : s;
    if (!isNaN(s) && s >= 1) {
      const start = Math.min(s, maxLine);
      const end = !isNaN(e) && e >= start ? Math.min(e, maxLine) : start;
      return [start, end];
    }
    return null;
  }

  applySegmentRange(segmentStr: string) {
    const bounds = this.parseSegmentBounds(segmentStr);
    if (bounds) {
      this.startLine = bounds[0];
      this.endLine = bounds[1];
    } else {
      this.startLine = null;
      this.endLine = null;
    }
  }

  async selectGuide(id: string, options?: { preservePractice?: boolean; preserveAnchor?: boolean }) {
    const preservePractice = options?.preservePractice ?? false;
    const preserveAnchor = options?.preserveAnchor ?? false;

    const guideChanged = this.selectedGuideId !== id;
    this.selectedGuideId = id;
    if (!preserveAnchor) {
      this.startLine = null;
      this.endLine = null;
    }
    if (guideChanged && !preservePractice) {
      this.resetBuilderForm(true);
    }

    try {
      this.selectedGuide = await GetGuide(this.graphDir, id);
      this.guideBodyLines = this.selectedGuide?.body ? this.selectedGuide.body.split('\n') : [];
      this.guideRenderedBlocks = parseMarkdownBlocksWithLines(this.selectedGuide?.body || '');
      this.quizItems = await GetQuizItems(this.graphDir, id);
      this.learningPath = await GetLearningPath(this.graphDir, id);

      // Rebuild practice deck only if practice tab is active AND practice scope is 'guide' AND not preserving practice
      if (!preservePractice && this.activeRightTab === 'practice' && this.practiceScope === 'guide') {
        await this.buildPracticeDeck('guide');
      }
    } catch (err: any) {
      this.statusMessage = `Failed to load guide: ${err}`;
    }
  }

  doesBlockOverlapSelection(blk: RenderedGuideBlock): boolean {
    if (this.startLine === null) return false;
    const selStart = this.startLine;
    const selEnd = this.endLine !== null ? this.endLine : this.startLine;
    return blk.startLine <= selEnd && blk.endLine >= selStart;
  }

  isLineInSelectedRange(lineNum: number): boolean {
    if (this.startLine === null) return false;
    const selEnd = this.endLine !== null ? this.endLine : this.startLine;
    return lineNum >= this.startLine && lineNum <= selEnd;
  }

  handleBlockAnchorClick(blk: RenderedGuideBlock, e: MouseEvent) {
    if (e.shiftKey && this.startLine !== null) {
      const curEnd = this.endLine ?? this.startLine;
      this.startLine = Math.min(this.startLine, blk.startLine);
      this.endLine = Math.max(curEnd, blk.endLine);
    } else {
      this.startLine = blk.startLine;
      this.endLine = blk.endLine;
    }
  }

  handleLineClick(lineNum: number) {
    if (this.startLine === null || (this.startLine !== null && this.endLine !== null)) {
      this.startLine = lineNum;
      this.endLine = null;
    } else {
      if (lineNum >= this.startLine) {
        this.endLine = lineNum;
      } else {
        this.endLine = this.startLine;
        this.startLine = lineNum;
      }
    }
  }

  getActiveSegment(): string {
    if (this.startLine !== null && this.endLine !== null) {
      return `${this.startLine}-${this.endLine}`;
    }
    if (this.startLine !== null) {
      return `${this.startLine}-${this.startLine}`;
    }
    const maxLine = Math.max(1, this.guideBodyLines.length);
    return `1-${maxLine}`;
  }

  // --- PRACTICE RUNNER METHODS ---

  async openPracticeTab(scope: 'guide' | 'graph' = 'guide') {
    this.activeRightTab = 'practice';
    this.rightPaneOpen = true;
    await this.buildPracticeDeck(scope);
  }

  async buildPracticeDeck(scope: 'guide' | 'graph') {
    this.practiceScope = scope;
    this.practiceIndex = 0;
    this.practiceChecked = false;
    this.practiceResult = null;

    if (scope === 'guide') {
      this.practiceDeck = [...this.quizItems];
    } else {
      try {
        const allItems = await GetQuizItems(this.graphDir, '');
        this.practiceDeck = allItems;
      } catch {
        this.practiceDeck = [...this.quizItems];
      }
    }

    this.resetPracticeCardAnswer();
    await this.syncPracticeCardAnchor();
  }

  async syncPracticeCardAnchor() {
    const current = this.practiceDeck[this.practiceIndex];
    if (!current) {
      this.startLine = null;
      this.endLine = null;
      return;
    }

    // In whole-graph mode, if question belongs to a different guide, load that guide first
    if (this.practiceScope === 'graph' && current.guide_id && current.guide_id !== this.selectedGuideId) {
      await this.selectGuide(current.guide_id, { preservePractice: true, preserveAnchor: true });
    }

    if (current.segment) {
      this.applySegmentRange(current.segment);
    } else {
      this.startLine = null;
      this.endLine = null;
    }
  }

  resetPracticeCardAnswer() {
    const current = this.practiceDeck[this.practiceIndex];
    this.practiceChecked = false;
    this.practiceResult = null;
    this.practiceAnswer = current ? initPracticeAnswer(current) : {};
  }

  async shufflePracticeDeck() {
    this.practiceDeck = shuffleArray(this.practiceDeck);
    this.practiceIndex = 0;
    this.resetPracticeCardAnswer();
    await this.syncPracticeCardAnchor();
  }

  async nextPracticeCard() {
    if (this.practiceIndex < this.practiceDeck.length - 1) {
      this.practiceIndex++;
      this.resetPracticeCardAnswer();
      await this.syncPracticeCardAnchor();
    }
  }

  async prevPracticeCard() {
    if (this.practiceIndex > 0) {
      this.practiceIndex--;
      this.resetPracticeCardAnswer();
      await this.syncPracticeCardAnchor();
    }
  }

  async submitPracticeAnswer() {
    const card = this.practiceDeck[this.practiceIndex];
    if (!card || this.practiceChecked) return;

    const result = gradePracticeAnswer(card, this.practiceAnswer);
    if (result.selfGraded) {
      this.practiceAnswer.revealed = true;
      return;
    }

    this.practiceChecked = true;
    this.practiceResult = { isCorrect: result.isCorrect, feedback: result.feedback };
    this.practiceSessionScore = {
      correct: this.practiceSessionScore.correct + (result.isCorrect ? 1 : 0),
      total: this.practiceSessionScore.total + 1
    };

    try {
      await RecordPracticeAttempt(this.graphDir, card.id, card.guide_id, result.isCorrect);
      await this.loadPracticeHistory();
    } catch (err: any) {
      this.statusMessage = `Practice record error: ${err}`;
    }
  }

  async selfGradeEssay(understood: boolean) {
    const card = this.practiceDeck[this.practiceIndex];
    if (!card) return;

    this.practiceChecked = true;
    this.practiceResult = {
      isCorrect: understood,
      feedback: understood ? 'Marked as Understood' : 'Marked for Review'
    };
    this.practiceSessionScore = {
      correct: this.practiceSessionScore.correct + (understood ? 1 : 0),
      total: this.practiceSessionScore.total + 1
    };

    try {
      await RecordPracticeAttempt(this.graphDir, card.id, card.guide_id, understood);
      await this.loadPracticeHistory();
    } catch (err: any) {
      this.statusMessage = `Practice record error: ${err}`;
    }
  }

  // --- QUESTION AUTHORING METHODS ---

  changeItemType(newType: string) {
    this.itemType = newType;
    if (newType === 'true_false' && this.itemChoices.length !== 2) {
      this.itemChoices = defaultTrueFalseChoices();
    } else if (newType === 'multiple_choice' || newType === 'multiple_answers') {
      if (this.itemChoices.length === 0) {
        this.itemChoices = defaultChoices();
      }
    }
  }

  loadBankItemIntoBuilder(item: main.QuizItemDTO) {
    this.editingItemId = item.id;
    this.itemTitle = item.title || item.id;
    this.itemPrompt = item.prompt;
    const rawType = item.type || 'multiple_choice';
    this.itemType = rawType === 'choice' ? 'multiple_choice' : rawType;

    // Load Choices
    if (item.choices && item.choices.length > 0) {
      this.itemChoices = item.choices.map((c, idx) => ({
        id: c.id || `c${idx + 1}`,
        text: c.text,
        correct: c.correct,
        feedback: c.feedback || ''
      }));
    } else {
      this.itemChoices = this.itemType === 'true_false' ? defaultTrueFalseChoices() : defaultChoices();
    }

    // Load Blanks
    if (item.blanks && Object.keys(item.blanks).length > 0) {
      this.itemBlanks = Object.entries(item.blanks).map(([token, b]) => ({
        token,
        type: (b.type as 'text' | 'dropdown') || 'text',
        answers: b.answers && b.answers.length > 0 ? [...b.answers] : [''],
        options: b.options && b.options.length > 0 ? [...b.options] : ['', ''],
        correct: b.correct || ''
      }));
    } else {
      this.itemBlanks = [{ token: 'b1', type: 'text', answers: [''], options: [], correct: '' }];
    }

    // Load Pairs & Distractors
    if (item.pairs && item.pairs.length > 0) {
      this.itemPairs = item.pairs.map(p => ({ prompt: p.prompt, target: p.target }));
    } else {
      this.itemPairs = [{ prompt: '', target: '' }, { prompt: '', target: '' }];
    }
    this.itemDistractors = item.distractors ? [...item.distractors] : [];

    // Load Ordering Items
    if (item.items && item.items.length > 0) {
      this.itemOrderItems = [...item.items];
    } else {
      this.itemOrderItems = ['', ''];
    }
    this.itemTopLabel = item.top_label || '';
    this.itemBottomLabel = item.bottom_label || '';

    // Load Essay Solution
    this.itemSolution = item.solution || '';

    this.builderNotice = null;
    this.applySegmentRange(item.segment);
    this.activeRightTab = 'questions';
    this.rightPaneOpen = true;
    this.statusMessage = `Editing ${this.itemType} question ${item.id}`;
  }

  resetBuilderForm(silent = false) {
    this.editingItemId = null;
    this.itemTitle = '';
    this.itemPrompt = '';
    this.itemChoices = this.itemType === 'true_false' ? defaultTrueFalseChoices() : defaultChoices();
    this.itemBlanks = [{ token: 'b1', type: 'text', answers: [''], options: [], correct: '' }];
    this.itemPairs = [{ prompt: '', target: '' }, { prompt: '', target: '' }];
    this.itemDistractors = [];
    this.itemOrderItems = ['', ''];
    this.itemTopLabel = '';
    this.itemBottomLabel = '';
    this.itemSolution = '';
    this.builderNotice = null;
    if (!silent) {
      this.statusMessage = 'Cleared builder for new question draft';
    }
  }

  // Choices Management
  private reindexChoices(choices: EditableChoice[]): EditableChoice[] {
    let hasCorrect = choices.some(c => c.correct);
    return choices.map((c, idx) => ({
      ...c,
      id: `c${idx + 1}`,
      correct: hasCorrect ? c.correct : idx === 0
    }));
  }

  addChoice() {
    this.itemChoices = this.reindexChoices([
      ...this.itemChoices,
      { id: '', text: '', correct: false, feedback: '' }
    ]);
  }

  removeChoice(index: number) {
    if (this.itemChoices.length > 2) {
      const filtered = this.itemChoices.filter((_, i) => i !== index);
      this.itemChoices = this.reindexChoices(filtered);
    }
  }

  setCorrectChoice(index: number) {
    this.itemChoices = this.itemChoices.map((c, i) => ({
      ...c,
      correct: i === index
    }));
  }

  toggleMultipleAnswerCorrect(index: number) {
    this.itemChoices = this.itemChoices.map((c, i) => ({
      ...c,
      correct: i === index ? !c.correct : c.correct
    }));
  }

  // Blanks Management
  addBlank() {
    const nextNum = this.itemBlanks.length + 1;
    this.itemBlanks = [
      ...this.itemBlanks,
      { token: `b${nextNum}`, type: 'text', answers: [''], options: [], correct: '' }
    ];
  }

  removeBlank(index: number) {
    if (this.itemBlanks.length > 1) {
      this.itemBlanks = this.itemBlanks.filter((_, i) => i !== index);
    }
  }

  addBlankAnswer(blankIndex: number) {
    this.itemBlanks = this.itemBlanks.map((b, i) => {
      if (i === blankIndex) {
        return { ...b, answers: [...b.answers, ''] };
      }
      return b;
    });
  }

  removeBlankAnswer(blankIndex: number, ansIndex: number) {
    this.itemBlanks = this.itemBlanks.map((b, i) => {
      if (i === blankIndex && b.answers.length > 1) {
        return { ...b, answers: b.answers.filter((_, j) => j !== ansIndex) };
      }
      return b;
    });
  }

  addBlankOption(blankIndex: number) {
    this.itemBlanks = this.itemBlanks.map((b, i) => {
      if (i === blankIndex) {
        return { ...b, options: [...b.options, ''] };
      }
      return b;
    });
  }

  removeBlankOption(blankIndex: number, optIndex: number) {
    this.itemBlanks = this.itemBlanks.map((b, i) => {
      if (i === blankIndex && b.options.length > 1) {
        return { ...b, options: b.options.filter((_, j) => j !== optIndex) };
      }
      return b;
    });
  }

  // Matching Pairs Management
  addMatchingPair() {
    this.itemPairs = [...this.itemPairs, { prompt: '', target: '' }];
  }

  removeMatchingPair(index: number) {
    if (this.itemPairs.length > 2) {
      this.itemPairs = this.itemPairs.filter((_, i) => i !== index);
    }
  }

  addDistractor() {
    this.itemDistractors = [...this.itemDistractors, ''];
  }

  removeDistractor(index: number) {
    this.itemDistractors = this.itemDistractors.filter((_, i) => i !== index);
  }

  // Ordering Items Management
  addOrderItem() {
    this.itemOrderItems = [...this.itemOrderItems, ''];
  }

  removeOrderItem(index: number) {
    if (this.itemOrderItems.length > 2) {
      this.itemOrderItems = this.itemOrderItems.filter((_, i) => i !== index);
    }
  }

  moveOrderItem(index: number, direction: 'up' | 'down') {
    const target = direction === 'up' ? index - 1 : index + 1;
    if (target < 0 || target >= this.itemOrderItems.length) return;
    const updated = [...this.itemOrderItems];
    const temp = updated[index];
    updated[index] = updated[target];
    updated[target] = temp;
    this.itemOrderItems = updated;
  }

  reorderOrderItems(fromIndex: number, toIndex: number) {
    if (fromIndex === toIndex || fromIndex < 0 || toIndex < 0 || fromIndex >= this.itemOrderItems.length || toIndex >= this.itemOrderItems.length) return;
    const updated = [...this.itemOrderItems];
    const [moved] = updated.splice(fromIndex, 1);
    updated.splice(toIndex, 0, moved);
    this.itemOrderItems = updated;
  }

  async saveItem() {
    if (!this.selectedGuide) return;

    const built = buildQuizItemPayload(this.selectedGuide, this.getActiveSegment(), {
      editingItemId: this.editingItemId,
      itemTitle: this.itemTitle,
      itemPrompt: this.itemPrompt,
      itemType: this.itemType,
      itemChoices: this.itemChoices,
      itemBlanks: this.itemBlanks,
      itemPairs: this.itemPairs,
      itemDistractors: this.itemDistractors,
      itemOrderItems: this.itemOrderItems,
      itemTopLabel: this.itemTopLabel,
      itemBottomLabel: this.itemBottomLabel,
      itemSolution: this.itemSolution
    });

    if (!built.ok) {
      this.statusMessage = built.error;
      this.builderNotice = { type: 'error', text: built.error };
      return;
    }

    const newItem = built.item;
    try {
      await SaveQuizItem(this.graphDir, newItem);
      this.editingItemId = newItem.id;
      this.quizItems = await GetQuizItems(this.graphDir, this.selectedGuide.id);
      const msg = `Saved ${newItem.type} question (${newItem.id}) to items.yaml`;
      this.statusMessage = msg;
      this.builderNotice = { type: 'success', text: msg };
    } catch (err: any) {
      const msg = `Save failed: ${err}`;
      this.statusMessage = msg;
      this.builderNotice = { type: 'error', text: msg };
    }
  }

  async deleteBankItem(itemId: string) {
    if (!this.selectedGuide || !itemId) return;
    try {
      await DeleteQuizItem(this.graphDir, itemId);
      if (this.editingItemId === itemId) {
        this.resetBuilderForm(true);
      }
      this.quizItems = await GetQuizItems(this.graphDir, this.selectedGuide.id);
      this.statusMessage = `Deleted ${itemId} from items.yaml`;
    } catch (err: any) {
      this.statusMessage = `Delete failed: ${err}`;
    }
  }
}

export const studio = new StudioState();
