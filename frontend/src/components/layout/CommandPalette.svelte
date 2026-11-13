<script lang="ts">
  import { studio } from '../../state.svelte';

  interface PaletteItem {
    id: string;
    label: string;
    category: 'guide' | 'action';
    detail?: string;
    action: () => void;
  }

  let commandQuery = $state('');
  let paletteSelectedIndex = $state(0);
  let paletteInputEl = $state<HTMLInputElement | null>(null);

  $effect(() => {
    if (studio.showCommandPalette) {
      commandQuery = '';
      paletteSelectedIndex = 0;
      setTimeout(() => {
        paletteInputEl?.focus();
      }, 20);
    }
  });

  function close() {
    studio.showCommandPalette = false;
  }

  function getPaletteItems(): PaletteItem[] {
    const items: PaletteItem[] = [
      {
        id: 'act-open',
        label: 'Open Directory...',
        category: 'action',
        detail: 'Ctrl+O',
        action: () => studio.handleNativeFolderPick()
      },
      {
        id: 'act-sidebar',
        label: studio.leftPaneOpen ? 'Hide Left Guide Navigator' : 'Show Left Guide Navigator',
        category: 'action',
        detail: 'Ctrl+B',
        action: () => { studio.leftPaneOpen = !studio.leftPaneOpen; }
      },
      {
        id: 'act-inspector',
        label: studio.rightPaneOpen ? 'Hide Right Panel' : 'Show Right Panel',
        category: 'action',
        action: () => { studio.rightPaneOpen = !studio.rightPaneOpen; }
      },
      {
        id: 'act-tab-inspector',
        label: 'Right Panel: Show Guide Details',
        category: 'action',
        action: () => {
          studio.activeRightTab = 'inspector';
          studio.rightPaneOpen = true;
        }
      },
      {
        id: 'act-tab-questions',
        label: 'Right Panel: Show Questions',
        category: 'action',
        action: () => {
          studio.activeRightTab = 'questions';
          studio.rightPaneOpen = true;
        }
      },
      {
        id: 'act-zoom-in',
        label: 'Zoom In',
        category: 'action',
        detail: 'Ctrl++',
        action: () => studio.zoomIn()
      },
      {
        id: 'act-zoom-out',
        label: 'Zoom Out',
        category: 'action',
        detail: 'Ctrl+-',
        action: () => studio.zoomOut()
      },
      {
        id: 'act-zoom-reset',
        label: 'Reset Zoom to 100%',
        category: 'action',
        detail: 'Ctrl+0',
        action: () => studio.resetZoom()
      }
    ];

    if (studio.startLine !== null) {
      items.push({
        id: 'act-clear-anchor',
        label: `Reset Segment Anchor (Lines ${studio.getActiveSegment()})`,
        category: 'action',
        action: () => {
          studio.startLine = null;
          studio.endLine = null;
        }
      });
    }

    if (studio.graph) {
      studio.graph.guides.forEach(g => {
        items.push({
          id: `guide-${g.id}`,
          label: g.id,
          category: 'guide',
          detail: `${g.scope} • ${g.line_count} lines`,
          action: () => studio.selectGuide(g.id)
        });
      });
    }

    if (!commandQuery.trim()) return items;

    const q = commandQuery.toLowerCase();
    return items.filter(it =>
      it.label.toLowerCase().includes(q) || (it.detail && it.detail.toLowerCase().includes(q))
    );
  }

  function handleKeydown(e: KeyboardEvent) {
    const items = getPaletteItems();
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      paletteSelectedIndex = (paletteSelectedIndex + 1) % Math.max(1, items.length);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      paletteSelectedIndex = (paletteSelectedIndex - 1 + items.length) % Math.max(1, items.length);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (items[paletteSelectedIndex]) {
        items[paletteSelectedIndex].action();
        close();
      }
    } else if (e.key === 'Escape') {
      e.preventDefault();
      close();
    }
  }
</script>

{#if studio.showCommandPalette}
  {@const items = getPaletteItems()}
  <div
    class="palette-backdrop"
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    onclick={(e) => { if (e.target === e.currentTarget) close(); }}
    onkeydown={handleKeydown}
  >
    <div class="palette-modal">
      <div class="palette-input-wrap">
        <span class="palette-search-label">Find</span>
        <input
          bind:this={paletteInputEl}
          type="text"
          bind:value={commandQuery}
          placeholder="Type a guide name or command..."
          class="palette-input"
          onkeydown={handleKeydown}
        />
        <kbd class="palette-esc-badge">Esc</kbd>
      </div>

      <div class="palette-results">
        {#if items.length === 0}
          <div class="palette-no-match">No matching guides or commands</div>
        {:else}
          {#each items as it, idx}
            <button
              type="button"
              class="palette-row {idx === paletteSelectedIndex ? 'selected' : ''}"
              onclick={() => { it.action(); close(); }}
              onmouseenter={() => paletteSelectedIndex = idx}
            >
              <div class="palette-row-left">
                <span class="palette-tag-badge tag-{it.category}">
                  {it.category}
                </span>
                <span class="palette-label">{it.label}</span>
              </div>
              {#if it.detail}
                <span class="palette-detail">{it.detail}</span>
              {/if}
            </button>
          {/each}
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .palette-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(2px);
    z-index: 9999;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    padding-top: 100px;
  }
  .palette-modal {
    width: 540px;
    max-width: 90vw;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    box-shadow: 0 20px 50px rgba(0, 0, 0, 0.85);
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }
  .palette-input-wrap {
    display: flex;
    align-items: center;
    padding: 10px 14px;
    gap: 10px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-card);
  }
  .palette-search-label {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    color: var(--text-dim);
    letter-spacing: 0.5px;
  }
  .palette-input {
    flex: 1;
    background: transparent;
    border: none;
    font-size: 12px;
    color: var(--text-main);
    outline: none;
  }
  .palette-esc-badge {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 1px 5px;
    font-size: 10px;
    color: var(--text-dim);
  }
  .palette-results {
    max-height: 320px;
    overflow-y: auto;
    padding: 4px;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }
  .palette-no-match {
    padding: 20px;
    text-align: center;
    color: var(--text-dim);
    font-size: 11px;
  }
  .palette-row {
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 6px 10px;
    border-radius: 3px;
    background: transparent;
    border: none;
    color: var(--text-muted);
    font-size: 11px;
    cursor: pointer;
    text-align: left;
    transition: background 0.08s ease;
  }
  .palette-row:hover,
  .palette-row.selected {
    background: var(--bg-card-hover);
    color: var(--text-main);
  }
  .palette-row-left {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }
  .palette-tag-badge {
    font-size: 8px;
    font-weight: 700;
    text-transform: uppercase;
    padding: 1px 4px;
    border-radius: 2px;
    letter-spacing: 0.4px;
  }
  .tag-guide {
    background: rgba(88, 166, 255, 0.15);
    color: #58a6ff;
  }
  .tag-action {
    background: rgba(163, 113, 247, 0.15);
    color: #d2a8ff;
  }
  .palette-label {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .palette-detail {
    font-size: 10px;
    color: var(--text-dim);
    flex-shrink: 0;
  }
</style>
