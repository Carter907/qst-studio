<script lang="ts">
  import { studio } from '../../state.svelte';
  import { getQuizRecordKey } from '../../stores/types';
  import { renderMarkdownWithLatex } from '../../markdown';
  import PracticeChoiceCard from '../practice/PracticeChoiceCard.svelte';
  import PracticeBlankCard from '../practice/PracticeBlankCard.svelte';
  import PracticeMatchingCard from '../practice/PracticeMatchingCard.svelte';
  import PracticeOrderingCard from '../practice/PracticeOrderingCard.svelte';
  import PracticeEssayCard from '../practice/PracticeEssayCard.svelte';

  const sampleQuizYaml = `title: Pointer Arithmetic Drills
description: Focused math and offset practice problems.
tags:
  - math
  - pointers
related_guides:
  - pointers
items:
  - id: QZ_PTR_01
    title: Offset address
    type: fill_in_the_blank
    prompt: "If \`int *p = 0x1000\` and \`sizeof(int) = 4\`, what is \`p + 3\` in hex? [b1]"
    blanks:
      b1:
        type: text
        answers: ["0x100c", "0x100C"]`;

  function guideExists(guideId: string): boolean {
    return Boolean(studio.graph?.guides.some(g => g.id === guideId));
  }

  function getMasteredCount(): number {
    if (!studio.selectedQuiz) return 0;
    let mastered = 0;
    for (const item of studio.selectedQuiz.items) {
      const rec = studio.quizHistory[getQuizRecordKey(studio.selectedQuiz.id, item.id)];
      if (rec && rec.last_correct) mastered++;
    }
    return mastered;
  }
</script>

