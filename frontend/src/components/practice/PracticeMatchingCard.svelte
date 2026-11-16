<script lang="ts">
  import { renderMarkdownWithLatex } from '../../markdown';
  import RichOptionSelect from './RichOptionSelect.svelte';
  import type { quiz } from '../../../wailsjs/go/models';

  let {
    card,
    answer,
    checked
  }: {
    card: quiz.QuizItemDTO;
    answer: Record<string, any>;
    checked: boolean;
  } = $props();

  function getAvailableMatchingTargets(item: quiz.QuizItemDTO): string[] {
    const set = new Set<string>();
    if (item.pairs) {
      for (const p of item.pairs) {
        if (p.target) set.add(p.target);
      }
    }
    if (item.distractors) {
      for (const d of item.distractors) {
        if (d) set.add(d);
      }
    }
    return Array.from(set);
  }

  let targets = $derived(answer.targets || getAvailableMatchingTargets(card));
</script>

<div class="practice-matching-container">
  <span class="field-label">Match each prompt to its target:</span>
  {#each card.pairs || [] as pair}
    <div class="practice-match-row">
      <span class="match-prompt-text qst-inline-md">{@html renderMarkdownWithLatex(pair.prompt, true)}</span>
      <span class="match-arrow">→</span>
      <RichOptionSelect
        bind:value={answer.pairs[pair.prompt]}
        options={targets}
        placeholder="Choose match..."
        disabled={checked}
      />
    </div>
  {/each}
</div>

<style>
  .practice-matching-container {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .practice-match-row {
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    gap: 6px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 6px 8px;
    border-radius: 4px;
  }
  .match-prompt-text {
    font-size: 11px;
    font-weight: 500;
  }
  .match-arrow {
    color: var(--text-dim);
  }
</style>
