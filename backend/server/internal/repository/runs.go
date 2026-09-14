package repository

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type Run struct {
	ID        string            `json:"id"`
	FetchedAt json.RawMessage   `json:"fetched_at"`
	Vacancies []json.RawMessage `json:"vacancies"`
}
type Runs struct {
	Runs   []Run    `json:"runs"`
	Errors []string `json:"errors"`
}
type Files struct{ Root string }

func (f Files) List() (Runs, error) {
	result := Runs{Runs: []Run{}, Errors: []string{}}
	paths, err := filepath.Glob(filepath.Join(f.Root, "jobs", "raw", "*.json"))
	if err != nil {
		return result, errors.Wrap(err, "list exports")
	}
	sort.Sort(sort.Reverse(sort.StringSlice(paths)))
	for _, path := range paths {
		run, err := readRun(path)
		if err != nil {
			result.Errors = append(result.Errors, filepath.Base(path))
			continue
		}
		result.Runs = append(result.Runs, run)
	}
	return result, nil
}
func readRun(path string) (Run, error) {
	run := Run{ID: strings.TrimSuffix(filepath.Base(path), ".json"), Vacancies: []json.RawMessage{}}
	data, err := os.ReadFile(path)
	if err != nil {
		return run, err
	}
	var export map[string]json.RawMessage
	if err := json.Unmarshal(data, &export); err != nil {
		return run, err
	}
	raw, ok := export["vacancies"]
	if !ok {
		return run, errors.New("missing vacancies")
	}
	var vacancies []json.RawMessage
	if err := json.Unmarshal(raw, &vacancies); err != nil {
		return run, err
	}
	for _, vacancy := range vacancies {
		if bytes.HasPrefix(bytes.TrimSpace(vacancy), []byte("{")) {
			run.Vacancies = append(run.Vacancies, vacancy)
		}
	}
	run.FetchedAt = export["fetched_at"]
	return run, nil
}
