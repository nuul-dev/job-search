package repository

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/pkg/errors"
)

var ErrResumeConflict = errors.New("Файл уже существует или изменился. Откройте актуальную версию либо сохраните копию")
var ErrResumeMissing = errors.New("Резюме не найдено")

type ResumeFiles struct {
	Root string
	mu   sync.Mutex
}

func ResumeVersion(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }
func (f *ResumeFiles) directory(create bool) (*os.Root, error) {
	root, err := os.OpenRoot(f.Root)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	if create {
		if err := root.MkdirAll("resumes", 0700); err != nil {
			return nil, err
		}
	}
	return root.OpenRoot("resumes")
}
func readResumeFile(dir *os.Root, name string) ([]byte, error) {
	if !resumeName(name) {
		return nil, errors.New("Некорректное имя резюме")
	}
	info, err := dir.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrResumeMissing
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("Резюме должно быть обычным файлом")
	}
	file, err := dir.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 8*1024*1024+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 8*1024*1024 {
		return nil, errors.New("Максимальный размер резюме 8 МБ")
	}
	return data, nil
}
func (f *ResumeFiles) Read(name string) ([]byte, error) {
	dir, err := f.directory(false)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrResumeMissing
	}
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	return readResumeFile(dir, name)
}
func (f *ResumeFiles) Write(name string, data []byte, version string, create bool) error {
	if !resumeName(name) {
		return errors.New("Некорректное имя резюме")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	dir, err := f.directory(true)
	if err != nil {
		return err
	}
	defer dir.Close()
	if !create {
		current, err := readResumeFile(dir, name)
		if err != nil {
			return err
		}
		if version == "" || ResumeVersion(current) != version {
			return ErrResumeConflict
		}
	}
	tmp := ".resume-" + NewDraftID()
	file, err := dir.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer dir.Remove(tmp)
	_, err = file.Write(data)
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if create {
		err = dir.Link(tmp, name)
	} else {
		// Recheck after preparing the replacement to catch edits during the write.
		current, readErr := readResumeFile(dir, name)
		if readErr != nil {
			return readErr
		}
		if ResumeVersion(current) != version {
			return ErrResumeConflict
		}
		err = dir.Rename(tmp, name)
	}
	if errors.Is(err, os.ErrExist) {
		return ErrResumeConflict
	}
	return err
}
