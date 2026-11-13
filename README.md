# Quest Studio (`qst-studio`)

A fast, lightweight desktop companion to **[`qst`](https://github.com/Carter907/qst)** for **reading** knowledge graphs, **authoring** line-anchored active-recall questions, and **learning** through graph-aware practice—built with **Wails (Go)** and **Svelte 5 (TypeScript)**.

---

## Installation

Download the latest prebuilt binary for your platform from **[GitHub Releases](https://github.com/Carter907/qst-studio/releases)**:

| Platform | Artifact | Notes |
| :--- | :--- | :--- |
| **Linux (`x86_64`)** | `qst-studio-linux-amd64.tar.gz` | Extract and run `./qst-studio`. Requires GTK 3 and WebKitGTK 4.1 (`libgtk-3-0t64` and `libwebkit2gtk-4.1-0` on Ubuntu 24.04+). |
| **Windows (`x86_64`)** | `qst-studio-windows-amd64-installer.exe` or `qst-studio-windows-amd64.zip` | Use the NSIS installer for Start Menu/Desktop shortcuts (includes WebView2 bootstrapper), or extract the portable `.zip`. |
| **macOS (`Universal`)** | `qst-studio-darwin-universal.zip` | Unzip and move `qst-studio.app` to Applications. Unsigned builds may require **Right-click → Open** (or clearing quarantine via `xattr -cr`) on first launch. |

---

## Quickstart: `qst` + `qst-studio` Workflow

`qst-studio` opens any knowledge graph directory created or validated with the [`qst`](https://github.com/Carter907/qst) CLI (or you can try the bundled [`examples/intro-graph`](examples/intro-graph) right away).

1. **Create and validate a graph with `qst`**:
   ```bash
   qst new my-topic
   qst add "Basics" --scope definition --dir my-topic
   qst add "Overview" --scope explanation --dir my-topic
   qst link "Overview" --subguide "Basics" --adherence strict --segment "1-1" --dir my-topic
   qst validate my-topic
   ```
2. **Open the workspace in `qst-studio`**:
   - Click **Open Directory...** (`Ctrl+O`) and select `my-topic` (or `examples/intro-graph`).
   - Browse guides by manifest scope, inspect prerequisite learning paths, and click line numbers in the reader gutter to anchor specific passages.
   - Edit guide `.md` files in your preferred editor and click **Reload** in the viewer to refresh from disk.
3. **Author questions & practice**:
   - Switch to the **Author** tab in the right pane to build line-anchored questions saved to `items.yaml`.
   - Switch to the **Practice** tab to drill cards by **Guide**, **Subtree**, **Learning Path**, **Tour**, or **Whole Graph**.

### Workspace Files

`qst-studio` reads your graph structure directly from disk and stores question banks and study progress alongside it (while recent workspace paths live in your OS user config directory under `qst-studio/recents.json`):

| File | Access | Purpose |
| :--- | :--- | :--- |
| `qst.yaml` / `manifest.yaml` / `quest.yaml` | Read | Graph title, declared `scopes` ladder, `adherences`, and `tours`. |
| `**/*.md` | Read | Tiered Markdown guides with YAML frontmatter (`scope`, `prerequisites`, `sub_guides`, `tags`). |
| `items.yaml` | Read / Write | Declarative question bank authored in the **Author** tab (or generated externally). |
| `.practice_history.json` | Read / Write | Per-question attempt counts, accuracy, and timestamps used for mastery badges. |

---

## Features

### 1. Graph-Aware Reading & Navigation
- **Dynamic Manifest Scopes**: Reads custom `scopes` from your manifest and filters the guide navigator accordingly.
- **Dependency Inspection**: Inspect upstream `prerequisites`, linked `sub_guides` with line segments, and topological prerequisite sequences in the **Details** tab.
- **Markdown + KaTeX Math**: High-fidelity rendering with inline (`$...$`) and display (`$$...$$`) math, plus clickable line-range anchors (`L#`) in the gutter.
- **Fast Workspace Switching**: Quick search via Command Palette (`Ctrl+P`), recent workspaces switcher, and adjustable UI zoom (`70%–200%`).

### 2. Line-Anchored Question Authoring (`items.yaml`)
- Interactive visual editor with live Markdown/KaTeX preview for **7 question types**:
  - **Multiple Choice** (`multiple_choice`)
  - **Multiple Answers** (`multiple_answers`)
  - **True / False** (`true_false`)
  - **Fill in the Blank** (`fill_in_the_blank` — free text & dropdown blanks)
  - **Matching** (`matching` — pairs with optional distractors)
  - **Ordering** (`ordering` — drag/reorder sequences)
  - **Essay** (`essay` — open response with reference solution & self-grading)
- Anchors each question to exact 1-based line ranges in its source guide.

### 3. Graph-Aware Active Recall
- **5 Deck Scopes**:
  - **Guide**: Practice questions anchored to the current guide.
  - **Subtree**: Recursively gather `sub_guides` bottom-up (foundational leaves first) to the target guide.
  - **Learning Path**: Follow upstream `prerequisites` in topological order from foundations up to the target guide.
  - **Tour**: Step through curated `tours` defined in your manifest.
  - **Whole Graph**: Drill all questions across the workspace.
- **Orthogonal Scope-Level Filtering**: Combine manifest `scopes` (e.g., `definition` + `description`) across any deck scope.
- **Auto-Anchored Reader**: Multi-guide decks automatically switch the reader to each card's guide and highlight its anchored lines.
- **Mastery Tracking**: Local `.practice_history.json` tracking with per-guide `correct/attempted` indicators in the sidebar.

---

## Roadmap

- **Inline Guide Editing**: Edit Markdown guides and frontmatter directly inside `qst-studio` (building on the existing backend persistence API).
- **Adaptive Practice & Spaced Repetition**: Smart review queues (*Needs Review*, *Unseen*, *Due*) and interval scheduling backed by `.practice_history.json`.

---

## Building from Source

### Prerequisites
- **Go** 1.27+
- **Node.js** 20+ (or 18+) and `npm`
- **Wails CLI v2** (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.1`)
- **Linux**: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev` (configured via `"build:tags": "webkit2_41"` in [`wails.json`](wails.json))

### Development & Production Builds
```bash
# Run with live reload
wails dev

# Build production binary (output: ./build/bin/qst-studio)
wails build
```

### Running Tests & Checks
```bash
# Build frontend assets first (required for go:embed in main.go)
cd frontend && npm ci && npm run check && npm run build
cd ..

# Run Go vet and tests
go vet -tags webkit2_41 ./...
go test -tags webkit2_41 -v -count=1 ./...
```

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for architecture details and pull request guidelines.

---

## License

Licensed under the **GNU General Public License v3.0** (`GPL-3.0-or-later`). See [`LICENSE`](LICENSE) for the full text.
