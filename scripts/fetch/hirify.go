package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

func fetchHirify(queries []string, cutoff time.Time) []Vacancy {
	var result []Vacancy
	dedup := map[string]bool{}

	for _, q := range queries {
		u := "https://api.hirify.me/api/vacancies?page=1&search=" + url.QueryEscape(q)
		body, err := get(u)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warn  Hirify %q: %v\n", q, err)
			continue
		}

		var items []map[string]any
		if err := json.Unmarshal(body, &items); err != nil {
			var wrapper map[string]any
			if err2 := json.Unmarshal(body, &wrapper); err2 != nil {
				continue
			}
			for _, key := range []string{"data", "vacancies", "results", "items"} {
				if raw, ok := wrapper[key].([]any); ok {
					for _, el := range raw {
						if m, ok := el.(map[string]any); ok {
							items = append(items, m)
						}
					}
					break
				}
			}
		}

		for _, v := range items {
			slug, _ := v["slug"].(string)
			if slug == "" {
				continue
			}
			link := "https://hirify.me/jobs/" + slug
			if dedup[link] {
				continue
			}
			fmtVal := fmt.Sprintf("%v", v["work_format"])
			if !strings.Contains(strings.ToLower(fmtVal), "remote") &&
				!strings.Contains(fmtVal, "удал") {
				continue
			}
			if updated, _ := v["updated_at"].(string); updated != "" {
				if t, err := time.Parse(time.RFC3339, updated); err == nil && t.Before(cutoff) {
					continue
				}
			}
			title, _ := v["title"].(string)
			salary, _ := v["salary"].(string)
			company := ""
			if c, ok := v["company"].(map[string]any); ok {
				company, _ = c["name"].(string)
			}
			desc := ""
			if d, _ := v["description"].(string); len(d) > 500 {
				desc = d[:500]
			} else {
				desc = d
			}
			dedup[link] = true
			result = append(result, Vacancy{
				Title:       strings.TrimSpace(title),
				URL:         link,
				Company:     company,
				Salary:      salary,
				Source:      "Hirify",
				Description: desc,
				Remote:      true,
			})
		}
	}
	return result
}
