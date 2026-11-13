import type { main } from '../../wailsjs/go/models';
import type { EditableBlank, EditableChoice, EditablePair } from './types';

export interface BuilderFormSnapshot {
  editingItemId: string | null;
  itemTitle: string;
  itemPrompt: string;
  itemType: string;
  itemChoices: EditableChoice[];
  itemBlanks: EditableBlank[];
  itemPairs: EditablePair[];
  itemDistractors: string[];
  itemOrderItems: string[];
  itemTopLabel: string;
  itemBottomLabel: string;
  itemSolution: string;
}

export type BuildQuizItemResult =
  | { ok: true; item: main.QuizItemDTO }
  | { ok: false; error: string };

/**
 * Validates the question builder form state and constructs a QuizItemDTO payload.
 */
export function buildQuizItemPayload(
  guide: main.GuideDetailDTO,
  segment: string,
  form: BuilderFormSnapshot
): BuildQuizItemResult {
  const trimmedPrompt = form.itemPrompt.trim();
  if (!trimmedPrompt) {
    return { ok: false, error: 'Question prompt cannot be empty.' };
  }

  const idToUse =
    form.editingItemId || `Q_${guide.id.toUpperCase()}_${Date.now().toString().slice(-4)}`;

  const newItem: main.QuizItemDTO = {
    id: idToUse,
    title: form.itemTitle.trim() || `${guide.id} (${segment})`,
    guide_id: guide.id,
    segment,
    scope: guide.scope,
    type: form.itemType,
    prompt: trimmedPrompt,
    created_at: new Date().toISOString()
  } as main.QuizItemDTO;

  if (
    form.itemType === 'multiple_choice' ||
    form.itemType === 'true_false' ||
    form.itemType === 'multiple_answers'
  ) {
    const nonEmptyChoices = form.itemChoices.filter(c => c.text.trim() !== '');
    if (nonEmptyChoices.length < 2) {
      return { ok: false, error: 'Please provide at least 2 non-empty answer choices.' };
    }
    if (!nonEmptyChoices.some(c => c.correct)) {
      return { ok: false, error: 'Please mark at least one choice as correct.' };
    }
    newItem.choices = nonEmptyChoices.map((c, idx) => ({
      id: `c${idx + 1}`,
      text: c.text.trim(),
      correct: c.correct,
      feedback: c.feedback ? c.feedback.trim() : ''
    }));
  } else if (form.itemType === 'fill_in_the_blank') {
    const blanksMap: Record<string, any> = {};
    for (const b of form.itemBlanks) {
      const tok = b.token.trim();
      if (!tok) continue;
      if (b.type === 'dropdown') {
        const opts = b.options.map(o => o.trim()).filter(o => o !== '');
        if (opts.length < 2) {
          return { ok: false, error: `Dropdown [${tok}] requires at least 2 non-empty options.` };
        }
        if (!b.correct || !opts.includes(b.correct)) {
          return { ok: false, error: `Dropdown [${tok}] must have one of its options selected as correct.` };
        }
        blanksMap[tok] = { type: 'dropdown', options: opts, correct: b.correct };
      } else {
        const answers = b.answers.map(a => a.trim()).filter(a => a !== '');
        if (answers.length === 0) {
          return { ok: false, error: `Blank [${tok}] requires at least one acceptable answer string.` };
        }
        blanksMap[tok] = { type: 'text', answers };
      }
    }
    if (Object.keys(blanksMap).length === 0) {
      return { ok: false, error: 'Fill-in-the-blank requires at least one blank definition (e.g. [b1]).' };
    }
    newItem.blanks = blanksMap;
  } else if (form.itemType === 'matching') {
    const cleanPairs = form.itemPairs
      .map(p => ({ prompt: p.prompt.trim(), target: p.target.trim() }))
      .filter(p => p.prompt !== '' && p.target !== '');
    if (cleanPairs.length < 2) {
      return { ok: false, error: 'Matching questions require at least 2 complete pairs.' };
    }
    newItem.pairs = cleanPairs;
    newItem.distractors = form.itemDistractors.map(d => d.trim()).filter(d => d !== '');
  } else if (form.itemType === 'ordering') {
    const cleanItems = form.itemOrderItems.map(i => i.trim()).filter(i => i !== '');
    if (cleanItems.length < 2) {
      return { ok: false, error: 'Ordering questions require at least 2 non-empty sequence items.' };
    }
    newItem.items = cleanItems;
    newItem.top_label = form.itemTopLabel.trim();
    newItem.bottom_label = form.itemBottomLabel.trim();
  } else if (form.itemType === 'essay') {
    newItem.solution = form.itemSolution.trim();
  }

  return { ok: true, item: newItem };
}
