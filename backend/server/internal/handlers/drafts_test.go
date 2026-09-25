package handlers

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"job-search/server/internal/repository"
)

func TestDraftGenerateSaveAndRecover(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "resumes"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "resumes", "cv.md"), []byte("Built Go API at Example"), 0600); err != nil {
		t.Fatal(err)
	}
	// A file blocks the archive directory to exercise a recoverable save failure.
	if err := os.WriteFile(filepath.Join(root, "applications"), []byte("blocked"), 0600); err != nil {
		t.Fatal(err)
	}
	d := NewDrafts(repository.DraftFiles{Root: root}, time.Second, func(_ context.Context, prompt string) (string, error) {
		if !strings.Contains(prompt, "Built Go API at Example") || !strings.Contains(prompt, "Go developer") {
			t.Error("missing resume or vacancy")
		}
		return "Здравствуйте! Разрабатывал API на Go.", nil
	})
	defer d.Close()
	req := DraftRequest{Resume: "cv.md", Vacancy: repository.Vacancy{Title: "Go developer", Company: "Example"}}
	job, err := d.Start(req)
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		job, err = d.Get(job.ID)
		if err != nil {
			t.Fatal(err)
		}
		if job.Status != "running" {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if job.Status != "failed" || job.Text == "" {
		t.Fatalf("lost generated text: %+v", job)
	}
	if err := os.Remove(filepath.Join(root, "applications")); err != nil {
		t.Fatal(err)
	}
	job, err = d.Save(job.ID, "Мой отредактированный отклик.")
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != "succeeded" || job.Path == "" {
		t.Fatal(job)
	}
	data, err := os.ReadFile(filepath.Join(root, job.Path))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "черновик") || !strings.Contains(string(data), job.Text) || strings.Contains(string(data), "отправлен") {
		t.Fatal(string(data))
	}
	path := job.Path
	job, err = d.Save(job.ID, "Вторая редакция.")
	if err != nil || job.Path != path {
		t.Fatal(job, err)
	}
	if _, err := d.Start(DraftRequest{Resume: "../cv.md", Vacancy: req.Vacancy}); err == nil {
		t.Fatal("accepted path traversal")
	}
}
