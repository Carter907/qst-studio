package main

import (
	"context"
	"fmt"
	"strings"

	"qst-studio/internal/graph"
	"qst-studio/internal/practice"
	"qst-studio/internal/quiz"
	"qst-studio/internal/workspace"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Type aliases preserve Wails frontend binding compatibility.
type (
	SubGuideDTO       = graph.SubGuideDTO
	GuideSummaryDTO   = graph.GuideSummaryDTO
	TourDTO           = graph.TourDTO
	GraphDTO          = graph.GraphDTO
	GuideDetailDTO    = graph.GuideDetailDTO
	QuizChoiceDTO     = quiz.QuizChoiceDTO
	BlankDTO          = quiz.BlankDTO
	MatchingPairDTO   = quiz.MatchingPairDTO
	QuizItemDTO         = quiz.QuizItemDTO
	PracticeDeckOptions = quiz.PracticeDeckOptions
	PracticeRecordDTO   = practice.PracticeRecordDTO
)

// App struct bound to the Wails frontend.
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the Wails application starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetRecentGraphs returns previously opened graph directory paths that still exist on disk.
func (a *App) GetRecentGraphs() ([]string, error) {
	return workspace.GetRecentGraphs()
}

// SaveRecentGraph persists a directory at the top of the recent graphs list (max 10).
func (a *App) SaveRecentGraph(dirPath string) error {
	return workspace.SaveRecentGraph(dirPath)
}

// SelectGraphDirectory opens the native OS directory picker dialog.
func (a *App) SelectGraphDirectory() (string, error) {
	selectedDir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Quest Knowledge Graph Directory",
	})
	if err != nil {
		return "", err
	}
	if selectedDir != "" {
		_ = workspace.SaveRecentGraph(selectedDir)
	}
	return selectedDir, nil
}

// LoadGraph parses and returns the full graph structure from a directory.
func (a *App) LoadGraph(dirPath string) (*GraphDTO, error) {
	g, err := graph.Load(dirPath)
	if err != nil {
		return nil, err
	}
	_ = workspace.SaveRecentGraph(dirPath)
	return g, nil
}

// GetGuide loads a specific guide's markdown body, raw content, and metadata.
func (a *App) GetGuide(dirPath string, guideID string) (*GuideDetailDTO, error) {
	return graph.GetGuide(dirPath, guideID)
}

// GetLearningPath computes the prerequisite sequence leading to a target guide.
func (a *App) GetLearningPath(dirPath string, targetID string) ([]string, error) {
	return graph.GetLearningPath(dirPath, targetID)
}

// GetSubtreePath computes the recursive sub-guide sequence bottom-up from foundational leaves to rootID.
func (a *App) GetSubtreePath(dirPath string, rootID string) ([]string, error) {
	return graph.GetSubtreePath(dirPath, rootID)
}

// GetPracticeDeck resolves quiz items for the requested deck options ("guide", "subtree", "learning_path", "tour", or "graph")
// and optionally filters by manifest scope levels.
func (a *App) GetPracticeDeck(dirPath string, opts PracticeDeckOptions) ([]QuizItemDTO, error) {
	var items []QuizItemDTO
	var err error

	switch opts.Scope {
	case "guide":
		if strings.TrimSpace(opts.AnchorGuideID) == "" {
			return []QuizItemDTO{}, nil
		}
		items, err = quiz.GetItems(dirPath, opts.AnchorGuideID)
	case "subtree":
		var guideIDs []string
		guideIDs, err = graph.GetSubtreePath(dirPath, opts.AnchorGuideID)
		if err == nil {
			items, err = quiz.GetItemsForGuides(dirPath, guideIDs)
		}
	case "learning_path":
		var guideIDs []string
		guideIDs, err = graph.GetLearningPath(dirPath, opts.AnchorGuideID)
		if err == nil {
			items, err = quiz.GetItemsForGuides(dirPath, guideIDs)
		}
	case "tour":
		var guideIDs []string
		guideIDs, err = graph.GetTourGuides(dirPath, opts.TourName)
		if err == nil {
			items, err = quiz.GetItemsForGuides(dirPath, guideIDs)
		}
	case "graph":
		items, err = quiz.GetItems(dirPath, "")
	default:
		return nil, fmt.Errorf("unsupported practice scope: %q", opts.Scope)
	}
	if err != nil {
		return nil, err
	}

	if len(opts.FilterScopes) > 0 {
		guideScopes, scopeErr := graph.GetGuideScopes(dirPath)
		if scopeErr != nil {
			return nil, scopeErr
		}
		items = quiz.FilterItemsByScopes(items, opts.FilterScopes, guideScopes)
	}

	return items, nil
}

// SaveGuideContent writes raw markdown content (frontmatter + body) directly to disk.
func (a *App) SaveGuideContent(filePath string, rawContent string) error {
	return graph.SaveGuideContent(filePath, rawContent)
}

// GetQuizItems returns all saved quiz items, optionally filtered by guideID.
func (a *App) GetQuizItems(dirPath string, guideID string) ([]QuizItemDTO, error) {
	return quiz.GetItems(dirPath, guideID)
}

// SaveQuizItem validates and saves or updates a quiz item in items.yaml.
func (a *App) SaveQuizItem(dirPath string, item QuizItemDTO) error {
	return quiz.SaveItem(dirPath, item)
}

// DeleteQuizItem removes a saved quiz item by ID from items.yaml.
func (a *App) DeleteQuizItem(dirPath string, itemID string) error {
	return quiz.DeleteItem(dirPath, itemID)
}

// GetPracticeHistory returns all saved question practice records.
func (a *App) GetPracticeHistory(dirPath string) (map[string]PracticeRecordDTO, error) {
	return practice.GetHistory(dirPath)
}

// GetPracticeRecord returns a practice record for a specific question item.
func (a *App) GetPracticeRecord(dirPath string, itemID string) (*PracticeRecordDTO, error) {
	return practice.GetRecord(dirPath, itemID)
}

// RecordPracticeAttempt updates practice attempts and accuracy for a question.
func (a *App) RecordPracticeAttempt(dirPath string, itemID string, guideID string, correct bool) error {
	return practice.RecordAttempt(dirPath, itemID, guideID, correct)
}
