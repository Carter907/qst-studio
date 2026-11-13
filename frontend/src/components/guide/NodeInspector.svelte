<script lang="ts">
  import { studio } from '../../state.svelte';
  import QuestionBuilder from '../builder/QuestionBuilder.svelte';
  import PracticeRunner from '../practice/PracticeRunner.svelte';
</script>

<section class="inspector-pane">
  <div class="inspector-tabs-header">
    <div class="tab-buttons">
      <button
        type="button"
        class="tab-btn {studio.activeRightTab === 'inspector' ? 'active' : ''}"
        onclick={() => studio.activeRightTab = 'inspector'}
      >
        Details
      </button>
      <button
        type="button"
        class="tab-btn {studio.activeRightTab === 'questions' ? 'active' : ''}"
        onclick={() => studio.activeRightTab = 'questions'}
      >
        Author ({studio.quizItems.length})
      </button>
      <button
        type="button"
        class="tab-btn {studio.activeRightTab === 'practice' ? 'active' : ''}"
        onclick={() => studio.openPracticeTab()}
      >
        Practice
      </button>
    </div>

    <button
      type="button"
      class="btn-text-sm"
      onclick={() => studio.rightPaneOpen = false}
      title="Close Right Panel"
    >
      Hide
    </button>
  </div>

  {#if studio.activeRightTab === 'inspector'}
    {#if studio.selectedGuide}
      <div class="inspector-body">
        <!-- Summary Card -->
        <div class="inspector-summary-card">
          <div class="inspector-title-row">
            <h3 class="inspector-title">{studio.selectedGuide.id}</h3>
            <span class="scope-pill active-scope-pill">{studio.selectedGuide.scope}</span>
          </div>
          <div class="inspector-meta-row">
            <span>{studio.selectedGuide.line_count} lines</span>
            <span>•</span>
            <span>{studio.quizItems.length} question{studio.quizItems.length === 1 ? '' : 's'}</span>
          </div>
        </div>

        <!-- Quick Switch to Practice or Author -->
        <div class="inspector-actions-box">
          {#if studio.quizItems.length > 0}
            <button
              type="button"
              class="btn btn-primary btn-block"
              onclick={() => studio.openPracticeTab('guide', studio.selectedGuide!.id)}
            >
              Practice Questions ({studio.quizItems.length}) →
            </button>
          {/if}
          {#if studio.selectedGuide.sub_guides && studio.selectedGuide.sub_guides.length > 0}
            <button
              type="button"
              class="btn btn-secondary btn-block"
              onclick={() => studio.openPracticeTab('subtree', studio.selectedGuide!.id)}
            >
              Practice Subtree →
            </button>
          {/if}
          {#if studio.selectedGuide.prerequisites && studio.selectedGuide.prerequisites.length > 0}
            <button
              type="button"
              class="btn btn-secondary btn-block"
              onclick={() => studio.openPracticeTab('learning_path', studio.selectedGuide!.id)}
            >
              Practice Learning Path →
            </button>
          {/if}
          <button
            type="button"
            class="btn btn-secondary btn-block"
            onclick={() => studio.activeRightTab = 'questions'}
          >
            + Author New Question
          </button>
        </div>

        <!-- Prerequisites -->
        <div class="inspector-section">
          <span class="field-label">Prerequisites ({studio.selectedGuide.prerequisites?.length || 0})</span>
          {#if studio.selectedGuide.prerequisites && studio.selectedGuide.prerequisites.length > 0}
            <div class="inspector-chip-list">
              {#each studio.selectedGuide.prerequisites as prereq}
                <button
                  type="button"
                  class="meta-chip clickable-chip"
                  onclick={() => studio.selectGuide(prereq)}
                >
                  {prereq}
                </button>
              {/each}
            </div>
          {:else}
            <div class="inspector-empty-note">Foundational leaf (no prerequisites)</div>
          {/if}
        </div>

        <!-- Linked Sub-guides -->
        {#if studio.selectedGuide.sub_guides && studio.selectedGuide.sub_guides.length > 0}
          <div class="inspector-section">
            <span class="field-label">Sub-Guides ({studio.selectedGuide.sub_guides.length})</span>
            <div class="inspector-chip-list">
              {#each studio.selectedGuide.sub_guides as sg}
                <button
                  type="button"
                  class="meta-chip clickable-chip"
                  onclick={() => studio.selectGuide(sg.guide)}
                >
                  <span>{sg.guide}</span>
                  <span class="chip-segment">{sg.segment}</span>
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Learning Path Sequence -->
        {#if studio.learningPath && studio.learningPath.length > 0}
          <div class="inspector-section">
            <span class="field-label">Prerequisite Dependency Order</span>
            <div class="sequence-list">
              {#each studio.learningPath as step, idx}
                <div class="sequence-item {step === studio.selectedGuide.id ? 'current' : ''}">
                  <span class="sequence-num">{idx + 1}.</span>
                  <button
                    type="button"
                    class="sequence-link"
                    onclick={() => studio.selectGuide(step)}
                  >
                    {step}
                  </button>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {:else}
      <div class="empty-state">Select a guide from the sidebar to inspect details.</div>
    {/if}
  {:else if studio.activeRightTab === 'questions'}
    <QuestionBuilder />
  {:else}
    <PracticeRunner />
  {/if}
</section>

<style>
  .inspector-pane {
    background: var(--bg-surface);
    border-left: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .inspector-tabs-header {
    height: 38px;
    padding: 0 10px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-card);
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-shrink: 0;
  }
  .tab-buttons {
    display: flex;
    gap: 3px;
  }
  .tab-btn {
    background: none;
    border: none;
    padding: 4px 7px;
    font-size: 11px;
    font-weight: 500;
    color: var(--text-muted);
    cursor: pointer;
    border-radius: 3px;
    transition: all 0.1s ease;
  }
  .tab-btn:hover {
    color: var(--text-main);
  }
  .tab-btn.active {
    color: var(--text-main);
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    font-weight: 600;
  }
  .btn-text-sm {
    background: none;
    border: none;
    color: var(--text-dim);
    font-size: 10px;
    cursor: pointer;
    padding: 2px 4px;
    border-radius: 2px;
  }
  .btn-text-sm:hover {
    color: var(--text-main);
  }
  .inspector-body {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .inspector-summary-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .inspector-title-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .inspector-title {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
    color: var(--text-main);
  }
  .inspector-meta-row {
    font-size: 11px;
    color: var(--text-dim);
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .inspector-actions-box {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .btn-block {
    width: 100%;
    padding: 6px;
    font-size: 11px;
    font-weight: 500;
  }
  .inspector-section {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .inspector-chip-list {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
  .clickable-chip {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 3px 6px;
    font-size: 11px;
    color: var(--text-main);
    cursor: pointer;
    transition: all 0.1s ease;
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }
  .clickable-chip:hover {
    border-color: var(--border-focus);
    background: var(--bg-card-hover);
  }
  .chip-segment {
    font-size: 9px;
    color: var(--text-dim);
  }
  .inspector-empty-note {
    font-size: 11px;
    color: var(--text-dim);
    font-style: italic;
  }
  .sequence-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 6px 8px;
  }
  .sequence-item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
  }
  .sequence-item.current {
    font-weight: 600;
    color: var(--accent-white);
  }
  .sequence-num {
    color: var(--text-dim);
    font-size: 10px;
    width: 14px;
  }
  .sequence-link {
    background: none;
    border: none;
    color: inherit;
    cursor: pointer;
    font-size: 11px;
    padding: 0;
    text-align: left;
  }
  .sequence-link:hover {
    text-decoration: underline;
  }
  .empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-dim);
    font-size: 12px;
    padding: 16px;
    text-align: center;
  }
</style>
