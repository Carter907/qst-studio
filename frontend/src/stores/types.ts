export type RightTab = 'inspector' | 'questions' | 'practice';
export type PracticeScope = 'guide' | 'subtree' | 'learning_path' | 'tour' | 'graph';

export interface EditableChoice {
  id: string;
  text: string;
  correct: boolean;
  feedback: string;
}

export interface EditableBlank {
  token: string;
  type: 'text' | 'dropdown';
  answers: string[];
  options: string[];
  correct: string;
}

export interface EditablePair {
  prompt: string;
  target: string;
}

export interface BuilderNotice {
  type: 'info' | 'error' | 'success';
  text: string;
}

export function defaultChoices(): EditableChoice[] {
  return [
    { id: 'c1', text: '', correct: true, feedback: '' },
    { id: 'c2', text: '', correct: false, feedback: '' },
    { id: 'c3', text: '', correct: false, feedback: '' }
  ];
}

export function defaultTrueFalseChoices(): EditableChoice[] {
  return [
    { id: 'c1', text: 'True', correct: true, feedback: '' },
    { id: 'c2', text: 'False', correct: false, feedback: '' }
  ];
}

export function shuffleArray<T>(items: T[]): T[] {
  const copy = [...items];
  for (let i = copy.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [copy[i], copy[j]] = [copy[j], copy[i]];
  }
  return copy;
}
