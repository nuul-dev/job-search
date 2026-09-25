package handlers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"job-search/server/internal/repository"
)

func TestResumeCreateEditAndUpload(t *testing.T) {
	root := t.TempDir()
	h := NewResumes(&repository.ResumeFiles{Root: root})
	first, err := h.Write("cv.md", "# Resume\n", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Write("cv.md", "overwrite", "", true); !errors.Is(err, repository.ErrResumeConflict) {
		t.Fatal("overwrote existing", err)
	}
	second, err := h.Write("cv.md", "# Edited\n", first.Version, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Write("cv.md", "stale", first.Version, false); !errors.Is(err, repository.ErrResumeConflict) {
		t.Fatal("accepted stale update", err)
	}
	if err := os.WriteFile(filepath.Join(root, "resumes", "cv.md"), []byte("external edit"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Write("cv.md", "stale", second.Version, false); !errors.Is(err, repository.ErrResumeConflict) {
		t.Fatal("overwrote external edit", err)
	}
	view, err := h.Get(context.Background(), "cv.md")
	if err != nil || view.Text != "external edit" {
		t.Fatal(view, err)
	}
	if _, err := h.Upload("../escape.md", []byte("text")); err == nil {
		t.Fatal("accepted traversal")
	}
	if _, err := h.Upload("bad.pdf", []byte("not pdf")); err == nil {
		t.Fatal("accepted invalid pdf")
	}
	pdf, err := h.Upload("resume.pdf", []byte("%PDF-1.7\nfixture"))
	if err != nil || pdf.Editable {
		t.Fatal(pdf, err)
	}
	if _, err := h.Upload("resume.pdf", []byte("%PDF-1.7\nfixture")); !errors.Is(err, repository.ErrResumeConflict) {
		t.Fatal("overwrote upload", err)
	}
}
