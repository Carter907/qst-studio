package quiz

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"
)

// NormalizeChoiceKey returns a canonical key for detecting duplicate answer choices.
func NormalizeChoiceKey(text string) string {
	s := strings.TrimSpace(text)
	s = strings.Trim(s, "`'\"$")
	s = strings.ToLower(strings.TrimSpace(s))

	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteRune(' ')
				prevSpace = true
			}
		} else {
			b.WriteRune(r)
			prevSpace = false
		}
	}
	return b.String()
}

// ParseLineRange parses a "start-end" segment string into 1-based inclusive [start, end] bounds.
func ParseLineRange(segment string, lineCount int) (int, int, bool) {
	if lineCount < 1 {
		lineCount = 1
	}
	segment = strings.TrimSpace(segment)
	if segment == "" || strings.EqualFold(segment, "ALL") {
		return 1, lineCount, false
	}
	if segment == "1-1" && lineCount > 1 {
		return 1, lineCount, false
	}

	parts := strings.Split(segment, "-")
	start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	end := start
	var err2 error
	if len(parts) > 1 {
		end, err2 = strconv.Atoi(strings.TrimSpace(parts[1]))
	}
	if err1 != nil || err2 != nil || start < 1 || start > lineCount {
		return 1, lineCount, false
	}
	if end < start {
		start, end = end, start
	}
	if end > lineCount {
		end = lineCount
	}
	return start, end, true
}

// NormalizeGuideSegment normalizes empty, "ALL", or default "1-1" (on multi-line guides) to "1-<lineCount>".
func NormalizeGuideSegment(segment string, lineCount int) string {
	start, end, _ := ParseLineRange(segment, lineCount)
	return fmt.Sprintf("%d-%d", start, end)
}

type QuizChoiceDTO struct {
	ID       string `json:"id" yaml:"id"`
	Text     string `json:"text" yaml:"text"`
	Correct  bool   `json:"correct" yaml:"correct"`
	Feedback string `json:"feedback,omitempty" yaml:"feedback,omitempty"`
}

type BlankDTO struct {
	Type    string   `json:"type" yaml:"type"`                           // "text" or "dropdown"
	Answers []string `json:"answers,omitempty" yaml:"answers,omitempty"` // for text blanks
	Options []string `json:"options,omitempty" yaml:"options,omitempty"` // for dropdowns
	Correct string   `json:"correct,omitempty" yaml:"correct,omitempty"` // correct dropdown option
}

type MatchingPairDTO struct {
	Prompt string `json:"prompt" yaml:"prompt"` // Left prompt
	Target string `json:"target" yaml:"target"` // Right matching target
}

type QuizItemDTO struct {
	ID          string              `json:"id" yaml:"id"`
	Title       string              `json:"title,omitempty" yaml:"title,omitempty"`
	GuideID     string              `json:"guide_id" yaml:"guide_id"`
	Segment     string              `json:"segment" yaml:"segment"`
	Scope       string              `json:"scope" yaml:"scope"`
	Type        string              `json:"type" yaml:"type"`
	Points      float64             `json:"points,omitempty" yaml:"points,omitempty"`
	Prompt      string              `json:"prompt" yaml:"prompt"`
	Feedback    string              `json:"feedback,omitempty" yaml:"feedback,omitempty"`
	Choices     []QuizChoiceDTO     `json:"choices,omitempty" yaml:"choices,omitempty"`
	Blanks      map[string]BlankDTO `json:"blanks,omitempty" yaml:"blanks,omitempty"`
	Pairs       []MatchingPairDTO   `json:"pairs,omitempty" yaml:"pairs,omitempty"`
	Distractors []string            `json:"distractors,omitempty" yaml:"distractors,omitempty"`
	Items       []string            `json:"items,omitempty" yaml:"items,omitempty"`
	TopLabel    string              `json:"top_label,omitempty" yaml:"top_label,omitempty"`
	BottomLabel string              `json:"bottom_label,omitempty" yaml:"bottom_label,omitempty"`
	Solution    string              `json:"solution,omitempty" yaml:"solution,omitempty"`
	CreatedAt   string              `json:"created_at,omitempty" yaml:"created_at,omitempty"`
}

type QuizBankFile struct {
	PackageID string        `yaml:"package_id"`
	Title     string        `yaml:"title"`
	Items     []QuizItemDTO `yaml:"items"`
}

type PracticeDeckOptions struct {
	Scope         string   `json:"scope"`
	AnchorGuideID string   `json:"anchor_guide_id"`
	TourName      string   `json:"tour_name"`
	FilterScopes  []string `json:"filter_scopes"`
}

