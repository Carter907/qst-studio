package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type SubGuideDTO struct {
	Guide     string `json:"guide" yaml:"guide"`
	Adherence string `json:"adherence" yaml:"adherence"`
	Segment   string `json:"segment" yaml:"segment"`
}

type GuideSummaryDTO struct {
	ID            string        `json:"id"`
	Scope         string        `json:"scope"`
	LineCount     int           `json:"line_count"`
	Prerequisites []string      `json:"prerequisites"`
	SubGuides     []SubGuideDTO `json:"sub_guides"`
}

type TourDTO struct {
	Name   string   `json:"name" yaml:"name"`
	Guides []string `json:"guides" yaml:"guides"`
}

type GraphDTO struct {
	Title            string            `json:"title" yaml:"title"`
	Description      string            `json:"description" yaml:"description"`
	Scopes           []string          `json:"scopes" yaml:"scopes"`
	Adherences       []string          `json:"adherences" yaml:"adherences"`
	RelaxedSubguides bool              `json:"relaxed_subguides" yaml:"relaxed_subguides"`
	StrictCoverage   bool              `json:"strict_coverage" yaml:"strict_coverage"`
	RequireSubguides bool              `json:"require_subguides" yaml:"require_subguides"`
	Guides           []GuideSummaryDTO `json:"guides" yaml:"-"`
	Tours            []TourDTO         `json:"tours" yaml:"tours"`
}

type GuideDetailDTO struct {
	ID            string        `json:"id" yaml:"-"`
	Scope         string        `json:"scope" yaml:"scope"`
	Path          string        `json:"path" yaml:"-"`
	Body          string        `json:"body" yaml:"-"`
	RawContent    string        `json:"raw_content" yaml:"-"`
	LineCount     int           `json:"line_count" yaml:"-"`
	Prerequisites []string      `json:"prerequisites" yaml:"prerequisites"`
	SubGuides     []SubGuideDTO `json:"sub_guides" yaml:"sub_guides"`
	Tags          []string      `json:"tags" yaml:"tags"`
}

// Load parses and returns the full graph structure from a directory.
func Load(dirPath string) (*GraphDTO, error) {
	graph, err := parseManifest(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	guidesMap, err := parseGraphDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse graph guides: %w", err)
	}

	summaries := make([]GuideSummaryDTO, 0, len(guidesMap))
	for _, g := range guidesMap {
		summaries = append(summaries, GuideSummaryDTO{
			ID:            g.ID,
			Scope:         g.Scope,
			LineCount:     g.LineCount,
			Prerequisites: g.Prerequisites,
			SubGuides:     g.SubGuides,
		})
	}
	graph.Guides = summaries
	return graph, nil
}

// GetGuide loads a specific guide's markdown body, raw content, and metadata.
func GetGuide(dirPath string, guideID string) (*GuideDetailDTO, error) {
	guidesMap, err := parseGraphDir(dirPath)
	if err != nil {
		return nil, err
	}

	g, exists := guidesMap[guideID]
	if !exists {
		return nil, fmt.Errorf("guide %q not found", guideID)
	}
	return &g, nil
}

// GetLearningPath computes the prerequisite sequence leading to a target guide.
func GetLearningPath(dirPath string, targetID string) ([]string, error) {
	guidesMap, err := parseGraphDir(dirPath)
	if err != nil {
		return nil, err
	}
	return computePrerequisitePath(targetID, guidesMap)
}

// SaveGuideContent writes raw markdown content (frontmatter + body) directly to disk.
func SaveGuideContent(filePath string, rawContent string) error {
	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}
	return os.WriteFile(filePath, []byte(rawContent), 0644)
}

func parseManifest(dirPath string) (*GraphDTO, error) {
	manifestCandidates := []string{"qst.yaml", "manifest.yaml", "quest.yaml"}
	var data []byte
	var err error
	for _, cand := range manifestCandidates {
		data, err = os.ReadFile(filepath.Join(dirPath, cand))
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("no qst.yaml, manifest.yaml, or quest.yaml found in %s", dirPath)
	}

	var graph GraphDTO
	if err := yaml.Unmarshal(data, &graph); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	if graph.Tours == nil {
		graph.Tours = []TourDTO{}
	}
	return &graph, nil
}

func parseGuideFile(filePath string) (*GuideDetailDTO, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	contentStr := strings.ReplaceAll(string(content), "\r\n", "\n")
	lines := strings.Split(contentStr, "\n")

	inFrontmatter := false
	var frontmatterLines []string
	bodyStart := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			bodyStart = i + 1
			break
		}
		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		}
	}

	var guide GuideDetailDTO
	if len(frontmatterLines) > 0 {
		fmRaw := strings.Join(frontmatterLines, "\n")
		if err := yaml.Unmarshal([]byte(fmRaw), &guide); err != nil {
			return nil, fmt.Errorf("invalid frontmatter in %s: %w", filePath, err)
		}
	}

	var bodyLines []string
	if bodyStart != -1 && bodyStart < len(lines) {
		bodyLines = lines[bodyStart:]
	}
	if guide.SubGuides == nil {
		guide.SubGuides = []SubGuideDTO{}
	}

	guide.ID = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	guide.Path = filePath
	guide.LineCount = len(bodyLines)
	guide.Body = strings.Join(bodyLines, "\n")
	guide.RawContent = string(content)
	return &guide, nil
}

func parseGraphDir(dirPath string) (map[string]GuideDetailDTO, error) {
	guides := make(map[string]GuideDetailDTO)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "public" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			g, err := parseGuideFile(path)
			if err != nil {
				return err
			}
			guides[g.ID] = *g
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return guides, nil
}

// computePrerequisitePath finds all upstream prerequisites in topological order using DFS.
func computePrerequisitePath(targetID string, guides map[string]GuideDetailDTO) ([]string, error) {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var order []string

	var dfs func(id string) error
	dfs = func(id string) error {
		if inStack[id] {
			return fmt.Errorf("cycle detected involving %s", id)
		}
		if visited[id] {
			return nil
		}

		inStack[id] = true
		if guide, exists := guides[id]; exists {
			for _, prereq := range guide.Prerequisites {
				if err := dfs(prereq); err != nil {
					return err
				}
			}
		}

		inStack[id] = false
		visited[id] = true
		order = append(order, id)
		return nil
	}

	if err := dfs(targetID); err != nil {
		return nil, err
	}
	return order, nil
}
