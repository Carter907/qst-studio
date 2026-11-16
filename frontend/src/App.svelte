<script lang="ts">
  import { onMount } from 'svelte';
  import { studio } from './state.svelte';
  import { getDirectoryName } from './stores/types';
  import TopBar from './components/layout/TopBar.svelte';
  import StatusBar from './components/layout/StatusBar.svelte';
  import CommandPalette from './components/layout/CommandPalette.svelte';
  import GuideSidebar from './components/guide/GuideSidebar.svelte';
  import GuideViewer from './components/guide/GuideViewer.svelte';
  import NodeInspector from './components/guide/NodeInspector.svelte';
  import QuizWorkspace from './components/quizzes/QuizWorkspace.svelte';

  onMount(() => {
    studio.init();
    window.addEventListener('keydown', handleGlobalKeydown);
    window.addEventListener('wheel', handleGlobalWheel, { passive: false });
    return () => {
      window.removeEventListener('keydown', handleGlobalKeydown);
      window.removeEventListener('wheel', handleGlobalWheel);
    };
  });

  function handleGlobalWheel(e: WheelEvent) {
    if (e.ctrlKey || e.metaKey) {
      e.preventDefault();
      if (e.deltaY < 0) {
        studio.zoomIn();
      } else if (e.deltaY > 0) {
        studio.zoomOut();
      }
    }
  }

  function handleGlobalKeydown(e: KeyboardEvent) {
    const isCmd = e.metaKey || e.ctrlKey;

    if (isCmd && (e.key === '=' || e.key === '+')) {
      e.preventDefault();
      studio.zoomIn();
      return;
    }

    if (isCmd && (e.key === '-' || e.key === '_')) {
      e.preventDefault();
      studio.zoomOut();
      return;
    }

    if (isCmd && e.key === '0') {
      e.preventDefault();
      studio.resetZoom();
      return;
    }

    if (isCmd && (e.key === 'p' || e.key === 'P')) {
      e.preventDefault();
      studio.showCommandPalette = !studio.showCommandPalette;
      return;
    }

    if (isCmd && (e.key === 'b' || e.key === 'B')) {
      e.preventDefault();
      if (studio.workspaceMode === 'quizzes') {
        if (studio.quizStage !== 'running') {
          studio.quizSidebarOpen = !studio.quizSidebarOpen;
        }
      } else {
        studio.leftPaneOpen = !studio.leftPaneOpen;
      }
      return;
    }

    if (isCmd && (e.key === 'o' || e.key === 'O')) {
      e.preventDefault();
      studio.handleNativeFolderPick();
      return;
    }

    if (isCmd && (e.key === 's' || e.key === 'S')) {
      e.preventDefault();
      if (studio.activeRightTab === 'questions') {
        studio.saveItem();
      }
      return;
    }
  }
</script>

<div class="studio-container">
  <TopBar />
  <CommandPalette />

  {#if !studio.graph}
    <div class="welcome-dashboard">
      <div class="welcome-card">
        <div class="welcome-brand">
          <span class="brand-badge-lg">Qst Studio</span>
          <h2>Knowledge Graph Editing and Review</h2>
          <p>Explore Quest knowledge graphs, read line-anchored guides, and author questions.</p>
        </div>

        <div class="welcome-actions">
          <button class="btn btn-primary welcome-btn" onclick={() => studio.handleNativeFolderPick()}>
            Open Directory...
          </button>
        </div>

        {#if studio.recentGraphs.length > 0}
          <div class="welcome-recents">
            <h3>Recent Workspaces</h3>
            <div class="welcome-recents-list">
              {#each studio.recentGraphs as rPath}
                <button type="button" class="recent-link" onclick={() => studio.handleLoadGraph(rPath)} title={rPath}>
                  <span class="recent-title">{getDirectoryName(rPath)}</span>
                  <span class="recent-path">{rPath}</span>
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <div class="welcome-shortcuts">
          <div class="shortcut-tag"><kbd>Ctrl+P</kbd> Search</div>
          <div class="shortcut-tag"><kbd>Ctrl+O</kbd> Open</div>
          <div class="shortcut-tag"><kbd>Ctrl+B</kbd> Sidebar</div>
          <div class="shortcut-tag"><kbd>Ctrl+±</kbd> Zoom</div>
        </div>
      </div>
    </div>
  {:else if studio.workspaceMode === 'quizzes'}
    <QuizWorkspace />
  {:else}
    <div
      class="workspace-stage {studio.leftPaneOpen ? 'has-left' : 'no-left'} {studio.rightPaneOpen ? 'has-right' : 'no-right'}"
    >
      {#if studio.leftPaneOpen}
        <GuideSidebar />
      {/if}

      <GuideViewer />

      {#if studio.rightPaneOpen}
        <NodeInspector />
      {/if}
    </div>
  {/if}

  <StatusBar />
</div>

<style>
  .studio-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    background: var(--bg-primary);
    overflow: hidden;
  }

  .workspace-stage {
    flex: 1;
    display: grid;
    overflow: hidden;
  }

  .workspace-stage.has-left.has-right {
    grid-template-columns: 220px 1fr 320px;
  }
  .workspace-stage.no-left.has-right {
    grid-template-columns: 1fr 320px;
  }
  .workspace-stage.has-left.no-right {
    grid-template-columns: 220px 1fr;
  }
  .workspace-stage.no-left.no-right {
    grid-template-columns: 1fr;
  }

  .welcome-dashboard {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: center;
    background: var(--bg-primary);
  }
  .welcome-card {
    width: 480px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 28px;
    display: flex;
    flex-direction: column;
    gap: 20px;
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.6);
  }
  .welcome-brand {
    text-align: center;
  }
  .brand-badge-lg {
    background: var(--accent-white);
    color: var(--text-dark);
    font-weight: 800;
    font-size: 10px;
    padding: 2px 7px;
    border-radius: 3px;
    letter-spacing: 0.8px;
  }
  .welcome-brand h2 {
    margin: 10px 0 4px 0;
    font-size: 18px;
    font-weight: 600;
  }
  .welcome-brand p {
    margin: 0;
    font-size: 12px;
    color: var(--text-dim);
    line-height: 1.5;
  }
  .welcome-actions {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .welcome-btn {
    height: 36px;
    font-size: 12px;
    font-weight: 500;
  }
  .welcome-recents h3 {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    color: var(--text-dim);
    letter-spacing: 0.5px;
    margin: 0 0 6px 0;
  }
  .welcome-recents-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
    max-height: 170px;
    overflow-y: auto;
  }
  .recent-link {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 7px 10px;
    border-radius: 4px;
    color: var(--text-main);
    cursor: pointer;
    font-size: 11px;
    text-align: left;
    transition: background 0.12s ease, border-color 0.12s ease;
  }
  .recent-link:hover {
    background: var(--bg-card-hover);
    border-color: var(--border-focus);
  }
  .recent-title {
    font-weight: 600;
    color: var(--text-main);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    width: 100%;
  }
  .recent-path {
    font-size: 9px;
    color: var(--text-dim);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    width: 100%;
  }
  .welcome-shortcuts {
    display: flex;
    justify-content: space-between;
    border-top: 1px solid var(--border-color);
    padding-top: 14px;
  }
  .shortcut-tag {
    font-size: 10px;
    color: var(--text-dim);
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .shortcut-tag kbd {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 1px 4px;
    color: var(--text-muted);
  }
</style>
