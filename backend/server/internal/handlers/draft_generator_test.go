package handlers

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexDraftCLI(t *testing.T) {
	dir := t.TempDir()
	// Fake executable only: no model call or candidate data leaves the test.
	script := `#!/bin/sh
printf '%s\n' "$@" > "$DRAFT_TEST_DIR/args"
cat > "$DRAFT_TEST_DIR/prompt"
while [ "$#" -gt 0 ]; do
 if [ "$1" = "--output-last-message" ]; then shift; output="$1"; fi
 shift
done
printf 'event log must not become letter' 
printf 'Здравствуйте! Мой отклик.' > "$output"
exit "${DRAFT_TEST_EXIT:-0}"
`
	if err := os.WriteFile(filepath.Join(dir, "codex"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("DRAFT_TEST_DIR", dir)
	got, err := CodexDraftGenerator("test-model")(context.Background(), "resume and vacancy fixture")
	if err != nil || got != "Здравствуйте! Мой отклик." {
		t.Fatal(got, err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "args"))
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{"exec\n", "--sandbox\nread-only\n", "--ignore-user-config\n", "--ephemeral\n", "--model\ntest-model\n", "--disable\nshell_tool\n", "--disable\nhooks\n", "--output-last-message\n"} {
		if !strings.Contains(string(data), part) {
			t.Fatalf("missing argument %q", part)
		}
	}
	prompt, err := os.ReadFile(filepath.Join(dir, "prompt"))
	if err != nil || string(prompt) != "resume and vacancy fixture" {
		t.Fatal(string(prompt), err)
	}
	t.Setenv("DRAFT_TEST_EXIT", "1")
	if _, err := CodexDraftGenerator("")(context.Background(), "fixture"); err == nil {
		t.Fatal("accepted unsuccessful CLI exit")
	}
	data, err = os.ReadFile(filepath.Join(dir, "args"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "--model\n") {
		t.Fatal("overrode default model")
	}
}
