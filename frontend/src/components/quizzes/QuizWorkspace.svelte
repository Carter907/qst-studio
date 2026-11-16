<script lang="ts">
  import { studio } from '../../state.svelte';
  import QuizSidebar from './QuizSidebar.svelte';
  import QuizRunner from './QuizRunner.svelte';
</script>

<div class="quiz-workspace-stage {studio.quizSidebarOpen ? 'has-left' : 'no-left'}">
  {#if studio.quizSidebarOpen}
    <QuizSidebar />
  {/if}

  <QuizRunner />
</div>

<style>
  .quiz-workspace-stage {
    flex: 1;
    display: grid;
    overflow: hidden;
  }
  .quiz-workspace-stage.has-left {
    grid-template-columns: 260px 1fr;
  }
  .quiz-workspace-stage.no-left {
    grid-template-columns: 1fr;
  }
</style>
