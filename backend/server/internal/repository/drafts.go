package repository

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"
)

type Vacancy struct {
	Title       string `json:"title"`
	Company     string `json:"company"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Salary      string `json:"salary"`
	Remote      bool   `json:"remote"`
}
type DraftFiles struct{ Root string }

func (f DraftFiles) Resumes() ([]string, error) {
	result := []string{}
	r, err := os.OpenRoot(f.Root)
	if err != nil {
		return result, err
	}
	defer r.Close()
	dir, err := r.OpenRoot("resumes")
	if errors.Is(err, os.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	defer dir.Close()
	entries, err := fs.ReadDir(dir.FS(), ".")
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		if entry.Type().IsRegular() && resumeName(entry.Name()) {
			result = append(result, entry.Name())
		}
	}
	sort.Strings(result)
	return result, nil
}
func resumeName(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return name != "" && filepath.Base(name) == name && !strings.ContainsAny(name, "/\\") && (ext == ".md" || ext == ".pdf")
}
func (f DraftFiles) ReadResume(name string) ([]byte, error) {
	if !resumeName(name) {
		return nil, errors.New("Выберите PDF или Markdown резюме из списка")
	}
	r, err := os.OpenRoot(f.Root)
	if err != nil {
		return nil, errors.New("Не удалось открыть папку резюме")
	}
	defer r.Close()
	dir, err := r.OpenRoot("resumes")
	if err != nil {
		return nil, errors.New("Не удалось открыть папку резюме")
	}
	defer dir.Close()
	info, err := dir.Lstat(name)
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("Резюме не найдено. Выберите файл из списка")
	}
	file, err := dir.Open(name)
	if err != nil {
		return nil, errors.New("Не удалось прочитать резюме")
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 8*1024*1024+1))
	if err != nil || len(data) > 8*1024*1024 {
		return nil, errors.New("Не удалось прочитать резюме: максимальный размер 8 МБ")
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, errors.New("Резюме пустое. Выберите другой файл")
	}
	return data, nil
}
func NewDraftID() string { return fmt.Sprintf("%x", rand.Text()) }
func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
		if b.Len() > 60 {
			break
		}
	}
	value := strings.Trim(b.String(), "-")
	if value == "" {
		return "unknown"
	}
	return value
}
func draftDocument(v Vacancy, resume, text string) []byte {
	line := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	return []byte(fmt.Sprintf("# %s — %s\n\n- **URL:** %s\n- **Источник:** %s\n- **Дата:** %s\n- **Статус:** черновик\n- **Резюме:** %s\n\n## Отклик\n\n%s\n", line(v.Title), line(v.Company), line(v.URL), line(v.Source), time.Now().Format("2006-01-02"), line(resume), text))
}

// Save writes a complete temporary file before publishing it. Creation uses a
// hard link that cannot replace an existing application; edits use atomic rename.
func (f DraftFiles) Save(id, path string, v Vacancy, resume, text string) (string, error) {
	r, err := os.OpenRoot(f.Root)
	if err != nil {
		return "", err
	}
	defer r.Close()
	if err := r.MkdirAll("applications", 0700); err != nil {
		return "", err
	}
	dir, err := r.OpenRoot("applications")
	if err != nil {
		return "", err
	}
	defer dir.Close()
	creating := path == ""
	name := filepath.Base(path)
	if creating {
		name = time.Now().Format("2006-01-02") + "-" + slug(v.Company) + "-" + slug(v.Title) + "-" + id + ".md"
	}
	if !creating && path != filepath.Join("applications", name) {
		return "", errors.New("invalid draft path")
	}
	tmp := ".draft-" + NewDraftID()
	file, err := dir.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", err
	}
	defer dir.Remove(tmp)
	_, writeErr := file.Write(draftDocument(v, resume, text))
	if writeErr == nil {
		writeErr = file.Sync()
	}
	closeErr := file.Close()
	if writeErr != nil {
		return "", writeErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	if creating {
		err = dir.Link(tmp, name)
	} else {
		err = dir.Rename(tmp, name)
	}
	if err != nil {
		return "", err
	}
	return filepath.Join("applications", name), nil
}
