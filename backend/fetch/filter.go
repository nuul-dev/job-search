package main

import "strings"

func filter(vacancies []Vacancy, seen map[string]bool, blacklist []string) []Vacancy {
	inRun := map[string]bool{}
	var out []Vacancy
	for _, v := range vacancies {
		if seen[v.URL] || inRun[v.URL] {
			continue
		}
		cl := strings.ToLower(v.Company)
		skip := false
		for _, bl := range blacklist {
			if bl != "" && strings.Contains(cl, strings.ToLower(bl)) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		inRun[v.URL] = true
		out = append(out, v)
	}
	return out
}
