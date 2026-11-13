<script lang="ts">
  import { studio } from '../../state.svelte';
  import { renderMarkdownWithLatex } from '../../markdown';
</script>

<div class="builder-preview-card">
  <div class="preview-header-row">
    <span class="preview-type-badge">{studio.itemType.replace(/_/g, ' ')}</span>
    {#if studio.itemTitle}
      <div class="turn-title">{studio.itemTitle}</div>
    {/if}
  </div>

  <div class="item-prompt qst-prose">
    {@html renderMarkdownWithLatex(studio.itemPrompt || '(No prompt entered yet)')}
  </div>

  {#if studio.itemType === 'multiple_choice' || studio.itemType === 'true_false' || studio.itemType === 'multiple_answers'}
    <div class="preview-options">
      {#each studio.itemChoices as ch (ch.id)}
        <div class="preview-choice-card {ch.correct ? 'is-correct' : ''}">
          <div class="preview-choice-row">
            <span class="option-mark">{ch.correct ? 'CORRECT' : 'OPTION'}</span>
            <div class="option-text qst-inline-md">
              {@html renderMarkdownWithLatex(ch.text || '(Empty option)', true)}
            </div>
          </div>
          {#if ch.feedback}
            <div class="preview-feedback qst-prose">
              {@html renderMarkdownWithLatex(ch.feedback, true)}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {:else if studio.itemType === 'fill_in_the_blank'}
    <div class="preview-blanks-summary">
      <span class="field-label">Declared Blanks</span>
      {#each studio.itemBlanks as b}
        <div class="blank-preview-item">
          <span class="blank-token-pill">[{b.token}]</span>
          {#if b.type === 'dropdown'}
            <span class="blank-preview-detail">Dropdown: <strong>{b.correct || '(none)'}</strong> (Options: {b.options.filter(o => o.trim()).join(', ')})</span>
          {:else}
            <span class="blank-preview-detail">Text Answers: <strong>{b.answers.filter(a => a.trim()).join(' | ')}</strong></span>
          {/if}
        </div>
      {/each}
    </div>
  {:else if studio.itemType === 'matching'}
    <div class="preview-matching-table">
      <span class="field-label">Matching Pairs</span>
      {#each studio.itemPairs as p}
        <div class="matching-pair-preview">
          <span class="match-left">{p.prompt || '(prompt)'}</span>
          <span class="match-arrow">→</span>
          <span class="match-right">{p.target || '(target)'}</span>
        </div>
      {/each}
      {#if studio.itemDistractors.length > 0}
        <div class="distractors-preview">
          <span class="field-label">Distractors:</span>
          <span class="distractors-list">{studio.itemDistractors.filter(d => d.trim()).join(', ')}</span>
        </div>
      {/if}
    </div>
  {:else if studio.itemType === 'ordering'}
    <div class="preview-ordering-card">
      {#if studio.itemTopLabel}
        <div class="ordering-label-bar">Top: {studio.itemTopLabel}</div>
      {/if}
      <div class="ordering-items-preview">
        {#each studio.itemOrderItems as it, idx}
          <div class="order-item-row">
            <span class="order-num">{idx + 1}.</span>
            <span>{it || '(empty item)'}</span>
          </div>
        {/each}
      </div>
      {#if studio.itemBottomLabel}
        <div class="ordering-label-bar">Bottom: {studio.itemBottomLabel}</div>
      {/if}
    </div>
  {:else if studio.itemType === 'essay'}
    {#if studio.itemSolution}
      <div class="preview-rubric-card">
        <span class="field-label">Reference Solution / Rubric:</span>
        <div class="qst-prose">
          {@html renderMarkdownWithLatex(studio.itemSolution)}
        </div>
      </div>
    {/if}
  {/if}
</div>

<style>
  .builder-preview-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .preview-header-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .preview-type-badge {
    font-size: 9px;
    font-weight: 700;
    text-transform: uppercase;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 2px;
    padding: 1px 5px;
    color: var(--text-muted);
  }
  .turn-title {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-main);
  }
  .preview-options {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .preview-choice-card {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 6px 8px;
  }
  .preview-choice-card.is-correct {
    border-color: var(--accent-white);
  }
  .preview-choice-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .option-mark {
    font-size: 8px;
    font-weight: 700;
    letter-spacing: 0.5px;
    padding: 1px 4px;
    border-radius: 2px;
    background: var(--bg-card);
    color: var(--text-dim);
  }
  .preview-choice-card.is-correct .option-mark {
    background: var(--accent-white);
    color: var(--text-dark);
  }
  .preview-feedback {
    margin-top: 3px;
    font-size: 10px;
    color: var(--text-dim);
    padding-left: 14px;
  }
  .preview-blanks-summary, .preview-matching-table, .preview-ordering-card {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .blank-preview-item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
  }
  .blank-token-pill {
    font-family: monospace;
    font-weight: 700;
    color: var(--text-main);
  }
  .matching-pair-preview {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 11px;
  }
  .match-left {
    font-weight: 600;
    color: var(--text-main);
  }
  .match-right {
    color: var(--text-muted);
  }
  .match-arrow {
    color: var(--text-dim);
    font-weight: 700;
  }
  .distractors-preview {
    margin-top: 4px;
    font-size: 10px;
    color: var(--text-dim);
  }
  .ordering-label-bar {
    font-size: 9px;
    font-weight: 700;
    text-transform: uppercase;
    color: var(--text-dim);
    padding: 2px 4px;
    background: var(--bg-card);
    border-radius: 2px;
  }
  .ordering-items-preview {
    display: flex;
    flex-direction: column;
    gap: 3px;
    padding: 2px 4px;
  }
  .order-item-row {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
  }
  .order-num {
    font-size: 11px;
    color: var(--text-dim);
    font-weight: 600;
    width: 16px;
  }
  .preview-rubric-card {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 8px;
    font-size: 11px;
  }
</style>