<main class="quiz-runner-stage">
  {#if studio.standaloneQuizzesError}
    <div class="quiz-empty-state">
      <div class="quiz-error-box">
        <h3>Failed to Load Standalone Quizzes</h3>
        <p class="quiz-error-msg">{studio.standaloneQuizzesError}</p>
        <button
          type="button"
          class="btn btn-primary"
          onclick={() => studio.reloadStandaloneQuizzes()}
        >
          Reload Quizzes
        </button>
      </div>
    </div>
  {:else if studio.standaloneQuizzes.length === 0}
    <div class="quiz-empty-state">
      <div class="quiz-onboarding-card">
        <h3>No Standalone Quizzes Found</h3>
        <p>
          Add YAML quiz files inside <code>quizzes/</code> in your workspace directory to practice focused problem sets without anchoring them to guide line segments.
        </p>
        <pre class="sample-yaml-block">{sampleQuizYaml}</pre>
        <div class="onboarding-actions">
          <button
            type="button"
            class="btn btn-primary"
            onclick={() => studio.reloadStandaloneQuizzes()}
          >
            Reload Quizzes
          </button>
          <button
            type="button"
            class="btn btn-secondary"
            onclick={() => studio.setWorkspaceMode('graph')}
          >
            Back to Graph
          </button>
        </div>
      </div>
    </div>
  {:else if !studio.selectedQuiz}
    <div class="quiz-empty-state">
      <p>Select a standalone quiz from the left sidebar to begin.</p>
    </div>
  {:else if studio.quizStage === 'overview'}
    {@const totalItems = studio.selectedQuiz.items.length}
    {@const masteredCount = getMasteredCount()}
    {@const allCount = studio.getFilteredQuizItems('all').length}
    {@const unseenCount = studio.getFilteredQuizItems('unseen').length}
    {@const reviewCount = studio.getFilteredQuizItems('needs_review').length}
    {@const selectedCount = studio.getFilteredQuizItems(studio.quizFilter).length}

    <div class="quiz-centered-shell">
      <div class="overview-card">
        <div class="overview-header">
          <div class="overview-title-wrap">
            <h2 class="overview-title">{studio.selectedQuiz.title}</h2>
            {#if studio.selectedQuiz.description}
              <div class="overview-desc qst-prose">
                {@html renderMarkdownWithLatex(studio.selectedQuiz.description)}
              </div>
            {/if}
          </div>
          <span class="quiz-total-badge">{totalItems} questions</span>
        </div>

        {#if totalItems > 0}
          <div class="overview-mastery">
            <div class="mastery-labels">
              <span class="field-label">Mastery</span>
              <span class="mastery-count">{masteredCount} / {totalItems} correct</span>
            </div>
            <div class="mastery-track">
              <div
                class="mastery-fill"
                style="width: {Math.round((masteredCount / totalItems) * 100)}%"
              ></div>
            </div>
          </div>
        {/if}

        {#if (studio.selectedQuiz.tags && studio.selectedQuiz.tags.length > 0) || (studio.selectedQuiz.related_guides && studio.selectedQuiz.related_guides.length > 0)}
          <div class="overview-meta">
            {#if studio.selectedQuiz.tags && studio.selectedQuiz.tags.length > 0}
              <div class="meta-group">
                <span class="field-label">Tags</span>
                {#each studio.selectedQuiz.tags as tag}
                  <span class="meta-chip">{tag}</span>
                {/each}
              </div>
            {/if}

            {#if studio.selectedQuiz.related_guides && studio.selectedQuiz.related_guides.length > 0}
              <div class="meta-group">
                <span class="field-label">Reference Guides</span>
                {#each studio.selectedQuiz.related_guides as gid}
                  {@const exists = guideExists(gid)}
                  <button
                    type="button"
                    class="meta-chip {exists ? 'clickable-chip' : 'missing-guide-chip'}"
                    disabled={!exists}
                    onclick={() => studio.jumpToRelatedGuide(gid)}
                    title={exists ? `Open guide "${gid}" in Graph mode` : `Guide "${gid}" not found in workspace`}
                  >
                    {gid} {exists ? '→' : '(missing)'}
                  </button>
                {/each}
              </div>
            {/if}
          </div>
        {/if}

        <div class="overview-setup">
          <div class="setup-row">
            <span class="field-label">Questions</span>
            <div class="filter-group">
              <button
                type="button"
                class="filter-pill {studio.quizFilter === 'all' ? 'active' : ''}"
                onclick={() => studio.quizFilter = 'all'}
              >
                All ({allCount})
              </button>
              <button
                type="button"
                class="filter-pill {studio.quizFilter === 'unseen' ? 'active' : ''}"
                onclick={() => studio.quizFilter = 'unseen'}
              >
                Unseen ({unseenCount})
              </button>
              <button
                type="button"
                class="filter-pill {studio.quizFilter === 'needs_review' ? 'active' : ''}"
                onclick={() => studio.quizFilter = 'needs_review'}
              >
                Needs Review ({reviewCount})
              </button>
            </div>
          </div>

          <div class="setup-row">
            <span class="field-label">Order</span>
            <div class="filter-group">
              <button
                type="button"
                class="filter-pill {!studio.quizShuffled ? 'active' : ''}"
                onclick={() => studio.quizShuffled = false}
              >
                Sequential
              </button>
              <button
                type="button"
                class="filter-pill {studio.quizShuffled ? 'active' : ''}"
                onclick={() => studio.quizShuffled = true}
              >
                Shuffled
              </button>
            </div>
          </div>
        </div>

        <div class="overview-start-row">
          <button
            type="button"
            class="btn btn-primary start-quiz-btn"
            disabled={selectedCount === 0}
            onclick={() => studio.startQuizRun()}
          >
            {selectedCount === 0
              ? 'No Matching Questions'
              : `Start Quiz (${selectedCount} question${selectedCount === 1 ? '' : 's'})`}
          </button>
        </div>
      </div>
    </div>
  {:else if studio.quizStage === 'running'}
    {@const card = studio.quizDeck[studio.quizIndex]}
    {@const total = studio.quizDeck.length}
    {@const pct = total > 0 ? Math.round(((studio.quizIndex + 1) / total) * 100) : 0}

    <div class="quiz-centered-shell">
      {#if card}
        {@const qType = card.type || 'multiple_choice'}
        <div class="runner-frame">
          <!-- Slim Distraction-Free Runner Header -->
          <div class="runner-slim-bar">
            <div class="runner-bar-left">
              <span class="runner-progress-text">Question {studio.quizIndex + 1} of {total}</span>
              <span class="type-pill">{qType.replace(/_/g, ' ')}</span>
            </div>
            <div class="runner-bar-right">
              {#if studio.quizSessionScore.total > 0}
                <span class="score-badge">
                  Score: {studio.quizSessionScore.correct}/{studio.quizSessionScore.total}
                </span>
              {/if}
              <button
                type="button"
                class="btn-exit-run"
                onclick={() => studio.exitQuizRun()}
                title="Exit back to quiz overview"
              >
                Exit
              </button>
            </div>
          </div>

          <div class="runner-progress-track">
            <div class="runner-progress-fill" style="width: {pct}%"></div>
          </div>

          <!-- Focused Question Card -->
          <div class="quiz-card-panel">
            {#if card.title}
              <h3 class="card-title">{card.title}</h3>
            {/if}

            <div class="card-prompt qst-prose">
              {@html renderMarkdownWithLatex(card.prompt)}
            </div>

            <div class="interaction-container">
              {#if qType === 'multiple_choice' || qType === 'true_false' || qType === 'multiple_answers'}
                <PracticeChoiceCard {card} {qType} answer={studio.quizAnswer} checked={studio.quizChecked} />
              {:else if qType === 'fill_in_the_blank'}
                <PracticeBlankCard {card} answer={studio.quizAnswer} checked={studio.quizChecked} />
              {:else if qType === 'matching'}
                <PracticeMatchingCard {card} answer={studio.quizAnswer} checked={studio.quizChecked} />
              {:else if qType === 'ordering'}
                <PracticeOrderingCard {card} answer={studio.quizAnswer} checked={studio.quizChecked} result={studio.quizResult} />
              {:else if qType === 'essay'}
                <PracticeEssayCard
                  {card}
                  answer={studio.quizAnswer}
                  checked={studio.quizChecked}
                  onSelfGrade={(understood) => studio.selfGradeQuizEssay(understood)}
                />
              {/if}
            </div>

            {#if studio.quizResult}
              <div class="result-banner {studio.quizResult.isCorrect ? 'result-correct' : 'result-wrong'}">
                <div class="result-title">
                  {studio.quizResult.isCorrect ? 'Correct' : 'Incorrect'}
                </div>
                {#if studio.quizResult.feedback}
                  <div class="result-feedback qst-prose">
                    {@html renderMarkdownWithLatex(studio.quizResult.feedback, true)}
                  </div>
                {/if}
              </div>
            {/if}

            <div class="quiz-actions-row">
              {#if !studio.quizChecked}
                {#if qType === 'essay'}
                  {#if !studio.quizAnswer.revealed}
                    <div class="nav-action-pair">
                      {#if studio.quizIndex > 0}
                        <button
                          type="button"
                          class="btn btn-secondary btn-nav"
                          onclick={() => studio.prevQuizCard()}
                        >
                          ← Previous
                        </button>
                      {/if}
                      <button
                        type="button"
                        class="btn btn-primary btn-next"
                        onclick={() => studio.quizAnswer.revealed = true}
                      >
                        Reveal Solution & Self-Grade
                      </button>
                    </div>
                  {/if}
                {:else}
                  <div class="nav-action-pair">
                    {#if studio.quizIndex > 0}
                      <button
                        type="button"
                        class="btn btn-secondary btn-nav"
                        onclick={() => studio.prevQuizCard()}
                      >
                        ← Previous
                      </button>
                    {/if}
                    <button
                      type="button"
                      class="btn btn-primary btn-next"
                      onclick={() => studio.submitQuizAnswer()}
                    >
                      Check Answer
                    </button>
                  </div>
                {/if}
              {:else}
                <div class="nav-action-pair">
                  <button
                    type="button"
                    class="btn btn-secondary btn-nav"
                    disabled={studio.quizIndex === 0}
                    onclick={() => studio.prevQuizCard()}
                  >
                    ← Previous
                  </button>
                  <button
                    type="button"
                    class="btn btn-primary btn-next"
                    onclick={() => {
                      if (studio.quizIndex < total - 1) {
                        studio.nextQuizCard();
                      } else {
                        studio.finishQuizRun();
                      }
                    }}
                  >
                    {studio.quizIndex < total - 1 ? 'Next Question →' : 'Finish Quiz'}
                  </button>
                </div>
              {/if}
            </div>
          </div>
        </div>
      {/if}
    </div>
  {:else}
    <!-- Quiz Completion Summary -->
    {@const totalAnswered = studio.quizSessionScore.total}
    {@const correctAnswered = studio.quizSessionScore.correct}
    {@const pct = totalAnswered > 0 ? Math.round((correctAnswered / totalAnswered) * 100) : 0}
    {@const missedCount = studio.quizMissedIds.length}

    <div class="quiz-centered-shell">
      <div class="summary-card">
        <div class="summary-header">
          <span class="summary-kicker">Quiz Complete</span>
          <h2 class="summary-title">{studio.selectedQuiz.title}</h2>
        </div>

        <div class="summary-stats-grid">
          <div class="summary-stat">
            <span class="stat-value">{correctAnswered}/{totalAnswered}</span>
            <span class="stat-label">Score</span>
          </div>
          <div class="summary-stat">
            <span class="stat-value">{pct}%</span>
            <span class="stat-label">Accuracy</span>
          </div>
          <div class="summary-stat">
            <span class="stat-value">{missedCount}</span>
            <span class="stat-label">Missed</span>
          </div>
        </div>

        <div class="summary-actions">
          {#if missedCount > 0}
            <button
              type="button"
              class="btn btn-primary summary-btn"
              onclick={() => studio.retryMissedQuestions()}
            >
              Retry Missed ({missedCount})
            </button>
          {/if}
          <button
            type="button"
            class="btn {missedCount === 0 ? 'btn-primary' : 'btn-secondary'} summary-btn"
            onclick={() => studio.startQuizRun()}
          >
            Restart Quiz
          </button>
          <button
            type="button"
            class="btn btn-secondary summary-btn"
            onclick={() => studio.exitQuizRun()}
          >
            Back to Overview
          </button>
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
  .quiz-runner-stage {
    flex: 1;
    height: 100%;
    overflow-y: auto;
    background: var(--bg-primary);
    display: flex;
    flex-direction: column;
  }
  .quiz-empty-state {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 24px;
    color: var(--text-dim);
    font-size: 12px;
  }
  .quiz-onboarding-card,
  .quiz-error-box {
    max-width: 560px;
    width: 100%;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 22px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .quiz-onboarding-card h3,
  .quiz-error-box h3 {
    margin: 0;
    font-size: 15px;
    color: var(--text-main);
  }
  .quiz-onboarding-card p,
  .quiz-error-msg {
    margin: 0;
    font-size: 12px;
    color: var(--text-muted);
    line-height: 1.5;
  }
  .quiz-error-msg {
    color: var(--accent-danger);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 11px;
  }
  .quiz-onboarding-card code {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 1px 4px;
    border-radius: 3px;
    color: var(--text-main);
  }
  .sample-yaml-block {
    margin: 0;
    padding: 10px 12px;
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 11px;
    line-height: 1.45;
    color: var(--text-body);
    overflow-x: auto;
    user-select: text;
  }
  .onboarding-actions {
    display: flex;
    gap: 8px;
  }

  .quiz-centered-shell {
    flex: 1;
    width: 100%;
    max-width: 700px;
    margin: 0 auto;
    padding: 28px 24px 40px 24px;
    display: flex;
    flex-direction: column;
    justify-content: flex-start;
  }

  /* Stage 1: Overview Card */
  .overview-card,
  .summary-card {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 22px 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .overview-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 12px;
  }
  .overview-title-wrap {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }
  .overview-title {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: var(--text-main);
  }
  .overview-desc {
    font-size: 12.5px;
    color: var(--text-muted);
    line-height: 1.5;
  }
  .quiz-total-badge {
    font-size: 10px;
    font-weight: 600;
    color: var(--text-muted);
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 3px 8px;
    border-radius: 4px;
    flex-shrink: 0;
  }
  .overview-mastery {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .mastery-labels {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .mastery-count {
    font-size: 11px;
    color: var(--text-muted);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }
  .overview-meta {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding-top: 12px;
    border-top: 1px solid var(--border-subtle);
  }
  .meta-group {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
  }
  .missing-guide-chip {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .overview-setup {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding-top: 12px;
    border-top: 1px solid var(--border-subtle);
  }
  .setup-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 8px;
  }
  .filter-group {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
  .filter-pill {
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-dim);
    font-size: 11px;
    font-weight: 500;
    padding: 4px 10px;
    cursor: pointer;
    transition: all 0.1s ease;
  }
  .filter-pill:hover {
    color: var(--text-main);
    border-color: var(--border-focus);
  }
  .filter-pill.active {
    background: var(--accent-white);
    color: var(--text-dark);
    border-color: var(--accent-white);
    font-weight: 600;
  }
  .overview-start-row {
    padding-top: 6px;
  }
  .start-quiz-btn {
    width: 100%;
    height: 36px;
    font-size: 12.5px;
  }

  /* Stage 2: Focused Runner */
  .runner-frame {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .runner-slim-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 10px;
  }
  .runner-bar-left,
  .runner-bar-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .runner-progress-text {
    font-size: 11.5px;
    font-weight: 600;
    color: var(--text-muted);
  }
  .type-pill {
    font-size: 9px;
    font-weight: 700;
    text-transform: uppercase;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    padding: 2px 6px;
    border-radius: 3px;
    color: var(--text-dim);
    letter-spacing: 0.4px;
  }
  .score-badge {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    padding: 2px 7px;
    border-radius: 3px;
    font-size: 10.5px;
    font-weight: 600;
    color: var(--text-main);
  }
  .btn-exit-run {
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 3px 9px;
    font-size: 11px;
    font-weight: 500;
    color: var(--text-muted);
    cursor: pointer;
    transition: all 0.12s ease;
  }
  .btn-exit-run:hover {
    border-color: var(--border-focus);
    color: var(--text-main);
    background: var(--bg-surface);
  }
  .runner-progress-track {
    width: 100%;
    height: 3px;
    background: var(--border-color);
    border-radius: 2px;
    overflow: hidden;
  }
  .runner-progress-fill {
    height: 100%;
    background: var(--accent-white);
    transition: width 0.2s ease;
  }

  .quiz-card-panel {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 22px 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .card-title {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    color: var(--text-main);
  }
  .card-prompt {
    font-size: 14px;
    line-height: 1.65;
  }
  .interaction-container {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .result-banner {
    padding: 10px 12px;
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
    font-size: 12px;
    color: var(--text-main);
  }
  .quiz-actions-row {
    padding-top: 4px;
  }
  .nav-action-pair {
    display: flex;
    gap: 8px;
  }
  .btn-nav {
    height: 36px;
    padding: 0 14px;
    font-size: 12px;
  }
  .btn-next {
    flex: 1;
    height: 36px;
    font-size: 12px;
  }

  /* Stage 3: Summary Card */
  .summary-header {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .summary-kicker {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.6px;
    color: var(--text-dim);
  }
  .summary-title {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: var(--text-main);
  }
  .summary-stats-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 10px;
  }
  .summary-stat {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 5px;
    padding: 14px 12px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }
  .stat-value {
    font-size: 20px;
    font-weight: 700;
    color: var(--text-main);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }
  .stat-label {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-dim);
    font-weight: 600;
  }
  .summary-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    padding-top: 4px;
  }
  .summary-btn {
    flex: 1;
    min-width: 140px;
    height: 36px;
    font-size: 12px;
  }
</style>
