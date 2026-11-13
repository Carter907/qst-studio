<script lang="ts">
  import { studio } from '../../state.svelte';

  let viewerEl = $state<HTMLElement | null>(null);

  // Auto-scroll to the anchored segment when selection changes
  $effect(() => {
    const _start = studio.startLine;
    if (viewerEl && _start !== null) {
      setTimeout(() => {
        const target = viewerEl?.querySelector('.block-selected');
        if (target && typeof (target as HTMLElement).scrollIntoView === 'function') {
          (target as HTMLElement).scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
      }, 30);
    }
  });
</script>

<main class="content-pane" bind:this={viewerEl}>
  {#if studio.selectedGuide}
    <div class="guide-header">
      <div class="guide-title-row">
        <div class="title-left">
          <h2>{studio.selectedGuide.id}</h2>
          <span class="scope-pill active-scope-pill">{studio.selectedGuide.scope}</span>
          <span class="guide-meta-lines">{studio.selectedGuide.line_count} lines</span>

          {#if studio.startLine !== null}
            <div class="active-anchor-tag">
              <span>Lines {studio.getActiveSegment()}</span>
              <button
                type="button"
                class="btn-reset-anchor"
                onclick={() => { studio.startLine = null; studio.endLine = null; }}
                title="Reset to full guide"
              >
                Clear
              </button>
            </div>
          {/if}
        </div>

        <div class="header-actions">
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            onclick={() => studio.selectGuide(studio.selectedGuideId)}
            title="Reload guide from disk"
          >
            Reload
          </button>
        </div>
      </div>
    </div>

    <div class="rendered-guide-viewer {studio.startLine !== null ? 'has-active-segment' : ''}">
      {#each studio.guideRenderedBlocks as blk, idx}
        {@const isSelected = studio.doesBlockOverlapSelection(blk)}
        {@const isPrevSelected = isSelected && idx > 0 && studio.doesBlockOverlapSelection(studio.guideRenderedBlocks[idx - 1])}
        {@const isNextSelected = isSelected && idx < studio.guideRenderedBlocks.length - 1 && studio.doesBlockOverlapSelection(studio.guideRenderedBlocks[idx + 1])}
        {@const isFirst = isSelected && !isPrevSelected}
        {@const isLast = isSelected && !isNextSelected}
        {@const isMiddle = isSelected && isPrevSelected && isNextSelected}
        {@const isSingle = isSelected && !isPrevSelected && !isNextSelected}
        <div
          class="rendered-block-row {isSelected ? 'block-selected' : ''} {isFirst ? 'segment-first' : ''} {isLast ? 'segment-last' : ''} {isMiddle ? 'segment-middle' : ''} {isSingle ? 'segment-single' : ''}"
        >
          <button
            type="button"
            class="block-line-gutter"
            onclick={(e) => studio.handleBlockAnchorClick(blk, e)}
            title="Anchor lines {blk.startLine}-{blk.endLine} (Shift+click to extend)"
          >
            {blk.startLine}{blk.endLine > blk.startLine ? `–${blk.endLine}` : ''}
          </button>
          <div class="rendered-block-content qst-prose">
            {@html blk.html}
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="empty-state">
      Select a guide from the sidebar to inspect its content and anchor segments.
    </div>
  {/if}
</main>

<style>
  .content-pane {
    background: var(--bg-primary);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .guide-header {
    padding: 10px 18px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-surface);
  }
  .guide-title-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .title-left {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }
  .title-left h2 {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    letter-spacing: -0.2px;
  }
  .guide-meta-lines {
    font-size: 11px;
    color: var(--text-dim);
  }
  .active-anchor-tag {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    background: var(--selection-bg);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 1px 7px;
    font-size: 10px;
    color: var(--accent-white);
    font-weight: 600;
  }
  .btn-reset-anchor {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 10px;
    cursor: pointer;
    padding: 0 2px;
  }
  .btn-reset-anchor:hover {
    color: var(--accent-white);
  }
  .header-actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .rendered-guide-viewer {
    flex: 1;
    overflow-y: auto;
    padding: 20px 24px;
    display: flex;
    flex-direction: column;
    gap: 0;
  }
  .rendered-block-row {
    display: grid;
    grid-template-columns: 44px 1fr;
    gap: 12px;
    align-items: start;
    padding: 4px 8px;
    margin-bottom: 4px;
    border-radius: 4px;
    border: 1px solid transparent;
    transition: background 0.1s ease, opacity 0.15s ease, border-color 0.1s ease;
  }
  .rendered-block-row:hover {
    background: rgba(255, 255, 255, 0.015);
  }
  .rendered-block-row:hover .block-line-gutter {
    opacity: 0.8;
    color: var(--text-main);
  }

  /* Focus Dimming when a segment is actively anchored */
  .rendered-guide-viewer.has-active-segment .rendered-block-row:not(.block-selected) {
    opacity: 0.65;
  }
  .rendered-guide-viewer.has-active-segment .rendered-block-row:not(.block-selected):hover {
    opacity: 1;
  }

  /* Continuous Unified Container for Selected Segment Blocks */
  .rendered-block-row.block-selected {
    background: var(--selection-bg);
    border-left: 3px solid var(--accent-white);
    border-right: 1px solid var(--border-color);
    border-top: none;
    border-bottom: none;
    border-radius: 0;
    margin-top: 0;
    margin-bottom: 0;
    padding-top: 5px;
    padding-bottom: 5px;
    opacity: 1 !important;
  }
  .rendered-block-row.block-selected .block-line-gutter {
    opacity: 1;
    color: var(--accent-white);
    font-weight: 600;
  }

  .rendered-block-row.segment-first {
    border-top: 1px solid var(--border-color);
    border-top-right-radius: 6px;
    padding-top: 8px;
    margin-top: 6px;
  }
  .rendered-block-row.segment-last {
    border-bottom: 1px solid var(--border-color);
    border-bottom-right-radius: 6px;
    padding-bottom: 8px;
    margin-bottom: 8px;
  }
  .rendered-block-row.segment-single {
    border-top: 1px solid var(--border-color);
    border-bottom: 1px solid var(--border-color);
    border-top-right-radius: 6px;
    border-bottom-right-radius: 6px;
    padding-top: 8px;
    padding-bottom: 8px;
    margin-top: 6px;
    margin-bottom: 8px;
  }
  .rendered-block-row.segment-middle {
    border-top: none;
    border-bottom: none;
    border-radius: 0;
    padding-top: 4px;
    padding-bottom: 4px;
  }

  .block-line-gutter {
    background: none;
    border: none;
    color: var(--text-dim);
    opacity: 0.35;
    font-size: 10px;
    font-family: var(--font-mono, monospace);
    padding: 2px 0;
    text-align: right;
    cursor: pointer;
    transition: all 0.12s ease;
    user-select: none;
  }
  .block-line-gutter:hover {
    opacity: 1;
    color: var(--text-main);
  }

  .rendered-block-content {
    min-width: 0;
    line-height: 1.6;
  }
  .empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-dim);
    font-size: 12px;
  }
</style>
