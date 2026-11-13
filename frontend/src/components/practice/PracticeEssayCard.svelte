<script lang="ts">
  import { studio } from '../../state.svelte';
  import { renderMarkdownWithLatex } from '../../markdown';
  import type { main } from '../../../wailsjs/go/models';

  let { card }: { card: main.QuizItemDTO } = $props();
</script>

<div class="practice-essay-container">
  <textarea
    class="input-field textarea"
    rows="4"
    placeholder="Write your answer or thoughts before revealing the solution..."
    bind:value={studio.practiceAnswer.responseText}
    disabled={studio.practiceAnswer.revealed}
  ></textarea>
  {#if studio.practiceAnswer.revealed}
    <div class="essay-solution-box">
      <span class="field-label">Reference Solution / Rubric:</span>
      <div class="qst-prose">
        {@html renderMarkdownWithLatex(card.solution || '(No reference solution provided by author)')}
      </div>
      {#if !studio.practiceChecked}
        <div class="essay-self-grade-actions">
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            onclick={() => studio.selfGradeEssay(true)}
          >
            Understood (Mark Correct)
          </button>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            onclick={() => studio.selfGradeEssay(false)}
          >
            Needs Review
          </button>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .practice-essay-container {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .essay-solution-box {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .essay-self-grade-actions {
    display: flex;
    gap: 6px;
    margin-top: 6px;
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
