<script lang="ts">
  import { studio } from '../../state.svelte';

  let builderDraggedOrderIdx = $state<number | null>(null);
  let builderDragOverOrderIdx = $state<number | null>(null);

  function handleBuilderDragStart(idx: number, e: DragEvent) {
    builderDraggedOrderIdx = idx;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/plain', String(idx));
    }
  }

  function handleBuilderDragOver(idx: number, e: DragEvent) {
    if (builderDraggedOrderIdx === null) return;
    e.preventDefault();
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move';
    }
    if (builderDragOverOrderIdx !== idx) {
      builderDragOverOrderIdx = idx;
    }
  }

  function handleBuilderDrop(targetIdx: number, e: DragEvent) {
    e.preventDefault();
    if (builderDraggedOrderIdx !== null && builderDraggedOrderIdx !== targetIdx) {
      studio.reorderOrderItems(builderDraggedOrderIdx, targetIdx);
    }
    builderDraggedOrderIdx = null;
    builderDragOverOrderIdx = null;
  }

  function handleBuilderDragEnd() {
    builderDraggedOrderIdx = null;
    builderDragOverOrderIdx = null;
  }
</script>

<div class="form-group">
  <div class="order-labels-row">
    <div>
      <label for="top-label-input" class="field-label">Top Label</label>
      <input
        id="top-label-input"
        type="text"
        bind:value={studio.itemTopLabel}
        class="input-field"
        placeholder="Earliest / Start..."
      />
    </div>
    <div>
      <label for="bottom-label-input" class="field-label">Bottom Label</label>
      <input
        id="bottom-label-input"
        type="text"
        bind:value={studio.itemBottomLabel}
        class="input-field"
        placeholder="Latest / End..."
      />
    </div>
  </div>

  <div class="choices-header">
    <span class="field-label">Items in Correct Sequence</span>
    <button type="button" class="btn btn-sm btn-secondary" onclick={() => studio.addOrderItem()}>+ Item</button>
  </div>

  <div class="order-items-list" role="list">
    {#each studio.itemOrderItems as item, oIdx}
      {@const isDragged = builderDraggedOrderIdx === oIdx}
      {@const isOver = builderDragOverOrderIdx === oIdx && builderDraggedOrderIdx !== oIdx}
      <div
        class="order-item-card {isDragged ? 'is-dragging' : ''} {isOver ? 'drop-target' : ''}"
        role="listitem"
        ondragover={(e) => handleBuilderDragOver(oIdx, e)}
        ondrop={(e) => handleBuilderDrop(oIdx, e)}
        ondragend={handleBuilderDragEnd}
      >
        <span
          class="order-drag-handle"
          title="Drag to reorder"
          role="button"
          tabindex="0"
          aria-label="Drag item {oIdx + 1} to reorder"
          draggable="true"
          ondragstart={(e) => handleBuilderDragStart(oIdx, e)}
        >
          <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor">
            <circle cx="5" cy="3.5" r="1.5" />
            <circle cx="11" cy="3.5" r="1.5" />
            <circle cx="5" cy="8" r="1.5" />
            <circle cx="11" cy="8" r="1.5" />
            <circle cx="5" cy="12.5" r="1.5" />
            <circle cx="11" cy="12.5" r="1.5" />
          </svg>
        </span>
        <span class="order-num">{oIdx + 1}.</span>
        <input
          type="text"
          bind:value={studio.itemOrderItems[oIdx]}
          class="input-field"
          placeholder="Step description..."
        />
        <div class="order-nav-btns">
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            disabled={oIdx === 0}
            onclick={() => studio.moveOrderItem(oIdx, 'up')}
            title="Move Up"
          >↑</button>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            disabled={oIdx === studio.itemOrderItems.length - 1}
            onclick={() => studio.moveOrderItem(oIdx, 'down')}
            title="Move Down"
          >↓</button>
          <button
            type="button"
            class="btn-delete"
            disabled={studio.itemOrderItems.length <= 2}
            onclick={() => studio.removeOrderItem(oIdx)}
          >Remove</button>
        </div>
      </div>
    {/each}
  </div>
</div>

<style>
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .choices-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .order-labels-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
    margin-bottom: 4px;
  }
  .order-items-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .order-item-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 6px 8px;
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 6px;
    transition: background 0.12s ease, border-color 0.12s ease;
  }
  .order-item-card.is-dragging {
    opacity: 0.35;
    border-style: dashed;
    border-color: var(--accent-white);
  }
  .order-item-card.drop-target {
    border-color: var(--accent-white);
    background: var(--selection-bg);
    box-shadow: 0 0 0 1px var(--accent-white);
  }
  .order-drag-handle {
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-dim);
    cursor: grab;
    flex-shrink: 0;
    user-select: none;
    padding: 0 2px;
  }
  .order-drag-handle:active {
    cursor: grabbing;
  }
  .order-drag-handle:hover {
    color: var(--text-muted);
  }
  .order-num {
    font-size: 11px;
    color: var(--text-dim);
    font-weight: 600;
    width: 16px;
  }
  .order-nav-btns {
    display: flex;
    align-items: center;
    gap: 3px;
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
  .btn-delete {
    background: none;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    font-size: 10px;
    padding: 1px 3px;
  }
  .btn-delete:hover:not(:disabled) {
    color: #f85149;
  }
</style>
