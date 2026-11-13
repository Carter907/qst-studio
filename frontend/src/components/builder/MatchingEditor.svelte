<script lang="ts">
  import { studio } from '../../state.svelte';
</script>

<div class="form-group">
  <div class="choices-header">
    <span class="field-label">Matching Pairs</span>
    <button type="button" class="btn btn-sm btn-secondary" onclick={() => studio.addMatchingPair()}>+ Pair</button>
  </div>

  <div class="pairs-list">
    {#each studio.itemPairs as pair, pIdx}
      <div class="pair-editor-card">
        <input
          type="text"
          bind:value={pair.prompt}
          class="input-field"
          placeholder="Prompt (Left)..."
        />
        <span class="match-arrow">→</span>
        <input
          type="text"
          bind:value={pair.target}
          class="input-field"
          placeholder="Target Match (Right)..."
        />
        <button
          type="button"
          class="btn-delete"
          disabled={studio.itemPairs.length <= 2}
          onclick={() => studio.removeMatchingPair(pIdx)}
        >Remove</button>
      </div>
    {/each}
  </div>

  <div class="distractors-group">
    <div class="choices-header">
      <span class="field-label">Unmatched Distractors (Optional)</span>
      <button type="button" class="btn btn-sm btn-secondary" onclick={() => studio.addDistractor()}>+ Distractor</button>
    </div>
    {#each studio.itemDistractors as dist, dIdx}
      <div class="distractor-row">
        <input
          type="text"
          bind:value={studio.itemDistractors[dIdx]}
          class="input-field"
          placeholder="Distractor option..."
        />
        <button
          type="button"
          class="btn-delete"
          onclick={() => studio.removeDistractor(dIdx)}
        >Remove</button>
      </div>
    {/each}
  </div>
</div>

<style>
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .choices-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .pairs-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .pair-editor-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 6px 8px;
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 6px;
  }
  .match-arrow {
    color: var(--text-dim);
    font-weight: 700;
  }
  .distractors-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-top: 6px;
  }
  .distractor-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .input-field {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 5px 7px;
    font-size: 11px;
    color: var(--text-main);
  }
  .input-field:focus {
    outline: none;
    border-color: var(--border-focus);
  }
  .btn-delete {
    background: none;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    font-size: 10px;
    padding: 1px 3px;
  }
  .btn-delete:hover:not(:disabled) {
    color: #f85149;
  }
</style>
