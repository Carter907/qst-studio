<script lang="ts">
  import { studio } from '../../state.svelte';
  import { renderMarkdownWithLatex } from '../../markdown';
  import PracticeChoiceCard from './PracticeChoiceCard.svelte';
  import PracticeBlankCard from './PracticeBlankCard.svelte';
  import PracticeMatchingCard from './PracticeMatchingCard.svelte';
  import PracticeOrderingCard from './PracticeOrderingCard.svelte';
  import PracticeEssayCard from './PracticeEssayCard.svelte';
</script>

<div class="practice-runner">
  <div class="deck-header">
    <div class="deck-scope-toggle">
      <button
        type="button"
        class="scope-btn {studio.practiceScope === 'guide' ? 'active' : ''}"
        onclick={() => studio.buildPracticeDeck('guide')}
        title="Practice questions in the anchored guide"
      >
        Guide
      </button>
      <button
        type="button"
        class="scope-btn {studio.practiceScope === 'subtree' ? 'active' : ''}"
        onclick={() => studio.buildPracticeDeck('subtree')}
        title="Practice recursive sub-guides bottom-up to this guide"
      >
        Subtree
      </button>
      <button
        type="button"
        class="scope-btn {studio.practiceScope === 'learning_path' ? 'active' : ''}"
        onclick={() => studio.buildPracticeDeck('learning_path')}
        title="Practice upstream prerequisites in topological order to this guide"
      >
        Path
      </button>
      {#if studio.graph?.tours && studio.graph.tours.length > 0}
        <button
          type="button"
          class="scope-btn {studio.practiceScope === 'tour' ? 'active' : ''}"
          onclick={() => studio.buildPracticeDeck('tour')}
          title="Practice guides along a manifest tour"
        >
          Tour
        </button>
      {/if}
      <button
        type="button"
        class="scope-btn {studio.practiceScope === 'graph' ? 'active' : ''}"
        onclick={() => studio.buildPracticeDeck('graph')}
        title="Practice all questions across the entire knowledge graph"
      >
        Graph
      </button>
    </div>

    {#if (studio.practiceScope === 'subtree' || studio.practiceScope === 'learning_path') && studio.practiceAnchorGuideId}
      <div class="deck-anchor-label">
        {studio.practiceScope === 'subtree' ? `Subtree: ${studio.practiceAnchorGuideId}` : `Path → ${studio.practiceAnchorGuideId}`}
      </div>
    {/if}

    {#if studio.practiceScope === 'tour' && studio.graph?.tours && studio.graph.tours.length > 0}
      <select
        class="tour-select"
        value={studio.practiceTourName}
        onchange={(e) => studio.selectPracticeTour(e.currentTarget.value)}
        aria-label="Select tour"
      >
        {#each studio.graph.tours as tour}
          <option value={tour.name}>{tour.name} ({tour.guides.length})</option>
        {/each}
      </select>
    {/if}

    {#if studio.graph?.scopes && studio.graph.scopes.length > 0}
      <div class="scope-filter-row" aria-label="Filter by scope level">
        <button
          type="button"
          class="scope-chip {studio.practiceFilterScopes.length === 0 ? 'active' : ''}"
          onclick={() => studio.clearPracticeScopeFilters()}
        >
          All
        </button>
        {#each studio.graph.scopes as sc}
          <button
            type="button"
            class="scope-chip {studio.practiceFilterScopes.includes(sc) ? 'active' : ''}"
            onclick={() => studio.togglePracticeScopeFilter(sc)}
            title="Toggle {sc} scope filter"
          >
            {sc}
          </button>
        {/each}
      </div>
    {/if}

    <div class="deck-status-row">
      {#if studio.practiceDeck.length > 0}
        <span class="card-counter">
          Card {studio.practiceIndex + 1} of {studio.practiceDeck.length}
        </span>
        {#if studio.practiceSessionScore.total > 0}
          <span class="score-badge">
            Score: {studio.practiceSessionScore.correct}/{studio.practiceSessionScore.total}
          </span>
        {/if}
        <button
          type="button"
          class="btn-text-action"
          onclick={() => studio.shufflePracticeDeck()}
          title="Shuffle deck"
        >
          Shuffle
        </button>
      {/if}
    </div>
  </div>

  {#if studio.practiceDeck.length === 0}
    <div class="empty-practice-state">
      <p>
        {studio.practiceFilterScopes.length > 0
          ? 'No questions match the selected scope filters.'
          : 'No questions found in this scope.'}
      </p>
      {#if studio.practiceFilterScopes.length > 0}
        <button
          type="button"
          class="btn btn-secondary btn-sm"
          onclick={() => studio.clearPracticeScopeFilters()}
        >
          Clear Scope Filters
        </button>
      {/if}
      <button
        type="button"
        class="btn btn-secondary btn-sm"
        onclick={() => studio.activeRightTab = 'questions'}
      >
        + Author Question
      </button>
    </div>
  {:else}
    {@const card = studio.practiceDeck[studio.practiceIndex]}
    {@const qType = card.type || 'multiple_choice'}

    <div class="practice-card-body">
      <div class="card-meta-top">
        <span class="type-pill">{qType.replace(/_/g, ' ')}</span>
        {#if card.segment}
          <span class="segment-pill">Lines {card.segment}</span>
        {/if}
        {#if studio.practiceScope !== 'guide'}
          <span class="guide-tag">{card.guide_id}</span>
        {/if}
      </div>

      {#if card.title}
        <h4 class="card-title">{card.title}</h4>
      {/if}
      <div class="card-prompt qst-prose">
        {@html renderMarkdownWithLatex(card.prompt)}
      </div>

      <div class="interaction-container">
        {#if qType === 'multiple_choice' || qType === 'true_false' || qType === 'multiple_answers'}
          <PracticeChoiceCard {card} {qType} answer={studio.practiceAnswer} checked={studio.practiceChecked} />
        {:else if qType === 'fill_in_the_blank'}
          <PracticeBlankCard {card} answer={studio.practiceAnswer} checked={studio.practiceChecked} />
        {:else if qType === 'matching'}
          <PracticeMatchingCard {card} answer={studio.practiceAnswer} checked={studio.practiceChecked} />
        {:else if qType === 'ordering'}
          <PracticeOrderingCard {card} answer={studio.practiceAnswer} checked={studio.practiceChecked} result={studio.practiceResult} />
        {:else if qType === 'essay'}
          <PracticeEssayCard
            {card}
            answer={studio.practiceAnswer}
            checked={studio.practiceChecked}
            onSelfGrade={(understood) => studio.selfGradeEssay(understood)}
          />
        {/if}
      </div>

      {#if studio.practiceResult}
        <div class="result-banner {studio.practiceResult.isCorrect ? 'result-correct' : 'result-wrong'}">
          <div class="result-title">
            {studio.practiceResult.isCorrect ? 'Correct' : 'Incorrect'}
          </div>
          {#if studio.practiceResult.feedback}
            <div class="result-feedback qst-prose">
              {@html renderMarkdownWithLatex(studio.practiceResult.feedback, true)}
            </div>
          {/if}
        </div>
      {/if}

      <div class="practice-actions-row">
        {#if !studio.practiceChecked}
          {#if qType === 'essay'}
            {#if !studio.practiceAnswer.revealed}
              <button
                type="button"
                class="btn btn-primary btn-block"
                onclick={() => studio.practiceAnswer.revealed = true}
              >
                Reveal Solution & Self-Grade
              </button>
            {/if}
          {:else}
            <button
              type="button"
              class="btn btn-primary btn-block"
              onclick={() => studio.submitPracticeAnswer()}
            >
              Check Answer
            </button>
          {/if}
        {:else}
          <div class="next-nav-group">
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              disabled={studio.practiceIndex === 0}
              onclick={() => studio.prevPracticeCard()}
            >
              Previous
            </button>
            <button
              type="button"
              class="btn btn-primary btn-sm btn-next"
              onclick={() => {
                if (studio.practiceIndex < studio.practiceDeck.length - 1) {
                  studio.nextPracticeCard();
                } else {
                  studio.shufflePracticeDeck();
                }
              }}
            >
              {studio.practiceIndex < studio.practiceDeck.length - 1 ? 'Next Question →' : 'Restart Deck ↻'}
            </button>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .practice-runner {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
  }
  .deck-header {
    padding: 8px 10px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-card);
    display: flex;
    flex-direction: column;
    gap: 6px;
    flex-shrink: 0;
  }
  .deck-scope-toggle {
    display: flex;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 2px;
  }
  .scope-btn {
    flex: 1;
    background: none;
    border: none;
    padding: 4px 6px;
    font-size: 10px;
    font-weight: 500;
    color: var(--text-dim);
    cursor: pointer;
    border-radius: 3px;
    transition: all 0.1s ease;
  }
  .scope-btn.active {
    background: var(--accent-white);
    color: var(--text-dark);
    font-weight: 600;
  }
  .deck-anchor-label {
    font-size: 10px;
    color: var(--text-muted);
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .tour-select {
    width: 100%;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 4px 6px;
    font-size: 11px;
    color: var(--text-main);
  }
  .tour-select:focus {
    outline: none;
    border-color: var(--border-focus);
  }
  .scope-filter-row {
    display: flex;
    flex-wrap: wrap;
    gap: 3px;
  }
  .scope-chip {
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: 3px;
    color: var(--text-dim);
    font-size: 9px;
    text-transform: capitalize;
    padding: 1px 5px;
    cursor: pointer;
    transition: all 0.1s ease;
  }
  .scope-chip:hover {
    color: var(--text-main);
    border-color: var(--border-focus);
  }
  .scope-chip.active {
    background: var(--accent-white);
    color: var(--text-dark);
    border-color: var(--accent-white);
    font-weight: 600;
  }
  .deck-status-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 10px;
    color: var(--text-dim);
  }
  .card-counter {
    font-weight: 600;
    color: var(--text-muted);
  }
  .score-badge {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    padding: 1px 5px;
    border-radius: 3px;
    font-weight: 600;
    color: var(--text-main);
  }
  .btn-text-action {
    background: none;
    border: none;
    color: var(--text-dim);
    font-size: 10px;
    cursor: pointer;
    text-decoration: underline;
  }
  .btn-text-action:hover {
    color: var(--text-main);
  }
  .empty-practice-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    padding: 20px;
    text-align: center;
    gap: 10px;
    color: var(--text-dim);
    font-size: 12px;
  }
  .practice-card-body {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .card-meta-top {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .type-pill {
    font-size: 8px;
    font-weight: 700;
    text-transform: uppercase;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 1px 5px;
    border-radius: 2px;
    color: var(--text-muted);
  }
  .segment-pill {
    font-size: 9px;
    color: var(--text-dim);
    font-family: monospace;
  }
  .guide-tag {
    font-size: 9px;
    color: var(--accent-white);
    margin-left: auto;
  }
  .card-title {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-main);
  }
  .card-prompt {
    font-size: 12px;
    line-height: 1.5;
  }
  .interaction-container {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .result-banner {
    padding: 8px 10px;
    border-radius: 4px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .result-correct {
    background: rgba(63, 185, 80, 0.12);
    border: 1px solid rgba(63, 185, 80, 0.4);
    color: #3fb950;
  }
  .result-wrong {
    background: rgba(248, 81, 73, 0.12);
    border: 1px solid rgba(248, 81, 73, 0.4);
    color: #f85149;
  }
  .result-title {
    font-weight: 700;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .result-feedback {
    font-size: 11px;
    color: var(--text-main);
  }
  .practice-actions-row {
    margin-top: auto;
    padding-top: 8px;
  }
  .btn-block {
    width: 100%;
    padding: 8px;
    font-size: 12px;
    font-weight: 600;
  }
  .next-nav-group {
    display: flex;
    gap: 8px;
  }
  .btn-next {
    flex: 1;
  }
</style>
