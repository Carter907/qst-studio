package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/yaml.v3"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// -------------------------------------------------------------
// Data Transfer Objects (DTOs) for Frontend
// -------------------------------------------------------------

type SubGuideDTO struct {
	Guide     string `json:"guide"`
	Adherence string `json:"adherence"`
	Segment   string `json:"segment"`
}

type GuideSummaryDTO struct {
	ID            string        `json:"id"`
	Scope         string        `json:"scope"`
	LineCount     int           `json:"line_count"`
	Prerequisites []string      `json:"prerequisites"`
	SubGuides     []SubGuideDTO `json:"sub_guides"`
}

type TourDTO struct {
	Name   string   `json:"name"`
	Guides []string `json:"guides"`
}

type GraphDTO struct {
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Scopes           []string          `json:"scopes"`
	Adherences       []string          `json:"adherences"`
	RelaxedSubguides bool              `json:"relaxed_subguides"`
	StrictCoverage   bool              `json:"strict_coverage"`
	RequireSubguides bool              `json:"require_subguides"`
	Guides           []GuideSummaryDTO `json:"guides"`
	Tours            []TourDTO         `json:"tours"`
}

type GuideDetailDTO struct {
	ID            string        `json:"id"`
	Scope         string        `json:"scope"`
	Path          string        `json:"path"`
	Body          string        `json:"body"`
	RawContent    string        `json:"raw_content"`
	LineCount     int           `json:"line_count"`
	Prerequisites []string      `json:"prerequisites"`
	SubGuides     []SubGuideDTO `json:"sub_guides"`
	Tags          []string      `json:"tags"`
}

type QuizChoiceDTO struct {
	ID       string `json:"id" yaml:"id"`
	Text     string `json:"text" yaml:"text"`
	Correct  bool   `json:"correct" yaml:"correct"`
	Feedback string `json:"feedback,omitempty" yaml:"feedback,omitempty"`
}

type QuizItemDTO struct {
	ID        string          `json:"id" yaml:"id"`
	Title     string          `json:"title,omitempty" yaml:"title,omitempty"`
	GuideID   string          `json:"guide_id" yaml:"guide_id"`
	Segment   string          `json:"segment" yaml:"segment"`
	Scope     string          `json:"scope" yaml:"scope"`
	Type      string          `json:"type" yaml:"type"` // e.g. "choice", "multiple_choice", "short_answer"
	Prompt    string          `json:"prompt" yaml:"prompt"`
	Choices   []QuizChoiceDTO `json:"choices" yaml:"choices"`
	CreatedAt string          `json:"created_at,omitempty" yaml:"created_at,omitempty"`
}

type QuizBankFile struct {
	PackageID string        `yaml:"package_id"`
	Title     string        `yaml:"title"`
	Items     []QuizItemDTO `yaml:"items"`
}

type LearnerMasteryDTO struct {
	GuideID        string    `json:"guide_id"`
	Segment        string    `json:"segment"`
	PMastery       float64   `json:"p_mastery"`
	StabilityDays  float64   `json:"stability_days"`
	LastReviewedAt time.Time `json:"last_reviewed_at"`
	DueForReview   bool      `json:"due_for_review"`
	AttemptsCount  int       `json:"attempts_count"`
	CorrectCount   int       `json:"correct_count"`
}

type LearnerStateStore struct {
	States map[string]LearnerMasteryDTO `json:"states"` // key: guide_id:segment
}

// -------------------------------------------------------------
// Workspace & File System Operations
// -------------------------------------------------------------

func getRecentsPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	dir := filepath.Join(configDir, "qst-studio")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "recents.json")
}

// GetRecentGraphs returns previously opened graph directory paths.
func (a *App) GetRecentGraphs() ([]string, error) {
	path := getRecentsPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []string{}, nil
	} else if err != nil {
		return nil, err
	}
	var recents []string
	if err := json.Unmarshal(data, &recents); err != nil {
		return []string{}, nil
	}
	existing := make([]string, 0, len(recents))
	for _, p := range recents {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			existing = append(existing, p)
		}
	}
	return existing, nil
}

// SaveRecentGraph persists a directory into the recent graphs list.
func (a *App) SaveRecentGraph(dirPath string) error {
	if dirPath == "" {
		return nil
	}
	recents, _ := a.GetRecentGraphs()
	updated := []string{dirPath}
	for _, p := range recents {
		if p != dirPath && len(updated) < 10 {
			updated = append(updated, p)
		}
	}
	data, err := json.MarshalIndent(updated, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getRecentsPath(), data, 0644)
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
		_ = a.SaveRecentGraph(selectedDir)
	}
	return selectedDir, nil
}

