<script lang="ts">
  import { studio } from '../../state.svelte';
  import { renderMarkdownWithLatex } from '../../markdown';
  import type { quiz } from '../../../wailsjs/go/models';

  let { card, qType }: { card: quiz.QuizItemDTO; qType: string } = $props();
</script>

<div class="practice-choices-list">
  {#if qType === 'multiple_answers'}
    {#each studio.practiceAnswer.choices || card.choices || [] as ch}
      {@const selectedList = studio.practiceAnswer.selectedIds || []}
      {@const isSelected = selectedList.includes(ch.id)}
      {@const isRevealed = studio.practiceChecked}
      <button
        type="button"
        class="practice-choice-btn {isSelected ? 'selected' : ''} {isRevealed ? (ch.correct ? 'choice-correct' : (isSelected ? 'choice-wrong' : '')) : ''}"
        disabled={studio.practiceChecked}
        onclick={() => {
          if (isSelected) {
            studio.practiceAnswer.selectedIds = selectedList.filter((id: string) => id !== ch.id);
          } else {
            studio.practiceAnswer.selectedIds = [...selectedList, ch.id];
          }
        }}
      >
        <span class="choice-indicator">{isSelected ? '[✓]' : '[ ]'}</span>
        <span class="choice-text qst-inline-md">{@html renderMarkdownWithLatex(ch.text, true)}</span>
      </button>
    {/each}
  {:else}
    {#each studio.practiceAnswer.choices || card.choices || [] as ch}
      {@const isSelected = studio.practiceAnswer.choiceId === ch.id}
      {@const isRevealed = studio.practiceChecked}
      <button
        type="button"
        class="practice-choice-btn {isSelected ? 'selected' : ''} {isRevealed ? (ch.correct ? 'choice-correct' : (isSelected ? 'choice-wrong' : '')) : ''}"
        disabled={studio.practiceChecked}
        onclick={() => studio.practiceAnswer.choiceId = ch.id}
      >
        <span class="choice-indicator">{isSelected ? '●' : '○'}</span>
        <span class="choice-text qst-inline-md">{@html renderMarkdownWithLatex(ch.text, true)}</span>
      </button>
    {/each}
  {/if}
</div>

<style>
  .practice-choices-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .practice-choice-btn {
    width: 100%;
    text-align: left;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 8px 10px;
    font-size: 11px;
    color: var(--text-main);
    cursor: pointer;
    display: flex;
    align-items: flex-start;
    gap: 8px;
    transition: all 0.1s ease;
  }
  .practice-choice-btn:hover:not(:disabled) {
    border-color: var(--border-focus);
    background: var(--bg-card-hover);
  }
  .practice-choice-btn.selected {
    border-color: var(--accent-white);
    background: var(--selection-bg);
  }
  .practice-choice-btn.choice-correct {
    border-color: #3fb950;
    background: rgba(63, 185, 80, 0.12);
  }
  .practice-choice-btn.choice-wrong {
    border-color: #f85149;
    background: rgba(248, 81, 73, 0.12);
  }
  .choice-indicator {
    font-size: 11px;
    font-family: monospace;
    color: var(--text-dim);
    flex-shrink: 0;
  }
  .choice-text {
    flex: 1;
    line-height: 1.4;
  }
</style>
