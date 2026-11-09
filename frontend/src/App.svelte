<script lang="ts">
  import { onMount } from 'svelte';
  import Graph3D from './Graph3D.svelte';
  import {
    renderMarkdownWithLatex,
    parseMarkdownBlocksWithLines,
    type RenderedGuideBlock
  } from './markdown';
  import {
    LoadGraph,
    GetGuide,
    GetLearningPath,
    GetQuizItems,
    SaveQuizItem,
    RecordPracticeResult,
    GetLearnerStates,
    SelectGraphDirectory,
    GetRecentGraphs,
    SaveGuideContent,
    GetAgentStatus,
    SaveAgentConfig,
    GenerateQuestionFromSegment,
    NextPracticeTurn
  } from '../wailsjs/go/main/App.js';
  import type { main } from '../wailsjs/go/models';

  // Workspace & Perspective State
  let graphDir = $state('/home/speer/Code/golang/qst/examples/C_programming');
  let graph = $state<main.GraphDTO | null>(null);
  let recentGraphs = $state<string[]>([]);
  let showRecentsMenu = $state(false);
  let perspective = $state<'explore' | 'author' | 'practice'>('author');
  let centerMode = $state<'rendered' | 'anchor' | 'edit'>('rendered');
  let leftPaneOpen = $state<boolean>(true);
  let rightPaneOpen = $state<boolean>(true);
  let builderPreviewMode = $state<boolean>(false);
  let statusMessage = $state<string>('Ready');

  // Command Palette State
  let showCommandPalette = $state(false);
  let commandQuery = $state('');
  let paletteSelectedIndex = $state(0);
  let paletteInputEl = $state<HTMLInputElement | null>(null);

  // Selected Guide & Navigator State
  let selectedGuideId = $state<string>('');
  let selectedGuide = $state<main.GuideDetailDTO | null>(null);
  let guideBodyLines = $state<string[]>([]);
  let guideRenderedBlocks = $state<RenderedGuideBlock[]>([]);
  let editableContent = $state<string>('');
  let isDirty = $state(false);
  let learningPath = $state<string[]>([]);
  let selectedScope = $state<string>('all');
  let guideSearchQuery = $state<string>('');
  let quizItems = $state<main.QuizItemDTO[]>([]);
  let learnerStates = $state<Record<string, main.LearnerMasteryDTO>>({});

  // Segment Anchoring
  let startLine = $state<number | null>(null);
  let endLine = $state<number | null>(null);

  // Question Builder Form
  let editingItemId = $state<string | null>(null);
  let itemTitle = $state('Variable Binding & Storage');
  let itemPrompt = $state('What happens in C when a variable declaration is not initialized?');
  let itemType = $state('choice');
  let itemChoices = $state([
    { id: 'c1', text: 'It contains indeterminate (garbage) values from stack memory.', correct: true, feedback: 'Correct. Uninitialized local variables retain whatever was in that memory location.' },
    { id: 'c2', text: 'The compiler automatically zeroes the allocated bytes.', correct: false, feedback: 'Incorrect. Unlike Java or Go, C does not zero-initialize automatic variables.' },
    { id: 'c3', text: 'A segmentation fault is generated at runtime immediately.', correct: false, feedback: 'Incorrect. Segmentation faults occur on illegal memory access, not uninitialized declarations.' }
  ]);

  // Continuous Practice Stream State
  let practiceScope = $state<'adaptive' | 'guide'>('adaptive');
  let currentTurn = $state<main.PracticeTurnDTO | null>(null);
  let prefetchedTurn = $state<main.PracticeTurnDTO | null>(null);
  let isLoadingTurn = $state<boolean>(false);
  let isPrefetching = $state<boolean>(false);
  let practiceError = $state<string>('');
  let seenItemIds = $state<string[]>([]);
  let practiceChoice = $state<string>('');
  let practiceSubmitted = $state(false);
  let lastAnswerCorrect = $state(false);
  let practiceResultMsg = $state('');
  let selectedChoiceFeedback = $state('');
  let correctChoiceFeedback = $state('');
  let lastMasteryUpdate = $state<main.LearnerMasteryDTO | null>(null);
  let sessionQuestionCount = $state<number>(0);
  let sessionCorrectCount = $state<number>(0);

  // Local Ollama Agent State
  let agentStatus = $state<main.AgentStatusDTO | null>(null);
  let selectedModel = $state<string>('');
  let isGenerating = $state<boolean>(false);

  onMount(() => {
    loadRecents();
    refreshAgentStatus();
    handleLoadGraph(graphDir);

    window.addEventListener('keydown', handleGlobalKeydown);
    return () => {
      window.removeEventListener('keydown', handleGlobalKeydown);
    };
  });

  async function refreshAgentStatus() {
    try {
      agentStatus = await GetAgentStatus();
      if (agentStatus && agentStatus.active_model) {
        selectedModel = agentStatus.active_model;
      }
    } catch {
      agentStatus = null;
    }
  }

  async function handleModelChange(newModel: string) {
    selectedModel = newModel;
    if (!agentStatus) return;
    try {
      await SaveAgentConfig({
        endpoint: agentStatus.endpoint || 'http://localhost:11434',
        model: newModel
      });
      agentStatus.active_model = newModel;
      statusMessage = `Ollama model set to ${newModel}`;
    } catch (err: any) {
      statusMessage = `Failed to save agent config: ${err}`;
    }
  }

  async function switchPerspective(next: 'explore' | 'author' | 'practice') {
    perspective = next;
    if (next === 'practice') {
      if (centerMode === 'edit') {
        centerMode = 'rendered';
      }
      if (!currentTurn && !isLoadingTurn) {
        await startNextPracticeTurn();
      } else if (currentTurn) {
        await syncGuideToTurn(currentTurn);
      }
    } else if (next === 'explore') {
      rightPaneOpen = true;
    }
  }

  async function practiceGuideNode(guideId: string) {
    if (!guideId) return;
    practiceScope = 'guide';
    prefetchedTurn = null;
    perspective = 'practice';
    await selectGuide(guideId, true);
    await startNextPracticeTurn(guideId);
  }

  async function authorGuideNode(guideId: string) {
    if (!guideId) return;
    await selectGuide(guideId);
    perspective = 'author';
    rightPaneOpen = true;
  }

  async function generateWithOllama() {
    if (!selectedGuide || isGenerating) return;
    isGenerating = true;
    const seg = getActiveSegment();
    const modelLabel = selectedModel || agentStatus?.active_model || 'Ollama';
    statusMessage = `Generating question from ${selectedGuide.id} (lines ${seg}) with ${modelLabel}...`;
    try {
      const draft = await GenerateQuestionFromSegment(graphDir, selectedGuide.id, seg);
      if (draft) {
        editingItemId = null;
        itemTitle = draft.title || itemTitle;
        itemPrompt = draft.prompt || itemPrompt;
        if (draft.choices && draft.choices.length >= 2) {
          itemChoices = draft.choices.map((c, idx) => ({
            id: c.id || `c${idx + 1}`,
            text: c.text,
            correct: c.correct,
            feedback: c.feedback || ''
          }));
        }
        perspective = 'author';
        rightPaneOpen = true;
        statusMessage = `Draft generated from ${selectedGuide.id} (lines ${seg})`;
      }
    } catch (err: any) {
      statusMessage = `Ollama error: ${err}`;
    } finally {
      isGenerating = false;
      refreshAgentStatus();
    }
  }

  async function loadRecents() {
    try {
      recentGraphs = await GetRecentGraphs();
    } catch {
      recentGraphs = [];
    }
  }

  async function handleLoadGraph(dir: string) {
    if (!dir) return;
    try {
      statusMessage = 'Loading graph...';
      graphDir = dir;
      graph = await LoadGraph(dir);
      learnerStates = await GetLearnerStates(dir);
      currentTurn = null;
      prefetchedTurn = null;
      seenItemIds = [];
      await loadRecents();
      showRecentsMenu = false;
      if (graph && graph.guides.length > 0) {
        await selectGuide(graph.guides[0].id);
      }
      statusMessage = `Loaded "${graph?.title}" (${graph?.guides.length} guides)`;
    } catch (err: any) {
      statusMessage = `Error: ${err}`;
    }
  }

  async function handleNativeFolderPick() {
    showRecentsMenu = false;
    try {
      const selected = await SelectGraphDirectory();
      if (selected) {
        await handleLoadGraph(selected);
      }
    } catch (err: any) {
      statusMessage = `Folder picker error: ${err}`;
    }
  }

  function applySegmentRange(segmentStr: string) {
    if (!segmentStr || segmentStr === 'ALL') {
      startLine = 1;
      endLine = Math.max(1, guideBodyLines.length);
      return;
    }
    const parts = segmentStr.split('-');
    const s = parseInt(parts[0], 10);
    const e = parts.length > 1 ? parseInt(parts[1], 10) : s;
    if (!isNaN(s) && s >= 1) {
      startLine = s;
      endLine = !isNaN(e) && e >= s ? e : s;
    } else {
      startLine = 1;
      endLine = Math.max(1, guideBodyLines.length);
    }
  }

  async function selectGuide(id: string, preservePracticeTurn = false) {
    selectedGuideId = id;
    if (!preservePracticeTurn) {
      startLine = null;
      endLine = null;
      editingItemId = null;
    }
    isDirty = false;

    try {
      selectedGuide = await GetGuide(graphDir, id);
      guideBodyLines = selectedGuide?.body ? selectedGuide.body.split('\n') : [];
      guideRenderedBlocks = parseMarkdownBlocksWithLines(selectedGuide?.body || '');
      editableContent = selectedGuide?.raw_content || selectedGuide?.body || '';
      quizItems = await GetQuizItems(graphDir, id);
      learningPath = await GetLearningPath(graphDir, id);

      if (!preservePracticeTurn && perspective === 'practice' && practiceScope === 'guide') {
        prefetchedTurn = null;
        await startNextPracticeTurn();
      }
    } catch (err: any) {
      statusMessage = `Failed to load guide: ${err}`;
    }
  }

  async function syncGuideToTurn(turn: main.PracticeTurnDTO) {
    const targetGuideId = turn.item.guide_id;
    if (targetGuideId && targetGuideId !== selectedGuideId) {
      await selectGuide(targetGuideId, true);
    } else if (targetGuideId) {
      quizItems = await GetQuizItems(graphDir, targetGuideId);
    }
    if (centerMode === 'edit') {
      centerMode = 'rendered';
    }
    applySegmentRange(turn.item.segment);
  }

  async function prefetchLookaheadTurn() {
    if (isPrefetching || !graphDir) return;
    isPrefetching = true;
    try {
      const nextSeen = currentTurn
        ? Array.from(new Set([...seenItemIds, currentTurn.item.id]))
        : [...seenItemIds];
      const excludeGuides =
        practiceScope === 'adaptive' && currentTurn ? [currentTurn.item.guide_id] : [];

      const nextTurn = await NextPracticeTurn({
        dir_path: graphDir,
        mode: practiceScope,
        focus_guide_id: selectedGuideId,
        focus_segment: getActiveSegment(),
        seen_item_ids: nextSeen,
        exclude_guides: excludeGuides
      });
      if (nextTurn) {
        prefetchedTurn = nextTurn;
        if (selectedGuideId && nextTurn.item.guide_id === selectedGuideId) {
          quizItems = await GetQuizItems(graphDir, selectedGuideId);
        }
      }
    } catch {
      // Background pre-fetch failure is ignored until requested in foreground
    } finally {
      isPrefetching = false;
    }
  }

  async function startNextPracticeTurn(overrideGuideId?: string) {
    if (isLoadingTurn) return;
    practiceSubmitted = false;
    practiceChoice = '';
    practiceResultMsg = '';
    selectedChoiceFeedback = '';
    correctChoiceFeedback = '';
    practiceError = '';

    if (!overrideGuideId && prefetchedTurn) {
      const readyTurn = prefetchedTurn;
      prefetchedTurn = null;
      currentTurn = readyTurn;
      if (!seenItemIds.includes(readyTurn.item.id)) {
        seenItemIds = [...seenItemIds, readyTurn.item.id];
      }
      await syncGuideToTurn(readyTurn);
      statusMessage = `Practice: ${readyTurn.item.guide_id} (${readyTurn.source === 'generated' ? 'Auto-Generated' : 'Bank'})`;
      prefetchLookaheadTurn();
      return;
    }

    isLoadingTurn = true;
    const effectiveMode = overrideGuideId ? 'guide' : practiceScope;
    const effectiveGuide = overrideGuideId || selectedGuideId;
    statusMessage =
      effectiveMode === 'adaptive'
        ? 'Resolving next adaptive practice question...'
        : `Resolving practice question for ${effectiveGuide}...`;

    try {
      const turn = await NextPracticeTurn({
        dir_path: graphDir,
        mode: effectiveMode,
        focus_guide_id: effectiveGuide,
        focus_segment: overrideGuideId ? '' : getActiveSegment(),
        seen_item_ids: seenItemIds,
        exclude_guides: []
      });
      if (turn) {
        currentTurn = turn;
        if (!seenItemIds.includes(turn.item.id)) {
          seenItemIds = [...seenItemIds, turn.item.id];
        }
        await syncGuideToTurn(turn);
        statusMessage = `Practice: ${turn.item.guide_id} (${turn.source === 'generated' ? 'Auto-Generated & Saved' : 'Loaded from Bank'})`;
        prefetchLookaheadTurn();
      }
    } catch (err: any) {
      practiceError = `${err}`;
      statusMessage = `Practice error: ${err}`;
    } finally {
      isLoadingTurn = false;
      refreshAgentStatus();
    }
  }

  async function switchPracticeScope(newScope: 'adaptive' | 'guide') {
    if (practiceScope === newScope && currentTurn) return;
    practiceScope = newScope;
    prefetchedTurn = null;
    await startNextPracticeTurn();
  }

  async function drillDownRemediation(targetGuideId: string) {
    if (!targetGuideId) return;
    prefetchedTurn = null;
    await startNextPracticeTurn(targetGuideId);
  }

  async function saveActiveGuideFile() {
    if (!selectedGuide?.path) return;
    try {
      await SaveGuideContent(selectedGuide.path, editableContent);
      isDirty = false;
      statusMessage = `Saved ${selectedGuide.id}.md to disk`;
      const updated = await GetGuide(graphDir, selectedGuide.id);
      selectedGuide = updated;
      guideBodyLines = updated.body.split('\n');
      guideRenderedBlocks = parseMarkdownBlocksWithLines(updated.body || '');
    } catch (err: any) {
      statusMessage = `Save error: ${err}`;
    }
  }

  function doesBlockOverlapSelection(blk: RenderedGuideBlock): boolean {
    if (startLine === null) return false;
    const selStart = startLine;
    const selEnd = endLine !== null ? endLine : startLine;
    return blk.startLine <= selEnd && blk.endLine >= selStart;
  }

  function handleBlockAnchorClick(blk: RenderedGuideBlock, e: MouseEvent) {
    if (e.shiftKey && startLine !== null) {
      const curEnd = endLine ?? startLine;
      startLine = Math.min(startLine, blk.startLine);
      endLine = Math.max(curEnd, blk.endLine);
    } else {
      startLine = blk.startLine;
      endLine = blk.endLine;
    }
  }

  function handleLineClick(lineNum: number) {
    if (startLine === null || (startLine !== null && endLine !== null)) {
      startLine = lineNum;
      endLine = null;
    } else {
      if (lineNum >= startLine) {
        endLine = lineNum;
      } else {
        endLine = startLine;
        startLine = lineNum;
      }
    }
  }

  function getActiveSegment(): string {
    if (startLine !== null && endLine !== null) {
      return `${startLine}-${endLine}`;
    }
    if (startLine !== null) {
      return `${startLine}-${startLine}`;
    }
    return '1-1';
  }

  function loadBankItemIntoBuilder(item: main.QuizItemDTO) {
    editingItemId = item.id;
    itemTitle = item.title || item.id;
    itemPrompt = item.prompt;
    itemType = item.type || 'choice';
    itemChoices = (item.choices || []).map((c, idx) => ({
      id: c.id || `c${idx + 1}`,
      text: c.text,
      correct: c.correct,
      feedback: c.feedback || ''
    }));
    if (centerMode === 'edit') {
      centerMode = 'rendered';
    }
    applySegmentRange(item.segment);
    statusMessage = `Editing bank item ${item.id} (lines ${item.segment})`;
  }

  async function practiceSpecificBankItem(item: main.QuizItemDTO) {
    currentTurn = {
      item,
      source: 'bank',
      reason: `Selected Bank Item • ${item.scope || selectedGuide?.scope || ''}`,
      linked_sub_guide: '',
      unmastered_prereq_id: '',
      target_mastery: getMastery(item.guide_id || selectedGuideId) / 100
    } as main.PracticeTurnDTO;
    practiceSubmitted = false;
    practiceChoice = '';
    perspective = 'practice';
    await syncGuideToTurn(currentTurn);
  }

  function resetBuilderForm() {
    editingItemId = null;
    itemTitle = '';
    itemPrompt = '';
    itemChoices = [
      { id: 'c1', text: '', correct: true, feedback: '' },
      { id: 'c2', text: '', correct: false, feedback: '' },
      { id: 'c3', text: '', correct: false, feedback: '' }
    ];
    statusMessage = 'Cleared builder for new question draft';
  }

  function addChoice() {
    const nextId = `c${itemChoices.length + 1}`;
    itemChoices = [...itemChoices, { id: nextId, text: '', correct: false, feedback: '' }];
  }

  function removeChoice(index: number) {
    if (itemChoices.length > 2) {
      itemChoices = itemChoices.filter((_, i) => i !== index);
    }
  }

  function setCorrectChoice(index: number) {
    itemChoices = itemChoices.map((c, i) => ({
      ...c,
      correct: i === index
    }));
  }

  async function saveItem() {
    if (!selectedGuide) return;
    const idToUse =
      editingItemId || `Q_${selectedGuide.id.toUpperCase()}_${Date.now().toString().slice(-4)}`;
    const newItem: main.QuizItemDTO = {
      id: idToUse,
      title: itemTitle,
      guide_id: selectedGuide.id,
      segment: getActiveSegment(),
      scope: selectedGuide.scope,
      type: itemType,
      prompt: itemPrompt,
      choices: itemChoices,
      created_at: new Date().toISOString()
    };

    try {
      await SaveQuizItem(graphDir, newItem);
      editingItemId = newItem.id;
      quizItems = await GetQuizItems(graphDir, selectedGuide.id);
      statusMessage = `Saved ${newItem.id} to items.yaml`;
    } catch (err: any) {
      statusMessage = `Save failed: ${err}`;
    }
  }

  async function submitPracticeAnswer(item: main.QuizItemDTO) {
    if (!practiceChoice) return;
    const targetGuideId = item.guide_id || selectedGuideId;
    if (!targetGuideId) return;

    practiceSubmitted = true;
    const chosen = item.choices.find(c => c.id === practiceChoice);
    const rightOpt = item.choices.find(c => c.correct);
    const isCorrect = chosen ? chosen.correct : false;

    lastAnswerCorrect = isCorrect;
    selectedChoiceFeedback = chosen?.feedback || '';
    correctChoiceFeedback = !isCorrect && rightOpt?.feedback ? rightOpt.feedback : '';
    sessionQuestionCount += 1;
    if (isCorrect) {
      sessionCorrectCount += 1;
    }

    try {
      lastMasteryUpdate = await RecordPracticeResult(
        graphDir,
        targetGuideId,
        item.segment || 'ALL',
        isCorrect
      );
      learnerStates = await GetLearnerStates(graphDir);
      practiceResultMsg = isCorrect ? 'Correct Assessment' : 'Incorrect Assessment';
    } catch (err: any) {
      statusMessage = `Scoring error: ${err}`;
    }
  }

  function getMastery(guideId: string): number {
    const key = `${guideId}:ALL`;
    if (learnerStates[key]) {
      return Math.round(learnerStates[key].p_mastery * 100);
    }
    for (const [k, v] of Object.entries(learnerStates)) {
      if (k.startsWith(`${guideId}:`)) {
        return Math.round(v.p_mastery * 100);
      }
    }
    return 10;
  }

  function getOverallGraphMastery(): number {
    if (!graph || !graph.guides || graph.guides.length === 0) return 10;
    const sum = graph.guides.reduce((acc, g) => acc + getMastery(g.id), 0);
    return Math.round(sum / graph.guides.length);
  }

  function getFilteredGuides(): main.GuideSummaryDTO[] {
    if (!graph) return [];
    const q = guideSearchQuery.trim().toLowerCase();
    return graph.guides.filter(g => {
      const matchesScope = selectedScope === 'all' || g.scope === selectedScope;
      const matchesQuery = !q || g.id.toLowerCase().includes(q) || g.scope.toLowerCase().includes(q);
      return matchesScope && matchesQuery;
    });
  }

  // -------------------------------------------------------------
  // Command Palette & Global Keyboard Shortcuts
  // -------------------------------------------------------------

  interface PaletteItem {
    id: string;
    label: string;
    category: 'guide' | 'action';
    detail?: string;
    action: () => void;
  }

  function getPaletteItems(): PaletteItem[] {
    const items: PaletteItem[] = [];

    items.push({
      id: 'act-p-explore',
      label: 'Perspective: 1. Explore 3D Graph & Inspector',
      category: 'action',
      detail: 'Ctrl+1',
      action: () => switchPerspective('explore')
    });

    items.push({
      id: 'act-p-author',
      label: 'Perspective: 2. Author & Anchor Studio',
      category: 'action',
      detail: 'Ctrl+2',
      action: () => switchPerspective('author')
    });

    items.push({
      id: 'act-p-practice',
      label: 'Perspective: 3. Focus Practice Stream',
      category: 'action',
      detail: 'Ctrl+3',
      action: () => switchPerspective('practice')
    });

    items.push({
      id: 'act-open',
      label: 'Open Knowledge Graph Directory...',
      category: 'action',
      detail: 'Ctrl+O',
      action: () => handleNativeFolderPick()
    });

    items.push({
      id: 'act-sidebar',
      label: leftPaneOpen ? 'Hide Left Guide Navigator' : 'Show Left Guide Navigator',
      category: 'action',
      detail: 'Ctrl+B',
      action: () => { leftPaneOpen = !leftPaneOpen; }
    });

    items.push({
      id: 'act-mode',
      label: centerMode === 'edit' ? 'Switch to Rendered Guide View' : 'Switch to Markdown Source Editor',
      category: 'action',
      detail: 'Ctrl+E',
      action: () => {
        perspective = 'author';
        centerMode = centerMode === 'edit' ? 'rendered' : 'edit';
      }
    });

    items.push({
      id: 'act-ollama-gen',
      label: `Generate Question from Segment (${getActiveSegment()}) with Ollama`,
      category: 'action',
      detail: 'Ctrl+G',
      action: () => generateWithOllama()
    });

    if (centerMode === 'edit' && isDirty) {
      items.push({
        id: 'act-save',
        label: `Save ${selectedGuideId}.md to disk`,
        category: 'action',
        detail: 'Ctrl+S',
        action: () => saveActiveGuideFile()
      });
    }

    if (graph) {
      graph.guides.forEach(g => {
        items.push({
          id: `guide-${g.id}`,
          label: g.id,
          category: 'guide',
          detail: `${g.scope} • ${getMastery(g.id)}%`,
          action: () => selectGuide(g.id)
        });
      });
    }

    if (!commandQuery.trim()) return items;

    const q = commandQuery.toLowerCase();
    return items.filter(it =>
      it.label.toLowerCase().includes(q) || (it.detail && it.detail.toLowerCase().includes(q))
    );
  }

  function openCommandPalette() {
    showCommandPalette = true;
    commandQuery = '';
    paletteSelectedIndex = 0;
    setTimeout(() => {
      paletteInputEl?.focus();
    }, 20);
  }

  function closeCommandPalette() {
    showCommandPalette = false;
  }

  function handleGlobalKeydown(e: KeyboardEvent) {
    const isCmd = e.metaKey || e.ctrlKey;

    if (isCmd && (e.key === 'p' || e.key === 'P')) {
      e.preventDefault();
      if (showCommandPalette) {
        closeCommandPalette();
      } else {
        openCommandPalette();
      }
      return;
    }

    if (isCmd && e.key === '1') {
      e.preventDefault();
      switchPerspective('explore');
      return;
    }

    if (isCmd && e.key === '2') {
      e.preventDefault();
      switchPerspective('author');
      return;
    }

    if (isCmd && e.key === '3') {
      e.preventDefault();
      switchPerspective('practice');
      return;
    }

    if (isCmd && (e.key === 'b' || e.key === 'B')) {
      e.preventDefault();
      leftPaneOpen = !leftPaneOpen;
      return;
    }

    if (isCmd && (e.key === 'o' || e.key === 'O')) {
      e.preventDefault();
      handleNativeFolderPick();
      return;
    }

    if (isCmd && (e.key === 's' || e.key === 'S')) {
      e.preventDefault();
      saveActiveGuideFile();
      return;
    }

    if (isCmd && (e.key === 'e' || e.key === 'E')) {
      e.preventDefault();
      perspective = 'author';
      centerMode = centerMode === 'edit' ? 'rendered' : 'edit';
      return;
    }

    if (isCmd && (e.key === 'g' || e.key === 'G')) {
      e.preventDefault();
      generateWithOllama();
      return;
    }

    if (e.key === 'Escape') {
      if (showCommandPalette) {
        closeCommandPalette();
      }
      showRecentsMenu = false;
    }
  }

  function handlePaletteKeydown(e: KeyboardEvent) {
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
        closeCommandPalette();
      }
    }
  }
