<script lang="ts">
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
</script>

<div class="practice-blanks-container">
  <span class="field-label">Fill in each blank token:</span>
  {#each Object.entries(card.blanks || {}) as [tok, def]}
    <div class="practice-blank-row">
      <span class="blank-tag">[{tok}]</span>
      {#if def.type === 'dropdown'}
        <RichOptionSelect
          bind:value={answer.blanks[tok]}
          options={(answer.dropdownOptions && answer.dropdownOptions[tok]) || def.options || []}
          placeholder="Select an option..."
          disabled={checked}
        />
      {:else}
        <input
          type="text"
          class="input-field"
          placeholder="Enter answer..."
          disabled={checked}
          bind:value={answer.blanks[tok]}
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
