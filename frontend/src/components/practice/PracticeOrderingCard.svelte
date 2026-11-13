<script lang="ts">
  import { studio } from '../../state.svelte';
  import type { quiz } from '../../../wailsjs/go/models';

  let { card }: { card: quiz.QuizItemDTO } = $props();

  let draggedIndex = $state<number | null>(null);
  let dragOverIndex = $state<number | null>(null);

  function handleDragStart(idx: number, e: DragEvent) {
    if (studio.practiceChecked) return;
    draggedIndex = idx;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/plain', String(idx));
    }
  }

  function handleDragOver(idx: number, e: DragEvent) {
    if (draggedIndex === null || studio.practiceChecked) return;
    e.preventDefault();
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move';
    }
    if (dragOverIndex !== idx) {
      dragOverIndex = idx;
    }
  }

  function handleDragEnter(idx: number) {
    if (draggedIndex === null || studio.practiceChecked) return;
    dragOverIndex = idx;
  }

  function handleDragLeave(idx: number) {
    if (dragOverIndex === idx) {
      dragOverIndex = null;
    }
  }

  function handleDrop(targetIdx: number, e: DragEvent) {
    e.preventDefault();
    if (draggedIndex === null || studio.practiceChecked) return;
    if (draggedIndex !== targetIdx) {
      const items = [...(studio.practiceAnswer.items || [])];
      const [moved] = items.splice(draggedIndex, 1);
      items.splice(targetIdx, 0, moved);
      studio.practiceAnswer.items = items;
    }
    draggedIndex = null;
    dragOverIndex = null;
  }

  function handleDragEnd() {
    draggedIndex = null;
    dragOverIndex = null;
  }
</script>

<div class="practice-ordering-container">
  {#if card.top_label}
    <div class="order-dir-label">▲ {card.top_label}</div>
  {/if}
  <div
    class="order-drag-list"
    role="list"
    ondragover={(e) => e.preventDefault()}
  >
    {#each studio.practiceAnswer.items || [] as item, idx}
      {@const isDragged = draggedIndex === idx}
      {@const isOver = dragOverIndex === idx && draggedIndex !== idx}
      <div
        class="practice-order-item {isDragged ? 'is-dragging' : ''} {isOver ? 'drop-target' : ''} {studio.practiceChecked ? 'is-disabled' : ''}"
        draggable={!studio.practiceChecked}
        role="listitem"
        ondragstart={(e) => handleDragStart(idx, e)}
        ondragover={(e) => handleDragOver(idx, e)}
        ondragenter={() => handleDragEnter(idx)}
        ondragleave={() => handleDragLeave(idx)}
        ondrop={(e) => handleDrop(idx, e)}
        ondragend={handleDragEnd}
      >
        {#if !studio.practiceChecked}
          <span class="order-drag-handle" title="Drag card to correct spot">
            <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor">
              <circle cx="5" cy="3.5" r="1.5" />
              <circle cx="11" cy="3.5" r="1.5" />
              <circle cx="5" cy="8" r="1.5" />
              <circle cx="11" cy="8" r="1.5" />
              <circle cx="5" cy="12.5" r="1.5" />
              <circle cx="11" cy="12.5" r="1.5" />
            </svg>
          </span>
        {/if}
        <span class="order-seq-num">{idx + 1}.</span>
        <span class="order-item-text">{item}</span>
        {#if studio.practiceChecked && card.items}
          {@const isSpotCorrect = card.items[idx] === item}
          <span class="order-spot-badge {isSpotCorrect ? 'spot-correct' : 'spot-wrong'}">
            {isSpotCorrect ? '✓' : '✗'}
          </span>
        {/if}
      </div>
    {/each}
  </div>
  {#if card.bottom_label}
    <div class="order-dir-label">▼ {card.bottom_label}</div>
  {/if}

  {#if studio.practiceChecked && studio.practiceResult && !studio.practiceResult.isCorrect && card.items}
    <div class="order-solution-box">
      <span class="field-label">Correct Sequence:</span>
      <ol class="order-solution-list">
        {#each card.items as solItem}
          <li>{solItem}</li>
        {/each}
      </ol>
    </div>
  {/if}
</div>

<style>
  .practice-ordering-container {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .order-dir-label {
    font-size: 9px;
    font-weight: 700;
    text-transform: uppercase;
    color: var(--text-dim);
    letter-spacing: 0.5px;
  }
  .order-drag-list {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .practice-order-item {
    display: flex;
    align-items: center;
    gap: 8px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 7px 10px;
    border-radius: 4px;
    font-size: 11px;
    cursor: grab;
    user-select: none;
    transition: background 0.12s ease, border-color 0.12s ease, box-shadow 0.12s ease, transform 0.1s ease;
  }
  .practice-order-item:hover:not(.is-disabled) {
    border-color: var(--border-focus);
    background: var(--bg-card-hover);
  }
  .practice-order-item:active:not(.is-disabled) {
    cursor: grabbing;
  }
  .practice-order-item.is-dragging {
    opacity: 0.35;
    border-style: dashed;
    border-color: var(--accent-white);
  }
  .practice-order-item.drop-target {
    border-color: var(--accent-white);
    background: var(--selection-bg);
    box-shadow: 0 0 0 1px var(--accent-white);
    transform: scale(1.01);
  }
  .practice-order-item.is-disabled {
    cursor: default;
  }
  .order-drag-handle {
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-dim);
    flex-shrink: 0;
  }
  .practice-order-item:hover .order-drag-handle {
    color: var(--text-muted);
  }
  .order-seq-num {
    color: var(--text-dim);
    font-size: 10px;
    font-weight: 700;
    width: 14px;
    flex-shrink: 0;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }
  .order-item-text {
    flex: 1;
    line-height: 1.4;
  }
  .order-spot-badge {
    font-size: 11px;
    font-weight: 700;
    padding: 1px 5px;
    border-radius: 3px;
  }
  .spot-correct {
    color: #4ade80;
    background: rgba(74, 222, 128, 0.1);
    border: 1px solid rgba(74, 222, 128, 0.2);
  }
  .spot-wrong {
    color: #f87171;
    background: rgba(248, 113, 113, 0.1);
    border: 1px solid rgba(248, 113, 113, 0.2);
  }
  .order-solution-box {
    margin-top: 6px;
    padding: 8px 10px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .order-solution-list {
    margin: 0;
    padding-left: 20px;
    font-size: 11px;
    color: var(--text-muted);
  }
  .order-solution-list li {
    margin: 2px 0;
  }
</style>
