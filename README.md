# Quest Studio (`qst-studio`)

A high-performance desktop studio for **Knowledge Graphs**, **Canvas-style Question Authoring**, and **Continual Practice (BKT + FSRS)**, built with **Wails (Go)** and **Svelte 5 (TypeScript)**.

---

## Features

1. **Knowledge Graph Exploration (`qst`)**:
   - Parses `manifest.yaml` and tiered Markdown guides with YAML frontmatter.
   - Respects scope ladders (`lesson` → `explanation` → `description` → `definition`).
   - Resolves horizontal prerequisite DAGs and topological learning paths.
2. **Line-Level Segment Anchoring**:
   - Gutter line selection directly maps questions to body line ranges (`segment: "start-end"`).
3. **Canvas-Style Question Studio**:
   - Interactive visual authoring of Multiple Choice, Distractors, Correct Answers, and Feedback.
   - Declarative persistence to `items.yaml` (ready for compilation into standard IMS QTI packages).
4. **Adaptive Continual Practice Engine**:
   - **Bayesian Knowledge Tracing (BKT)**: Tracks latent mastery probability $P(L)$ across learning events.
   - **FSRS Spaced Repetition**: Calculates memory retrievability $R(t) = \exp(-t / S)$ and flags when nodes are due for review.
   - State persisted locally in `.learner_state.json`.

---

## Quickstart

### Prerequisites
- Go 1.22+
- Node.js 18+ and npm
- Linux: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev` (configured automatically via `"build:tags": "webkit2_41"` in `wails.json`)

### Development Mode (Live Reload)
```bash
cd /home/speer/Code/golang/qst-studio
wails dev
```

### Production Build
```bash
wails build
# Output executable:
./build/bin/qst-studio
```

---

## Project Structure

```text
qst-studio/
├── app.go              # Go Backend: Wails RPC bindings for Graph, Items, BKT & FSRS
├── graph.go            # qst parser (manifest.yaml, guide frontmatter, prerequisite pathfinding)
├── main.go             # Wails window initialization and startup hooks
├── wails.json          # Configuration with webkit2_41 tags and build pipelines
└── frontend/
    ├── src/
    │   ├── App.svelte  # Three-pane Studio: Navigator, Guide Viewer, & Canvas Builder/Runner
    │   ├── main.ts     # Svelte mounting
    │   └── style.css   # Dark theme CSS variables
    └── wailsjs/        # Auto-generated Go-to-TypeScript type definitions
```
