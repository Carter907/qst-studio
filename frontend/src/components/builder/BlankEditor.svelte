<script lang="ts">
  import { studio } from '../../state.svelte';
</script>

<div class="form-group">
  <div class="choices-header">
    <span class="field-label">Blanks Configuration (e.g. [b1], [d1])</span>
    <button type="button" class="btn btn-sm btn-secondary" onclick={() => studio.addBlank()}>+ Blank</button>
  </div>

  <div class="blanks-list">
    {#each studio.itemBlanks as blank, bIdx}
      <div class="blank-editor-card">
        <div class="blank-card-top">
          <div class="blank-token-input-wrap">
            <span class="field-label">Token:</span>
            <input
              type="text"
              bind:value={blank.token}
              class="input-field blank-token-input"
              placeholder="b1"
            />
          </div>
          <select bind:value={blank.type} class="input-field blank-type-select">
            <option value="text">Text Entry</option>
            <option value="dropdown">Dropdown</option>
          </select>
          <button
            type="button"
            class="btn-delete"
            disabled={studio.itemBlanks.length <= 1}
            onclick={() => studio.removeBlank(bIdx)}
          >Remove</button>
        </div>

        {#if blank.type === 'dropdown'}
          <div class="blank-dropdown-editor">
            <div class="choices-header">
              <span class="field-label">Options:</span>
              <button type="button" class="btn btn-sm btn-secondary" onclick={() => studio.addBlankOption(bIdx)}>+ Option</button>
            </div>
            {#each blank.options as opt, oIdx}
              <div class="dropdown-opt-row">
                <input
                  type="radio"
                  name="blank_correct_{bIdx}"
                  checked={blank.correct === opt && opt !== ''}
                  onchange={() => blank.correct = opt}
                />
                <input
                  type="text"
                  bind:value={blank.options[oIdx]}
                  class="input-field opt-input"
                  placeholder="Option text..."
                  oninput={(e) => {
                    if (blank.correct === opt) {
                      blank.correct = (e.target as HTMLInputElement).value;
                    }
                  }}
                />
                <button
                  type="button"
                  class="btn-delete"
                  disabled={blank.options.length <= 2}
                  onclick={() => studio.removeBlankOption(bIdx, oIdx)}
                >Remove</button>
              </div>
            {/each}
          </div>
        {:else}
          <div class="blank-text-editor">
            <div class="choices-header">
              <span class="field-label">Acceptable Answers:</span>
              <button type="button" class="btn btn-sm btn-secondary" onclick={() => studio.addBlankAnswer(bIdx)}>+ Answer</button>
            </div>
            {#each blank.answers as ans, aIdx}
              <div class="text-ans-row">
                <input
                  type="text"
                  bind:value={blank.answers[aIdx]}
                  class="input-field"
                  placeholder="Acceptable string..."
                />
                <button
                  type="button"
                  class="btn-delete"
                  disabled={blank.answers.length <= 1}
                  onclick={() => studio.removeBlankAnswer(bIdx, aIdx)}
                >Remove</button>
              </div>
            {/each}
          </div>
        {/if}
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
  .blanks-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .blank-editor-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    padding: 6px 8px;
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .blank-card-top {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .blank-token-input-wrap {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .blank-token-input {
    width: 50px;
    font-family: monospace;
  }
  .blank-type-select {
    font-size: 10px;
    padding: 3px 6px;
  }
  .blank-dropdown-editor, .blank-text-editor {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding-top: 4px;
    border-top: 1px dashed var(--border-color);
  }
  .dropdown-opt-row, .text-ans-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .opt-input {
    flex: 1;
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
