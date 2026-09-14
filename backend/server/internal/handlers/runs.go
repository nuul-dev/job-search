package handlers

import "job-search/server/internal/repository"

type RunsRepository interface {
	List() (repository.Runs, error)
}
type Runs struct{ repo RunsRepository }

func NewRuns(repo RunsRepository) *Runs        { return &Runs{repo: repo} }
func (h *Runs) List() (repository.Runs, error) { return h.repo.List() }