</script>

<div class="studio-container">
  <!-- Top Perspective & Workspace Bar (44px) -->
  <header class="app-header">
    <!-- Left Cluster: Brand + Unified Workspace Dropdown + Search -->
    <div class="header-left">
      <span class="brand-badge">QST</span>

      <div class="recents-wrapper">
        <button
          type="button"
          class="workspace-Dropdown-btn"
          onclick={() => showRecentsMenu = !showRecentsMenu}
          title="Switch or Open Knowledge Graph Workspace"
        >
          <span class="ws-folder-icon">📁</span>
          <span class="ws-name">{graph ? graph.title : (graphDir.split('/').pop() || 'Workspace')}</span>
          <span class="ws-caret">▾</span>
        </button>

        {#if showRecentsMenu}
          <div class="recents-dropdown">
            <button type="button" class="dropdown-action-item" onclick={handleNativeFolderPick}>
              <span>+ Open Knowledge Graph Directory...</span>
              <kbd>Ctrl+O</kbd>
            </button>
            <div class="dropdown-divider"></div>
            <div class="dropdown-header">Recent Workspaces</div>
            {#if recentGraphs.length === 0}
              <div class="dropdown-empty">No recent projects</div>
            {:else}
              {#each recentGraphs as rPath}
                <button
                  type="button"
                  class="dropdown-item {rPath === graphDir ? 'active' : ''}"
                  onclick={() => handleLoadGraph(rPath)}
                >
                  <span class="path-title">{rPath.split('/').pop()}</span>
                  <span class="path-full">{rPath}</span>
                </button>
              {/each}
            {/if}
          </div>
        {/if}
      </div>

      <button class="btn btn-secondary btn-sm cmd-btn" onclick={openCommandPalette} title="Command Palette (Ctrl+P)">
        <span>Search</span>
        <kbd>Ctrl+P</kbd>
      </button>
    </div>

    <!-- Center Cluster: Segmented Perspective Switcher -->
    <nav class="perspective-switcher" aria-label="Workspace Perspectives">
      <button
        type="button"
        class="perspective-btn {perspective === 'explore' ? 'active' : ''}"
        onclick={() => switchPerspective('explore')}
        title="Explore 3D Knowledge Graph & Live Node Inspector (Ctrl+1)"
      >
        <span class="p-num">1</span>
        <span>Explore 3D</span>
      </button>
      <button
        type="button"
        class="perspective-btn {perspective === 'author' ? 'active' : ''}"
        onclick={() => switchPerspective('author')}
        title="Author Guides, Anchor Segments & Build Questions (Ctrl+2)"
      >
        <span class="p-num">2</span>
        <span>Author & Anchor</span>
      </button>
      <button
        type="button"
        class="perspective-btn {perspective === 'practice' ? 'active' : ''}"
        onclick={() => switchPerspective('practice')}
        title="Distraction-Free Adaptive Practice Stream (Ctrl+3)"
      >
        <span class="p-num">3</span>
        <span>Focus Practice</span>
      </button>
    </nav>

    <!-- Right Cluster: Contextual Pane Collapse Controls -->
    <div class="header-right">
      {#if perspective === 'author' || perspective === 'explore'}
        <button
          type="button"
          class="pane-toggle-btn {leftPaneOpen ? 'is-open' : ''}"
          onclick={() => leftPaneOpen = !leftPaneOpen}
          title="Toggle Left Guide Navigator (Ctrl+B)"
        >
          ◧ Guides
        </button>
        <button
          type="button"
          class="pane-toggle-btn {rightPaneOpen ? 'is-open' : ''}"
          onclick={() => rightPaneOpen = !rightPaneOpen}
          title="Toggle Right {perspective === 'explore' ? 'Inspector' : 'Question Studio'}"
        >
          ◨ {perspective === 'explore' ? 'Inspector' : 'Studio'}
        </button>
      {:else}
        <button
          type="button"
          class="pane-toggle-btn {leftPaneOpen ? 'is-open' : ''}"
          onclick={() => leftPaneOpen = !leftPaneOpen}
          title="Toggle Guide List Drawer (Ctrl+B)"
        >
          ◧ Guides
        </button>
        <span class="focus-split-badge">50/50 Focus Split</span>
      {/if}
    </div>
  </header>

  <!-- Command Palette Modal (Ctrl+P) -->
  {#if showCommandPalette}
    <div
      role="button"
      tabindex="-1"
      class="palette-backdrop"
      onclick={closeCommandPalette}
      onkeydown={(e) => { if (e.key === 'Escape') closeCommandPalette(); }}
    >
      <div
        role="dialog"
        tabindex="-1"
        class="palette-dialog"
        onclick={(e) => e.stopPropagation()}
        onkeydown={(e) => e.stopPropagation()}
      >
        <div class="palette-input-wrap">
          <span class="palette-icon">⌕</span>
          <input
            bind:this={paletteInputEl}
            type="text"
            bind:value={commandQuery}
            placeholder="Jump to a guide, switch perspective, or run a studio command..."
            class="palette-input"
            onkeydown={handlePaletteKeydown}
          />
          <kbd class="palette-esc">ESC</kbd>
        </div>

        <div class="palette-results">
          {#each getPaletteItems() as item, idx}
            <button
              type="button"
              class="palette-item {idx === paletteSelectedIndex ? 'selected' : ''}"
              onclick={() => { item.action(); closeCommandPalette(); }}
              onmouseenter={() => paletteSelectedIndex = idx}
            >
              <div class="palette-item-main">
                <span class="palette-category-badge">{item.category}</span>
                <span class="palette-label">{item.label}</span>
              </div>
              {#if item.detail}
                <span class="palette-detail">{item.detail}</span>
              {/if}
            </button>
          {/each}
        </div>
      </div>
    </div>
  {/if}

  <!-- Main Workspace Area -->
  {#if !graph}
    <div class="welcome-dashboard">
      <div class="welcome-card">
        <div class="welcome-brand">
          <span class="brand-badge-lg">QST STUDIO</span>
          <h2>Knowledge Graph Studio</h2>
          <p>Explore 3D topology, author line-anchored guides, and practice with adaptive AI tutoring.</p>
        </div>

        <div class="welcome-actions">
          <button class="btn btn-primary welcome-btn" onclick={handleNativeFolderPick}>
            📁 Open Knowledge Graph Directory
          </button>
          <button class="btn btn-secondary welcome-btn" onclick={() => handleLoadGraph('/home/speer/Code/golang/qst/examples/C_programming')}>
            ⚡ Load C Programming Example Graph
          </button>
        </div>

        {#if recentGraphs.length > 0}
          <div class="welcome-recents">
            <h3>Recent Workspaces</h3>
            <div class="welcome-recents-list">
              {#each recentGraphs as rPath}
                <button type="button" class="recent-link" onclick={() => handleLoadGraph(rPath)}>
                  <span>📁 {rPath.split('/').pop()}</span>
                  <span class="recent-path">{rPath}</span>
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <div class="welcome-shortcuts">
          <div class="shortcut-tag"><kbd>Ctrl+1/2/3</kbd> Perspectives</div>
          <div class="shortcut-tag"><kbd>Ctrl+P</kbd> Command Palette</div>
          <div class="shortcut-tag"><kbd>Ctrl+G</kbd> Ollama Draft</div>
          <div class="shortcut-tag"><kbd>Ctrl+B</kbd> Toggle Sidebar</div>
        </div>
      </div>
    </div>
  {:else}
    <div
      class="workspace-stage perspective-{perspective} {leftPaneOpen ? 'has-left' : 'no-left'} {rightPaneOpen ? 'has-right' : 'no-right'}"
    >
      <!-- Shared Collapsible Left Guide Navigator -->
      {#if leftPaneOpen}
        <aside class="sidebar-pane">
          <div class="pane-header">
            <div class="pane-header-top">
              <span class="pane-title">Guides ({getFilteredGuides().length}/{graph.guides.length})</span>
              <button
                type="button"
                class="btn-collapse-inline"
                onclick={() => leftPaneOpen = false}
                title="Collapse Guide Navigator (Ctrl+B)"
              >
                ✕
              </button>
            </div>

            <input
              type="text"
              bind:value={guideSearchQuery}
              placeholder="Filter guides by name..."
              class="guide-search-input"
            />

            <div class="scope-tags">
              <button
                class="scope-btn {selectedScope === 'all' ? 'active' : ''}"
                onclick={() => selectedScope = 'all'}
              >All</button>
              {#each graph.scopes as sc}
                <button
                  class="scope-btn {selectedScope === sc ? 'active' : ''}"
                  onclick={() => selectedScope = sc}
                >{sc}</button>
              {/each}
            </div>
          </div>

          <div class="guide-list">
            {#each getFilteredGuides() as guide}
              <button
                type="button"
                class="guide-card {guide.id === selectedGuideId ? 'selected' : ''}"
                onclick={() => selectGuide(guide.id)}
              >
                <div class="guide-card-top">
                  <span class="guide-name">{guide.id}</span>
                  <span class="scope-pill">{guide.scope}</span>
                </div>
                <div class="guide-card-bottom">
                  <div class="mastery-track">
                    <div
                      class="mastery-fill"
                      style="width: {getMastery(guide.id)}%; opacity: {0.3 + (getMastery(guide.id) / 100) * 0.7};"
                    ></div>
                  </div>
                  <span class="mastery-label">{getMastery(guide.id)}%</span>
                </div>
              </button>
            {/each}
          </div>
        </aside>
      {/if}

      <!-- ========================================================= -->
      <!-- PERSPECTIVE 1: EXPLORE 3D + LIVE NODE INSPECTOR SPLIT     -->
      <!-- ========================================================= -->
      {#if perspective === 'explore'}
        <main class="explore-canvas-pane">
          <Graph3D
            graph={graph}
            learnerStates={learnerStates}
            selectedGuideId={selectedGuideId}
            onNodeSelect={(id) => { selectGuide(id); rightPaneOpen = true; }}
          />
        </main>

        {#if rightPaneOpen}
          <section class="studio-pane inspector-pane">
            <div class="inspector-header">
              <div>
                <div class="field-label">3D Node Inspector</div>
                <h3 class="inspector-title">{selectedGuide ? selectedGuide.id : 'Select a Node'}</h3>
              </div>
              <button
                type="button"
                class="btn-collapse-inline"
                onclick={() => rightPaneOpen = false}
                title="Close Inspector Drawer"
              >
                ✕
              </button>
            </div>

            {#if selectedGuide}
              <div class="inspector-body">
                <!-- Mastery & Scope Summary Card -->
                <div class="inspector-summary-card">
                  <div class="inspector-meta-top">
                    <span class="scope-pill active-scope-pill">{selectedGuide.scope}</span>
                    <span class="inspector-stat">{selectedGuide.line_count} lines • {quizItems.length} bank items</span>
                  </div>
                  <div class="inspector-mastery-row">
                    <span class="field-label">BKT Mastery</span>
                    <span class="inspector-mastery-val">{getMastery(selectedGuide.id)}%</span>
                  </div>
                  <div class="mastery-track">
                    <div
                      class="mastery-fill"
                      style="width: {getMastery(selectedGuide.id)}%;"
                    ></div>
                  </div>
                </div>

                <!-- 1-Click Perspective Jump Actions -->
                <div class="inspector-actions">
                  <button
                    type="button"
                    class="btn btn-primary"
                    onclick={() => practiceGuideNode(selectedGuide!.id)}
                  >
                    ⚡ Practice Node →
                  </button>
                  <button
                    type="button"
                    class="btn btn-secondary"
                    onclick={() => authorGuideNode(selectedGuide!.id)}
                  >
                    ✎ Author & Anchor →
                  </button>
                </div>

                <!-- Upstream Prerequisites -->
                <div class="inspector-section">
                  <span class="field-label">Prerequisites ({selectedGuide.prerequisites?.length || 0})</span>
                  {#if selectedGuide.prerequisites && selectedGuide.prerequisites.length > 0}
                    <div class="inspector-chip-list">
                      {#each selectedGuide.prerequisites as prereq}
                        <button
                          type="button"
                          class="meta-chip clickable-chip"
                          onclick={() => selectGuide(prereq)}
                        >
                          <span>{prereq}</span>
                          <span class="chip-pct">{getMastery(prereq)}%</span>
                        </button>
                      {/each}
                    </div>
                  {:else}
                    <div class="inspector-empty-note">Foundational leaf node (no prerequisites)</div>
                  {/if}
                </div>

                <!-- Linked Sub-guides -->
                {#if selectedGuide.sub_guides && selectedGuide.sub_guides.length > 0}
                  <div class="inspector-section">
                    <span class="field-label">Sub-Guides ({selectedGuide.sub_guides.length})</span>
                    <div class="inspector-chip-list">
                      {#each selectedGuide.sub_guides as sg}
                        <button
                          type="button"
                          class="meta-chip clickable-chip"
                          onclick={() => selectGuide(sg.guide)}
                        >
                          <span>{sg.guide}</span>
                          <span class="chip-pct">lines {sg.segment}</span>
                        </button>
                      {/each}
                    </div>
                  </div>
                {/if}

                <!-- Guide Markdown Preview -->
                <div class="inspector-section inspector-preview-section">
                  <span class="field-label">Rendered Guide Preview</span>
                  <div class="inspector-rendered-preview">
                    {#each guideRenderedBlocks as blk}
                      <div class="inspector-rendered-block">
                        <span class="block-line-pill">
                          L{blk.startLine}{blk.endLine > blk.startLine ? `–${blk.endLine}` : ''}
                        </span>
                        <div class="qst-prose">{@html blk.html}</div>
                      </div>
                    {/each}
                  </div>
                </div>
              </div>
            {:else}
              <div class="empty-state">Click any node in the 3D graph to inspect its contents and mastery.</div>
            {/if}
          </section>
        {/if}

      <!-- ========================================================= -->
      <!-- PERSPECTIVE 2: AUTHOR & ANCHOR STUDIO                     -->
      <!-- ========================================================= -->
      {:else if perspective === 'author'}
        <main class="content-pane">
          {#if selectedGuide}
            <div class="guide-header">
              <div class="guide-title-row">
                <div class="title-left">
                  <h2>{selectedGuide.id}</h2>
                  <span class="scope-pill active-scope-pill">{selectedGuide.scope}</span>
                  {#if isDirty}
                    <span class="dirty-badge">● Modified</span>
                  {/if}
                </div>

                <div class="center-mode-toggle">
                  <button
                    class="toggle-btn {centerMode === 'rendered' ? 'active' : ''}"
                    onclick={() => centerMode = 'rendered'}
                    title="Rendered Markdown & LaTeX View with Block Line Anchoring"
                  >
                    📖 Rendered
                  </button>
                  <button
                    class="toggle-btn {centerMode === 'anchor' ? 'active' : ''}"
                    onclick={() => centerMode = 'anchor'}
                    title="Exact Single-Line Gutter View"
                  >
                    🎯 Line Gutter
                  </button>
                  <button
                    class="toggle-btn {centerMode === 'edit' ? 'active' : ''}"
                    onclick={() => centerMode = 'edit'}
                    title="Direct Markdown Editor (Ctrl+E)"
                  >
                    📝 Edit Source
                  </button>
                  {#if centerMode === 'edit'}
                    <button
                      class="btn btn-primary btn-sm save-btn"
                      disabled={!isDirty}
                      onclick={saveActiveGuideFile}
                      title="Save to disk (Ctrl+S)"
                    >
                      Save (Ctrl+S)
                    </button>
                  {/if}
                </div>
              </div>

              <div class="meta-row">
                {#if selectedGuide.prerequisites && selectedGuide.prerequisites.length > 0}
                  <div class="meta-group">
                    <span class="meta-label">Prerequisites:</span>
                    {#each selectedGuide.prerequisites as p}
                      <button type="button" class="meta-chip clickable-chip" onclick={() => selectGuide(p)}>
                        {p}
                      </button>
                    {/each}
                  </div>
                {/if}

                {#if learningPath.length > 0}
                  <div class="meta-group">
                    <span class="meta-label">Sequence:</span>
                    <span class="path-chain">{learningPath.join(' → ')}</span>
                  </div>
                {/if}
              </div>

              {#if centerMode !== 'edit'}
                <div class="segment-banner">
                  <span>Anchored Segment: <strong>lines {getActiveSegment()}</strong></span>
                  <span class="segment-hint">
                    {centerMode === 'rendered'
                      ? 'Click any L# badge (or Shift+click) to anchor block lines'
                      : 'Click gutter line numbers to anchor exact line range'}
                  </span>
                </div>
              {/if}
            </div>

            {#if centerMode === 'rendered'}
              <div class="rendered-guide-viewer">
                {#each guideRenderedBlocks as blk}
                  {@const isSelected = doesBlockOverlapSelection(blk)}
                  <div class="rendered-block-row {isSelected ? 'block-selected' : ''}">
                    <button
                      type="button"
                      class="block-line-badge"
                      onclick={(e) => handleBlockAnchorClick(blk, e)}
                      title="Click to anchor lines {blk.startLine}-{blk.endLine} (Shift+click to extend range)"
                    >
                      L{blk.startLine}{blk.endLine > blk.startLine ? `–${blk.endLine}` : ''}
                    </button>
                    <div class="rendered-block-content qst-prose">
                      {@html blk.html}
                    </div>
                  </div>
                {/each}
              </div>
            {:else if centerMode === 'anchor'}
              <div class="guide-body-viewer">
                {#each guideBodyLines as line, i}
                  {@const lineNum = i + 1}
                  {@const inSelectedRange = startLine !== null && (endLine !== null ? (lineNum >= startLine && lineNum <= endLine) : lineNum === startLine)}
                  <div class="code-line {inSelectedRange ? 'line-selected' : ''}">
                    <button
                      type="button"
                      class="gutter-num"
                      onclick={() => handleLineClick(lineNum)}
                    >
                      {lineNum}
                    </button>
                    <div class="line-text">{line || ' '}</div>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="editor-container">
                <textarea
                  class="raw-markdown-editor"
                  bind:value={editableContent}
                  oninput={() => isDirty = true}
                  placeholder="Edit Markdown source code (supports $...$ and $$...$$ LaTeX)..."
                  spellcheck="false"
                ></textarea>
              </div>
            {/if}
          {:else}
            <div class="empty-state">Select a guide from the sidebar to author or anchor segments.</div>
          {/if}
        </main>

        {#if rightPaneOpen}
          <section class="studio-pane">
            <div class="studio-pane-header">
              <div class="studio-pane-title-group">
                <span class="field-label">Question Builder</span>
                <span class="builder-mode-pill">{editingItemId ? `Editing ${editingItemId}` : 'New Item'}</span>
              </div>
              <div class="studio-pane-actions">
                <button
                  type="button"
                  class="btn btn-secondary btn-sm {builderPreviewMode ? 'btn-active-toggle' : ''}"
                  onclick={() => builderPreviewMode = !builderPreviewMode}
                  title="Toggle Rendered Markdown & LaTeX Question Preview"
                >
                  {builderPreviewMode ? '✎ Edit Fields' : '👁 Preview'}
                </button>
                {#if editingItemId}
                  <button type="button" class="btn btn-secondary btn-sm" onclick={resetBuilderForm}>
                    + New
                  </button>
                {/if}
                <button
                  type="button"
                  class="btn btn-primary btn-sm"
                  onclick={generateWithOllama}
                  disabled={isGenerating || !selectedGuideId}
                  title="Generate question draft from anchored segment with Ollama (Ctrl+G)"
                >
                  {isGenerating ? 'Generating...' : '⚡ Ollama Draft'}
                </button>
              </div>
            </div>

            <div class="tab-content">
              <div class="builder-form">
                <div class="form-group">
                  <span class="field-label">Target Guide & Segment</span>
                  <div class="anchor-badge">
                    <span>{selectedGuideId}</span>
                    <span class="badge-separator">/</span>
                    <span>lines {getActiveSegment()}</span>
                    <span class="badge-separator">/</span>
                    <span>{selectedGuide?.scope}</span>
                  </div>
                </div>

                {#if builderPreviewMode}
                  <!-- Live Rendered Markdown + KaTeX Question Preview -->
                  <div class="builder-preview-card">
                    <div class="field-label">Rendered Markdown & LaTeX Preview</div>
                    {#if itemTitle}
                      <div class="turn-title">{itemTitle}</div>
                    {/if}
                    <div class="item-prompt qst-prose">
                      {@html renderMarkdownWithLatex(itemPrompt)}
                    </div>
                    <div class="practice-options">
                      {#each itemChoices as ch}
                        <div class="preview-choice-card {ch.correct ? 'is-correct' : ''}">
                          <div class="preview-choice-row">
                            <span class="option-mark">{ch.correct ? '✓' : '○'}</span>
                            <div class="option-text qst-inline-md">
                              {@html renderMarkdownWithLatex(ch.text || '(Empty option)', true)}
                            </div>
                          </div>
                          {#if ch.feedback}
                            <div class="preview-feedback qst-prose">
                              {@html renderMarkdownWithLatex(ch.feedback, true)}
                            </div>
                          {/if}
                        </div>
                      {/each}
                    </div>
                  </div>
                {:else}
                  <div class="form-group">
                    <label for="item-title-input" class="field-label">Question Title</label>
                    <input id="item-title-input" type="text" bind:value={itemTitle} class="input-field" />
                  </div>

                  <div class="form-group">
                    <label for="item-prompt-input" class="field-label">Prompt (Markdown & $LaTeX$ supported)</label>
                    <textarea id="item-prompt-input" rows="3" bind:value={itemPrompt} class="input-field textarea"></textarea>
                  </div>

                  <div class="form-group">
                    <div class="choices-header">
                      <span class="field-label">Answer Choices</span>
                      <button class="btn btn-sm btn-secondary" onclick={addChoice}>+ Option</button>
                    </div>

                    <div class="choices-list">
                      {#each itemChoices as choice, idx}
                        <div class="choice-editor-card {choice.correct ? 'is-correct' : ''}">
                          <div class="choice-top">
                            <label class="radio-label">
                              <input
                                type="radio"
                                name="correct_choice"
                                checked={choice.correct}
                                onchange={() => setCorrectChoice(idx)}
                              />
                              <span>{choice.correct ? 'Correct Answer' : 'Distractor'}</span>
                            </label>
                            <button
                              class="btn-delete"
                              disabled={itemChoices.length <= 2}
                              onclick={() => removeChoice(idx)}
                            >✕</button>
                          </div>

                          <input
                            type="text"
                            bind:value={choice.text}
                            class="input-field"
                            placeholder="Option text (supports `code` and $math$)..."
                          />
                          <input
                            type="text"
                            bind:value={choice.feedback}
                            class="input-field feedback-input"
                            placeholder="Diagnostic feedback..."
                          />
                        </div>
                      {/each}
                    </div>
                  </div>
                {/if}

                <button class="btn btn-primary btn-block" onclick={saveItem}>
                  {editingItemId ? `Update ${editingItemId} in items.yaml` : 'Save Item to items.yaml'}
                </button>

                <!-- Saved Question Bank Drawer for Active Guide -->
                <div class="bank-drawer">
                  <div class="bank-drawer-header">
                    <span class="field-label">Saved Bank Items ({quizItems.length})</span>
                    <span class="bank-subtitle">Click to load & edit</span>
                  </div>

                  {#if quizItems.length === 0}
                    <div class="bank-empty">
                      No saved items for <strong>{selectedGuideId}</strong> yet. Use <strong>⚡ Ollama Draft</strong> or practice mode to auto-grow.
                    </div>
                  {:else}
                    <div class="bank-item-list">
                      {#each quizItems as qItem}
                        <div class="bank-item-row {editingItemId === qItem.id ? 'active' : ''}">
                          <button
                            type="button"
                            class="bank-item-select"
                            onclick={() => loadBankItemIntoBuilder(qItem)}
                          >
                            <div class="bank-item-top">
                              <span class="bank-item-id">{qItem.title || qItem.id}</span>
                              <span class="bank-item-seg">lines {qItem.segment}</span>
                            </div>
                            <div class="bank-item-prompt qst-inline-md">
                              {@html renderMarkdownWithLatex(qItem.prompt, true)}
                            </div>
                          </button>
                          <button
                            type="button"
                            class="btn btn-secondary btn-sm bank-practice-btn"
                            onclick={() => practiceSpecificBankItem(qItem)}
                            title="Practice this question"
                          >
                            ▶
                          </button>
                        </div>
                      {/each}
                    </div>
                  {/if}
                </div>
              </div>
            </div>
          </section>
        {/if}

      <!-- ========================================================= -->
      <!-- PERSPECTIVE 3: FOCUS PRACTICE (DISTRACTION-FREE 2-PANE)   -->
      <!-- ========================================================= -->
      {:else}
        <!-- Left Half: Active Question Stage -->
        <section class="practice-stage-pane">
          <div class="practice-stage-header">
            <div class="practice-scope-switch">
              <button
                type="button"
                class="scope-mode-btn {practiceScope === 'adaptive' ? 'active' : ''}"
                onclick={() => switchPracticeScope('adaptive')}
                title="Adaptive Graph Queue: prioritizes prerequisite readiness, FSRS due reviews, and lowest BKT mastery"
              >
                Adaptive Graph Queue
              </button>
              <button
                type="button"
                class="scope-mode-btn {practiceScope === 'guide' ? 'active' : ''}"
                onclick={() => switchPracticeScope('guide')}
                title="Current Guide: locks continuous practice to the selected guide"
              >
                Locked Guide ({selectedGuideId})
              </button>
            </div>

            <div class="practice-status-pills">
              {#if sessionQuestionCount > 0}
                <span class="session-score-pill" title="Session Accuracy">
                  Score: {sessionCorrectCount}/{sessionQuestionCount}
                </span>
              {/if}
              {#if isPrefetching}
                <span class="prefetch-pill prefetching" title="Generating/staging next question in background">
                  ◌ Pre-fetching Next
                </span>
              {:else if prefetchedTurn}
                <span class="prefetch-pill ready" title="Next question staged in memory (0ms wait)">
                  ● Next Ready
                </span>
              {/if}
            </div>
          </div>

          <div class="practice-stage-body">
            {#if isLoadingTurn}
              <div class="practice-loading-card">
                <div class="loading-spinner-line"></div>
                <div class="loading-title">Synthesizing On-the-Fly Assessment...</div>
                <div class="loading-sub">
                  Calibrating to <strong>{practiceScope === 'adaptive' ? 'Adaptive Frontier' : selectedGuideId}</strong> via local Ollama ({selectedModel || agentStatus?.active_model || 'local model'})
                </div>
              </div>
            {:else if practiceError}
              <div class="practice-card">
                <div class="item-prompt">Unable to load practice turn</div>
                <div class="agent-offline-msg">{practiceError}</div>
                <button class="btn btn-primary btn-sm" onclick={() => startNextPracticeTurn()}>
                  Retry Practice Turn
                </button>
              </div>
            {:else if currentTurn}
              {@const item = currentTurn.item}
              <div class="practice-card focus-practice-card">
                <div class="turn-header">
                  <div class="turn-badges">
                    <span class="turn-guide-badge">{item.guide_id}</span>
                    <span class="scope-pill">{item.scope || selectedGuide?.scope}</span>
                    <span class="item-segment">lines {item.segment}</span>
                  </div>
                  <span class="turn-source-badge {currentTurn.source === 'generated' ? 'is-generated' : 'is-bank'}">
                    {currentTurn.source === 'generated' ? '⚡ AUTO-GEN' : '📦 BANK'}
                  </span>
                </div>

                {#if currentTurn.reason}
                  <div class="turn-reason-bar">
                    <span>Schedule: <strong>{currentTurn.reason}</strong></span>
                    <span>Guide Mastery: <strong>{getMastery(item.guide_id)}%</strong></span>
                  </div>
                {/if}

                {#if item.title}
                  <div class="turn-title">{item.title}</div>
                {/if}

                <div class="item-prompt focus-prompt qst-prose">
                  {@html renderMarkdownWithLatex(item.prompt)}
                </div>

                <div class="practice-options">
                  {#each item.choices as opt}
                    {@const isSelected = practiceChoice === opt.id}
                    {@const showRight = practiceSubmitted && opt.correct}
                    {@const showWrong = practiceSubmitted && isSelected && !opt.correct}
                    <label
                      class="option-pill {isSelected ? 'option-selected' : ''} {showRight ? 'option-correct' : ''} {showWrong ? 'option-wrong' : ''}"
                    >
                      <input
                        type="radio"
                        name="practice_answer_{item.id}"
                        value={opt.id}
                        bind:group={practiceChoice}
                        disabled={practiceSubmitted}
                      />
                      <span class="option-text qst-inline-md">
                        {@html renderMarkdownWithLatex(opt.text, true)}
                      </span>
                      {#if showRight}
                        <span class="option-mark">✓</span>
                      {:else if showWrong}
                        <span class="option-mark">✕</span>
                      {/if}
                    </label>
                  {/each}
                </div>

                {#if !practiceSubmitted}
                  <div class="practice-actions-row">
                    <button
                      class="btn btn-primary btn-submit-turn"
                      disabled={!practiceChoice}
                      onclick={() => submitPracticeAnswer(item)}
                    >
                      Submit Response
                    </button>
                    <button
                      class="btn btn-secondary"
                      onclick={() => startNextPracticeTurn()}
                      title="Skip to next question"
                    >
                      Skip →
                    </button>
                  </div>
                {:else}
                  <div class="result-banner {lastAnswerCorrect ? 'result-correct' : 'result-incorrect'}">
                    <div class="result-header-row">
                      <span class="result-title">
                        {lastAnswerCorrect ? '✓ Correct Assessment' : '✕ Incorrect Assessment'}
                      </span>
                      {#if lastMasteryUpdate}
                        <span class="result-mastery-pill">
                          P(L): {Math.round(lastMasteryUpdate.p_mastery * 100)}% • S: {lastMasteryUpdate.stability_days.toFixed(1)}d
                        </span>
                      {/if}
                    </div>

                    {#if selectedChoiceFeedback}
                      <div class="diagnostic-feedback qst-prose">
                        {@html renderMarkdownWithLatex(selectedChoiceFeedback)}
                      </div>
                    {/if}

                    {#if correctChoiceFeedback}
                      <div class="diagnostic-feedback correct-note qst-prose">
                        <strong>Key Concept:</strong> {@html renderMarkdownWithLatex(correctChoiceFeedback, true)}
                      </div>
                    {/if}

                    {#if !lastAnswerCorrect && (currentTurn.linked_sub_guide || currentTurn.unmastered_prereq_id)}
                      <div class="remediation-box">
                        <span class="remediation-label">Recommended Remediation:</span>
                        <div class="remediation-buttons">
                          {#if currentTurn.linked_sub_guide && currentTurn.linked_sub_guide !== item.guide_id}
                            <button
                              type="button"
                              class="btn btn-secondary btn-sm"
                              onclick={() => drillDownRemediation(currentTurn!.linked_sub_guide)}
                            >
                              ↳ Drill Down Sub-guide: {currentTurn.linked_sub_guide}
                            </button>
                          {/if}
                          {#if currentTurn.unmastered_prereq_id && currentTurn.unmastered_prereq_id !== currentTurn.linked_sub_guide}
                            <button
                              type="button"
                              class="btn btn-secondary btn-sm"
                              onclick={() => drillDownRemediation(currentTurn!.unmastered_prereq_id)}
                            >
                              ↳ Review Prerequisite: {currentTurn.unmastered_prereq_id}
                            </button>
                          {/if}
                        </div>
                      </div>
                    {/if}

                    <div class="next-turn-row">
                      <button
                        type="button"
                        class="btn btn-primary btn-block next-turn-btn"
                        onclick={() => startNextPracticeTurn()}
                      >
                        <span>Next Question →</span>
                        {#if prefetchedTurn}
                          <span class="instant-badge">Instant</span>
                        {/if}
                      </button>
                    </div>
                  </div>
                {/if}
              </div>
            {:else}
              <div class="empty-state">
                <p>Start an adaptive continuous practice stream across your knowledge graph.</p>
                <button class="btn btn-primary" onclick={() => startNextPracticeTurn()}>
                  Start Continuous Practice
                </button>
              </div>
            {/if}
          </div>
        </section>

        <!-- Right Half: Synchronized Source Guide Viewer with Active Recall Cloak -->
        <main class="content-pane practice-source-pane">
          {#if selectedGuide}
            <div class="guide-header">
              <div class="guide-title-row">
                <div class="title-left">
                  <h2>{selectedGuide.id}</h2>
                  <span class="scope-pill active-scope-pill">{selectedGuide.scope}</span>
                </div>
                <div class="center-mode-toggle">
                  <button
                    class="toggle-btn {centerMode === 'rendered' ? 'active' : ''}"
                    onclick={() => centerMode = 'rendered'}
                  >
                    📖 Rendered
                  </button>
                  <button
                    class="toggle-btn {centerMode === 'anchor' ? 'active' : ''}"
                    onclick={() => centerMode = 'anchor'}
                  >
                    🎯 Line Gutter
                  </button>
                  <button
                    type="button"
                    class="toggle-btn"
                    onclick={() => switchPerspective('author')}
                    title="Open in Author & Anchor Perspective"
                  >
                    ✎ Edit
                  </button>
                </div>
              </div>

              <div class="segment-banner {currentTurn ? (practiceSubmitted ? 'banner-revealed' : 'banner-cloaked') : ''}">
                {#if currentTurn && !practiceSubmitted}
                  <span>🔒 Active Recall Cloak: <strong>lines {getActiveSegment()}</strong> blurred while answering</span>
                  <span class="segment-hint">Submit answer to reveal source</span>
                {:else if currentTurn && practiceSubmitted}
                  <span>🔓 Source Revealed: <strong>lines {getActiveSegment()}</strong> ({currentTurn.item.guide_id})</span>
                  <span class="segment-hint">Verify explanation in context</span>
                {:else}
                  <span>Anchored Segment: <strong>lines {getActiveSegment()}</strong></span>
                {/if}
              </div>
            </div>

            {#if centerMode === 'anchor'}
              <div class="guide-body-viewer">
                {#each guideBodyLines as line, i}
                  {@const lineNum = i + 1}
                  {@const inSelectedRange = startLine !== null && (endLine !== null ? (lineNum >= startLine && lineNum <= endLine) : lineNum === startLine)}
                  {@const isCloaked = currentTurn !== null && !practiceSubmitted && inSelectedRange}
                  {@const isRevealed = currentTurn !== null && practiceSubmitted && inSelectedRange}
                  <div class="code-line {inSelectedRange ? 'line-selected' : ''} {isCloaked ? 'line-cloaked' : ''} {isRevealed ? 'line-revealed' : ''}">
                    <span class="gutter-num">{lineNum}</span>
                    <div class="line-text">{line || ' '}</div>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="rendered-guide-viewer">
                {#each guideRenderedBlocks as blk}
                  {@const inRange = doesBlockOverlapSelection(blk)}
                  {@const isCloaked = currentTurn !== null && !practiceSubmitted && inRange}
                  {@const isRevealed = currentTurn !== null && practiceSubmitted && inRange}
                  <div class="rendered-block-row {inRange ? 'block-selected' : ''} {isCloaked ? 'rendered-block-cloaked' : ''} {isRevealed ? 'rendered-block-revealed' : ''}">
                    <span class="block-line-badge">
                      L{blk.startLine}{blk.endLine > blk.startLine ? `–${blk.endLine}` : ''}
                    </span>
                    <div class="rendered-block-content qst-prose">
                      {@html blk.html}
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          {:else}
            <div class="empty-state">Synchronizing guide context...</div>
          {/if}
        </main>
      {/if}
    </div>
  {/if}

  <!-- Slim 24px Bottom Telemetry & Agent Status Bar -->
  <footer class="app-statusbar">
    <div class="statusbar-left">
      <div class="statusbar-agent">
        <span class="agent-dot {agentStatus?.available ? 'online' : 'offline'}"></span>
        {#if agentStatus?.available && agentStatus.models && agentStatus.models.length > 0}
          <span class="statusbar-label">Ollama:</span>
          <select
            class="statusbar-model-select"
            value={selectedModel}
            onchange={(e) => handleModelChange((e.target as HTMLSelectElement).value)}
            aria-label="Active Ollama Model"
          >
            {#each agentStatus.models as m}
              <option value={m}>{m}</option>
            {/each}
          </select>
        {:else}
          <span class="statusbar-label">Ollama Offline</span>
        {/if}
        <button
          type="button"
          class="statusbar-refresh-btn"
          onclick={refreshAgentStatus}
          title="Refresh local Ollama connection"
        >
          ↻
        </button>
      </div>

      {#if graph}
        <span class="statusbar-sep">|</span>
        <span class="statusbar-metric">
          Graph Mastery: <strong>{getOverallGraphMastery()}%</strong>
        </span>
        <span class="statusbar-sep">|</span>
        <span class="statusbar-metric">
          {graph.guides.length} Guides • {quizItems.length} Items ({selectedGuideId || 'none'})
        </span>
      {/if}
    </div>

    <div class="statusbar-right">
      <span class="statusbar-msg">{statusMessage}</span>
    </div>
  </footer>
</div>

<style>
  .studio-container {
    display: flex;
    flex-direction: column;
    height: 100vh;
    width: 100vw;
    background: var(--bg-primary);
    overflow: hidden;
  }

  /* Top Perspective Header (44px) */
  .app-header {
    height: 44px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-color);
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    padding: 0 12px;
    gap: 12px;
    flex-shrink: 0;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .brand-badge {
    background: var(--accent-white);
    color: var(--text-dark);
    font-weight: 800;
    font-size: 10px;
    padding: 2px 6px;
    border-radius: 4px;
    letter-spacing: 0.5px;
  }

  .recents-wrapper {
    position: relative;
  }
  .workspace-Dropdown-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    height: 26px;
    padding: 0 9px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    color: var(--text-main);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: border-color 0.12s ease;
  }
  .workspace-Dropdown-btn:hover {
    border-color: var(--border-focus);
  }
  .ws-folder-icon {
    font-size: 11px;
  }
  .ws-caret {
    color: var(--text-dim);
    font-size: 10px;
  }
  .recents-dropdown {
    position: absolute;
    top: 32px;
    left: 0;
    width: 300px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 6px;
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.85);
    z-index: 100;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .dropdown-action-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 7px 8px;
    border-radius: 4px;
    border: none;
    background: transparent;
    color: var(--text-main);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
  }
  .dropdown-action-item:hover {
    background: var(--bg-card-hover);
  }
  .dropdown-action-item kbd {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    font-size: 9px;
    padding: 1px 4px;
    color: var(--text-dim);
  }
  .dropdown-divider {
    height: 1px;
    background: var(--border-color);
    margin: 4px 0;
  }
  .dropdown-header {
    font-size: 9px;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--text-dim);
    padding: 3px 8px;
    letter-spacing: 0.5px;
  }
  .dropdown-empty {
    font-size: 11px;
    color: var(--text-dim);
    padding: 8px;
    text-align: center;
  }
  .dropdown-item {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    padding: 6px 8px;
    border-radius: 4px;
    border: none;
    background: transparent;
    color: var(--text-main);
    cursor: pointer;
    text-align: left;
  }
  .dropdown-item:hover {
    background: var(--bg-card-hover);
  }
  .dropdown-item.active {
    background: var(--selection-bg);
  }
  .path-title {
    font-weight: 600;
    font-size: 11px;
  }
  .path-full {
    font-size: 10px;
    color: var(--text-dim);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    width: 100%;
  }

  .cmd-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    height: 26px;
  }
  .cmd-btn kbd {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    font-size: 9px;
    padding: 1px 4px;
    color: var(--text-dim);
  }

  /* Centered Perspective Switcher */
  .perspective-switcher {
    display: flex;
    align-items: center;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 2px;
    gap: 2px;
  }
  .perspective-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    background: transparent;
    border: none;
    padding: 4px 12px;
    border-radius: 4px;
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.12s ease;
  }
  .perspective-btn:hover {
    color: var(--text-main);
  }
  .perspective-btn.active {
    background: var(--accent-white);
    color: var(--text-dark);
    font-weight: 600;
  }
  .p-num {
    font-family: monospace;
    font-size: 9px;
    opacity: 0.65;
  }

  .header-right {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    gap: 6px;
  }
  .pane-toggle-btn {
    height: 26px;
    padding: 0 8px;
    border-radius: 4px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    color: var(--text-dim);
    font-size: 10px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.12s ease;
  }
  .pane-toggle-btn:hover {
    color: var(--text-main);
    border-color: var(--border-focus);
  }
  .pane-toggle-btn.is-open {
    color: var(--text-main);
    border-color: var(--border-focus);
    background: var(--selection-bg);
  }
  .focus-split-badge {
    font-size: 10px;
    font-weight: 600;
    color: var(--text-dim);
    padding: 3px 8px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-card);
  }

  /* Dynamic Workspace Grid Layouts per Perspective */
  .workspace-stage {
    flex: 1;
    display: grid;
    overflow: hidden;
  }

  /* Author & Explore Perspectives */
  .workspace-stage.perspective-author.has-left.has-right,
  .workspace-stage.perspective-explore.has-left.has-right {
    grid-template-columns: 260px 1fr 390px;
  }
  .workspace-stage.perspective-author.no-left.has-right,
  .workspace-stage.perspective-explore.no-left.has-right {
    grid-template-columns: 1fr 390px;
  }
  .workspace-stage.perspective-author.has-left.no-right,
  .workspace-stage.perspective-explore.has-left.no-right {
    grid-template-columns: 260px 1fr;
  }
  .workspace-stage.perspective-author.no-left.no-right,
  .workspace-stage.perspective-explore.no-left.no-right {
    grid-template-columns: 1fr;
  }

  /* Practice Perspective (50/50 Focus Split, with optional Left Guide Drawer) */
  .workspace-stage.perspective-practice.has-left {
    grid-template-columns: 250px 1fr 1fr;
  }
  .workspace-stage.perspective-practice.no-left {
    grid-template-columns: 1fr 1fr;
  }

  /* Left Guide Navigator Sidebar */
  .sidebar-pane {
    background: var(--bg-surface);
    border-right: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .pane-header {
    padding: 10px 12px;
    border-bottom: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    gap: 8px;
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
    letter-spacing: 0.6px;
    color: var(--text-dim);
  }
  .btn-collapse-inline {
    background: transparent;
    border: none;
    color: var(--text-dim);
    font-size: 11px;
    cursor: pointer;
    padding: 2px 4px;
  }
  .btn-collapse-inline:hover {
    color: var(--text-main);
  }
  .guide-search-input {
    width: 100%;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 5px 8px;
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
    gap: 4px;
  }
  .scope-btn {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-muted);
    font-size: 10px;
    padding: 2px 6px;
    cursor: pointer;
    transition: all 0.1s ease;
  }
  .scope-btn:hover {
    border-color: var(--border-focus);
    color: var(--text-main);
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
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .guide-card {
    display: block;
    width: 100%;
    text-align: left;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 8px 10px;
    cursor: pointer;
    transition: all 0.12s ease;
    color: inherit;
  }
  .guide-card:hover {
    background: var(--bg-card-hover);
    border-color: var(--border-focus);
  }
  .guide-card.selected {
    border-color: var(--accent-white);
    background: var(--selection-bg);
  }
  .guide-card-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 5px;
  }
  .guide-name {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-main);
  }
  .guide-card-bottom {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .mastery-track {
    flex: 1;
    height: 4px;
    background: var(--border-color);
    border-radius: 2px;
    overflow: hidden;
  }
  .mastery-fill {
    height: 100%;
    background: var(--accent-white);
    border-radius: 2px;
    transition: width 0.3s ease;
  }
  .mastery-label {
    font-size: 10px;
    font-weight: 600;
    color: var(--text-dim);
    min-width: 26px;
    text-align: right;
  }

  .scope-pill {
    font-size: 9px;
    font-weight: 600;
    text-transform: uppercase;
    padding: 2px 5px;
    border-radius: 3px;
    letter-spacing: 0.4px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    color: var(--text-muted);
  }
  .active-scope-pill {
    background: var(--bg-card);
    border-color: var(--border-focus);
    color: var(--text-main);
  }

  /* Explore 3D Perspective Canvas & Live Node Inspector */
  .explore-canvas-pane {
    width: 100%;
    height: 100%;
    overflow: hidden;
    position: relative;
  }
  .inspector-pane {
    display: flex;
    flex-direction: column;
  }
  .inspector-header {
    padding: 12px 14px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-card);
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
  }
  .inspector-title {
    margin: 3px 0 0 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--text-main);
  }
  .inspector-body {
    flex: 1;
    overflow-y: auto;
    padding: 14px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .inspector-summary-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .inspector-meta-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .inspector-stat {
    font-size: 10px;
    color: var(--text-dim);
    font-family: monospace;
  }
  .inspector-mastery-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .inspector-mastery-val {
    font-size: 12px;
    font-weight: 700;
    color: var(--text-main);
    font-family: monospace;
  }
  .inspector-actions {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }
  .inspector-section {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .inspector-chip-list {
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
  }
  .clickable-chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
    transition: all 0.12s ease;
  }
  .clickable-chip:hover {
    border-color: var(--accent-white);
    color: var(--text-main);
  }
  .chip-pct {
    font-size: 9px;
    color: var(--text-dim);
    font-family: monospace;
  }
  .inspector-empty-note {
    font-size: 11px;
    color: var(--text-dim);
  }
  .inspector-preview-section {
    flex: 1;
    min-height: 180px;
  }
  .inspector-rendered-preview {
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 8px 10px;
    max-height: 360px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .inspector-rendered-block {
    display: flex;
    flex-direction: column;
    gap: 3px;
    padding-bottom: 6px;
    border-bottom: 1px dashed rgba(255, 255, 255, 0.06);
  }
  .inspector-rendered-block:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }
  .block-line-pill {
    align-self: flex-start;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 9px;
    color: var(--text-dim);
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 0 5px;
  }
  .inspector-code-preview {
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 8px 0;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 11px;
    line-height: 18px;
    max-height: 340px;
    overflow-y: auto;
  }
  .preview-line {
    display: flex;
    padding: 0 10px;
  }
  .preview-ln {
    width: 28px;
    color: var(--text-dim);
    text-align: right;
    padding-right: 8px;
    user-select: none;
  }
  .preview-txt {
    flex: 1;
    color: var(--text-muted);
    white-space: pre-wrap;
    word-break: break-word;
  }

  /* Center Content Pane */
  .content-pane {
    background: var(--bg-primary);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .guide-header {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-surface);
  }
  .guide-title-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }
  .title-left {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .title-left h2 {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    letter-spacing: -0.2px;
  }
  .dirty-badge {
    font-size: 10px;
    color: var(--accent-white);
    background: var(--border-color);
    padding: 2px 6px;
    border-radius: 3px;
  }

  .center-mode-toggle {
    display: flex;
    align-items: center;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 2px;
    gap: 2px;
  }
  .toggle-btn {
    background: transparent;
    border: none;
    padding: 3px 8px;
    border-radius: 4px;
    color: var(--text-muted);
    font-size: 10px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.1s ease;
  }
  .toggle-btn.active {
    background: var(--accent-white);
    color: var(--text-dark);
    font-weight: 600;
  }
  .save-btn {
    margin-left: 4px;
  }

  .meta-row {
    display: flex;
    flex-direction: column;
    gap: 5px;
    margin-bottom: 8px;
  }
  .meta-group {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
    font-size: 11px;
  }
  .meta-label {
    color: var(--text-dim);
    font-weight: 500;
  }
  .meta-chip {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 1px 6px;
    border-radius: 3px;
    color: var(--text-muted);
    font-size: 10px;
  }
  .path-chain {
    color: var(--text-muted);
    font-family: monospace;
    font-size: 11px;
  }
  .segment-banner {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 5px 10px;
    border-radius: 4px;
    font-size: 11px;
    color: var(--text-muted);
    transition: all 0.2s ease;
  }
  .segment-banner.banner-cloaked {
    border-color: var(--border-focus);
    background: rgba(255, 255, 255, 0.04);
    color: var(--text-main);
  }
  .segment-banner.banner-revealed {
    border-color: var(--accent-white);
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-main);
  }
  .segment-hint {
    color: var(--text-dim);
    font-size: 10px;
  }

  /* Rendered Guide Viewer (Markdown + KaTeX Blocks with Source Line Anchors) */
  .rendered-guide-viewer {
    flex: 1;
    overflow-y: auto;
    padding: 14px 18px 28px 14px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .rendered-block-row {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 6px 10px 6px 6px;
    border-radius: 5px;
    border-left: 2px solid transparent;
    transition: background 0.16s ease, border-color 0.16s ease;
  }
  .rendered-block-row:hover {
    background: rgba(255, 255, 255, 0.025);
  }
  .rendered-block-row.block-selected {
    background: var(--selection-bg);
    border-left-color: var(--accent-white);
  }
  .rendered-block-row.rendered-block-cloaked .rendered-block-content {
    filter: blur(6px);
    opacity: 0.35;
    user-select: none;
    transition: filter 0.25s ease, opacity 0.25s ease;
  }
  .rendered-block-row.rendered-block-revealed {
    background: rgba(255, 255, 255, 0.09);
    border-left-color: var(--accent-white);
  }
  .rendered-block-row.rendered-block-revealed .rendered-block-content {
    filter: none;
    opacity: 1;
    transition: filter 0.25s ease, opacity 0.25s ease;
  }
  .block-line-badge {
    flex-shrink: 0;
    min-width: 46px;
    text-align: right;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 2px 6px;
    margin-top: 2px;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 10px;
    color: var(--text-dim);
    cursor: pointer;
    user-select: none;
    transition: all 0.12s ease;
  }
  .block-line-badge:hover {
    border-color: var(--accent-white);
    color: var(--text-main);
  }
  .rendered-block-row.block-selected .block-line-badge {
    border-color: var(--accent-white);
    color: var(--text-main);
    background: var(--bg-card);
  }
  .rendered-block-content {
    flex: 1;
    min-width: 0;
  }

  .guide-body-viewer {
    flex: 1;
    overflow-y: auto;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 12px;
    line-height: 20px;
    padding: 10px 0;
  }
  .code-line {
    display: flex;
    align-items: flex-start;
    transition: background 0.2s ease;
  }
  .code-line:hover {
    background: rgba(255, 255, 255, 0.02);
  }
  .code-line.line-selected {
    background: var(--selection-bg);
  }
  .code-line.line-cloaked .line-text {
    filter: blur(5px);
    opacity: 0.4;
    user-select: none;
    transition: filter 0.25s ease, opacity 0.25s ease;
  }
  .code-line.line-revealed {
    background: rgba(255, 255, 255, 0.11);
    border-left: 2px solid var(--accent-white);
  }
  .code-line.line-revealed .line-text {
    color: var(--text-main);
    filter: none;
    opacity: 1;
    transition: filter 0.25s ease, opacity 0.25s ease;
  }
  .gutter-num {
    background: transparent;
    border: none;
    font-family: inherit;
    font-size: inherit;
    width: 40px;
    text-align: right;
    padding-right: 12px;
    color: var(--text-dim);
    user-select: none;
    cursor: pointer;
  }
  .gutter-num:hover {
    color: var(--text-main);
  }
  .line-text {
    flex: 1;
    padding-right: 16px;
    white-space: pre-wrap;
    word-break: break-all;
    color: var(--text-muted);
  }

  .editor-container {
    flex: 1;
    display: flex;
    overflow: hidden;
  }
  .raw-markdown-editor {
    flex: 1;
    background: var(--bg-primary);
    border: none;
    outline: none;
    color: var(--text-main);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 13px;
    line-height: 1.6;
    padding: 14px 18px;
    resize: none;
    tab-size: 2;
  }

  /* Right Studio Pane (Author Perspective) */
  .studio-pane {
    background: var(--bg-surface);
    border-left: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .studio-pane-header {
    padding: 10px 14px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-card);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .studio-pane-title-group {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .builder-mode-pill {
    font-size: 9px;
    font-family: monospace;
    color: var(--text-muted);
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    padding: 1px 5px;
    border-radius: 3px;
  }
  .studio-pane-actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .tab-content {
    flex: 1;
    overflow-y: auto;
    padding: 14px;
  }

  .builder-form {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .field-label {
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--text-dim);
    letter-spacing: 0.5px;
  }
  .anchor-badge {
    display: flex;
    align-items: center;
    gap: 6px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 5px 8px;
    border-radius: 4px;
    font-size: 11px;
    color: var(--text-muted);
  }
  .badge-separator {
    color: var(--border-color);
  }
  .input-field {
    width: 100%;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 7px 9px;
    font-size: 12px;
    color: var(--text-main);
    transition: border-color 0.12s ease;
  }
  .input-field:focus {
    outline: none;
    border-color: var(--border-focus);
  }
  .textarea {
    resize: vertical;
    line-height: 1.4;
  }

  .choices-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .choices-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .choice-editor-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 7px;
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .choice-editor-card.is-correct {
    border-color: var(--border-focus);
    background: var(--selection-bg);
  }
  .choice-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .radio-label {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 10px;
    color: var(--text-muted);
    cursor: pointer;
  }
  .btn-delete {
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    font-size: 11px;
  }
  .btn-delete:hover {
    color: var(--accent-danger);
  }
  .feedback-input {
    font-size: 11px;
    color: var(--text-dim);
  }
  .btn-active-toggle {
    border-color: var(--accent-white) !important;
    color: var(--text-main) !important;
    background: var(--selection-bg) !important;
  }
  .builder-preview-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 9px 11px;
    color: var(--text-main);
    min-height: 36px;
  }
  .preview-choice-card {
    gap: 6px;
  }
  .preview-choice-row {
    display: flex;
    align-items: baseline;
    gap: 8px;
  }
  .preview-feedback {
    margin-top: 2px;
    font-size: 11px;
  }

  /* Saved Bank Items Drawer in Author Mode */
  .bank-drawer {
    margin-top: 6px;
    padding-top: 12px;
    border-top: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .bank-drawer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .bank-subtitle {
    font-size: 10px;
    color: var(--text-dim);
  }
  .bank-empty {
    font-size: 11px;
    color: var(--text-dim);
    background: var(--bg-card);
    border: 1px dashed var(--border-color);
    border-radius: 5px;
    padding: 10px;
    line-height: 1.4;
  }
  .bank-item-list {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .bank-item-row {
    display: flex;
    align-items: stretch;
    gap: 4px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 4px;
    transition: border-color 0.12s ease;
  }
  .bank-item-row:hover,
  .bank-item-row.active {
    border-color: var(--accent-white);
  }
  .bank-item-select {
    flex: 1;
    background: transparent;
    border: none;
    text-align: left;
    padding: 4px 6px;
    cursor: pointer;
    color: inherit;
    overflow: hidden;
  }
  .bank-item-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2px;
  }
  .bank-item-id {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-main);
  }
  .bank-item-seg {
    font-family: monospace;
    font-size: 9px;
    color: var(--text-dim);
  }
  .bank-item-prompt {
    font-size: 10px;
    color: var(--text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .bank-practice-btn {
    align-self: center;
  }

  /* Focus Practice Perspective (50/50 Split Stage) */
  .practice-stage-pane {
    background: var(--bg-surface);
    border-right: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .practice-stage-header {
    padding: 12px 18px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-card);
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 10px;
  }
  .practice-stage-body {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  .focus-practice-card {
    width: 100%;
    max-width: 600px;
    padding: 20px;
    gap: 14px;
  }
  .focus-prompt {
    font-size: 15px;
    line-height: 1.5;
  }
  .practice-source-pane {
    background: var(--bg-primary);
  }

  .practice-scope-switch {
    display: flex;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 2px;
    gap: 2px;
  }
  .scope-mode-btn {
    background: transparent;
    border: none;
    padding: 4px 10px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 500;
    color: var(--text-dim);
    cursor: pointer;
    transition: all 0.12s ease;
  }
  .scope-mode-btn:hover {
    color: var(--text-main);
  }
  .scope-mode-btn.active {
    background: var(--accent-white);
    color: var(--text-dark);
    font-weight: 600;
  }
  .practice-status-pills {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .session-score-pill {
    font-family: monospace;
    font-size: 10px;
    font-weight: 600;
    padding: 3px 7px;
    border-radius: 4px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    color: var(--text-muted);
  }
  .prefetch-pill {
    font-size: 10px;
    font-weight: 600;
    padding: 3px 7px;
    border-radius: 4px;
    border: 1px solid var(--border-color);
    background: var(--bg-surface);
    color: var(--text-dim);
  }
  .prefetch-pill.ready {
    border-color: var(--border-focus);
    color: var(--text-main);
  }
  .practice-loading-card {
    width: 100%;
    max-width: 540px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 32px 20px;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 10px;
  }
  .loading-spinner-line {
    width: 42px;
    height: 2px;
    background: var(--accent-white);
  }
  .loading-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-main);
  }
  .loading-sub {
    font-size: 11px;
    color: var(--text-dim);
    line-height: 1.4;
  }
  .practice-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .turn-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 6px;
  }
  .turn-badges {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
  }
  .turn-guide-badge {
    font-size: 12px;
    font-weight: 700;
    color: var(--text-main);
  }
  .item-segment {
    font-family: monospace;
    font-size: 10px;
    color: var(--text-dim);
  }
  .turn-source-badge {
    font-size: 9px;
    font-weight: 700;
    padding: 2px 6px;
    border-radius: 3px;
    letter-spacing: 0.4px;
    border: 1px solid var(--border-color);
    color: var(--text-muted);
    background: var(--bg-surface);
  }
  .turn-source-badge.is-generated {
    border-color: var(--accent-white);
    color: var(--text-main);
  }
  .turn-reason-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 10px;
    color: var(--text-dim);
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 5px 9px;
  }
  .turn-reason-bar strong {
    color: var(--text-muted);
  }
  .turn-title {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.4px;
    color: var(--text-muted);
  }
  .item-prompt {
    font-size: 13px;
    font-weight: 600;
    line-height: 1.45;
    color: var(--text-main);
  }
  .practice-options {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .option-pill {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 10px 12px;
    font-size: 12px;
    cursor: pointer;
    transition: all 0.1s ease;
    color: var(--text-muted);
  }
  .option-pill:hover {
    border-color: var(--border-focus);
    color: var(--text-main);
  }
  .option-pill.option-selected {
    border-color: var(--accent-white);
    background: var(--selection-bg);
    color: var(--text-main);
  }
  .option-pill.option-correct {
    border-color: var(--accent-white);
    background: rgba(255, 255, 255, 0.12);
    color: var(--text-main);
    font-weight: 600;
  }
  .option-pill.option-wrong {
    border-color: var(--border-focus);
    opacity: 0.65;
  }
  .option-text {
    flex: 1;
    line-height: 1.4;
  }
  .option-mark {
    font-weight: 800;
    font-size: 12px;
    color: var(--text-main);
  }
  .practice-actions-row {
    display: flex;
    gap: 8px;
  }
  .btn-submit-turn {
    flex: 1;
  }
  .result-banner {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 9px;
  }
  .result-banner.result-correct {
    border-color: var(--accent-white);
  }
  .result-header-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .result-title {
    font-size: 12px;
    font-weight: 700;
    color: var(--text-main);
  }
  .result-mastery-pill {
    font-family: monospace;
    font-size: 10px;
    color: var(--text-muted);
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 2px 6px;
    border-radius: 3px;
  }
  .diagnostic-feedback {
    font-size: 11px;
    line-height: 1.45;
    color: var(--text-muted);
    background: var(--bg-card);
    border-left: 2px solid var(--border-focus);
    padding: 7px 10px;
    border-radius: 0 4px 4px 0;
  }
  .diagnostic-feedback.correct-note {
    border-left-color: var(--accent-white);
    color: var(--text-main);
  }
  .remediation-box {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding-top: 6px;
    border-top: 1px dashed var(--border-color);
  }
  .remediation-label {
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.4px;
    color: var(--text-dim);
  }
  .remediation-buttons {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .next-turn-row {
    display: flex;
  }
  .next-turn-btn {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 8px;
    height: 34px;
  }
  .instant-badge {
    font-size: 9px;
    font-weight: 700;
    text-transform: uppercase;
    padding: 1px 5px;
    border-radius: 3px;
    background: rgba(0, 0, 0, 0.25);
    color: var(--text-dark);
  }
  .agent-offline-msg {
    font-size: 11px;
    color: var(--text-dim);
    line-height: 1.4;
  }

  /* Bottom 24px Status Bar */
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
  }
  .statusbar-agent {
    display: flex;
    align-items: center;
    gap: 5px;
  }
  .agent-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--text-dim);
  }
  .agent-dot.online {
    background: var(--accent-white);
    box-shadow: 0 0 6px rgba(255, 255, 255, 0.8);
  }
  .statusbar-label {
    color: var(--text-muted);
    font-weight: 500;
  }
  .statusbar-model-select {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    color: var(--text-main);
    font-size: 10px;
    padding: 0 4px;
    height: 18px;
    cursor: pointer;
  }
  .statusbar-refresh-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    font-size: 10px;
    cursor: pointer;
    padding: 0 3px;
  }
  .statusbar-refresh-btn:hover {
    color: var(--text-main);
  }
  .statusbar-sep {
    color: var(--border-color);
  }
  .statusbar-metric strong {
    color: var(--text-main);
  }
  .statusbar-right {
    display: flex;
    align-items: center;
    gap: 8px;
    overflow: hidden;
  }
  .statusbar-msg {
    color: var(--text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 460px;
  }

  /* Command Palette Modal */
  .palette-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.65);
    backdrop-filter: blur(4px);
    z-index: 1000;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    padding-top: 15vh;
  }
  .palette-dialog {
    width: 580px;
    max-width: 90vw;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.9);
    display: flex;
    flex-direction: column;
  }
  .palette-input-wrap {
    display: flex;
    align-items: center;
    padding: 12px 14px;
    border-bottom: 1px solid var(--border-color);
    gap: 10px;
  }
  .palette-icon {
    font-size: 14px;
    color: var(--text-muted);
  }
  .palette-input {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    font-size: 14px;
    color: var(--text-main);
  }
  .palette-esc {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    font-size: 9px;
    padding: 2px 5px;
    color: var(--text-dim);
  }
  .palette-results {
    max-height: 340px;
    overflow-y: auto;
    padding: 6px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .palette-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 10px;
    background: transparent;
    border: none;
    border-radius: 5px;
    color: var(--text-muted);
    font-size: 13px;
    cursor: pointer;
    text-align: left;
  }
  .palette-item:hover,
  .palette-item.selected {
    background: var(--selection-bg);
    color: var(--text-main);
  }
  .palette-item-main {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .palette-category-badge {
    font-size: 9px;
    text-transform: uppercase;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 1px 4px;
    border-radius: 3px;
    color: var(--text-dim);
  }
  .palette-label {
    font-weight: 500;
  }
  .palette-detail {
    font-size: 11px;
    color: var(--text-dim);
    font-family: monospace;
  }

  /* Welcome Dashboard */
  .welcome-dashboard {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: center;
    background: var(--bg-primary);
  }
  .welcome-card {
    width: 520px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 32px;
    display: flex;
    flex-direction: column;
    gap: 24px;
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.6);
  }
  .welcome-brand {
    text-align: center;
  }
  .brand-badge-lg {
    background: var(--accent-white);
    color: var(--text-dark);
    font-weight: 800;
    font-size: 11px;
    padding: 3px 8px;
    border-radius: 4px;
    letter-spacing: 0.8px;
  }
  .welcome-brand h2 {
    margin: 12px 0 6px 0;
    font-size: 20px;
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
    gap: 10px;
  }
  .welcome-btn {
    height: 38px;
    font-size: 13px;
  }
  .welcome-recents h3 {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--text-dim);
    letter-spacing: 0.5px;
    margin: 0 0 8px 0;
  }
  .welcome-recents-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
    max-height: 140px;
    overflow-y: auto;
  }
  .recent-link {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 8px 12px;
    border-radius: 5px;
    color: var(--text-main);
    cursor: pointer;
    font-size: 12px;
    text-align: left;
  }
  .recent-link:hover {
    background: var(--bg-card-hover);
    border-color: var(--border-focus);
  }
  .recent-path {
    font-size: 10px;
    color: var(--text-dim);
  }
  .welcome-shortcuts {
    display: flex;
    justify-content: space-between;
    border-top: 1px solid var(--border-color);
    padding-top: 16px;
  }
  .shortcut-tag {
    font-size: 10px;
    color: var(--text-dim);
    display: flex;
    align-items: center;
    gap: 5px;
  }
  .shortcut-tag kbd {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 1px 4px;
    color: var(--text-muted);
  }

  /* Minimalist High-Contrast Buttons */
  .btn {
    height: 30px;
    padding: 0 10px;
    border-radius: 5px;
    border: none;
    font-weight: 600;
    font-size: 11px;
    cursor: pointer;
    transition: all 0.12s ease;
  }
  .btn-sm {
    height: 24px;
    padding: 0 8px;
    font-size: 10px;
  }
  .btn-primary {
    background: var(--accent-white);
    color: var(--text-dark);
  }
  .btn-primary:hover {
    background: var(--accent-silver);
  }
  .btn-primary:disabled {
    background: var(--border-color);
    color: var(--text-dim);
    cursor: not-allowed;
  }
  .btn-secondary {
    background: var(--bg-card);
    color: var(--text-muted);
    border: 1px solid var(--border-color);
  }
  .btn-secondary:hover {
    border-color: var(--border-focus);
    color: var(--text-main);
  }
  .btn-block {
    width: 100%;
    margin-top: 4px;
  }

  .empty-state {
    padding: 30px 16px;
    text-align: center;
    color: var(--text-dim);
    font-size: 12px;
    line-height: 1.5;
  }
</style>
