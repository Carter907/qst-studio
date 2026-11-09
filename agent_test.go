package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractSegmentLines(t *testing.T) {
	body := strings.Join([]string{
		"Line 1: Introduction to C variables.",
		"Line 2: Variables must be declared with a type.",
		"Line 3: Local variables are allocated on the stack.",
		"Line 4: Uninitialized local variables hold indeterminate garbage.",
		"Line 5: Static variables are zero-initialized.",
	}, "\n")

	got := extractSegmentLines(body, "3-4")
	want := "Line 3: Local variables are allocated on the stack.\nLine 4: Uninitialized local variables hold indeterminate garbage."
	if got != want {
		t.Fatalf("extractSegmentLines(3-4) = %q, want %q", got, want)
	}

	// Default "1-1" on a multi-line body falls back to the full body so the user doesn't have to click lines first
	gotFull := extractSegmentLines(body, "1-1")
	if gotFull != body {
		t.Fatalf("extractSegmentLines(1-1) expected full body fallback, got %q", gotFull)
	}

	// Single specific line > 1
	gotLine2 := extractSegmentLines(body, "2-2")
	if gotLine2 != "Line 2: Variables must be declared with a type." {
		t.Fatalf("extractSegmentLines(2-2) = %q", gotLine2)
	}
}

func TestBuildScopeRubric(t *testing.T) {
	scopes := []string{"definition", "description", "explanation", "lesson"}
	for _, sc := range scopes {
		r := buildScopeRubric(sc)
		if !strings.Contains(strings.ToLower(r), sc) {
			t.Errorf("buildScopeRubric(%q) = %q, expected to mention scope", sc, r)
		}
	}
}

