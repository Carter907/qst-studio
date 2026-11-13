import type { quiz } from '../../wailsjs/go/models';
import { shuffleArray } from './types';

export interface GradeResult {
  isCorrect: boolean;
  feedback: string;
  selfGraded?: boolean;
}

/**
 * Initializes the interactive answer state for a practice card, shuffling
 * choices, dropdown options, matching targets, or ordering items as needed.
 */
export function initPracticeAnswer(card: quiz.QuizItemDTO): Record<string, any> {
  const qType = card.type || 'multiple_choice';

  if (qType === 'multiple_choice') {
    return {
      choiceId: '',
      choices: shuffleArray(card.choices || [])
    };
  }
  if (qType === 'true_false') {
    return {
      choiceId: '',
      choices: [...(card.choices || [])]
    };
  }
  if (qType === 'multiple_answers') {
    return {
      selectedIds: [],
      choices: shuffleArray(card.choices || [])
    };
  }
  if (qType === 'fill_in_the_blank') {
    const blanksAns: Record<string, string> = {};
    const dropdownOptions: Record<string, string[]> = {};
    if (card.blanks) {
      for (const [tok, def] of Object.entries(card.blanks)) {
        blanksAns[tok] = '';
        if (def && def.type === 'dropdown' && def.options) {
          dropdownOptions[tok] = shuffleArray(def.options);
        }
      }
    }
    return { blanks: blanksAns, dropdownOptions };
  }
  if (qType === 'matching') {
    const pairsAns: Record<string, string> = {};
    const targetSet = new Set<string>();
    if (card.pairs) {
      for (const p of card.pairs) {
        pairsAns[p.prompt] = '';
        if (p.target) targetSet.add(p.target);
      }
    }
    if (card.distractors) {
      for (const d of card.distractors) {
        if (d) targetSet.add(d);
      }
    }
    return {
      pairs: pairsAns,
      targets: shuffleArray(Array.from(targetSet))
    };
  }
  if (qType === 'ordering') {
    return { items: card.items ? shuffleArray(card.items) : [] };
  }
  if (qType === 'essay') {
    return { responseText: '', revealed: false };
  }
  return {};
}

/**
 * Grades a user's answer against a practice card.
 * Returns `{ selfGraded: true }` for essay items so the caller can reveal the rubric.
 */
export function gradePracticeAnswer(card: quiz.QuizItemDTO, answer: Record<string, any>): GradeResult {
  const qType = card.type || 'multiple_choice';
  let isCorrect = false;
  let feedback = card.feedback || '';

  if (qType === 'multiple_choice' || qType === 'true_false') {
    const matchingChoice = (card.choices || []).find(c => c.id === answer.choiceId);
    isCorrect = matchingChoice ? matchingChoice.correct : false;
    if (matchingChoice?.feedback) {
      feedback = matchingChoice.feedback;
    }
  } else if (qType === 'multiple_answers') {
    const selected = new Set<string>(answer.selectedIds || []);
    const allChoices = card.choices || [];
    const correctIds = allChoices.filter(c => c.correct).map(c => c.id);
    const incorrectIds = allChoices.filter(c => !c.correct).map(c => c.id);
    isCorrect = correctIds.every(id => selected.has(id)) && incorrectIds.every(id => !selected.has(id));
  } else if (qType === 'fill_in_the_blank') {
    const userBlanks = answer.blanks || {};
    const cardBlanks = card.blanks || {};
    isCorrect = Object.entries(cardBlanks).every(([tok, def]) => {
      const userVal = (userBlanks[tok] || '').trim().toLowerCase();
      if (def.type === 'dropdown') {
        return userVal === (def.correct || '').trim().toLowerCase();
      }
      return (def.answers || []).some(a => a.trim().toLowerCase() === userVal);
    });
  } else if (qType === 'matching') {
    const userPairs = answer.pairs || {};
    const cardPairs = card.pairs || [];
    isCorrect = cardPairs.length > 0 && cardPairs.every(p => userPairs[p.prompt] === p.target);
  } else if (qType === 'ordering') {
    const userItems = answer.items || [];
    const cardItems = card.items || [];
    isCorrect =
      userItems.length === cardItems.length &&
      userItems.every((item: string, idx: number) => item === cardItems[idx]);
  } else if (qType === 'essay') {
    return { isCorrect: false, feedback, selfGraded: true };
  }

  return { isCorrect, feedback };
}
