package quizzes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListMissingQuizzesDirReturnsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	list, err := List(tmpDir)
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 quizzes, got %d", len(list))
	}
}

func TestListAndGetStandaloneQuizzes(t *testing.T) {
	tmpDir := t.TempDir()
	mathDir := filepath.Join(tmpDir, "quizzes", "math")
	if err := os.MkdirAll(mathDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Hidden file should be ignored.
	if err := os.WriteFile(filepath.Join(tmpDir, "quizzes", ".draft.yaml"), []byte("title: Ignored\n"), 0644); err != nil {
		t.Fatalf("write hidden yaml: %v", err)
	}

	// Quiz with omitted id and omitted item ids (exercises defaults + normalization).
	quadYaml := `title: Quadratic Roots
description: Drill factoring and discriminant problems.
tags:
  - algebra
related_guides:
  - polynomials
items:
  - prompt: "Solve $x^2 - 9 = 0$ for $x > 0$: [b1]"
    type: fill_in_the_blank
    blanks:
      b1:
        type: text
        answers: ["3"]
  - prompt: "The discriminant of $x^2 + 2x + 1 = 0$ is $0$."
    choices:
      - text: "True"
        correct: true
      - text: "False"
        correct: false
`
	if err := os.WriteFile(filepath.Join(mathDir, "quadratics.yml"), []byte(quadYaml), 0644); err != nil {
		t.Fatalf("write quadratics.yml: %v", err)
	}

	summaries, err := List(tmpDir)
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 quiz summary, got %d", len(summaries))
	}
	if summaries[0].ID != "math/quadratics" {
		t.Fatalf("expected derived id 'math/quadratics', got %q", summaries[0].ID)
	}
	if summaries[0].ItemCount != 2 {
		t.Fatalf("expected ItemCount=2, got %d", summaries[0].ItemCount)
	}

	detail, err := Get(tmpDir, "math/quadratics")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if detail.Items[0].ID != "math/quadratics__1" || detail.Items[1].ID != "math/quadratics__2" {
		t.Fatalf("unexpected fallback item IDs: %q, %q", detail.Items[0].ID, detail.Items[1].ID)
	}
	if detail.Items[1].Type != "multiple_choice" {
		t.Fatalf("expected defaulted type 'multiple_choice', got %q", detail.Items[1].Type)
	}
	if detail.Items[1].Choices[0].ID != "c1" {
		t.Fatalf("expected defaulted choice ID 'c1', got %q", detail.Items[1].Choices[0].ID)
	}
}

func TestGetStandaloneQuizValidationErrors(t *testing.T) {
	tmpDir := t.TempDir()
	qDir := filepath.Join(tmpDir, "quizzes")
	if err := os.MkdirAll(qDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Duplicate item IDs should fail cleanly.
	dupYaml := `id: dup-items
title: Duplicate Items
items:
  - id: Q1
    prompt: First
    choices:
      - text: A
        correct: true
      - text: B
        correct: false
  - id: Q1
    prompt: Second
    choices:
      - text: A
        correct: true
      - text: B
        correct: false
`
	if err := os.WriteFile(filepath.Join(qDir, "dup.yaml"), []byte(dupYaml), 0644); err != nil {
		t.Fatalf("write dup.yaml: %v", err)
	}

	if _, err := List(tmpDir); err == nil || !strings.Contains(err.Error(), "duplicate item id") {
		t.Fatalf("expected duplicate item id error, got %v", err)
	}
}