// SaveGuideContent writes raw markdown content (frontmatter + body) directly to disk.
func (a *App) SaveGuideContent(filePath string, rawContent string) error {
	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}
	return os.WriteFile(filePath, []byte(rawContent), 0644)
}

// -------------------------------------------------------------
// App Methods (Bound to Svelte Frontend)
// -------------------------------------------------------------

// LoadGraph parses and returns the full graph structure from a directory.
func (a *App) LoadGraph(dirPath string) (*GraphDTO, error) {
	manifest, err := parseManifest(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	guidesMap, err := parseGraphDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse graph guides: %w", err)
	}

	guideSummaries := make([]GuideSummaryDTO, 0, len(guidesMap))
	for _, g := range guidesMap {
		subguides := make([]SubGuideDTO, 0, len(g.Metadata.SubGuides))
		for _, sg := range g.Metadata.SubGuides {
			subguides = append(subguides, SubGuideDTO{
				Guide:     sg.Guide,
				Adherence: sg.Adherence,
				Segment:   sg.Segment,
			})
		}

		guideSummaries = append(guideSummaries, GuideSummaryDTO{
			ID:            g.ID,
			Scope:         g.Metadata.Scope,
			LineCount:     g.LineCount,
			Prerequisites: g.Metadata.Prerequisites,
			SubGuides:     subguides,
		})
	}

	tours := make([]TourDTO, 0, len(manifest.Tours))
	for _, t := range manifest.Tours {
		tours = append(tours, TourDTO{
			Name:   t.Name,
			Guides: t.Guides,
		})
	}

	_ = a.SaveRecentGraph(dirPath)

	return &GraphDTO{
		Title:            manifest.Title,
		Description:      manifest.Description,
		Scopes:           manifest.Scopes,
		Adherences:       manifest.Adherences,
		RelaxedSubguides: manifest.RelaxedSubguides,
		StrictCoverage:   manifest.StrictCoverage,
		RequireSubguides: manifest.RequireSubguides,
		Guides:           guideSummaries,
		Tours:            tours,
	}, nil
}

// GetGuide loads a specific guide's markdown body, raw content, and metadata.
func (a *App) GetGuide(dirPath string, guideID string) (*GuideDetailDTO, error) {
	guidesMap, err := parseGraphDir(dirPath)
	if err != nil {
		return nil, err
	}

	g, exists := guidesMap[guideID]
	if !exists {
		return nil, fmt.Errorf("guide %q not found", guideID)
	}

	subguides := make([]SubGuideDTO, 0, len(g.Metadata.SubGuides))
	for _, sg := range g.Metadata.SubGuides {
		subguides = append(subguides, SubGuideDTO{
			Guide:     sg.Guide,
			Adherence: sg.Adherence,
			Segment:   sg.Segment,
		})
	}

	return &GuideDetailDTO{
		ID:            g.ID,
		Scope:         g.Metadata.Scope,
		Path:          g.Path,
		Body:          g.Body,
		RawContent:    g.RawContent,
		LineCount:     g.LineCount,
		Prerequisites: g.Metadata.Prerequisites,
		SubGuides:     subguides,
		Tags:          g.Metadata.Tags,
	}, nil
}

// GetLearningPath computes the prerequisite sequence leading to a target guide.
func (a *App) GetLearningPath(dirPath string, targetID string) ([]string, error) {
	guidesMap, err := parseGraphDir(dirPath)
	if err != nil {
		return nil, err
	}

	return computePrerequisitePath(targetID, guidesMap)
}

// -------------------------------------------------------------
// Quiz Bank & Items (YAML storage)
// -------------------------------------------------------------

func getItemsFilePath(dirPath string) string {
	return filepath.Join(dirPath, "items.yaml")
}

