<script lang="ts">
  import { studio } from '../../state.svelte';
  import { getDirectoryName } from '../../stores/types';

  let showRecentsMenu = $state(false);

  async function selectRecent(path: string) {
    showRecentsMenu = false;
    await studio.handleLoadGraph(path);
  }

  async function pickFolder() {
    showRecentsMenu = false;
    await studio.handleNativeFolderPick();
  }
</script>

<header class="app-header">
  <!-- Left Cluster: Workspace Breadcrumb + Quick Search -->
  <div class="header-left">
    <div class="recents-wrapper">
      <button
        type="button"
        class="workspace-breadcrumb-btn"
        onclick={() => showRecentsMenu = !showRecentsMenu}
        title="Switch or Open Knowledge Graph Workspace"
      >
        <svg class="ws-icon" width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
          <path d="M1.75 2.5h3.69l1.5 1.5h7.31c.41 0 .75.34.75.75v7.5a.75.75 0 01-.75.75H1.75a.75.75 0 01-.75-.75V3.25c0-.41.34-.75.75-.75zm0 1.5v7.5h12.5v-6H6.56l-1.5-1.5H1.75z" />
        </svg>
        <span class="ws-title">{studio.graph ? studio.graph.title : (getDirectoryName(studio.graphDir) || 'No Workspace')}</span>
        <span class="ws-chevron">▾</span>
      </button>

      {#if showRecentsMenu}
        <div class="recents-dropdown">
          <button type="button" class="dropdown-action-item" onclick={pickFolder}>
            <span>Open Directory...</span>
            <kbd>Ctrl+O</kbd>
          </button>
          <div class="dropdown-divider"></div>
          <div class="dropdown-header">Recent Workspaces</div>
          {#if studio.recentGraphs.length === 0}
            <div class="dropdown-empty">No recent workspaces</div>
          {:else}
            {#each studio.recentGraphs as rPath}
              <button
                type="button"
                class="dropdown-item {rPath === studio.graphDir ? 'active' : ''}"
                onclick={() => selectRecent(rPath)}
                title={rPath}
              >
                <span class="path-title">{getDirectoryName(rPath)}</span>
                <span class="path-full">{rPath}</span>
              </button>
            {/each}
          {/if}
        </div>
      {/if}
    </div>

    <button
      type="button"
      class="cmd-search-btn"
      onclick={() => studio.showCommandPalette = true}
      title="Command Palette (Ctrl+P)"
    >
      <svg class="search-icon" width="12" height="12" viewBox="0 0 16 16" fill="currentColor">
        <path d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1 1 0 0 0 1.415-1.414l-3.85-3.85a1.007 1.007 0 0 0-.115-.1zM12 6.5a5.5 5.5 0 1 1-11 0 5.5 5.5 0 0 1 11 0z" />
      </svg>
      <span class="search-label">Search</span>
      <kbd class="search-kbd">Ctrl+P</kbd>
    </button>
  </div>

  <!-- Center: Active Guide Indicator -->
  <div class="header-center">
    {#if studio.selectedGuide}
      <span class="center-guide-id">{studio.selectedGuide.id}</span>
      <span class="center-sep">•</span>
      <span class="scope-pill-sm">{studio.selectedGuide.scope}</span>
      <span class="center-sep">•</span>
      <span class="center-line-count">{studio.selectedGuide.line_count} lines</span>
    {/if}
  </div>

  <!-- Right Cluster: Borderless Zoom + Tab-Style Pane Toggles -->
  <div class="header-right">
    <div class="zoom-controls" title="UI Zoom Level">
      <button
        type="button"
        class="zoom-btn"
        disabled={studio.zoomLevel <= 70}
        onclick={() => studio.zoomOut()}
        title="Zoom Out (Ctrl+-)"
      >
        −
      </button>
      <button
        type="button"
        class="zoom-level-btn"
        onclick={() => studio.resetZoom()}
        title="Reset Zoom to 100% (Ctrl+0)"
      >
        {studio.zoomLevel}%
      </button>
      <button
        type="button"
        class="zoom-btn"
        disabled={studio.zoomLevel >= 200}
        onclick={() => studio.zoomIn()}
        title="Zoom In (Ctrl++)"
      >
        +
      </button>
    </div>

    <div class="header-vsep"></div>

    <div class="pane-tabs-group">
      <button
        type="button"
        class="pane-tab-btn {studio.leftPaneOpen ? 'is-open' : ''}"
        onclick={() => studio.leftPaneOpen = !studio.leftPaneOpen}
        title="Toggle Left Guide Navigator (Ctrl+B)"
      >
        Guides
      </button>
      <button
        type="button"
        class="pane-tab-btn {studio.rightPaneOpen ? 'is-open' : ''}"
        onclick={() => studio.rightPaneOpen = !studio.rightPaneOpen}
        title="Toggle Right Inspector & Questions Panel"
      >
        Inspector
      </button>
    </div>
  </div>
</header>

<style>
  .app-header {
    height: 40px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-color);
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    padding: 0 14px 0 14px;
    gap: 16px;
    flex-shrink: 0;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .recents-wrapper {
    position: relative;
  }
  .workspace-breadcrumb-btn {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    height: 28px;
    padding: 0 8px;
    background: transparent;
    border: none;
    border-radius: 4px;
    color: var(--text-main);
    font-size: 12px;
    cursor: pointer;
    transition: background 0.12s ease;
  }
  .workspace-breadcrumb-btn:hover {
    background: rgba(255, 255, 255, 0.05);
  }
  .ws-icon {
    color: var(--text-muted);
    flex-shrink: 0;
  }
  .ws-title {
    font-weight: 600;
    max-width: 220px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .ws-chevron {
    color: var(--text-dim);
    font-size: 9px;
  }

  .recents-dropdown {
    position: absolute;
    top: 32px;
    left: 0;
    width: 280px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 4px;
    box-shadow: 0 10px 24px rgba(0, 0, 0, 0.7);
    z-index: 100;
  }
  .dropdown-action-item {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: none;
    border: none;
    color: var(--text-main);
    font-size: 11px;
    font-weight: 600;
    padding: 6px 8px;
    border-radius: 3px;
    cursor: pointer;
  }
  .dropdown-action-item:hover {
    background: var(--bg-card-hover);
  }
  .dropdown-divider {
    height: 1px;
    background: var(--border-color);
    margin: 4px 0;
  }
  .dropdown-header {
    font-size: 9px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-muted);
    padding: 4px 8px;
  }
  .dropdown-empty {
    font-size: 11px;
    color: var(--text-dim);
    padding: 8px;
    text-align: center;
  }
  .dropdown-item {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
    background: none;
    border: none;
    padding: 5px 8px;
    border-radius: 3px;
    cursor: pointer;
    text-align: left;
    transition: background 0.1s ease;
  }
  .dropdown-item:hover {
    background: var(--bg-card-hover);
  }
  .dropdown-item.active {
    background: var(--bg-card);
    border-left: 2px solid var(--accent-white);
  }
  .path-title {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-main);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    width: 100%;
  }
  .path-full {
    font-size: 9px;
    color: var(--text-dim);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    width: 100%;
  }

  .cmd-search-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 28px;
    padding: 0 8px;
    background: transparent;
    border: none;
    border-radius: 4px;
    color: var(--text-muted);
    font-size: 11px;
    cursor: pointer;
    transition: background 0.12s ease, color 0.12s ease;
  }
  .cmd-search-btn:hover {
    background: rgba(255, 255, 255, 0.05);
    color: var(--text-main);
  }
  .search-icon {
    color: var(--text-dim);
  }
  .search-label {
    font-size: 11px;
  }
  .search-kbd {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: var(--text-dim);
    font-size: 9px;
    padding: 1px 4px;
    border-radius: 3px;
    font-family: inherit;
  }

  .header-center {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    font-size: 12px;
  }
  .center-guide-id {
    font-weight: 600;
    color: var(--text-main);
    letter-spacing: -0.1px;
  }
  .center-sep {
    color: var(--border-color);
    font-size: 9px;
  }
  .scope-pill-sm {
    font-size: 9px;
    font-weight: 700;
    text-transform: uppercase;
    padding: 2px 6px;
    border-radius: 3px;
    background: rgba(255, 255, 255, 0.06);
    color: var(--text-muted);
    border: none;
    letter-spacing: 0.4px;
  }
  .center-line-count {
    font-size: 11px;
    color: var(--text-dim);
  }

  .header-right {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    height: 100%;
    gap: 12px;
  }

  /* Borderless Zoom Group */
  .zoom-controls {
    display: inline-flex;
    align-items: center;
    background: transparent;
    border: none;
    gap: 1px;
  }
  .zoom-btn {
    height: 24px;
    width: 22px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    border-radius: 4px;
    color: var(--text-muted);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.12s ease;
    padding: 0;
  }
  .zoom-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-main);
  }
  .zoom-btn:disabled {
    opacity: 0.25;
    cursor: not-allowed;
  }
  .zoom-level-btn {
    height: 24px;
    padding: 0 6px;
    background: transparent;
    border: none;
    border-radius: 4px;
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 600;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    cursor: pointer;
    min-width: 40px;
    text-align: center;
    transition: all 0.12s ease;
  }
  .zoom-level-btn:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-main);
  }

  .header-vsep {
    width: 1px;
    height: 14px;
    background: rgba(255, 255, 255, 0.1);
  }

  /* Tab-Style Pane Toggles with Active Bottom Borders */
  .pane-tabs-group {
    display: flex;
    align-items: stretch;
    height: 100%;
    gap: 4px;
  }
  .pane-tab-btn {
    height: 100%;
    padding: 0 12px;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-dim);
    font-size: 11.5px;
    font-weight: 600;
    cursor: pointer;
    display: flex;
    align-items: center;
    transition: all 0.12s ease;
    position: relative;
    top: 1px;
  }
  .pane-tab-btn:hover {
    color: var(--text-main);
    background: rgba(255, 255, 255, 0.03);
  }
  .pane-tab-btn.is-open {
    color: var(--text-main);
    border-bottom-color: var(--accent-white);
  }
</style>
