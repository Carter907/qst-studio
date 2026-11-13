<script lang="ts">
  import { studio } from '../../state.svelte';
  import type { quiz } from '../../../wailsjs/go/models';

  let { card }: { card: quiz.QuizItemDTO } = $props();

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

  let targets = $derived(studio.practiceAnswer.targets || getAvailableMatchingTargets(card));
</script>

<div class="practice-matching-container">
  <span class="field-label">Match each prompt to its target:</span>
  {#each card.pairs || [] as pair}
    <div class="practice-match-row">
      <span class="match-prompt-text">{pair.prompt}</span>
      <span class="match-arrow">→</span>
      <select
        class="input-field select-field match-select"
        disabled={studio.practiceChecked}
        bind:value={studio.practiceAnswer.pairs[pair.prompt]}
      >
        <option value="">Choose match...</option>
        {#each targets as tgt}
          <option value={tgt}>{tgt}</option>
        {/each}
      </select>
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
  .input-field {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-main);
  }
  .match-select {
    font-size: 10px;
    padding: 3px 6px;
  }
</style>
