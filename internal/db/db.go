package db

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type SearchHistory struct {
	Timestamp time.Time `json:"timestamp"`
	Pattern   string    `json:"pattern"`
	Directory string    `json:"directory"`
	Results   int       `json:"results"`
}

const historyFile = "search_history.json"

func SaveSearchHistory(pattern, directory string, resultCount int) error {
	history := SearchHistory{
		Timestamp: time.Now(),
		Pattern:   pattern,
		Directory: directory,
		Results:   resultCount,
	}

	// Load existing history
	var histories []SearchHistory
	if data, err := os.ReadFile(historyFile); err == nil {
		json.Unmarshal(data, &histories)
	}

	// Append new history
	histories = append(histories, history)

	// Keep only last 100 entries
	if len(histories) > 100 {
		histories = histories[len(histories)-100:]
	}

	// Save back to file
	data, err := json.MarshalIndent(histories, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	return os.WriteFile(historyFile, data, 0644)
}

func LoadSearchHistory() ([]SearchHistory, error) {
	var histories []SearchHistory

	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return histories, nil // Return empty slice if file doesn't exist
		}
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	if err := json.Unmarshal(data, &histories); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return histories, nil
}

func ClearSearchHistory() error {
	return os.Remove(historyFile)
}
