<script lang="ts">
  import { studio } from '../../state.svelte';
</script>

<div class="form-group">
  <div class="choices-header">
    <span class="field-label">Answer Choices</span>
    {#if studio.itemType !== 'true_false'}
      <button type="button" class="btn btn-sm btn-secondary" onclick={() => studio.addChoice()}>+ Option</button>
    {/if}
  </div>

  <div class="choices-list">
    {#each studio.itemChoices as choice, idx (choice.id)}
      <div class="choice-editor-card {choice.correct ? 'is-correct' : ''}">
        <div class="choice-top">
          {#if studio.itemType === 'multiple_answers'}
            <label class="radio-label">
              <input
                type="checkbox"
                checked={choice.correct}
                onchange={() => studio.toggleMultipleAnswerCorrect(idx)}
              />
              <span>{choice.correct ? 'Correct Answer' : 'Distractor'} ({choice.id})</span>
            </label>
          {:else}
            <label class="radio-label">
              <input
                type="radio"
                name="correct_choice"
                checked={choice.correct}
                onchange={() => studio.setCorrectChoice(idx)}
              />
              <span>{choice.correct ? 'Correct Answer' : 'Distractor'} ({choice.id})</span>
            </label>
          {/if}

          {#if studio.itemType !== 'true_false'}
            <button
              type="button"
              class="btn-delete"
              disabled={studio.itemChoices.length <= 2}
              onclick={() => studio.removeChoice(idx)}
              title="Remove choice"
            >Remove</button>
          {/if}
        </div>

        <input
          type="text"
          bind:value={choice.text}
          class="input-field"
          placeholder="Option text..."
          disabled={studio.itemType === 'true_false'}
        />
        <input
          type="text"
          bind:value={choice.feedback}
          class="input-field feedback-input"
          placeholder="Diagnostic feedback (optional)..."
        />
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
  .choices-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .choice-editor-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 6px 8px;
    display: flex;
    flex-direction: column;
    gap: 5px;
    transition: border-color 0.1s ease;
  }
  .choice-editor-card.is-correct {
    border-color: var(--accent-white);
  }
  .choice-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .radio-label {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 10px;
    color: var(--text-muted);
    cursor: pointer;
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
  .feedback-input {
    font-size: 10px;
    color: var(--text-muted);
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
