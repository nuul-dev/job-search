package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func parseSection(text, header string) []string {
	var out []string
	inside := false
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## "+header {
			inside = true
			continue
		}
		if inside {
			if strings.HasPrefix(trimmed, "## ") {
				break
			}
			s := strings.TrimPrefix(trimmed, "- ")
			s = strings.TrimSpace(s)
			if s != "" && !strings.HasPrefix(s, "_") && !strings.HasPrefix(s, "#") {
				out = append(out, s)
			}
		}
	}
	return out
}

func loadProfile(repoDir string) (queries, blacklist []string, err error) {
	b, err := os.ReadFile(filepath.Join(repoDir, "user-profile.md"))
	if err != nil {
		return nil, nil, fmt.Errorf("user-profile.md not found: %w", err)
	}
	text := string(b)
	queries = parseSection(text, "Поисковые запросы")
	if len(queries) == 0 {
		return nil, nil, fmt.Errorf("'## Поисковые запросы' section is empty — add search terms to user-profile.md")
	}
	blacklist = parseSection(text, "Не хочу от этих компаний")
	return queries, blacklist, nil
}

func loadSeen(repoDir string) map[string]bool {
	seen := map[string]bool{}
	f, err := os.Open(filepath.Join(repoDir, "seen-vacancies.txt"))
	if err != nil {
		return seen
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if parts := strings.SplitN(sc.Text(), " ", 2); len(parts) == 2 {
			seen[parts[1]] = true
		}
	}
	return seen
}
