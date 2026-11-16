export type RightTab = 'inspector' | 'questions' | 'practice';
export type PracticeScope = 'guide' | 'subtree' | 'learning_path' | 'tour' | 'graph';
export type WorkspaceMode = 'graph' | 'quizzes';
export type QuizDeckFilter = 'all' | 'unseen' | 'needs_review';
export type QuizStage = 'overview' | 'running' | 'summary';

export interface PracticeResultState {
  isCorrect: boolean;
  feedback?: string;
}

export interface QuizCardSnapshot {
  answer: Record<string, any>;
  checked: boolean;
  result: PracticeResultState | null;
}

/** Builds the canonical quiz history lookup key matching internal/practice.QuizRecordKey. */
export function getQuizRecordKey(quizId: string, itemId: string): string {
  return `${quizId}::${itemId}`;
}

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

/** Extracts the final directory name from a POSIX or Windows path. */
export function getDirectoryName(dirPath: string): string {
  const trimmed = dirPath.replace(/[/\\]+$/, '');
  if (!trimmed) return dirPath;
  const parts = trimmed.split(/[/\\]/);
  return parts[parts.length - 1] || dirPath;
}
