package practice

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type PracticeRecordDTO struct {
	ItemID          string `json:"item_id"`
	GuideID         string `json:"guide_id"`
	Attempts        int    `json:"attempts"`
	Correct         int    `json:"correct"`
	LastCorrect     bool   `json:"last_correct"`
	LastPracticedAt string `json:"last_practiced_at"`
}

type PracticeHistoryStore struct {
	Records map[string]PracticeRecordDTO `json:"records"`
}

func getPracticeHistoryPath(dirPath string) string {
	return filepath.Join(dirPath, ".practice_history.json")
}

func loadHistory(dirPath string) (*PracticeHistoryStore, error) {
	data, err := os.ReadFile(getPracticeHistoryPath(dirPath))
	if os.IsNotExist(err) {
		return &PracticeHistoryStore{Records: make(map[string]PracticeRecordDTO)}, nil
	} else if err != nil {
		return nil, err
	}

	var store PracticeHistoryStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	if store.Records == nil {
		store.Records = make(map[string]PracticeRecordDTO)
	}
	return &store, nil
}

// GetHistory returns all saved question practice records.
func GetHistory(dirPath string) (map[string]PracticeRecordDTO, error) {
	store, err := loadHistory(dirPath)
	if err != nil {
		return nil, err
	}
	return store.Records, nil
}

// GetRecord returns a practice record for a specific question item.
func GetRecord(dirPath string, itemID string) (*PracticeRecordDTO, error) {
	store, err := loadHistory(dirPath)
	if err != nil {
		return nil, err
	}
	if rec, exists := store.Records[itemID]; exists {
		return &rec, nil
	}
	return nil, nil
}

// RecordAttempt updates practice attempts and accuracy for a question.
func RecordAttempt(dirPath string, itemID string, guideID string, correct bool) error {
	store, err := loadHistory(dirPath)
	if err != nil {
		return err
	}

	rec, exists := store.Records[itemID]
	if !exists {
		rec = PracticeRecordDTO{
			ItemID:  itemID,
			GuideID: guideID,
		}
	}

	rec.Attempts++
	if correct {
		rec.Correct++
	}
	rec.LastCorrect = correct
	rec.LastPracticedAt = time.Now().UTC().Format(time.RFC3339)
	store.Records[itemID] = rec

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getPracticeHistoryPath(dirPath), data, 0644)
}
