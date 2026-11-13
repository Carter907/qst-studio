package main

import (
	"os"
	"path/filepath"
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
