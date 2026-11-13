<script lang="ts">
  import { studio } from '../../state.svelte';
  import BuilderPreview from './BuilderPreview.svelte';
  import ChoiceEditor from './ChoiceEditor.svelte';
  import BlankEditor from './BlankEditor.svelte';
  import MatchingEditor from './MatchingEditor.svelte';
  import OrderingEditor from './OrderingEditor.svelte';
  import SavedItemsList from './SavedItemsList.svelte';

  let builderPreviewMode = $state<boolean>(false);
</script>

<div class="builder-container">
  <div class="builder-sub-header">
    <div class="builder-mode-group">
      <span class="field-label">Author</span>
      <span class="builder-mode-pill">{studio.editingItemId ? `Editing ${studio.editingItemId}` : 'New'}</span>
    </div>
    <div class="builder-actions">
      <button
        type="button"
        class="btn btn-secondary btn-sm {builderPreviewMode ? 'btn-active-toggle' : ''}"
        onclick={() => builderPreviewMode = !builderPreviewMode}
        title="Toggle Preview"
      >
        {builderPreviewMode ? 'Edit Form' : 'Preview'}
      </button>
      {#if studio.editingItemId || studio.itemPrompt.trim() !== ''}
        <button type="button" class="btn btn-secondary btn-sm" onclick={() => studio.resetBuilderForm()}>
          New
        </button>
      {/if}
    </div>
  </div>

  <div class="builder-body">
    <div class="form-group">
      <span class="field-label">Target Segment</span>
      <div class="anchor-badge">
        <span>{studio.selectedGuideId}</span>
        <span class="badge-separator">/</span>
        <span>Lines {studio.getActiveSegment()}</span>
        <span class="badge-separator">/</span>
        <span>{studio.selectedGuide?.scope}</span>
      </div>
    </div>

    <div class="form-group">
      <label for="item-type-select" class="field-label">Question Type</label>
      <select
        id="item-type-select"
        class="input-field select-field"
        value={studio.itemType}
        onchange={(e) => studio.changeItemType((e.target as HTMLSelectElement).value)}
      >
        <option value="multiple_choice">Multiple Choice (Single Select)</option>
        <option value="multiple_answers">Multiple Answers (Multi-Select)</option>
        <option value="true_false">True / False</option>
        <option value="fill_in_the_blank">Fill in the Blank (Text & Dropdown)</option>
        <option value="matching">Matching (Pairs & Distractors)</option>
        <option value="ordering">Ordering (Sequencing)</option>
        <option value="essay">Essay / Open Response</option>
      </select>
    </div>

    {#if studio.builderNotice}
      <div class="builder-notice notice-{studio.builderNotice.type}">
        <span>{studio.builderNotice.text}</span>
        <button
          type="button"
          class="notice-close"
          onclick={() => studio.builderNotice = null}
          title="Dismiss notice"
        >Dismiss</button>
      </div>
    {/if}

    {#if builderPreviewMode}
      <BuilderPreview />
    {:else}
      <div class="form-group">
        <label for="item-title-input" class="field-label">Question Title</label>
        <input
          id="item-title-input"
          type="text"
          bind:value={studio.itemTitle}
          class="input-field"
          placeholder="Concept title (optional)..."
        />
      </div>

      <div class="form-group">
        <label for="item-prompt-input" class="field-label">Question Prompt (Markdown & $LaTeX$ supported)</label>
        <textarea
          id="item-prompt-input"
          rows="3"
          bind:value={studio.itemPrompt}
          class="input-field textarea"
          placeholder={studio.itemType === 'fill_in_the_blank' ? "Stem with tokens like [b1] or [d1]..." : "Enter question stem..."}
        ></textarea>
      </div>

      {#if studio.itemType === 'multiple_choice' || studio.itemType === 'true_false' || studio.itemType === 'multiple_answers'}
        <ChoiceEditor />
      {:else if studio.itemType === 'fill_in_the_blank'}
        <BlankEditor />
      {:else if studio.itemType === 'matching'}
        <MatchingEditor />
      {:else if studio.itemType === 'ordering'}
        <OrderingEditor />
      {:else if studio.itemType === 'essay'}
        <div class="form-group">
          <label for="item-solution-input" class="field-label">Reference Solution / Grading Rubric (Optional)</label>
          <textarea
            id="item-solution-input"
            rows="3"
            bind:value={studio.itemSolution}
            class="input-field textarea"
            placeholder="Reference explanation or solution points for grading..."
          ></textarea>
        </div>
      {/if}
    {/if}

    <button type="button" class="btn btn-primary btn-block" onclick={() => studio.saveItem()}>
      {studio.editingItemId ? `Update ${studio.editingItemId}` : 'Save Question'}
    </button>

    <SavedItemsList />
  </div>
</div>

<style>
  .builder-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
  }
  .builder-sub-header {
    padding: 6px 10px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-card);
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-shrink: 0;
  }
  .builder-mode-group {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .builder-mode-pill {
    font-size: 9px;
    font-family: monospace;
    color: var(--text-muted);
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    padding: 1px 4px;
    border-radius: 2px;
  }
  .builder-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .btn-active-toggle {
    background: var(--accent-white);
    color: var(--text-dark);
    font-weight: 600;
  }
  .builder-body {
    flex: 1;
    overflow-y: auto;
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .select-field {
    cursor: pointer;
  }
  .anchor-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 3px 6px;
    font-size: 10px;
    color: var(--text-muted);
  }
  .badge-separator {
    color: var(--border-color);
  }
  .builder-notice {
    padding: 5px 8px;
    border-radius: 3px;
    font-size: 11px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .notice-info {
    background: rgba(88, 166, 255, 0.1);
    color: #58a6ff;
    border: 1px solid rgba(88, 166, 255, 0.3);
  }
  .notice-success {
    background: rgba(63, 185, 80, 0.1);
    color: #3fb950;
    border: 1px solid rgba(63, 185, 80, 0.3);
  }
  .notice-error {
    background: rgba(248, 81, 73, 0.1);
    color: #f85149;
    border: 1px solid rgba(248, 81, 73, 0.3);
  }
  .notice-close {
    background: none;
    border: none;
    color: inherit;
    cursor: pointer;
    font-size: 10px;
    text-decoration: underline;
    padding: 0 2px;
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
  .textarea {
    resize: vertical;
    font-family: inherit;
  }
  .btn-block {
    width: 100%;
    padding: 7px;
    font-size: 11px;
    font-weight: 600;
  }
</style>
