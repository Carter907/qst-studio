<script lang="ts">
  import { studio } from '../../state.svelte';
  import type { graph } from '../../../wailsjs/go/models';

  let selectedScope = $state<string>('all');
  let guideSearchQuery = $state<string>('');

  function getFilteredGuides(): graph.GuideSummaryDTO[] {
    if (!studio.graph) return [];
    const q = guideSearchQuery.trim().toLowerCase();
    return studio.graph.guides.filter(g => {
      const matchesScope = selectedScope === 'all' || g.scope === selectedScope;
      const matchesQuery = !q || g.id.toLowerCase().includes(q) || g.scope.toLowerCase().includes(q);
      return matchesScope && matchesQuery;
    });
  }

  function getGuidePracticeStats(guideId: string) {
    let attempted = 0;
    let correct = 0;
    if (studio.practiceHistory) {
      for (const rec of Object.values(studio.practiceHistory)) {
        if (rec.guide_id === guideId) {
          if (rec.attempts > 0) attempted++;
          if (rec.last_correct) correct++;
        }
      }
    }
    return { attempted, correct };
  }
</script>

{#if studio.graph}
  {@const filteredGuides = getFilteredGuides()}
  <aside class="sidebar-pane">
    <div class="pane-header">
      <div class="pane-header-top">
        <span class="pane-title">Guides ({filteredGuides.length})</span>
        <button
          type="button"
          class="btn-text-sm"
          onclick={() => studio.leftPaneOpen = false}
          title="Collapse Guide Navigator (Ctrl+B)"
        >
          Hide
        </button>
      </div>

      <input
        type="text"
        bind:value={guideSearchQuery}
        placeholder="Filter guides..."
        class="guide-search-input"
      />

      <div class="scope-tags">
        <button
          class="scope-btn {selectedScope === 'all' ? 'active' : ''}"
          onclick={() => selectedScope = 'all'}
        >All</button>
        {#each studio.graph.scopes as sc}
          <button
            class="scope-btn {selectedScope === sc ? 'active' : ''}"
            onclick={() => selectedScope = sc}
          >{sc}</button>
        {/each}
      </div>
    </div>

    <div class="guide-list">
      {#each filteredGuides as guide}
        {@const stats = getGuidePracticeStats(guide.id)}
        <button
          type="button"
          class="guide-row {guide.id === studio.selectedGuideId ? 'selected' : ''}"
          onclick={() => studio.selectGuide(guide.id)}
        >
          <span class="guide-name">{guide.id}</span>
          <div class="guide-badges">
            {#if stats.attempted > 0}
              <span
                class="guide-practice-stat {stats.correct === stats.attempted ? 'mastered' : ''}"
                title="{stats.correct}/{stats.attempted} questions mastered"
              >
                {stats.correct}/{stats.attempted}
              </span>
            {/if}
            <span class="guide-scope-label">{guide.scope}</span>
          </div>
        </button>
      {/each}
    </div>
  </aside>
{/if}

<style>
  .sidebar-pane {
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
  .guide-search-input {
    width: 100%;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 4px 7px;
    font-size: 11px;
    color: var(--text-main);
  }
  .guide-search-input:focus {
    outline: none;
    border-color: var(--border-focus);
  }
  .scope-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 3px;
  }
  .scope-btn {
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: 3px;
    color: var(--text-dim);
    font-size: 9px;
    text-transform: capitalize;
    padding: 1px 5px;
    cursor: pointer;
    transition: all 0.1s ease;
  }
  .scope-btn:hover {
    color: var(--text-main);
    border-color: var(--border-focus);
  }
  .scope-btn.active {
    background: var(--accent-white);
    color: var(--text-dark);
    border-color: var(--accent-white);
    font-weight: 600;
  }

  .guide-list {
    flex: 1;
    overflow-y: auto;
    padding: 4px;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }
  .guide-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    text-align: left;
    background: transparent;
    border: 1px solid transparent;
    border-radius: 4px;
    padding: 5px 8px;
    cursor: pointer;
    transition: background 0.08s ease;
    color: inherit;
  }
  .guide-row:hover {
    background: var(--bg-card-hover);
  }
  .guide-row.selected {
    background: var(--selection-bg);
    border-color: var(--border-color);
  }
  .guide-name {
    font-size: 11px;
    font-weight: 500;
    color: var(--text-main);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin-right: 6px;
  }
  .guide-row.selected .guide-name {
    font-weight: 600;
    color: var(--accent-white);
  }
  .guide-badges {
    display: flex;
    align-items: center;
    gap: 5px;
    flex-shrink: 0;
  }
  .guide-practice-stat {
    font-size: 8px;
    font-family: var(--font-mono);
    color: var(--text-dim);
    padding: 0 4px;
    border-radius: 2px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    line-height: 1.4;
  }
  .guide-practice-stat.mastered {
    color: var(--color-correct);
    border-color: rgba(62, 199, 111, 0.3);
  }
  .guide-scope-label {
    font-size: 9px;
    color: var(--text-dim);
    text-transform: uppercase;
    font-weight: 600;
    flex-shrink: 0;
  }
</style>
