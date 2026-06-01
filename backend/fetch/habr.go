package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

func fetchHabr(queries []string) []Vacancy {
	var result []Vacancy
	dedup := map[string]bool{}

	for _, q := range queries {
		u := "https://career.habr.com/api/frontend/vacancies?q=" + url.QueryEscape(q) + "&sort=date&type=all"
		body, err := get(u)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warn  Habr %q: %v\n", q, err)
			continue
		}
		var wrapper map[string]any
		if err := json.Unmarshal(body, &wrapper); err != nil {
			continue
		}
		list, _ := wrapper["list"].([]any)
		for _, el := range list {
			v, ok := el.(map[string]any)
			if !ok {
				continue
			}
			href, _ := v["href"].(string)
			if href == "" {
				href, _ = v["url"].(string)
			}
			if href == "" {
				continue
			}
			if !strings.HasPrefix(href, "http") {
				href = "https://career.habr.com" + href
			}
			if dedup[href] {
				continue
			}
			title, _ := v["title"].(string)
			company := ""
			if c, ok := v["company"].(map[string]any); ok {
				company, _ = c["title"].(string)
			}
			salary := ""
			if s, ok := v["salary"].(map[string]any); ok {
				from, _ := s["from"].(float64)
				to, _ := s["to"].(float64)
				cur, _ := s["currency"].(string)
				if from > 0 || to > 0 {
					salary = fmt.Sprintf("%.0f–%.0f %s", from, to, cur)
				}
			}
			dedup[href] = true
			result = append(result, Vacancy{
				Title:   strings.TrimSpace(title),
				URL:     href,
				Company: company,
				Salary:  salary,
				Source:  "Habr Career",
			})
		}
	}
	return result
}
