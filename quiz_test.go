package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMultipleQuestionTypesPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	app := NewApp()

	// 1. Multiple Choice
	mc := QuizItemDTO{
		ID:      "Q_MC_1",
		GuideID: "pointers",
		Type:    "multiple_choice",
		Prompt:  "What is a pointer?",
		Choices: []QuizChoiceDTO{
			{Text: "Memory address variable", Correct: true},
			{Text: "Hardware device", Correct: false},
		},
	}
	if err := app.SaveQuizItem(tmpDir, mc); err != nil {
		t.Fatalf("SaveQuizItem multiple_choice failed: %v", err)
	}

	// 2. Multiple Answers
	ma := QuizItemDTO{
		ID:      "Q_MA_1",
		GuideID: "pointers",
		Type:    "multiple_answers",
		Prompt:  "Which are valid pointer operators in C?",
		Choices: []QuizChoiceDTO{
			{Text: "* (dereference)", Correct: true},
			{Text: "& (address-of)", Correct: true},
			{Text: "@@ (at-at)", Correct: false},
		},
	}
	if err := app.SaveQuizItem(tmpDir, ma); err != nil {
		t.Fatalf("SaveQuizItem multiple_answers failed: %v", err)
	}

	// 3. True / False
	tf := QuizItemDTO{
		ID:      "Q_TF_1",
		GuideID: "pointers",
		Type:    "true_false",
		Prompt:  "A pointer can point to another pointer.",
		Choices: []QuizChoiceDTO{
			{Text: "True", Correct: true},
			{Text: "False", Correct: false},
		},
	}
	if err := app.SaveQuizItem(tmpDir, tf); err != nil {
		t.Fatalf("SaveQuizItem true_false failed: %v", err)
	}

	// 4. Fill in the Blank
	fib := QuizItemDTO{
		ID:      "Q_FIB_1",
		GuideID: "pointers",
		Type:    "fill_in_the_blank",
		Prompt:  "Dereferencing a pointer is done using the [b1] operator.",
		Blanks: map[string]BlankDTO{
			"b1": {
				Type:    "text",
				Answers: []string{"*", "asterisk"},
			},
		},
	}
	if err := app.SaveQuizItem(tmpDir, fib); err != nil {
		t.Fatalf("SaveQuizItem fill_in_the_blank failed: %v", err)
	}

	// 5. Matching
	match := QuizItemDTO{
		ID:      "Q_MATCH_1",
		GuideID: "pointers",
		Type:    "matching",
		Prompt:  "Match the operator to its purpose.",
		Pairs: []MatchingPairDTO{
			{Prompt: "*p", Target: "Dereference"},
			{Prompt: "&x", Target: "Address-of"},
		},
		Distractors: []string{"Bitwise NOT"},
	}
	if err := app.SaveQuizItem(tmpDir, match); err != nil {
		t.Fatalf("SaveQuizItem matching failed: %v", err)
	}

	// 6. Ordering
	order := QuizItemDTO{
		ID:          "Q_ORDER_1",
		GuideID:     "pointers",
		Type:        "ordering",
		Prompt:      "Order the steps to initialize a pointer.",
		TopLabel:    "First step",
		BottomLabel: "Last step",
		Items: []string{
			"Declare integer variable x",
			"Declare pointer variable p",
			"Assign address of x to p: p = &x",
		},
	}
	if err := app.SaveQuizItem(tmpDir, order); err != nil {
		t.Fatalf("SaveQuizItem ordering failed: %v", err)
	}

	// 7. Essay
	essay := QuizItemDTO{
		ID:       "Q_ESSAY_1",
		GuideID:  "pointers",
		Type:     "essay",
		Prompt:   "Explain how pointer arithmetic behaves when incrementing a pointer to int vs char.",
		Solution: "Pointer arithmetic scales by the size of the underlying type: int increments by sizeof(int), char increments by 1 byte.",
	}
	if err := app.SaveQuizItem(tmpDir, essay); err != nil {
		t.Fatalf("SaveQuizItem essay failed: %v", err)
	}

	// Fetch all and verify count
	items, err := app.GetQuizItems(tmpDir, "pointers")
	if err != nil {
		t.Fatalf("GetQuizItems failed: %v", err)
	}
	if len(items) != 7 {
		t.Fatalf("expected 7 items, got %d", len(items))
	}

	// Verify delete
	if err := app.DeleteQuizItem(tmpDir, "Q_ESSAY_1"); err != nil {
		t.Fatalf("DeleteQuizItem failed: %v", err)
	}
	itemsAfter, _ := app.GetQuizItems(tmpDir, "pointers")
	if len(itemsAfter) != 6 {
		t.Fatalf("expected 6 items after delete, got %d", len(itemsAfter))
	}
}

