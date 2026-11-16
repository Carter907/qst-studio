<script lang="ts">
  import { renderMarkdownWithLatex } from '../../markdown';

  let {
    value = $bindable(''),
    options = [],
    placeholder = 'Select an option...',
    disabled = false
  }: {
    value: string;
    options: string[];
    placeholder?: string;
    disabled?: boolean;
  } = $props();

  let open = $state(false);
  let containerEl = $state<HTMLDivElement | null>(null);

  function selectOption(opt: string) {
    if (disabled) return;
    value = opt;
    open = false;
  }

  function handleWindowClick(e: MouseEvent) {
    if (!open || !containerEl) return;
    if (!containerEl.contains(e.target as Node)) {
      open = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (open && e.key === 'Escape') {
      e.preventDefault();
      open = false;
    }
  }
</script>

<svelte:window onclick={handleWindowClick} onkeydown={handleKeydown} />

<div class="rich-select" bind:this={containerEl}>
  <button
    type="button"
    class="rich-select-trigger {open ? 'is-open' : ''}"
    {disabled}
    aria-haspopup="listbox"
    aria-expanded={open}
    onclick={() => {
      if (!disabled) open = !open;
    }}
  >
    <span class="rich-select-value qst-inline-md {!value ? 'is-placeholder' : ''}">
      {#if value}
        {@html renderMarkdownWithLatex(value, true)}
      {:else}
        {placeholder}
      {/if}
    </span>
    <span class="rich-select-chevron">▾</span>
  </button>

  {#if open && !disabled}
    <div class="rich-select-menu" role="listbox">
      {#if value}
        <button
          type="button"
          class="rich-select-option clear-option"
          role="option"
          aria-selected={false}
          onclick={() => selectOption('')}
        >
          {placeholder}
        </button>
      {/if}
      {#each options as opt}
        <button
          type="button"
          class="rich-select-option {opt === value ? 'is-selected' : ''}"
          role="option"
          aria-selected={opt === value}
          onclick={() => selectOption(opt)}
        >
          <span class="qst-inline-md">{@html renderMarkdownWithLatex(opt, true)}</span>
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .rich-select {
    position: relative;
    width: 100%;
    min-width: 0;
  }
  .rich-select-trigger {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 5px 8px;
    font-size: 11px;
    color: var(--text-main);
    text-align: left;
    cursor: pointer;
    min-height: 28px;
    transition: border-color 0.12s ease;
  }
  .rich-select-trigger:hover:not(:disabled),
  .rich-select-trigger.is-open {
    border-color: var(--border-focus);
  }
  .rich-select-trigger:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }
  .rich-select-value {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    line-height: 1.35;
  }
  .rich-select-value.is-placeholder {
    color: var(--text-dim);
  }
  .rich-select-chevron {
    color: var(--text-dim);
    font-size: 9px;
    flex-shrink: 0;
  }
  .rich-select-menu {
    position: absolute;
    top: calc(100% + 3px);
    left: 0;
    right: 0;
    max-height: 220px;
    overflow-y: auto;
    background: var(--bg-surface);
    border: 1px solid var(--border-focus);
    border-radius: 4px;
    padding: 3px;
    box-shadow: 0 8px 20px rgba(0, 0, 0, 0.65);
    z-index: 60;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .rich-select-option {
    width: 100%;
    text-align: left;
    background: transparent;
    border: none;
    border-radius: 3px;
    padding: 5px 8px;
    font-size: 11px;
    color: var(--text-main);
    cursor: pointer;
    line-height: 1.4;
  }
  .rich-select-option:hover {
    background: var(--bg-card-hover);
  }
  .rich-select-option.is-selected {
    background: var(--selection-bg);
    font-weight: 600;
  }
  .rich-select-option.clear-option {
    color: var(--text-dim);
    font-size: 10px;
    border-bottom: 1px solid var(--border-subtle);
    border-radius: 0;
    margin-bottom: 1px;
  }
</style>
