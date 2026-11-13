<script lang="ts">
  import { studio } from '../../state.svelte';
</script>

<footer class="app-statusbar">
  <div class="statusbar-left">
    <span class="statusbar-msg">{studio.statusMessage}</span>
  </div>

  <div class="statusbar-right">
    {#if studio.graph}
      <span class="statusbar-metric">
        <strong>{studio.graph.guides.length}</strong> Guides
      </span>
      <span class="statusbar-sep">•</span>
      <span class="statusbar-metric">
        <strong>{studio.quizItems.length}</strong> Questions ({studio.selectedGuideId || 'none'})
      </span>
      {#if studio.startLine !== null}
        <span class="statusbar-sep">•</span>
        <span class="statusbar-metric">
          Segment: <strong>lines {studio.getActiveSegment()}</strong>
        </span>
      {/if}
    {/if}
  </div>
</footer>

<style>
  .app-statusbar {
    height: 24px;
    background: var(--bg-surface);
    border-top: 1px solid var(--border-color);
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 12px;
    font-size: 10px;
    color: var(--text-dim);
    flex-shrink: 0;
    user-select: none;
  }
  .statusbar-left {
    display: flex;
    align-items: center;
    gap: 8px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .statusbar-right {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }
  .statusbar-metric {
    color: var(--text-muted);
  }
  .statusbar-metric strong {
    color: var(--text-main);
    font-weight: 600;
  }
  .statusbar-sep {
    color: var(--border-color);
  }
  .statusbar-msg {
    color: var(--text-muted);
  }
</style>