func TestManifestAndGraphLoading(t *testing.T) {
	tmpDir := t.TempDir()

	manifestContent := `title: Test Graph
description: A mock graph for testing
scopes: [definition, explanation]
adherences: [strict, loose]
`
	if err := os.WriteFile(filepath.Join(tmpDir, "qst.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to write qst.yaml: %v", err)
	}

	guidesDir := filepath.Join(tmpDir, "guides")
	if err := os.MkdirAll(guidesDir, 0755); err != nil {
		t.Fatalf("failed to create guides dir: %v", err)
	}

	guide1 := `---
scope: definition
prerequisites: []
---
# Variables
Variables hold values.
`
	if err := os.WriteFile(filepath.Join(guidesDir, "variables.md"), []byte(guide1), 0644); err != nil {
		t.Fatalf("failed to write variables.md: %v", err)
	}

	app := NewApp()
	graph, err := app.LoadGraph(tmpDir)
	if err != nil {
		t.Fatalf("LoadGraph failed: %v", err)
	}
	if graph.Title != "Test Graph" {
		t.Errorf("expected title 'Test Graph', got %q", graph.Title)
	}
	if len(graph.Guides) != 1 {
		t.Fatalf("expected 1 guide, got %d", len(graph.Guides))
	}
}

func TestPracticeHistory(t *testing.T) {
	tmpDir := t.TempDir()
	app := NewApp()

	if err := app.RecordPracticeAttempt(tmpDir, "Q_TEST_1", "pointers", true); err != nil {
		t.Fatalf("RecordPracticeAttempt failed: %v", err)
	}

	hist, err := app.GetPracticeHistory(tmpDir)
	if err != nil {
		t.Fatalf("GetPracticeHistory failed: %v", err)
	}
	if len(hist) != 1 {
		t.Fatalf("expected 1 record, got %d", len(hist))
	}
	rec := hist["Q_TEST_1"]
	if rec.Attempts != 1 || rec.Correct != 1 || !rec.LastCorrect {
		t.Errorf("unexpected record values: %+v", rec)
	}

	// Record an incorrect attempt
	if err := app.RecordPracticeAttempt(tmpDir, "Q_TEST_1", "pointers", false); err != nil {
		t.Fatalf("RecordPracticeAttempt 2 failed: %v", err)
	}
	histAfter, _ := app.GetPracticeHistory(tmpDir)
	recAfter := histAfter["Q_TEST_1"]
	if recAfter.Attempts != 2 || recAfter.Correct != 1 || recAfter.LastCorrect {
		t.Errorf("unexpected record after 2nd attempt: %+v", recAfter)
	}

	// Test GetPracticeRecord
	singleRec, err := app.GetPracticeRecord(tmpDir, "Q_TEST_1")
	if err != nil {
		t.Fatalf("GetPracticeRecord failed: %v", err)
	}
	if singleRec == nil || singleRec.Attempts != 2 {
		t.Errorf("unexpected single record: %+v", singleRec)
	}

	nonExistent, err := app.GetPracticeRecord(tmpDir, "DOES_NOT_EXIST")
	if err != nil {
		t.Fatalf("GetPracticeRecord non-existent failed: %v", err)
	}
	if nonExistent != nil {
		t.Errorf("expected nil record for non-existent item, got %+v", nonExistent)
	}
}

