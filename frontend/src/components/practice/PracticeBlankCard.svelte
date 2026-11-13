<script lang="ts">
  import { studio } from '../../state.svelte';
  import type { quiz } from '../../../wailsjs/go/models';

  let { card }: { card: quiz.QuizItemDTO } = $props();
</script>

<div class="practice-blanks-container">
  <span class="field-label">Fill in each blank token:</span>
  {#each Object.entries(card.blanks || {}) as [tok, def]}
    <div class="practice-blank-row">
      <span class="blank-tag">[{tok}]</span>
      {#if def.type === 'dropdown'}
        <select
          class="input-field select-field"
          disabled={studio.practiceChecked}
          bind:value={studio.practiceAnswer.blanks[tok]}
        >
          <option value="">Select an option...</option>
          {#each (studio.practiceAnswer.dropdownOptions && studio.practiceAnswer.dropdownOptions[tok]) || def.options || [] as opt}
            <option value={opt}>{opt}</option>
          {/each}
        </select>
      {:else}
        <input
          type="text"
          class="input-field"
          placeholder="Enter answer..."
          disabled={studio.practiceChecked}
          bind:value={studio.practiceAnswer.blanks[tok]}
        />
      {/if}
    </div>
  {/each}
</div>

<style>
  .practice-blanks-container {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .practice-blank-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .blank-tag {
    font-family: monospace;
    font-weight: 700;
    font-size: 11px;
    color: var(--text-dim);
    width: 32px;
  }
  .input-field {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 6px 8px;
    font-size: 11px;
    color: var(--text-main);
  }
</style>
