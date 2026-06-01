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
	hirifyCutoff := cutoff.Add(-48 * time.Hour)

	var result []Vacancy
	dedup := map[string]bool{}

	for _, q := range queries {
		for page := 1; page <= 2; page++ {
			items := fetchHirifyPage(q, page)
			if len(items) == 0 {
				break
			}
			pageHadFresh := false
			for _, v := range items {
				slug, _ := v["slug"].(string)
				if slug == "" {
					continue
				}
				link := "https://hirify.me/jobs/" + slug
				if dedup[link] {
					continue
				}
				if !hirifyIsRemote(v["work_format"]) {
					continue
				}
				if updated, _ := v["updated_at"].(string); updated != "" {
					if t, err := time.Parse(time.RFC3339, updated); err == nil {
						if t.Before(hirifyCutoff) {
							continue
						}
					}
				}
				pageHadFresh = true
				title, _ := v["title"].(string)
				company, _ := v["company_title"].(string)
				salary := ""
				if s, ok := v["salary"].(map[string]any); ok {
					min, _ := s["min"].(float64)
					max, _ := s["max"].(float64)
					cur, _ := s["currency"].(string)
					if min > 0 || max > 0 {
						salary = fmt.Sprintf("%.0f–%.0f %s", min, max, cur)
					}
				}
				desc, _ := v["tldr"].(string)
				if len(desc) > 500 {
					desc = desc[:500]
				}
				dedup[link] = true
				result = append(result, Vacancy{
					Title:       strings.TrimSpace(title),
					URL:         link,
					Company:     strings.TrimSpace(company),
					Salary:      salary,
					Source:      "Hirify",
					Description: desc,
					Remote:      true,
				})
			}
			if !pageHadFresh {
				break
			}
		}
	}
	return result
}

func fetchHirifyPage(q string, page int) []map[string]any {
	u := fmt.Sprintf("https://api.hirify.me/api/vacancies?page=%d&search=%s", page, url.QueryEscape(q))
	body, err := get(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  warn  Hirify %q p%d: %v\n", q, page, err)
		return nil
	}
	var items []map[string]any
	if err := json.Unmarshal(body, &items); err == nil {
		return items
	}
	var wrapper map[string]any
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil
	}
	for _, key := range []string{"data", "vacancies", "results", "items"} {
		if raw, ok := wrapper[key].([]any); ok {
			for _, el := range raw {
				if m, ok := el.(map[string]any); ok {
					items = append(items, m)
				}
			}
			return items
		}
	}
	return nil
}

func hirifyIsRemote(v any) bool {
	arr, ok := v.([]any)
	if !ok {
		return false
	}
	for _, el := range arr {
		s, _ := el.(string)
		if strings.EqualFold(s, "remote") {
			return true
		}
	}
	return false
}
