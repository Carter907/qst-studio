<script lang="ts">
  import { studio } from '../../state.svelte';
  import { getQuizRecordKey } from '../../stores/types';
  import type { quizzes } from '../../../wailsjs/go/models';

  let searchQuery = $state<string>('');

  function getFilteredQuizzes(): quizzes.StandaloneQuizSummaryDTO[] {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return studio.standaloneQuizzes;
    return studio.standaloneQuizzes.filter(item =>
      item.title.toLowerCase().includes(q) ||
      item.id.toLowerCase().includes(q) ||
      (item.tags || []).some(t => t.toLowerCase().includes(q))
    );
  }

  function getQuizStats(quizItem: quizzes.StandaloneQuizSummaryDTO) {
    let attempted = 0;
    let correct = 0;
    const prefix = getQuizRecordKey(quizItem.id, '');
    for (const rec of Object.values(studio.quizHistory)) {
      if (rec.key.startsWith(prefix) || rec.quiz_id === quizItem.id) {
        if (rec.attempts > 0) attempted++;
        if (rec.last_correct) correct++;
      }
    }
    return { attempted, correct };
  }
</script>

{#if studio.graph}
  {@const filteredQuizzes = getFilteredQuizzes()}
  <aside class="quiz-sidebar-pane">
    <div class="pane-header">
      <div class="pane-header-top">
        <span class="pane-title">Quizzes ({filteredQuizzes.length})</span>
        <div class="pane-header-actions">
          <button
            type="button"
            class="btn-text-sm"
            onclick={() => studio.reloadStandaloneQuizzes()}
            title="Reload standalone quizzes from quizzes/"
          >
            Reload
          </button>
          <button
            type="button"
            class="btn-text-sm"
            onclick={() => studio.quizSidebarOpen = false}
            title="Hide Quizzes Sidebar (Ctrl+B)"
          >
            Hide
          </button>
        </div>
      </div>

      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Filter quizzes or tags..."
        class="quiz-search-input"
      />
    </div>

    <div class="quiz-list">
      {#if filteredQuizzes.length === 0}
        <div class="sidebar-empty">
          {studio.standaloneQuizzes.length === 0
            ? 'No YAML files in quizzes/'
            : 'No quizzes match your filter.'}
        </div>
      {:else}
        {#each filteredQuizzes as qSummary}
          {@const stats = getQuizStats(qSummary)}
          <button
            type="button"
            class="quiz-row {qSummary.id === studio.selectedQuizId ? 'selected' : ''}"
            onclick={() => studio.selectStandaloneQuiz(qSummary.id)}
            title={qSummary.description || qSummary.title}
          >
            <span class="quiz-row-title">{qSummary.title}</span>
            <div class="quiz-row-badges">
              {#if stats.attempted > 0}
                <span
                  class="quiz-stat-pill {stats.correct === qSummary.item_count && qSummary.item_count > 0 ? 'mastered' : ''}"
                  title="{stats.correct}/{qSummary.item_count} mastered"
                >
                  {stats.correct}/{qSummary.item_count}
                </span>
              {/if}
              <span class="quiz-count-pill">{qSummary.item_count}q</span>
            </div>
          </button>
        {/each}
      {/if}
    </div>
  </aside>
{/if}

<style>
  .quiz-sidebar-pane {
    background: var(--bg-surface);
    border-right: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .pane-header {
    padding: 8px 10px;
    border-bottom: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .pane-header-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .pane-header-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .pane-title {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-dim);
  }
  .btn-text-sm {
    background: none;
    border: none;
    color: var(--text-dim);
    font-size: 10px;
    cursor: pointer;
    padding: 1px 4px;
    border-radius: 2px;
  }
  .btn-text-sm:hover {
    color: var(--text-main);
  }
  .quiz-search-input {
    width: 100%;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 4px 7px;
    font-size: 11px;
    color: var(--text-main);
  }
  .quiz-search-input:focus {
    outline: none;
    border-color: var(--border-focus);
  }
  .quiz-list {
    flex: 1;
    overflow-y: auto;
    padding: 4px;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }
  .sidebar-empty {
    padding: 20px 10px;
    text-align: center;
    font-size: 11px;
    color: var(--text-dim);
  }
  .quiz-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 6px;
    width: 100%;
    text-align: left;
    background: transparent;
    border: 1px solid transparent;
    border-radius: 4px;
    padding: 6px 8px;
    cursor: pointer;
    transition: background 0.08s ease;
    color: inherit;
  }
  .quiz-row:hover {
    background: var(--bg-card-hover);
  }
  .quiz-row.selected {
    background: var(--selection-bg);
    border-color: var(--border-color);
  }
  .quiz-row-title {
    font-size: 11px;
    font-weight: 500;
    color: var(--text-main);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .quiz-row.selected .quiz-row-title {
    font-weight: 600;
    color: var(--accent-white);
  }
  .quiz-row-badges {
    display: flex;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
  }
  .quiz-stat-pill {
    font-size: 8px;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    color: var(--text-dim);
    padding: 0 4px;
    border-radius: 2px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    line-height: 1.4;
  }
  .quiz-stat-pill.mastered {
    color: #3fb950;
    border-color: rgba(63, 185, 80, 0.35);
  }
  .quiz-count-pill {
    font-size: 9px;
    color: var(--text-dim);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }
</style>