func getItemsFilePath(dirPath string) string {
	return filepath.Join(dirPath, "items.yaml")
}

// GetItems returns all saved quiz items, optionally filtered by guideID.
func GetItems(dirPath string, guideID string) ([]QuizItemDTO, error) {
	data, err := os.ReadFile(getItemsFilePath(dirPath))
	if os.IsNotExist(err) {
		return []QuizItemDTO{}, nil
	} else if err != nil {
		return nil, err
	}

	var bank QuizBankFile
	if err := yaml.Unmarshal(data, &bank); err != nil {
		return nil, fmt.Errorf("failed to parse items.yaml: %w", err)
	}

	NormalizeLoadedItems(bank.Items)

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

// NormalizeLoadedItems populates default item types and choice IDs for loaded quiz items.
func NormalizeLoadedItems(items []QuizItemDTO) {
	for i := range items {
		if items[i].Type == "" || items[i].Type == "choice" {
			items[i].Type = "multiple_choice"
		}
		for j := range items[i].Choices {
			if strings.TrimSpace(items[i].Choices[j].ID) == "" {
				items[i].Choices[j].ID = fmt.Sprintf("c%d", j+1)
			}
		}
	}
}

// GetItemsForGuides returns saved quiz items matching guideIDs, ordered by guideIDs then items.yaml order.
func GetItemsForGuides(dirPath string, guideIDs []string) ([]QuizItemDTO, error) {
	if len(guideIDs) == 0 {
		return []QuizItemDTO{}, nil
	}

	allItems, err := GetItems(dirPath, "")
	if err != nil {
		return nil, err
	}

	byGuide := make(map[string][]QuizItemDTO)
	for _, item := range allItems {
		byGuide[item.GuideID] = append(byGuide[item.GuideID], item)
	}

	result := make([]QuizItemDTO, 0)
	seen := make(map[string]bool, len(guideIDs))
	for _, gid := range guideIDs {
		if seen[gid] {
			continue
		}
		seen[gid] = true
		result = append(result, byGuide[gid]...)
	}
	return result, nil
}

// FilterItemsByScopes filters quiz items by allowed manifest scopes, falling back to guideScopes when item.Scope is empty.
func FilterItemsByScopes(items []QuizItemDTO, allowedScopes []string, guideScopes map[string]string) []QuizItemDTO {
	if len(allowedScopes) == 0 {
		return items
	}

	allowed := make(map[string]bool, len(allowedScopes))
	for _, sc := range allowedScopes {
		trimmed := strings.TrimSpace(sc)
		if trimmed != "" {
			allowed[trimmed] = true
		}
	}
	if len(allowed) == 0 {
		return items
	}

	filtered := make([]QuizItemDTO, 0, len(items))
	for _, item := range items {
		effScope := strings.TrimSpace(item.Scope)
		if effScope == "" && guideScopes != nil {
			effScope = strings.TrimSpace(guideScopes[item.GuideID])
		}
		if allowed[effScope] {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// SaveItem validates and saves or updates a quiz item in items.yaml.
func SaveItem(dirPath string, item QuizItemDTO) error {
	if err := ValidateAndNormalizeItem(&item); err != nil {
		return err
	}

	filePath := getItemsFilePath(dirPath)
	var bank QuizBankFile

	if data, err := os.ReadFile(filePath); err == nil {
		_ = yaml.Unmarshal(data, &bank)
	}

	if bank.PackageID == "" {
		base := filepath.Base(dirPath)
		bank.PackageID = base + "-assessments"
		bank.Title = base + " Question Bank"
	}

	if item.CreatedAt == "" {
		item.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

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

// ValidateAndNormalizeItem validates a quiz item's prompt, type, and type-specific fields, normalizing defaults.
func ValidateAndNormalizeItem(item *QuizItemDTO) error {
	if strings.TrimSpace(item.Prompt) == "" {
		return fmt.Errorf("question prompt cannot be empty")
	}
	if item.Type == "" || item.Type == "choice" {
		item.Type = "multiple_choice"
	}
	switch item.Type {
	case "multiple_choice", "true_false":
		validChoices, err := validateChoices(item.Choices, true)
		if err != nil {
			return err
		}
		item.Choices = validChoices

	case "multiple_answers":
		validChoices, err := validateChoices(item.Choices, false)
		if err != nil {
			return err
		}
		item.Choices = validChoices

	case "fill_in_the_blank":
		return validateBlanks(item.Blanks)

	case "matching":
		cleanPairs, err := validatePairs(item.Pairs)
		if err != nil {
			return err
		}
		item.Pairs = cleanPairs

	case "ordering":
		cleanItems, err := validateOrdering(item.Items)
		if err != nil {
			return err
		}
		item.Items = cleanItems

	case "essay":
		// Only requires non-empty prompt (checked by caller)

	default:
		return fmt.Errorf("unsupported question type: %q", item.Type)
	}
	return nil
}

func validateBlanks(blanks map[string]BlankDTO) error {
	if len(blanks) == 0 {
		return fmt.Errorf("fill_in_the_blank questions require at least one blank token (e.g. [b1] or [d1])")
	}
	for tok, blk := range blanks {
		if strings.TrimSpace(tok) == "" {
			return fmt.Errorf("blank token cannot be empty")
		}
		if blk.Type == "dropdown" {
			if len(blk.Options) < 2 {
				return fmt.Errorf("dropdown [%s] requires at least 2 options", tok)
			}
			if strings.TrimSpace(blk.Correct) == "" {
				return fmt.Errorf("dropdown [%s] requires a correct option selected", tok)
			}
			continue
		}
		cleanAnswers := make([]string, 0)
		for _, ans := range blk.Answers {
			if t := strings.TrimSpace(ans); t != "" {
				cleanAnswers = append(cleanAnswers, t)
			}
		}
		if len(cleanAnswers) == 0 {
			return fmt.Errorf("blank [%s] requires at least one acceptable answer string", tok)
		}
		blk.Answers = cleanAnswers
		blanks[tok] = blk
	}
	return nil
}

func validatePairs(pairs []MatchingPairDTO) ([]MatchingPairDTO, error) {
	if len(pairs) < 2 {
		return nil, fmt.Errorf("matching questions require at least 2 pairs")
	}
	cleanPairs := make([]MatchingPairDTO, 0, len(pairs))
	for _, p := range pairs {
		prompt := strings.TrimSpace(p.Prompt)
		target := strings.TrimSpace(p.Target)
		if prompt == "" || target == "" {
			continue
		}
		cleanPairs = append(cleanPairs, MatchingPairDTO{Prompt: prompt, Target: target})
	}
	if len(cleanPairs) < 2 {
		return nil, fmt.Errorf("matching questions require at least 2 complete pairs with non-empty prompt and target")
	}
	return cleanPairs, nil
}

func validateOrdering(items []string) ([]string, error) {
	cleanItems := make([]string, 0, len(items))
	for _, it := range items {
		if t := strings.TrimSpace(it); t != "" {
			cleanItems = append(cleanItems, t)
		}
	}
	if len(cleanItems) < 2 {
		return nil, fmt.Errorf("ordering questions require at least 2 items to sequence")
	}
	return cleanItems, nil
}

func validateChoices(choices []QuizChoiceDTO, singleCorrect bool) ([]QuizChoiceDTO, error) {
	validChoices := make([]QuizChoiceDTO, 0, len(choices))
	seenKeys := make(map[string]bool)
	correctCount := 0

	for _, c := range choices {
		text := strings.TrimSpace(c.Text)
		if text == "" {
			continue
		}
		key := NormalizeChoiceKey(text)
		if seenKeys[key] {
			return nil, fmt.Errorf("duplicate answer choice text %q: all choices must have distinct text", text)
		}
		seenKeys[key] = true

		isCorrect := c.Correct
		if singleCorrect && isCorrect && correctCount >= 1 {
			isCorrect = false
		}
		if isCorrect {
			correctCount++
		}

		validChoices = append(validChoices, QuizChoiceDTO{
			ID:       fmt.Sprintf("c%d", len(validChoices)+1),
			Text:     text,
			Correct:  isCorrect,
			Feedback: strings.TrimSpace(c.Feedback),
		})
	}

	if len(validChoices) < 2 {
		return nil, fmt.Errorf("question must have at least 2 non-empty choices")
	}
	if correctCount == 0 {
		return nil, fmt.Errorf("question must have at least one choice marked as correct")
	}
	return validChoices, nil
}

// DeleteItem removes a saved quiz item by ID from items.yaml.
func DeleteItem(dirPath string, itemID string) error {
	filePath := getItemsFilePath(dirPath)
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	var bank QuizBankFile
	if err := yaml.Unmarshal(data, &bank); err != nil {
		return fmt.Errorf("failed to parse items.yaml: %w", err)
	}

	filtered := make([]QuizItemDTO, 0, len(bank.Items))
	for _, it := range bank.Items {
		if it.ID != itemID {
			filtered = append(filtered, it)
		}
	}
	bank.Items = filtered

	out, err := yaml.Marshal(&bank)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, out, 0644)
}
