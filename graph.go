package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type SubGuideRelation struct {
	Guide     string `yaml:"guide"`
	Adherence string `yaml:"adherence"`
	Segment   string `yaml:"segment"`
}

type GuideMetadata struct {
	Prerequisites []string           `yaml:"prerequisites"`
	SubGuides     []SubGuideRelation `yaml:"sub_guides"`
	Scope         string             `yaml:"scope"`
	Tags          []string           `yaml:"tags"`
}

type Guide struct {
	ID         string
	Path       string
	Metadata   GuideMetadata
	LineCount  int
	Body       string
	RawContent string
}

type Manifest struct {
	Title            string       `yaml:"title"`
	Description      string       `yaml:"description"`
	Scopes           []string     `yaml:"scopes"`
	Adherences       []string     `yaml:"adherences"`
	RelaxedSubguides bool         `yaml:"relaxed_subguides"`
	StrictCoverage   bool         `yaml:"strict_coverage"`
	RequireSubguides bool         `yaml:"require_subguides"`
	Tours            []TourConfig `yaml:"tours"`
}

type TourConfig struct {
	Name   string   `yaml:"name"`
	Guides []string `yaml:"guides"`
}

func parseManifest(dirPath string) (*Manifest, error) {
	manifestPath := filepath.Join(dirPath, "manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		// Fallback to quest.yaml
		manifestPath = filepath.Join(dirPath, "quest.yaml")
		data, err = os.ReadFile(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("no manifest.yaml or quest.yaml found in %s", dirPath)
		}
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	return &m, nil
}

func parseGuideFile(filePath string) (*Guide, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	contentStr := strings.ReplaceAll(string(content), "\r\n", "\n")
	lines := strings.Split(contentStr, "\n")

	inFrontmatter := false
	frontmatterLines := []string{}
	bodyStart := -1

	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				inFrontmatter = false
				bodyStart = i + 1
				break
			}
		}
		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		}
	}

	var meta GuideMetadata
	if len(frontmatterLines) > 0 {
		fmRaw := strings.Join(frontmatterLines, "\n")
		if err := yaml.Unmarshal([]byte(fmRaw), &meta); err != nil {
			return nil, fmt.Errorf("invalid frontmatter in %s: %w", filePath, err)
		}
	}

	bodyLines := []string{}
	if bodyStart != -1 && bodyStart < len(lines) {
		bodyLines = lines[bodyStart:]
	}

	id := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	return &Guide{
		ID:         id,
		Path:       filePath,
		Metadata:   meta,
		LineCount:  len(bodyLines),
		Body:       strings.Join(bodyLines, "\n"),
		RawContent: string(content),
	}, nil
}

func parseGraphDir(dirPath string) (map[string]Guide, error) {
	guides := make(map[string]Guide)

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

// computePrerequisitePath finds all upstream prerequisites in topological order using DFS
func computePrerequisitePath(targetID string, guides map[string]Guide) ([]string, error) {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	order := []string{}

	var dfs func(id string) error
	dfs = func(id string) error {
		if inStack[id] {
			return fmt.Errorf("cycle detected involving %s", id)
		}
		if visited[id] {
			return nil
		}

		inStack[id] = true
		guide, exists := guides[id]
		if exists {
			for _, prereq := range guide.Metadata.Prerequisites {
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
