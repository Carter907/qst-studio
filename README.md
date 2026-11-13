# Quest Studio (`qst-studio`)

A lightweight, high-performance desktop studio for browsing **Quest Knowledge Graphs**, **Authoring Line-Anchored Questions**, and **Practicing Question Banks**, built with **Wails (Go)** and **Svelte 5 (TypeScript)**.

---

## Features

1. **Knowledge Graph Exploration (`qst`)**:
   - Parses `qst.yaml`, `manifest.yaml`, or `quest.yaml` manifests and tiered Markdown guides with YAML frontmatter.
   - Respects scope ladders (`lesson` → `explanation` → `description` → `definition`).
   - Resolves prerequisite DAGs and topological learning paths.
2. **Rendered Guide Viewer with KaTeX Math**:
   - High-fidelity Markdown and LaTeX rendering ($inline$ and $$display$$ math).
   - Clickable line-range badges (`L#`) for direct segment anchoring.
3. **Line-Anchored Question Authoring**:
   - Interactive visual authoring of Multiple Choice, Multiple Answers, True/False, Fill in the Blank (Text & Dropdown), Matching, Ordering, and Essay questions.
   - Live Markdown and LaTeX question preview.
   - Declarative persistence to `items.yaml`.
4. **Interactive Practice Mode**:
   - Practice cards scoped to the current guide or the entire knowledge graph, with local attempt history tracking (`.practice_history.json`).
5. **Clean, Fast, Minimal Baseline**:
   - No heavy 3D rendering engines or external LLM daemon dependencies.

---

## Quickstart

### Prerequisites
- Go 1.27+
- Node.js 18+ and npm
- Linux: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev` (configured automatically via `"build:tags": "webkit2_41"` in `wails.json`)

### Development Mode (Live Reload)
```bash
wails dev
```
You can open `examples/intro-graph` using the folder picker to explore a sample knowledge graph right away.

### Production Build
```bash
wails build
# Output executable:
./build/bin/qst-studio
```

### Running Tests & Checks
```bash
go test -v ./...
cd frontend && npm run check && npm run build
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for architecture details and pull request guidelines.

---

## Project Structure

```text
qst-studio/
├── main.go                     # Wails window bootstrap and embedded frontend assets
├── app.go                      # Thin Wails RPC adapter delegating to internal/
├── quiz_test.go                # Integration tests for App RPCs
├── wails.json                  # Wails configuration with webkit2_41 tags and build pipelines
├── examples/intro-graph/       # Sample Quest knowledge graph for local testing
├── internal/
│   ├── graph/                  # Manifest & guide parser, topological pathfinder
│   ├── practice/               # Practice history JSON store & attempt recording
│   ├── quiz/                   # Quiz item models, validation, and items.yaml persistence
│   └── workspace/              # Recent workspaces persistence
└── frontend/
    ├── src/
    │   ├── App.svelte          # Root layout shell, welcome dashboard, and global shortcuts
    │   ├── state.svelte.ts     # Shared Svelte 5 reactive state & Wails RPC actions
    │   ├── markdown.ts         # Marked + KaTeX rendering and line-mapped block parser
    │   ├── style.css           # Global theme variables, typography, and shared UI primitives
    │   ├── stores/             # Pure TypeScript types, builder validation, and practice grading
    │   └── components/
    │       ├── layout/         # TopBar, StatusBar, CommandPalette
    │       ├── guide/          # GuideSidebar, GuideViewer, NodeInspector
    │       ├── builder/        # QuestionBuilder, BuilderPreview, SavedItemsList, type editors
    │       └── practice/       # PracticeRunner and per-type practice cards
    └── wailsjs/                # Auto-generated Go-to-TypeScript bindings
```

---

## License

Licensed under the **GNU General Public License v3.0** (`GPL-3.0-or-later`). See [LICENSE](LICENSE) for the full text.
