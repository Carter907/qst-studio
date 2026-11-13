<script lang="ts">
  import { studio } from '../../state.svelte';
  import { renderMarkdownWithLatex } from '../../markdown';
</script>

<div class="bank-drawer">
  <div class="bank-drawer-header">
    <span class="field-label">Saved Questions ({studio.quizItems.length})</span>
    <span class="bank-subtitle">Click to edit</span>
  </div>

  {#if studio.quizItems.length === 0}
    <div class="bank-empty">
      No saved questions for <strong>{studio.selectedGuideId}</strong>.
    </div>
  {:else}
    <div class="bank-item-list">
      {#each studio.quizItems as qItem}
        <div class="bank-item-row {studio.editingItemId === qItem.id ? 'active' : ''}">
          <button
            type="button"
            class="bank-item-select"
            onclick={() => studio.loadBankItemIntoBuilder(qItem)}
          >
            <div class="bank-item-top">
              <span class="bank-item-id">{qItem.title || qItem.id}</span>
              <span class="bank-item-type">{(qItem.type || 'multiple_choice').replace(/_/g, ' ')}</span>
              <span class="bank-item-seg">Lines {qItem.segment}</span>
            </div>
            <div class="bank-item-prompt qst-inline-md">
              {@html renderMarkdownWithLatex(qItem.prompt, true)}
            </div>
          </button>
          <div class="bank-item-actions">
            <button
              type="button"
              class="btn-delete bank-delete-btn"
              onclick={() => studio.deleteBankItem(qItem.id)}
              title="Delete item"
            >
              Delete
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .bank-drawer {
    margin-top: 6px;
    border-top: 1px solid var(--border-color);
    padding-top: 8px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .bank-drawer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .bank-subtitle {
    font-size: 9px;
    color: var(--text-dim);
  }
  .bank-empty {
    font-size: 11px;
    color: var(--text-dim);
    text-align: center;
    padding: 10px;
  }
  .bank-item-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .bank-item-row {
    display: flex;
    align-items: center;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    overflow: hidden;
    transition: border-color 0.1s ease;
  }
  .bank-item-row:hover {
    border-color: var(--border-focus);
  }
  .bank-item-row.active {
    border-color: var(--accent-white);
  }
  .bank-item-select {
    flex: 1;
    background: none;
    border: none;
    padding: 6px 8px;
    text-align: left;
    cursor: pointer;
    min-width: 0;
    color: inherit;
  }
  .bank-item-top {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 2px;
  }
  .bank-item-id {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-main);
  }
  .bank-item-type {
    font-size: 8px;
    font-weight: 700;
    text-transform: uppercase;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    padding: 0 4px;
    border-radius: 2px;
    color: var(--text-dim);
  }
  .bank-item-seg {
    font-size: 9px;
    color: var(--text-dim);
    margin-left: auto;
  }
  .bank-item-prompt {
    font-size: 10px;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .bank-item-actions {
    padding: 0 4px;
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
  .bank-delete-btn {
    font-size: 9px;
    padding: 3px 5px;
  }
</style>
