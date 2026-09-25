package handlers

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
	"job-search/server/internal/repository"
)

type ResumeView struct {
	Name     string `json:"name"`
	Text     string `json:"text"`
	Editable bool   `json:"editable"`
	Version  string `json:"version"`
}
type Resumes struct{ files *repository.ResumeFiles }

func NewResumes(files *repository.ResumeFiles) *Resumes { return &Resumes{files: files} }
func validResume(name string, data []byte) error {
	if len(name) > 200 || strings.TrimSpace(name) != name || name == "" || strings.HasPrefix(name, ".") || filepath.Base(name) != name || strings.ContainsAny(name, "/\\") {
		return errors.New("Укажите имя файла без пути, до 200 байт")
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return errors.New("Некорректное имя файла")
		}
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".md":
		if len(data) > 100000 || !utf8.Valid(data) || strings.TrimSpace(string(data)) == "" || bytes.ContainsRune(data, 0) {
			return errors.New("Markdown должен содержать UTF-8 текст до 100 КБ")
		}
	case ".pdf":
		if len(data) > 8*1024*1024 || !bytes.HasPrefix(data, []byte("%PDF-")) {
			return errors.New("Выберите корректный PDF размером до 8 МБ")
		}
	default:
		return errors.New("Поддерживаются только PDF и Markdown (.md)")
	}
	return nil
}
func (h *Resumes) Get(ctx context.Context, name string) (ResumeView, error) {
	data, err := h.files.Read(name)
	if err != nil {
		return ResumeView{}, err
	}
	text, err := resumeText(ctx, filepath.Ext(name), data)
	if err != nil {
		return ResumeView{}, err
	}
	if strings.EqualFold(filepath.Ext(name), ".md") {
		text = string(data)
	}
	return ResumeView{Name: name, Text: text, Editable: strings.EqualFold(filepath.Ext(name), ".md"), Version: repository.ResumeVersion(data)}, nil
}
func (h *Resumes) Write(name, text, version string, create bool) (ResumeView, error) {
	if !strings.EqualFold(filepath.Ext(name), ".md") {
		return ResumeView{}, errors.New("Редактировать можно только Markdown. Создайте копию PDF в формате .md")
	}
	return h.write(name, []byte(text), version, create)
}
func (h *Resumes) Upload(name string, data []byte) (ResumeView, error) {
	return h.write(name, data, "", true)
}
func (h *Resumes) write(name string, data []byte, version string, create bool) (ResumeView, error) {
	if err := validResume(name, data); err != nil {
		return ResumeView{}, err
	}
	if err := h.files.Write(name, data, version, create); err != nil {
		return ResumeView{}, err
	}
	view := ResumeView{Name: name, Editable: strings.EqualFold(filepath.Ext(name), ".md"), Version: repository.ResumeVersion(data)}
	if view.Editable {
		view.Text = string(data)
	}
	return view, nil
}
func (h *Resumes) Download(name string) ([]byte, error) { return h.files.Read(name) }
