package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultOllamaEndpoint = "http://localhost:11434"

var ollamaGenMu sync.Mutex

// AgentConfigDTO represents the persisted local agent configuration.
type AgentConfigDTO struct {
	Endpoint string `json:"endpoint"`
	Model    string `json:"model"`
}

// AgentStatusDTO represents the live status and discovered models from Ollama.
type AgentStatusDTO struct {
	Available   bool     `json:"available"`
	Endpoint    string   `json:"endpoint"`
	ActiveModel string   `json:"active_model"`
	Models      []string `json:"models"`
	Error       string   `json:"error,omitempty"`
}

type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

type ollamaChatMessage struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	Thinking string `json:"thinking,omitempty"`
}

type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []ollamaChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
	Think    *bool               `json:"think,omitempty"`
	Format   any                 `json:"format,omitempty"`
	Options  map[string]any      `json:"options,omitempty"`
}

type ollamaChatResponse struct {
	Message ollamaChatMessage `json:"message"`
	Done    bool              `json:"done,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type generatedQuestionPayload struct {
	Title   string `json:"title"`
	Prompt  string `json:"prompt"`
	Choices []struct {
		Text     string `json:"text"`
		Correct  bool   `json:"correct"`
		Feedback string `json:"feedback"`
	} `json:"choices"`
}

func getAgentConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	dir := filepath.Join(configDir, "qst-studio")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "agent_config.json")
}

func loadAgentConfig() AgentConfigDTO {
	cfg := AgentConfigDTO{
		Endpoint: defaultOllamaEndpoint,
	}
	data, err := os.ReadFile(getAgentConfigPath())
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}
	cfg.Endpoint = strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/")
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultOllamaEndpoint
	}
	return cfg
}

// SaveAgentConfig persists the Ollama endpoint and active model to disk.
func (a *App) SaveAgentConfig(cfg AgentConfigDTO) error {
	cfg.Endpoint = strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/")
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultOllamaEndpoint
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getAgentConfigPath(), data, 0644)
}

// GetAgentStatus checks Ollama connectivity and discovers installed local models via /api/tags.
func (a *App) GetAgentStatus() (*AgentStatusDTO, error) {
	cfg := loadAgentConfig()
	status := &AgentStatusDTO{
		Available:   false,
		Endpoint:    cfg.Endpoint,
		ActiveModel: cfg.Model,
		Models:      []string{},
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(cfg.Endpoint + "/api/tags")
	if err != nil {
		status.Error = fmt.Sprintf("Ollama unreachable at %s", cfg.Endpoint)
		return status, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		status.Error = fmt.Sprintf("Ollama returned HTTP %d", resp.StatusCode)
		return status, nil
	}

	var tags ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		status.Error = "Failed to decode /api/tags response"
		return status, nil
	}

	for _, m := range tags.Models {
		if m.Name != "" {
			status.Models = append(status.Models, m.Name)
		}
	}

	status.Available = true
	if len(status.Models) > 0 {
		found := false
		for _, m := range status.Models {
			if m == status.ActiveModel {
				found = true
				break
			}
		}
		if !found {
			status.ActiveModel = status.Models[0]
			cfg.Model = status.ActiveModel
			_ = a.SaveAgentConfig(cfg)
		}
	}

	return status, nil
}

// extractSegmentLines extracts 1-based inclusive lines from a guide body.
// If the segment is "1-1", empty, or produces blank text, it falls back to the full body.
func extractSegmentLines(body string, segment string) string {
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 {
		return ""
	}

	segment = strings.TrimSpace(segment)
	if segment == "" || segment == "ALL" || segment == "1-1" {
		if segment == "1-1" && len(lines) >= 1 && strings.TrimSpace(lines[0]) != "" && len(lines) == 1 {
			return lines[0]
		}
		if segment != "1-1" {
			return strings.TrimSpace(normalized)
		}
	}

	parts := strings.Split(segment, "-")
	start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	end := start
	var err2 error
	if len(parts) > 1 {
		end, err2 = strconv.Atoi(strings.TrimSpace(parts[1]))
	}
	if err1 != nil || err2 != nil || start < 1 {
		return strings.TrimSpace(normalized)
	}
	if end < start {
		start, end = end, start
	}
	if start > len(lines) {
		return strings.TrimSpace(normalized)
	}
	if end > len(lines) {
		end = len(lines)
	}

	// If default "1-1" is still unchanged and the guide has multiple lines, use the full body
	// unless line 1 is substantial on its own.
	if start == 1 && end == 1 && len(lines) > 1 {
		return strings.TrimSpace(normalized)
	}

	extracted := strings.TrimSpace(strings.Join(lines[start-1:end], "\n"))
	if extracted == "" {
		return strings.TrimSpace(normalized)
	}
	return extracted
}

// buildScopeRubric returns cognitive calibration instructions matched to the guide's scope tier.
func buildScopeRubric(scope string) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "definition":
		return "Scope Tier: DEFINITION (Leaf Tier 0). Focus on precise terminology, core definitions, syntax rules, or atomic facts."
	case "description":
		return "Scope Tier: DESCRIPTION (Tier 1). Focus on step-by-step mechanisms, operational behavior, or what happens during execution."
	case "explanation":
		return "Scope Tier: EXPLANATION (Tier 2). Focus on conceptual reasoning ('why'), cause-and-effect relationships, and architectural principles."
	case "lesson":
		return "Scope Tier: LESSON (Top Tier 3). Focus on applied scenario synthesis, comparing trade-offs, or diagnosing a realistic problem."
	default:
		return fmt.Sprintf("Scope Tier: %s. Align question depth with this abstraction level.", scope)
	}
}

func quizItemJSONSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]any{
				"type":        "string",
				"description": "Concise 3-7 word title summarizing the concept tested",
			},
			"prompt": map[string]any{
				"type":        "string",
				"description": "Clear, unambiguous multiple-choice question stem grounded in the target segment",
			},
			"choices": map[string]any{
				"type":     "array",
				"minItems": 3,
				"maxItems": 4,
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"text": map[string]any{
							"type":        "string",
							"description": "Answer option text",
						},
						"correct": map[string]any{
							"type":        "boolean",
							"description": "True for the single correct answer, false for distractors",
						},
						"feedback": map[string]any{
							"type":        "string",
							"description": "Diagnostic explanation shown to the learner when this option is selected",
						},
					},
					"required": []string{"text", "correct", "feedback"},
				},
			},
		},
		"required": []string{"title", "prompt", "choices"},
	}
}

// extractJSONObject strips optional markdown code fences or thinking preambles and extracts the outermost JSON object.
func extractJSONObject(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if idx := strings.Index(trimmed, "</think>"); idx >= 0 {
		trimmed = strings.TrimSpace(trimmed[idx+len("</think>"):])
	}
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		return trimmed[start : end+1]
	}
	return ""
}

func runOllamaChatStream(client *http.Client, endpoint string, chatReq ollamaChatRequest) (string, string, error) {
	reqBytes, err := json.Marshal(chatReq)
	if err != nil {
		return "", "", err
	}

	resp, err := client.Post(endpoint+"/api/chat", "application/json", bytes.NewReader(reqBytes))
	if err != nil {
		return "", "", fmt.Errorf("ollama /api/chat request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("ollama returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var contentBuilder strings.Builder
	var thinkingBuilder strings.Builder
	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var chunk ollamaChatResponse
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return "", "", fmt.Errorf("failed to decode Ollama stream chunk: %w", err)
		}
		if chunk.Error != "" {
			return "", "", fmt.Errorf("ollama error: %s", chunk.Error)
		}
		contentBuilder.WriteString(chunk.Message.Content)
		thinkingBuilder.WriteString(chunk.Message.Thinking)
		if chunk.Done {
			break
		}
	}
	return contentBuilder.String(), thinkingBuilder.String(), nil
}

// GenerateQuestionFromSegment uses the local Ollama model to generate a scope-calibrated question from the anchored segment.
func (a *App) GenerateQuestionFromSegment(dirPath string, guideID string, segment string) (*QuizItemDTO, error) {
	// Serialize generation calls so background pre-fetches never stack in Ollama's queue and cause header timeouts
	ollamaGenMu.Lock()
	defer ollamaGenMu.Unlock()

	status, err := a.GetAgentStatus()
	if err != nil {
		return nil, err
	}
	if !status.Available {
		return nil, fmt.Errorf("ollama server is offline at %s (run 'ollama serve')", status.Endpoint)
	}
	if status.ActiveModel == "" {
		return nil, fmt.Errorf("no local Ollama models found (run 'ollama pull llama3.2')")
	}

	guide, err := a.GetGuide(dirPath, guideID)
	if err != nil {
		return nil, fmt.Errorf("failed to load guide %s: %w", guideID, err)
	}

	existingItems, _ := a.GetQuizItems(dirPath, guideID)
	targetText := extractSegmentLines(guide.Body, segment)
	if targetText == "" {
		return nil, fmt.Errorf("guide %q has no body text to generate questions from", guideID)
	}

	var existingPrompts []string
	for _, it := range existingItems {
		if strings.TrimSpace(it.Prompt) != "" {
			existingPrompts = append(existingPrompts, "- "+strings.TrimSpace(it.Prompt))
			if len(existingPrompts) >= 5 {
				break
			}
		}
	}

	var subguideNames []string
	for _, sg := range guide.SubGuides {
		subguideNames = append(subguideNames, fmt.Sprintf("%s (lines %s, %s)", sg.Guide, sg.Segment, sg.Adherence))
	}

	systemPrompt := strings.Join([]string{
		"You are an expert assessment designer for an adaptive Intelligent Tutoring System built on Quest (qst) Knowledge Graphs and the IMS QTI specification.",
		buildScopeRubric(guide.Scope),
		"Rules:",
		"1. Base the question strictly on the Target Segment Text provided.",
		"2. Create exactly 3 or 4 plausible choices where ONLY ONE choice has \"correct\": true.",
		"3. Distractors (\"correct\": false) must represent realistic conceptual misconceptions rather than obvious jokes.",
		"4. Every choice must include a concise \"feedback\" explanation clarifying why it is correct or what misconception it represents.",
		"5. Use Markdown inline code (`...`) for code identifiers/syntax and LaTeX ($...$ or $$...$$) for mathematical or complexity expressions where appropriate.",
		"6. Output ONLY a single valid JSON object with keys \"title\", \"prompt\", and \"choices\" (each choice having \"text\", \"correct\", \"feedback\").",
	}, "\n")

	var userBuilder strings.Builder
	fmt.Fprintf(&userBuilder, "Guide ID: %s\n", guide.ID)
	fmt.Fprintf(&userBuilder, "Scope Tier: %s\n", guide.Scope)
	fmt.Fprintf(&userBuilder, "Anchored Segment: lines %s\n", segment)
	if len(guide.Prerequisites) > 0 {
		fmt.Fprintf(&userBuilder, "Prerequisites: %s\n", strings.Join(guide.Prerequisites, ", "))
	}
	if len(subguideNames) > 0 {
		fmt.Fprintf(&userBuilder, "Linked Sub-guides: %s\n", strings.Join(subguideNames, ", "))
	}

	fullBody := strings.TrimSpace(guide.Body)
	if fullBody != "" && fullBody != targetText {
		contextSnippet := fullBody
		if len(contextSnippet) > 2500 {
			contextSnippet = contextSnippet[:2500] + "\n...[truncated]..."
		}
		fmt.Fprintf(&userBuilder, "\n--- Surrounding Guide Context ---\n%s\n", contextSnippet)
	}
	fmt.Fprintf(&userBuilder, "\n--- Target Segment Text (Focus the question here) ---\n%s\n", targetText)

	if len(existingPrompts) > 0 {
		fmt.Fprintf(&userBuilder, "\n--- Existing Questions on this Guide (Do NOT duplicate these) ---\n%s\n", strings.Join(existingPrompts, "\n"))
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	messages := []ollamaChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userBuilder.String()},
	}

	// Attempt 1: With Think=false and structured JSON schema
	thinkDisabled := false
	rawContent, rawThinking, err := runOllamaChatStream(client, status.Endpoint, ollamaChatRequest{
		Model:    status.ActiveModel,
		Messages: messages,
		Stream:   true,
		Think:    &thinkDisabled,
		Format:   quizItemJSONSchema(),
		Options: map[string]any{
			"temperature": 0.4,
			"num_ctx":     4096,
		},
	})
	if err != nil {
		return nil, err
	}

	var generated generatedQuestionPayload
	content := extractJSONObject(rawContent)
	if content == "" && rawThinking != "" {
		content = extractJSONObject(rawThinking)
	}

	// If the model produced empty output or truncated JSON under Think=false + GBNF schema (common on Qwen3.5/reasoning models),
	// retry without the GBNF Format constraint so the model can complete its natural template and emit JSON.
	if content == "" || json.Unmarshal([]byte(content), &generated) != nil {
		rawContent2, rawThinking2, err2 := runOllamaChatStream(client, status.Endpoint, ollamaChatRequest{
			Model:    status.ActiveModel,
			Messages: messages,
			Stream:   true,
			Think:    &thinkDisabled,
			Options: map[string]any{
				"temperature": 0.4,
				"num_ctx":     4096,
			},
		})
		if err2 == nil {
			content2 := extractJSONObject(rawContent2)
			if content2 == "" && rawThinking2 != "" {
				content2 = extractJSONObject(rawThinking2)
			}
			if content2 != "" {
				content = content2
			}
		}
	}

	if content == "" {
		return nil, fmt.Errorf("model %q returned an empty response", status.ActiveModel)
	}
	if err := json.Unmarshal([]byte(content), &generated); err != nil {
		return nil, fmt.Errorf("model returned invalid JSON payload: %w", err)
	}

	if strings.TrimSpace(generated.Prompt) == "" || len(generated.Choices) < 2 {
		return nil, fmt.Errorf("model generated an incomplete question")
	}

	choices := make([]QuizChoiceDTO, 0, len(generated.Choices))
	correctAssigned := false
	for i, c := range generated.Choices {
		isCorrect := c.Correct && !correctAssigned
		if isCorrect {
			correctAssigned = true
		}
		choices = append(choices, QuizChoiceDTO{
			ID:       fmt.Sprintf("c%d", i+1),
			Text:     strings.TrimSpace(c.Text),
			Correct:  isCorrect,
			Feedback: strings.TrimSpace(c.Feedback),
		})
	}
	if !correctAssigned && len(choices) > 0 {
		choices[0].Correct = true
	}

	title := strings.TrimSpace(generated.Title)
	if title == "" {
		title = fmt.Sprintf("%s (%s)", guide.ID, segment)
	}

	return &QuizItemDTO{
		ID:        fmt.Sprintf("Q_%s_%d", strings.ToUpper(strings.ReplaceAll(guide.ID, " ", "_")), time.Now().UnixNano()%1000000),
		Title:     title,
		GuideID:   guide.ID,
		Segment:   segment,
		Scope:     guide.Scope,
		Type:      "choice",
		Prompt:    strings.TrimSpace(generated.Prompt),
		Choices:   choices,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// PracticeTurnRequest specifies parameters for resolving the next practice question on the fly.
type PracticeTurnRequest struct {
	DirPath       string   `json:"dir_path"`
	Mode          string   `json:"mode"`           // "adaptive" (whole graph) or "guide" (locked to FocusGuideID)
	FocusGuideID  string   `json:"focus_guide_id"` // Used when Mode == "guide" or remediation drill-down
	FocusSegment  string   `json:"focus_segment"`  // Used when Mode == "guide"
	SeenItemIDs   []string `json:"seen_item_ids"`  // Items already practiced or pre-fetched in this session
	ExcludeGuides []string `json:"exclude_guides"` // Guides to de-prioritize on lookahead pre-fetch
}

// PracticeTurnDTO represents a resolved practice turn (from bank or newly generated) with adaptive metadata.
type PracticeTurnDTO struct {
	Item               QuizItemDTO `json:"item"`
	Source             string      `json:"source"`               // "bank" or "generated"
	Reason             string      `json:"reason"`               // Human-readable scheduling explanation
	LinkedSubGuide     string      `json:"linked_sub_guide"`     // Sub-guide anchored to this segment, if any
	UnmasteredPrereqID string      `json:"unmastered_prereq_id"` // First prerequisite with mastery < 0.70, if any
	TargetMastery      float64     `json:"target_mastery"`       // Current BKT mastery P(L) of the target
}

func getGuideMasteryScore(states map[string]LearnerMasteryDTO, guideID string) float64 {
	if st, ok := states[guideID+":ALL"]; ok {
		return st.PMastery
	}
	return 0.10
}

func getSegmentMasteryScore(states map[string]LearnerMasteryDTO, guideID string, segment string) float64 {
	if st, ok := states[guideID+":"+segment]; ok {
		return st.PMastery
	}
	return getGuideMasteryScore(states, guideID)
}

func findLinkedSubGuide(guide Guide, segment string) string {
	for _, sg := range guide.Metadata.SubGuides {
		if strings.TrimSpace(sg.Segment) == strings.TrimSpace(segment) && sg.Guide != "" {
			return sg.Guide
		}
	}
	if len(guide.Metadata.SubGuides) > 0 {
		return guide.Metadata.SubGuides[0].Guide
	}
	return ""
}

func findUnmasteredPrereq(guide Guide, states map[string]LearnerMasteryDTO) string {
	lowestID := ""
	lowestMastery := 0.70
	for _, prereqID := range guide.Metadata.Prerequisites {
		m := getGuideMasteryScore(states, prereqID)
		if m < lowestMastery {
			lowestMastery = m
			lowestID = prereqID
		}
	}
	return lowestID
}

func selectBestSegmentForGuide(guide Guide, states map[string]LearnerMasteryDTO) string {
	if len(guide.Metadata.SubGuides) > 0 {
		bestSeg := ""
		lowestMastery := 2.0
		for _, sg := range guide.Metadata.SubGuides {
			seg := strings.TrimSpace(sg.Segment)
			if seg == "" {
				continue
			}
			m := getSegmentMasteryScore(states, guide.ID, seg)
			if m < lowestMastery {
				lowestMastery = m
				bestSeg = seg
			}
		}
		if bestSeg != "" {
			return bestSeg
		}
	}
	lineCount := guide.LineCount
	if lineCount < 1 {
		lineCount = 1
	}
	return fmt.Sprintf("1-%d", lineCount)
}

// NextPracticeTurn resolves the next optimal practice question using the hybrid bank + local Ollama engine.
func (a *App) NextPracticeTurn(req PracticeTurnRequest) (*PracticeTurnDTO, error) {
	if strings.TrimSpace(req.DirPath) == "" {
		return nil, fmt.Errorf("graph directory path is empty")
	}

	manifest, err := parseManifest(req.DirPath)
	if err != nil {
		return nil, err
	}
	guidesMap, err := parseGraphDir(req.DirPath)
	if err != nil {
		return nil, err
	}
	if len(guidesMap) == 0 {
		return nil, fmt.Errorf("no guides found in %s", req.DirPath)
	}

	states, err := a.GetLearnerStates(req.DirPath)
	if err != nil {
		states = make(map[string]LearnerMasteryDTO)
	}

	seenSet := make(map[string]bool, len(req.SeenItemIDs))
	for _, id := range req.SeenItemIDs {
		seenSet[id] = true
	}

	excludedGuides := make(map[string]bool, len(req.ExcludeGuides))
	for _, gID := range req.ExcludeGuides {
		excludedGuides[gID] = true
	}

	scopeRank := make(map[string]int, len(manifest.Scopes))
	for idx, sc := range manifest.Scopes {
		scopeRank[sc] = idx
	}

	var targetGuide Guide
	var targetSegment string
	var reason string

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "guide" && req.FocusGuideID != "" {
		if g, exists := guidesMap[req.FocusGuideID]; exists {
			targetGuide = g
			seg := strings.TrimSpace(req.FocusSegment)
			if seg == "" || (seg == "1-1" && g.LineCount > 1) {
				seg = selectBestSegmentForGuide(g, states)
			}
			targetSegment = seg
			reason = fmt.Sprintf("Focused Guide • %s", g.Metadata.Scope)
		}
	}

	if targetGuide.ID == "" {
		// Adaptive Graph Scheduling across all guides
		guideIDs := make([]string, 0, len(guidesMap))
		for id := range guidesMap {
			guideIDs = append(guideIDs, id)
		}
		sort.Strings(guideIDs)

		bestScore := -1e9
		for _, id := range guideIDs {
			g := guidesMap[id]
			pMastery := getGuideMasteryScore(states, g.ID)

			unmetPrereqs := 0
			for _, prereqID := range g.Metadata.Prerequisites {
				if getGuideMasteryScore(states, prereqID) < 0.70 {
					unmetPrereqs++
				}
			}

			dueForReview := false
			if st, ok := states[g.ID+":ALL"]; ok && st.DueForReview {
				dueForReview = true
			}

			score := (1.0 - pMastery) * 100.0
			if unmetPrereqs == 0 {
				score += 200.0
			} else {
				score -= float64(unmetPrereqs) * 80.0
			}
			if dueForReview {
				score += 150.0
			}
			if pMastery < 0.85 {
				score += 50.0
			}
			score -= float64(scopeRank[g.Metadata.Scope]) * 8.0

			if excludedGuides[g.ID] && len(guideIDs) > 1 {
				score -= 250.0
			}

			if score > bestScore {
				bestScore = score
				targetGuide = g
				if dueForReview {
					reason = "FSRS Review Due (R < 85%)"
				} else if len(g.Metadata.Prerequisites) == 0 && pMastery < 0.85 {
					reason = fmt.Sprintf("Foundational %s • Unmastered", g.Metadata.Scope)
				} else if unmetPrereqs == 0 && pMastery < 0.85 {
					reason = "Frontier Ready • Prerequisites Met"
				} else {
					reason = fmt.Sprintf("Adaptive Reinforcement • %d%% Mastery", int(math.Round(pMastery*100)))
				}
			}
		}
		targetSegment = selectBestSegmentForGuide(targetGuide, states)
	}

	// 1. Check items.yaml for an unseen item on targetGuide
	guideItems, _ := a.GetQuizItems(req.DirPath, targetGuide.ID)
	var chosenBankItem *QuizItemDTO
	for i := range guideItems {
		it := guideItems[i]
		if !seenSet[it.ID] && it.Segment == targetSegment {
			chosenBankItem = &it
			break
		}
	}
	if chosenBankItem == nil {
		for i := range guideItems {
			it := guideItems[i]
			if !seenSet[it.ID] {
				chosenBankItem = &it
				break
			}
		}
	}

	if chosenBankItem != nil {
		seg := chosenBankItem.Segment
		if seg == "" {
			seg = targetSegment
		}
		return &PracticeTurnDTO{
			Item:               *chosenBankItem,
			Source:             "bank",
			Reason:             reason,
			LinkedSubGuide:     findLinkedSubGuide(targetGuide, seg),
			UnmasteredPrereqID: findUnmasteredPrereq(targetGuide, states),
			TargetMastery:      getSegmentMasteryScore(states, targetGuide.ID, seg),
		}, nil
	}

	// 2. No unseen item in bank -> generate a new question via local Ollama and auto-save to items.yaml
	generated, genErr := a.GenerateQuestionFromSegment(req.DirPath, targetGuide.ID, targetSegment)
	if genErr == nil && generated != nil {
		_ = a.SaveQuizItem(req.DirPath, *generated)
		return &PracticeTurnDTO{
			Item:               *generated,
			Source:             "generated",
			Reason:             reason,
			LinkedSubGuide:     findLinkedSubGuide(targetGuide, targetSegment),
			UnmasteredPrereqID: findUnmasteredPrereq(targetGuide, states),
			TargetMastery:      getSegmentMasteryScore(states, targetGuide.ID, targetSegment),
		}, nil
	}

	// 3. Offline / Error Fallback: If Ollama is unavailable, serve from existing bank items if any exist
	if len(guideItems) > 0 {
		fallback := guideItems[0]
		return &PracticeTurnDTO{
			Item:               fallback,
			Source:             "bank",
			Reason:             reason + " (Bank Cycle)",
			LinkedSubGuide:     findLinkedSubGuide(targetGuide, fallback.Segment),
			UnmasteredPrereqID: findUnmasteredPrereq(targetGuide, states),
			TargetMastery:      getSegmentMasteryScore(states, targetGuide.ID, fallback.Segment),
		}, nil
	}

	if mode != "guide" {
		allItems, _ := a.GetQuizItems(req.DirPath, "")
		for _, it := range allItems {
			if !seenSet[it.ID] {
				g := guidesMap[it.GuideID]
				return &PracticeTurnDTO{
					Item:               it,
					Source:             "bank",
					Reason:             "Cached Bank Item (Ollama Offline)",
					LinkedSubGuide:     findLinkedSubGuide(g, it.Segment),
					UnmasteredPrereqID: findUnmasteredPrereq(g, states),
					TargetMastery:      getSegmentMasteryScore(states, it.GuideID, it.Segment),
				}, nil
			}
		}
		if len(allItems) > 0 {
			it := allItems[0]
			g := guidesMap[it.GuideID]
			return &PracticeTurnDTO{
				Item:               it,
				Source:             "bank",
				Reason:             "Cached Bank Cycle (Ollama Offline)",
				LinkedSubGuide:     findLinkedSubGuide(g, it.Segment),
				UnmasteredPrereqID: findUnmasteredPrereq(g, states),
				TargetMastery:      getSegmentMasteryScore(states, it.GuideID, it.Segment),
			}, nil
		}
	}

	return nil, genErr
}