// GetQuizItems returns all saved quiz items, optionally filtered by guideID.
func (a *App) GetQuizItems(dirPath string, guideID string) ([]QuizItemDTO, error) {
	filePath := getItemsFilePath(dirPath)
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return []QuizItemDTO{}, nil
	} else if err != nil {
		return nil, err
	}

	var bank QuizBankFile
	if err := yaml.Unmarshal(data, &bank); err != nil {
		return nil, fmt.Errorf("failed to parse items.yaml: %w", err)
	}

	if guideID == "" {
		return bank.Items, nil
	}

	filtered := make([]QuizItemDTO, 0)
	for _, item := range bank.Items {
		if item.GuideID == guideID {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

// SaveQuizItem saves or updates a quiz item in items.yaml.
func (a *App) SaveQuizItem(dirPath string, item QuizItemDTO) error {
	filePath := getItemsFilePath(dirPath)
	var bank QuizBankFile

	data, err := os.ReadFile(filePath)
	if err == nil {
		_ = yaml.Unmarshal(data, &bank)
	}

	if bank.PackageID == "" {
		bank.PackageID = filepath.Base(dirPath) + "-assessments"
		bank.Title = filepath.Base(dirPath) + " Question Bank"
	}

	if item.CreatedAt == "" {
		item.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	// Upsert item
	found := false
	for i, existing := range bank.Items {
		if existing.ID == item.ID {
			bank.Items[i] = item
			found = true
			break
		}
	}
	if !found {
		bank.Items = append(bank.Items, item)
	}

	out, err := yaml.Marshal(&bank)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, out, 0644)
}

// -------------------------------------------------------------
// Continual Practice & Learner Mastery (BKT + FSRS)
// -------------------------------------------------------------

func getLearnerStatePath(dirPath string) string {
	return filepath.Join(dirPath, ".learner_state.json")
}

func (a *App) loadLearnerState(dirPath string) (*LearnerStateStore, error) {
	filePath := getLearnerStatePath(dirPath)
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return &LearnerStateStore{States: make(map[string]LearnerMasteryDTO)}, nil
	} else if err != nil {
		return nil, err
	}

	var store LearnerStateStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	if store.States == nil {
		store.States = make(map[string]LearnerMasteryDTO)
	}
	return &store, nil
}

func (a *App) saveLearnerState(dirPath string, store *LearnerStateStore) error {
	filePath := getLearnerStatePath(dirPath)
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func applyBKTUpdate(state LearnerMasteryDTO, guideID string, segment string, correct bool, exists bool) LearnerMasteryDTO {
	if !exists {
		state = LearnerMasteryDTO{
			GuideID:        guideID,
			Segment:        segment,
			PMastery:       0.10, // Initial prior P(L0)
			StabilityDays:  1.0,  // Initial stability 1 day
			LastReviewedAt: time.Now(),
		}
	}

	// Bayesian Knowledge Tracing (BKT) Parameters
	const (
		pTransit = 0.20 // P(T)
		pGuess   = 0.25 // P(G)
		pSlip    = 0.10 // P(S)
	)

	pKnown := state.PMastery
	var pPosterior float64
	if correct {
		pPosterior = (pKnown * (1.0 - pSlip)) / (pKnown*(1.0-pSlip) + (1.0-pKnown)*pGuess)
		state.CorrectCount++
		state.StabilityDays = math.Max(1.0, state.StabilityDays*2.2)
	} else {
		pPosterior = (pKnown * pSlip) / (pKnown*pSlip + (1.0-pKnown)*(1.0-pGuess))
		state.StabilityDays = math.Max(0.5, state.StabilityDays*0.5)
	}

	state.PMastery = pPosterior + (1.0-pPosterior)*pTransit
	if state.PMastery > 0.999 {
		state.PMastery = 0.999
	}
	if state.PMastery < 0.001 {
		state.PMastery = 0.001
	}

	state.LastReviewedAt = time.Now()
	state.AttemptsCount++
	state.DueForReview = false
	return state
}

// RecordPracticeResult updates BKT probability and FSRS spaced repetition stability.
func (a *App) RecordPracticeResult(dirPath string, guideID string, segment string, correct bool) (*LearnerMasteryDTO, error) {
	store, err := a.loadLearnerState(dirPath)
	if err != nil {
		return nil, err
	}

	if segment == "" {
		segment = "ALL"
	}
	key := fmt.Sprintf("%s:%s", guideID, segment)
	prev, exists := store.States[key]
	state := applyBKTUpdate(prev, guideID, segment, correct, exists)
	store.States[key] = state

	// Roll up segment practice to guide-level mastery ("guideID:ALL")
	if segment != "ALL" {
		allKey := fmt.Sprintf("%s:ALL", guideID)
		allPrev, allExists := store.States[allKey]
		store.States[allKey] = applyBKTUpdate(allPrev, guideID, "ALL", correct, allExists)
	}

	if err := a.saveLearnerState(dirPath, store); err != nil {
		return nil, err
	}

	return &state, nil
}

// GetLearnerStates retrieves all mastery states, evaluating whether each is due for review.
func (a *App) GetLearnerStates(dirPath string) (map[string]LearnerMasteryDTO, error) {
	store, err := a.loadLearnerState(dirPath)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for k, st := range store.States {
		daysSinceReview := now.Sub(st.LastReviewedAt).Hours() / 24.0
		// Ebbinghaus memory retrievability R(t) = exp(-t / S)
		retrievability := math.Exp(-daysSinceReview / math.Max(0.1, st.StabilityDays))
		// Due for review if retrievability drops below 85%
		st.DueForReview = retrievability < 0.85
		store.States[k] = st
	}

	return store.States, nil
}
