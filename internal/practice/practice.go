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

type QuizPracticeRecordDTO struct {
	Key             string `json:"key"`
	QuizID          string `json:"quiz_id"`
	ItemID          string `json:"item_id"`
	Attempts        int    `json:"attempts"`
	Correct         int    `json:"correct"`
	LastCorrect     bool   `json:"last_correct"`
	LastPracticedAt string `json:"last_practiced_at"`
}

type PracticeHistoryStore struct {
	Records     map[string]PracticeRecordDTO     `json:"records"`
	QuizRecords map[string]QuizPracticeRecordDTO `json:"quiz_records,omitempty"`
}

func getPracticeHistoryPath(dirPath string) string {
	return filepath.Join(dirPath, ".practice_history.json")
}

func loadHistory(dirPath string) (*PracticeHistoryStore, error) {
	data, err := os.ReadFile(getPracticeHistoryPath(dirPath))
	if os.IsNotExist(err) {
		return &PracticeHistoryStore{
			Records:     make(map[string]PracticeRecordDTO),
			QuizRecords: make(map[string]QuizPracticeRecordDTO),
		}, nil
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
	if store.QuizRecords == nil {
		store.QuizRecords = make(map[string]QuizPracticeRecordDTO)
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

// QuizRecordKey builds the canonical key used to store standalone quiz practice records.
func QuizRecordKey(quizID string, itemID string) string {
	return quizID + "::" + itemID
}

// GetQuizHistory returns all saved standalone quiz practice records.
func GetQuizHistory(dirPath string) (map[string]QuizPracticeRecordDTO, error) {
	store, err := loadHistory(dirPath)
	if err != nil {
		return nil, err
	}
	return store.QuizRecords, nil
}

// GetQuizRecord returns a practice record for a specific standalone quiz question.
func GetQuizRecord(dirPath string, quizID string, itemID string) (*QuizPracticeRecordDTO, error) {
	store, err := loadHistory(dirPath)
	if err != nil {
		return nil, err
	}
	if rec, exists := store.QuizRecords[QuizRecordKey(quizID, itemID)]; exists {
		return &rec, nil
	}
	return nil, nil
}

// RecordQuizAttempt updates practice attempts and accuracy for a standalone quiz question.
func RecordQuizAttempt(dirPath string, quizID string, itemID string, correct bool) error {
	store, err := loadHistory(dirPath)
	if err != nil {
		return err
	}

	key := QuizRecordKey(quizID, itemID)
	rec, exists := store.QuizRecords[key]
	if !exists {
		rec = QuizPracticeRecordDTO{
			Key:    key,
			QuizID: quizID,
			ItemID: itemID,
		}
	}

	rec.Attempts++
	if correct {
		rec.Correct++
	}
	rec.LastCorrect = correct
	rec.LastPracticedAt = time.Now().UTC().Format(time.RFC3339)
	store.QuizRecords[key] = rec

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getPracticeHistoryPath(dirPath), data, 0644)
}
