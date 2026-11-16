package quizzes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"qst-studio/internal/quiz"

	"gopkg.in/yaml.v3"
)

type StandaloneQuizSummaryDTO struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	FilePath      string   `json:"file_path"`
	Tags          []string `json:"tags"`
	RelatedGuides []string `json:"related_guides"`
	ItemCount     int      `json:"item_count"`
}

type StandaloneQuizDetailDTO struct {
	ID            string             `json:"id"`
	Title         string             `json:"title"`
	Description   string             `json:"description"`
	FilePath      string             `json:"file_path"`
	Tags          []string           `json:"tags"`
	RelatedGuides []string           `json:"related_guides"`
	Items         []quiz.QuizItemDTO `json:"items"`
}

type standaloneQuizFile struct {
	ID            string             `yaml:"id"`
	Title         string             `yaml:"title"`
	Description   string             `yaml:"description"`
	Tags          []string           `yaml:"tags"`
	RelatedGuides []string           `yaml:"related_guides"`
	Items         []quiz.QuizItemDTO `yaml:"items"`
}

func quizzesDir(dirPath string) string {
	return filepath.Join(dirPath, "quizzes")
}

func loadQuizFile(rootDir string, filePath string) (*StandaloneQuizDetailDTO, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var raw standaloneQuizFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse quiz %s: %w", filePath, err)
	}

	relPath, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		relPath = filepath.Base(filePath)
	}
	stem := strings.TrimSuffix(filepath.ToSlash(relPath), filepath.Ext(relPath))

	quizID := strings.TrimSpace(raw.ID)
	if quizID == "" {
		quizID = stem
	}

	title := strings.TrimSpace(raw.Title)
	if title == "" {
		title = quizID
	}

	tags := raw.Tags
	if tags == nil {
		tags = []string{}
	}
	relatedGuides := raw.RelatedGuides
	if relatedGuides == nil {
		relatedGuides = []string{}
	}
	items := raw.Items
	if items == nil {
		items = []quiz.QuizItemDTO{}
	}

	quiz.NormalizeLoadedItems(items)
	seenItemIDs := make(map[string]bool, len(items))
	for i := range items {
		items[i].ID = strings.TrimSpace(items[i].ID)
		if items[i].ID == "" {
			items[i].ID = fmt.Sprintf("%s__%d", quizID, i+1)
		}
		if seenItemIDs[items[i].ID] {
			return nil, fmt.Errorf("quiz %q (%s): duplicate item id %q", quizID, relPath, items[i].ID)
		}
		seenItemIDs[items[i].ID] = true

		if err := quiz.ValidateAndNormalizeItem(&items[i]); err != nil {
			return nil, fmt.Errorf("quiz %q (%s), item %q: %w", quizID, relPath, items[i].ID, err)
		}
	}

	return &StandaloneQuizDetailDTO{
		ID:            quizID,
		Title:         title,
		Description:   strings.TrimSpace(raw.Description),
		FilePath:      filePath,
		Tags:          tags,
		RelatedGuides: relatedGuides,
		Items:         items,
	}, nil
}

func loadAll(dirPath string) ([]StandaloneQuizDetailDTO, error) {
	qDir := quizzesDir(dirPath)
	info, err := os.Stat(qDir)
	if os.IsNotExist(err) || (err == nil && !info.IsDir()) {
		return []StandaloneQuizDetailDTO{}, nil
	} else if err != nil {
		return nil, err
	}

	results := make([]StandaloneQuizDetailDTO, 0)
	seenIDs := make(map[string]string)

	err = filepath.WalkDir(qDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := d.Name()
		if d.IsDir() {
			if path != qDir && strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(name, ".") {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		detail, loadErr := loadQuizFile(qDir, path)
		if loadErr != nil {
			return loadErr
		}
		if prevPath, exists := seenIDs[detail.ID]; exists {
			return fmt.Errorf("duplicate standalone quiz id %q in %s and %s", detail.ID, prevPath, path)
		}
		seenIDs[detail.ID] = path
		results = append(results, *detail)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		li := strings.ToLower(results[i].Title)
		lj := strings.ToLower(results[j].Title)
		if li != lj {
			return li < lj
		}
		return results[i].ID < results[j].ID
	})

	return results, nil
}

// List returns summaries for all standalone quizzes in <dirPath>/quizzes/.
func List(dirPath string) ([]StandaloneQuizSummaryDTO, error) {
	all, err := loadAll(dirPath)
	if err != nil {
		return nil, err
	}
	summaries := make([]StandaloneQuizSummaryDTO, 0, len(all))
	for _, q := range all {
		summaries = append(summaries, StandaloneQuizSummaryDTO{
			ID:            q.ID,
			Title:         q.Title,
			Description:   q.Description,
			FilePath:      q.FilePath,
			Tags:          q.Tags,
			RelatedGuides: q.RelatedGuides,
			ItemCount:     len(q.Items),
		})
	}
	return summaries, nil
}

// Get loads and validates a specific standalone quiz by ID from <dirPath>/quizzes/.
func Get(dirPath string, quizID string) (*StandaloneQuizDetailDTO, error) {
	all, err := loadAll(dirPath)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].ID == quizID {
			return &all[i], nil
		}
	}
	return nil, fmt.Errorf("standalone quiz %q not found", quizID)
}
