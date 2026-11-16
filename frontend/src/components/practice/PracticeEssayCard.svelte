<script lang="ts">
  import { renderMarkdownWithLatex } from '../../markdown';
  import type { quiz } from '../../../wailsjs/go/models';

  let {
    card,
    answer,
    checked,
    onSelfGrade
  }: {
    card: quiz.QuizItemDTO;
    answer: Record<string, any>;
    checked: boolean;
    onSelfGrade: (understood: boolean) => void;
  } = $props();
</script>

<div class="practice-essay-container">
  <textarea
    class="input-field textarea"
    rows="4"
    placeholder="Write your answer or thoughts before revealing the solution..."
    bind:value={answer.responseText}
    disabled={answer.revealed}
  ></textarea>
  {#if answer.revealed}
    <div class="essay-solution-box">
      <span class="field-label">Reference Solution / Rubric:</span>
      <div class="qst-prose">
        {@html renderMarkdownWithLatex(card.solution || '(No reference solution provided by author)')}
      </div>
      {#if !checked}
        <div class="essay-self-grade-actions">
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            onclick={() => onSelfGrade(true)}
          >
            Understood (Mark Correct)
          </button>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            onclick={() => onSelfGrade(false)}
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