func TestOllamaAgentDiscoveryAndGeneration(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// Set up a mock Ollama HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"models": []map[string]any{
					{"name": "qwen2.5-coder:7b"},
					{"name": "llama3.2:latest"},
				},
			})
		case "/api/chat":
			var req ollamaChatRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			payload := generatedQuestionPayload{
				Title:  "Stack Variable Initialization",
				Prompt: "What value does an uninitialized local variable hold in C?",
				Choices: []struct {
					Text     string `json:"text"`
					Correct  bool   `json:"correct"`
					Feedback string `json:"feedback"`
				}{
					{
						Text:     "Indeterminate garbage bytes from stack memory.",
						Correct:  true,
						Feedback: "Correct! Automatic stack variables are not zeroed.",
					},
					{
						Text:     "Zero by default.",
						Correct:  false,
						Feedback: "Incorrect. Only static/global variables are zeroed by default.",
					},
					{
						Text:     "Compiler error at link time.",
						Correct:  false,
						Feedback: "Incorrect. It compiles cleanly unless strict warnings-as-errors are enabled.",
					},
				},
			}
			contentBytes, _ := json.Marshal(payload)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(ollamaChatResponse{
				Message: ollamaChatMessage{
					Role:    "assistant",
					Content: string(contentBytes),
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	// Create temporary graph directory
	tmpDir := t.TempDir()
	manifestContent := "title: Test C Graph\nscopes:\n  - definition\n  - description\nadherences:\n  - strict\ntours: []\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}
	guideContent := "---\nscope: definition\nprerequisites: []\nsub_guides: []\ntags: []\n---\nLine 1: Variables in C.\nLine 2: Uninitialized local variables contain indeterminate stack values.\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "Variable.md"), []byte(guideContent), 0644); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	origCfg := loadAgentConfig()
	defer func() {
		_ = app.SaveAgentConfig(origCfg)
	}()

	if err := app.SaveAgentConfig(AgentConfigDTO{
		Endpoint: mockServer.URL,
		Model:    "qwen2.5-coder:7b",
	}); err != nil {
		t.Fatal(err)
	}

	status, err := app.GetAgentStatus()
	if err != nil {
		t.Fatalf("GetAgentStatus error: %v", err)
	}
	if !status.Available || len(status.Models) != 2 {
		t.Fatalf("unexpected status: %+v", status)
	}

	item, err := app.GenerateQuestionFromSegment(tmpDir, "Variable", "2-2")
	if err != nil {
		t.Fatalf("GenerateQuestionFromSegment error: %v", err)
	}
	if item.GuideID != "Variable" || item.Segment != "2-2" || len(item.Choices) != 3 {
		t.Fatalf("unexpected generated item: %+v", item)
	}
	if !item.Choices[0].Correct || item.Choices[1].Correct {
		t.Fatalf("unexpected choice correctness flags: %+v", item.Choices)
	}
}

func TestNextPracticeTurnAdaptiveAndAutoGrow(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	callCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"models": []map[string]any{
					{"name": "qwen2.5-coder:7b"},
				},
			})
		case "/api/chat":
			callCount++
			payload := generatedQuestionPayload{
				Title:  "Auto-Generated Question",
				Prompt: "Which concept is foundational in this segment?",
				Choices: []struct {
					Text     string `json:"text"`
					Correct  bool   `json:"correct"`
					Feedback string `json:"feedback"`
				}{
					{Text: "Option A", Correct: true, Feedback: "Correct explanation."},
					{Text: "Option B", Correct: false, Feedback: "Misconception B."},
					{Text: "Option C", Correct: false, Feedback: "Misconception C."},
				},
			}
			contentBytes, _ := json.Marshal(payload)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(ollamaChatResponse{
				Message: ollamaChatMessage{
					Role:    "assistant",
					Content: string(contentBytes),
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	tmpDir := t.TempDir()
	manifestContent := "title: Adaptive Test Graph\nscopes:\n  - definition\n  - description\nadherences:\n  - strict\ntours: []\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Guide 1: Leaf definition (no prerequisites)
	leafGuide := "---\nscope: definition\nprerequisites: []\nsub_guides: []\ntags: []\n---\nLine 1: A pointer stores a memory address.\nLine 2: Dereferencing accesses the target value.\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "Pointer.md"), []byte(leafGuide), 0644); err != nil {
		t.Fatal(err)
	}

	// Guide 2: Higher-tier description that depends on Pointer
	advGuide := "---\nscope: description\nprerequisites:\n  - Pointer\nsub_guides:\n  - guide: Pointer\n    adherence: strict\n    segment: 1-2\ntags: []\n---\nLine 1: Dynamic allocation returns a pointer.\nLine 2: Always check for NULL before dereferencing.\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "Malloc.md"), []byte(advGuide), 0644); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.SaveAgentConfig(AgentConfigDTO{
		Endpoint: mockServer.URL,
		Model:    "qwen2.5-coder:7b",
	}); err != nil {
		t.Fatal(err)
	}

	// Turn 1: Adaptive mode should pick "Pointer" first because "Malloc" has an unmastered prerequisite ("Pointer"),
	// and since items.yaml does not exist yet, it should generate via Ollama and auto-save to items.yaml.
	turn1, err := app.NextPracticeTurn(PracticeTurnRequest{
		DirPath: tmpDir,
		Mode:    "adaptive",
	})
	if err != nil {
		t.Fatalf("NextPracticeTurn Turn 1 error: %v", err)
	}
	if turn1.Item.GuideID != "Pointer" {
		t.Fatalf("expected Turn 1 to prioritize foundational prerequisite 'Pointer', got %q", turn1.Item.GuideID)
	}
	if turn1.Source != "generated" {
		t.Fatalf("expected Turn 1 Source='generated', got %q", turn1.Source)
	}

	// Verify item was automatically persisted to items.yaml
	savedItems, err := app.GetQuizItems(tmpDir, "Pointer")
	if err != nil || len(savedItems) != 1 {
		t.Fatalf("expected 1 auto-saved item in items.yaml, got %d (err=%v)", len(savedItems), err)
	}

	// Turn 2 (without marking Turn 1 as seen): Should serve the cached item from items.yaml without calling Ollama again
	prevCalls := callCount
	turn2, err := app.NextPracticeTurn(PracticeTurnRequest{
		DirPath: tmpDir,
		Mode:    "adaptive",
	})
	if err != nil {
		t.Fatalf("NextPracticeTurn Turn 2 error: %v", err)
	}
	if turn2.Source != "bank" || turn2.Item.ID != turn1.Item.ID || callCount != prevCalls {
		t.Fatalf("expected Turn 2 to serve unseen item from bank without LLM call; got Source=%q, callCount=%d", turn2.Source, callCount)
	}

	// Record mastery on "Pointer" segment "1-2" and verify it rolls up to "Pointer:ALL"
	for i := 0; i < 4; i++ {
		if _, err := app.RecordPracticeResult(tmpDir, "Pointer", "1-2", true); err != nil {
			t.Fatal(err)
		}
	}
	states, err := app.GetLearnerStates(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if states["Pointer:ALL"].PMastery < 0.75 {
		t.Fatalf("expected Pointer:ALL mastery roll-up >= 0.75, got %f", states["Pointer:ALL"].PMastery)
	}

	// Turn 3: Now that "Pointer" is mastered (> 0.70) and excluded/mastered, Adaptive mode should unlock "Malloc"
	// and populate LinkedSubGuide ("Pointer")!
	turn3, err := app.NextPracticeTurn(PracticeTurnRequest{
		DirPath:     tmpDir,
		Mode:        "adaptive",
		SeenItemIDs: []string{turn1.Item.ID},
	})
	if err != nil {
		t.Fatalf("NextPracticeTurn Turn 3 error: %v", err)
	}
	if turn3.Item.GuideID != "Malloc" {
		t.Fatalf("expected Turn 3 to advance to unlocked 'Malloc', got %q", turn3.Item.GuideID)
	}
	if turn3.LinkedSubGuide != "Pointer" {
		t.Fatalf("expected Turn 3 LinkedSubGuide='Pointer', got %q", turn3.LinkedSubGuide)
	}
}