func TestSubtreeAndLearningPathPracticeDecks(t *testing.T) {
	tmpDir := t.TempDir()
	app := NewApp()

	manifest := `title: Graph Deck Test
scopes: [definition, description, explanation]
adherences: [strict]
`
	if err := os.WriteFile(filepath.Join(tmpDir, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	guidesDir := filepath.Join(tmpDir, "guides")
	if err := os.MkdirAll(guidesDir, 0755); err != nil {
		t.Fatalf("mkdir guides: %v", err)
	}

	// Subtree DAG:
	//   lesson_root (explanation, 0 direct items) -> [desc_a, desc_b]
	//   desc_a (description) -> [def_shared, def_a]
	//   desc_b (description) -> [def_shared]
	//   Prerequisites on def_b: [def_shared] -> [def_prereq_mid] -> [def_b]
	files := map[string]string{
		"def_shared.md": "---\nscope: definition\nprerequisites: []\nsub_guides: []\n---\n# Shared\nLine 1\n",
		"def_a.md":      "---\nscope: definition\nprerequisites: []\nsub_guides: []\n---\n# Def A\nLine 1\n",
		"desc_a.md": "---\nscope: description\nprerequisites: []\nsub_guides:\n  - guide: def_shared\n    adherence: strict\n    segment: \"1-1\"\n  - guide: def_a\n    adherence: strict\n    segment: \"1-1\"\n---\n# Desc A\nLine 1\n",
		"desc_b.md": "---\nscope: description\nprerequisites: []\nsub_guides:\n  - guide: def_shared\n    adherence: strict\n    segment: \"1-1\"\n---\n# Desc B\nLine 1\n",
		"lesson_root.md": "---\nscope: explanation\nprerequisites: []\nsub_guides:\n  - guide: desc_a\n    adherence: strict\n    segment: \"1-1\"\n  - guide: desc_b\n    adherence: strict\n    segment: \"1-1\"\n---\n# Lesson Root\nLine 1\n",
		"def_prereq_mid.md": "---\nscope: definition\nprerequisites: [def_shared]\nsub_guides: []\n---\n# Mid\nLine 1\n",
		"def_b.md":          "---\nscope: definition\nprerequisites: [def_prereq_mid]\nsub_guides: []\n---\n# Target\nLine 1\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(guidesDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	// Save quiz items in non-topological order to verify guide-order + intra-guide order preservation.
	items := []QuizItemDTO{
		{ID: "Q_DESCA_1", GuideID: "desc_a", Type: "essay", Prompt: "Desc A 1"},
		{ID: "Q_SHARED_1", GuideID: "def_shared", Type: "essay", Prompt: "Shared 1"},
		{ID: "Q_SHARED_2", GuideID: "def_shared", Type: "essay", Prompt: "Shared 2"},
		{ID: "Q_DEFB_1", GuideID: "def_b", Type: "essay", Prompt: "Def B 1"},
		{ID: "Q_DEFA_1", GuideID: "def_a", Type: "essay", Prompt: "Def A 1"},
		{ID: "Q_MID_1", GuideID: "def_prereq_mid", Type: "essay", Prompt: "Mid 1"},
		{ID: "Q_DESCB_1", GuideID: "desc_b", Type: "essay", Prompt: "Desc B 1"},
	}
	for _, it := range items {
		if err := app.SaveQuizItem(tmpDir, it); err != nil {
			t.Fatalf("SaveQuizItem %s: %v", it.ID, err)
		}
	}

	// 1. Subtree path is bottom-up and deduplicates diamond descendant def_shared.
	subtreePath, err := app.GetSubtreePath(tmpDir, "lesson_root")
	if err != nil {
		t.Fatalf("GetSubtreePath failed: %v", err)
	}
	wantSubtreePath := []string{"def_shared", "def_a", "desc_a", "desc_b", "lesson_root"}
	if !reflect.DeepEqual(subtreePath, wantSubtreePath) {
		t.Fatalf("GetSubtreePath = %v, want %v", subtreePath, wantSubtreePath)
	}

	// 2. Subtree deck preserves bottom-up guide order and items.yaml order within def_shared,
	// even though lesson_root itself has 0 direct questions.
	subtreeDeck, err := app.GetPracticeDeck(tmpDir, PracticeDeckOptions{Scope: "subtree", AnchorGuideID: "lesson_root"})
	if err != nil {
		t.Fatalf("GetPracticeDeck(subtree) failed: %v", err)
	}
	gotSubtreeIDs := make([]string, len(subtreeDeck))
	for i, it := range subtreeDeck {
		gotSubtreeIDs[i] = it.ID
	}
	wantSubtreeIDs := []string{"Q_SHARED_1", "Q_SHARED_2", "Q_DEFA_1", "Q_DESCA_1", "Q_DESCB_1"}
	if !reflect.DeepEqual(gotSubtreeIDs, wantSubtreeIDs) {
		t.Fatalf("subtree deck IDs = %v, want %v", gotSubtreeIDs, wantSubtreeIDs)
	}

	// 3. Learning-path deck follows prerequisites topologically: def_shared -> def_prereq_mid -> def_b.
	pathDeck, err := app.GetPracticeDeck(tmpDir, PracticeDeckOptions{Scope: "learning_path", AnchorGuideID: "def_b"})
	if err != nil {
		t.Fatalf("GetPracticeDeck(learning_path) failed: %v", err)
	}
	gotPathIDs := make([]string, len(pathDeck))
	for i, it := range pathDeck {
		gotPathIDs[i] = it.ID
	}
	wantPathIDs := []string{"Q_SHARED_1", "Q_SHARED_2", "Q_MID_1", "Q_DEFB_1"}
	if !reflect.DeepEqual(gotPathIDs, wantPathIDs) {
		t.Fatalf("learning_path deck IDs = %v, want %v", gotPathIDs, wantPathIDs)
	}

	// 4. Guide & Graph scopes + unsupported scope error.
	guideDeck, err := app.GetPracticeDeck(tmpDir, PracticeDeckOptions{Scope: "guide", AnchorGuideID: "def_shared"})
	if err != nil || len(guideDeck) != 2 {
		t.Fatalf("GetPracticeDeck(guide) = %d items, err=%v", len(guideDeck), err)
	}
	graphDeck, err := app.GetPracticeDeck(tmpDir, PracticeDeckOptions{Scope: "graph"})
	if err != nil || len(graphDeck) != 7 {
		t.Fatalf("GetPracticeDeck(graph) = %d items, err=%v", len(graphDeck), err)
	}
	if _, err := app.GetPracticeDeck(tmpDir, PracticeDeckOptions{Scope: "unknown", AnchorGuideID: "def_shared"}); err == nil {
		t.Fatal("expected error for unsupported practice scope")
	}

	// 5. Subtree cycle detection.
	cycleDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(cycleDir, "guides"), 0755)
	_ = os.WriteFile(filepath.Join(cycleDir, "guides", "a.md"), []byte("---\nscope: explanation\nsub_guides:\n  - guide: b\n---\n"), 0644)
	_ = os.WriteFile(filepath.Join(cycleDir, "guides", "b.md"), []byte("---\nscope: description\nsub_guides:\n  - guide: a\n---\n"), 0644)
	if _, err := app.GetSubtreePath(cycleDir, "a"); err == nil {
		t.Fatal("expected cycle error in GetSubtreePath")
	}
}

func TestTourAndScopeLevelFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	app := NewApp()

	manifest := `title: Tour and Scope Test
scopes: [definition, description, explanation]
adherences: [strict]
tours:
  - name: Core Tour
    guides: [exp_intro, missing_guide, def_basics, desc_flow]
`
	if err := os.WriteFile(filepath.Join(tmpDir, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	guidesDir := filepath.Join(tmpDir, "guides")
	if err := os.MkdirAll(guidesDir, 0755); err != nil {
		t.Fatalf("mkdir guides: %v", err)
	}

	guides := map[string]string{
		"def_basics.md": "---\nscope: definition\nprerequisites: []\nsub_guides: []\n---\n# Basics\n",
		"desc_flow.md":  "---\nscope: description\nprerequisites: []\nsub_guides:\n  - guide: def_basics\n---\n# Flow\n",
		"exp_intro.md":  "---\nscope: explanation\nprerequisites: []\nsub_guides:\n  - guide: desc_flow\n---\n# Intro\n",
	}
	for name, content := range guides {
		if err := os.WriteFile(filepath.Join(guidesDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	items := []QuizItemDTO{
		{ID: "Q_DEF_EXPLICIT", GuideID: "def_basics", Scope: "definition", Type: "essay", Prompt: "Def explicit"},
		{ID: "Q_DEF_FALLBACK", GuideID: "def_basics", Scope: "", Type: "essay", Prompt: "Def fallback"},
		{ID: "Q_DESC_1", GuideID: "desc_flow", Scope: "", Type: "essay", Prompt: "Desc fallback"},
		{ID: "Q_EXP_1", GuideID: "exp_intro", Scope: "explanation", Type: "essay", Prompt: "Exp explicit"},
	}
	for _, it := range items {
		if err := app.SaveQuizItem(tmpDir, it); err != nil {
			t.Fatalf("SaveQuizItem %s: %v", it.ID, err)
		}
	}

	// 1. Tour preserves manifest tour.guides order and skips missing guides cleanly.
	tourDeck, err := app.GetPracticeDeck(tmpDir, PracticeDeckOptions{
		Scope:    "tour",
		TourName: "Core Tour",
	})
	if err != nil {
		t.Fatalf("GetPracticeDeck(tour) failed: %v", err)
	}
	gotTourIDs := make([]string, len(tourDeck))
	for i, it := range tourDeck {
		gotTourIDs[i] = it.ID
	}
	wantTourIDs := []string{"Q_EXP_1", "Q_DEF_EXPLICIT", "Q_DEF_FALLBACK", "Q_DESC_1"}
	if !reflect.DeepEqual(gotTourIDs, wantTourIDs) {
		t.Fatalf("tour deck IDs = %v, want %v", gotTourIDs, wantTourIDs)
	}

	// 2. Unknown tour returns error.
	if _, err := app.GetPracticeDeck(tmpDir, PracticeDeckOptions{Scope: "tour", TourName: "Does Not Exist"}); err == nil {
		t.Fatal("expected error for unknown tour name")
	}

	// 3. Single and multi-scope filtering on tour & subtree, with fallback to guide frontmatter scope.
	filteredTour, err := app.GetPracticeDeck(tmpDir, PracticeDeckOptions{
		Scope:        "tour",
		TourName:     "Core Tour",
		FilterScopes: []string{"definition"},
	})
	if err != nil {
		t.Fatalf("GetPracticeDeck(tour filtered) failed: %v", err)
	}
	gotFilteredTour := make([]string, len(filteredTour))
	for i, it := range filteredTour {
		gotFilteredTour[i] = it.ID
	}
	wantFilteredTour := []string{"Q_DEF_EXPLICIT", "Q_DEF_FALLBACK"}
	if !reflect.DeepEqual(gotFilteredTour, wantFilteredTour) {
		t.Fatalf("filtered tour IDs = %v, want %v", gotFilteredTour, wantFilteredTour)
	}

	filteredSubtree, err := app.GetPracticeDeck(tmpDir, PracticeDeckOptions{
		Scope:         "subtree",
		AnchorGuideID: "exp_intro",
		FilterScopes:  []string{"description", "explanation"},
	})
	if err != nil {
		t.Fatalf("GetPracticeDeck(subtree filtered) failed: %v", err)
	}
	gotFilteredSubtree := make([]string, len(filteredSubtree))
	for i, it := range filteredSubtree {
		gotFilteredSubtree[i] = it.ID
	}
	wantFilteredSubtree := []string{"Q_DESC_1", "Q_EXP_1"}
	if !reflect.DeepEqual(gotFilteredSubtree, wantFilteredSubtree) {
		t.Fatalf("filtered subtree IDs = %v, want %v", gotFilteredSubtree, wantFilteredSubtree)
	}

	// 4. Omitted scopes in manifest.yaml initializes to non-nil empty slice.
	noScopesDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(noScopesDir, "manifest.yaml"), []byte("title: No Scopes\n"), 0644); err != nil {
		t.Fatalf("write no-scopes manifest: %v", err)
	}
	loaded, err := app.LoadGraph(noScopesDir)
	if err != nil {
		t.Fatalf("LoadGraph(noScopesDir) failed: %v", err)
	}
	if loaded.Scopes == nil {
		t.Fatal("expected non-nil empty Scopes slice when omitted from manifest.yaml")
	}
}
