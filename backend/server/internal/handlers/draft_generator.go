package handlers

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"
)

type boundedBuffer struct {
	buffer bytes.Buffer
	limit  int
}

func (b *boundedBuffer) Bytes() []byte { return b.buffer.Bytes() }

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.buffer.Len() {
		return 0, errors.New("output exceeds limit")
	}
	return b.buffer.Write(p)
}
func groupCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = 3 * time.Second
	cmd.Stderr = io.Discard
	return cmd
}
func resumeText(ctx context.Context, ext string, data []byte) (string, error) {
	if strings.EqualFold(ext, ".pdf") {
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		cmd := groupCommand(ctx, "pdftotext", "-layout", "-", "-")
		cmd.Stdin = bytes.NewReader(data)
		out := &boundedBuffer{limit: 100000}
		cmd.Stdout = out
		if err := cmd.Run(); err != nil {
			return "", errors.New("Не удалось прочитать PDF. Установите pdftotext или выберите Markdown-резюме")
		}
		data = out.Bytes()
	}
	text := strings.TrimSpace(string(data))
	if text == "" || !utf8.ValidString(text) {
		return "", errors.New("В резюме нет читаемого текста. Выберите другой файл")
	}
	if len(text) > 100000 {
		return "", errors.New("Резюме слишком длинное. Выберите файл до 100 КБ текста")
	}
	return text, nil
}
func CodexDraftGenerator(model string) DraftGenerator {
	return func(ctx context.Context, prompt string) (string, error) {
		dir, err := os.MkdirTemp("", "job-draft-")
		if err != nil {
			return "", errors.New("Не удалось подготовить генерацию")
		}
		defer os.RemoveAll(dir)
		output := filepath.Join(dir, "letter.txt")
		args := []string{"exec", "--sandbox", "read-only", "--ephemeral", "--skip-git-repo-check", "--ignore-user-config", "--ignore-rules", "--color", "never", "--output-last-message", output, "-c", `web_search="disabled"`}
		// Supported feature names are checked against the installed CLI. Authentication
		// stays in the normal CODEX_HOME; user configuration and plugins are not loaded.
		for _, feature := range []string{"shell_tool", "unified_exec", "hooks", "apps", "plugins", "browser_use", "browser_use_external", "computer_use", "image_generation", "multi_agent", "multi_agent_v2", "skill_search", "code_mode_host"} {
			args = append(args, "--disable", feature)
		}
		if modelName := strings.TrimSpace(model); modelName != "" {
			args = append(args, "--model", modelName)
		}
		args = append(args, "-")
		cmd := groupCommand(ctx, "codex", args...)
		cmd.Dir = dir
		cmd.Stdin = strings.NewReader(prompt)
		cmd.Stdout = io.Discard
		if err := cmd.Run(); err != nil {
			return "", errors.New("Не удалось создать письмо. Проверьте установку Codex CLI и вход через codex login, затем повторите")
		}
		file, err := os.Open(output)
		if err != nil {
			return "", errors.New("Codex CLI не вернул текст письма. Попробуйте ещё раз")
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, 20001))
		text := strings.TrimSpace(string(data))
		if err != nil || len(data) > 20000 || text == "" || !utf8.Valid(data) {
			return "", errors.New("Codex CLI вернул пустой или слишком длинный текст письма. Попробуйте ещё раз")
		}
		return text, nil
	}
}
