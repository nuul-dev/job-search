// Package searchoptions defines filters shared by the web server and fetcher.
package searchoptions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

const Env = "JOB_SEARCH_FILTERS"

type Options struct {
	Direction  string   `json:"direction"`
	Levels     []string `json:"levels"`
	Query      string   `json:"query"`
	RemoteOnly bool     `json:"remote_only"`
}

var directions = map[string][]string{
	"profile":   nil,
	"backend":   {"backend", "бэкенд", "back-end", "golang", "java developer", "python developer", "php developer", ".net developer"},
	"frontend":  {"frontend", "фронтенд", "front-end", "react developer", "vue developer", "angular developer"},
	"fullstack": {"fullstack", "full-stack", "фулстек"},
	"mobile":    {"mobile", "android", "ios", "мобильный разработчик"},
	"devops":    {"devops", "sre"},
	"qa":        {"qa", "тестировщик"},
	"data":      {"data engineer", "data analyst", "аналитик данных"},
	"ai":        {"machine learning", "ml engineer", "ai engineer", "data scientist"},
	"any":       {"developer", "engineer", "разработчик", "аналитик", "тестировщик"},
}
var grades = map[string][]string{
	"junior": {"junior", "jr", "джуниор", "джун", "младший"},
	"middle": {"middle", "mid", "мидл", "миддл"},
	"senior": {"senior", "sr", "сеньор", "синьор", "старший"},
	"lead":   {"lead", "teamlead", "techlead", "тимлид", "техлид", "ведущий", "руководитель"},
}

func Decode(data []byte) (Options, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return Options{}, fmt.Errorf("Ожидается JSON-объект с параметрами поиска")
	}
	for _, raw := range fields {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return Options{}, fmt.Errorf("Параметры поиска не могут быть null")
		}
	}
	var o Options
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&o); err != nil {
		return o, fmt.Errorf("Некорректные параметры поиска: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return o, fmt.Errorf("Ожидается один JSON-объект")
	}
	return o.Normalize()
}

func (o Options) Normalize() (Options, error) {
	if o.Direction == "" {
		o.Direction = "profile"
	}
	if _, ok := directions[o.Direction]; !ok {
		return o, fmt.Errorf("Неизвестное направление поиска")
	}
	seen := map[string]bool{}
	levels := []string{}
	for _, level := range o.Levels {
		if _, ok := grades[level]; !ok {
			return o, fmt.Errorf("Неизвестный уровень вакансии")
		}
		if !seen[level] {
			levels = append(levels, level)
			seen[level] = true
		}
	}
	o.Levels = levels
	o.Query = strings.TrimSpace(o.Query)
	if !utf8.ValidString(o.Query) || utf8.RuneCountInString(o.Query) > 120 {
		return o, fmt.Errorf("Запрос должен быть не длиннее 120 символов")
	}
	for _, r := range o.Query {
		if unicode.IsControl(r) {
			return o, fmt.Errorf("Запрос содержит управляющие символы")
		}
	}
	return o, nil
}

func (o Options) Clone() Options { o.Levels = append([]string{}, o.Levels...); return o }

func (o Options) Queries(profile []string) []string {
	bases := directions[o.Direction]
	if o.Direction == "profile" {
		bases = profile
	}
	if o.Direction == "any" && o.Query != "" {
		bases = []string{""}
	}
	var result []string
	seen := map[string]bool{}
	for _, base := range bases {
		if o.Direction == "profile" && len(o.Levels) > 0 {
			base = withoutGrades(base)
		}
		base = strings.TrimSpace(base + " " + o.Query)
		variants := []string{base}
		if len(o.Levels) > 0 {
			variants = nil
			for _, level := range o.Levels {
				variants = append(variants, strings.TrimSpace(base+" "+level))
			}
		}
		for _, q := range variants {
			if q != "" && !seen[q] {
				result = append(result, q)
				seen[q] = true
			}
		}
	}
	return result
}

func tokens(s string) []string {
	words := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '+' && r != '#' })
	for i, word := range words {
		if word != "c++" {
			word = strings.TrimRight(word, "+")
		}
		switch word {
		case "go":
			word = "golang"
		case "js":
			word = "javascript"
		case "ts":
			word = "typescript"
		}
		words[i] = word
	}
	return words
}
func contains(text, phrase string) bool {
	hay, needle := tokens(text), tokens(phrase)
	if len(needle) == 0 {
		return false
	}
	for i := 0; i+len(needle) <= len(hay); i++ {
		match := true
		for j := range needle {
			if hay[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
func withoutGrades(query string) string {
	var out []string
	for _, word := range strings.Fields(query) {
		isGrade := false
		for _, aliases := range grades {
			for _, alias := range aliases {
				if contains(word, alias) {
					isGrade = true
				}
			}
		}
		if !isGrade {
			out = append(out, word)
		}
	}
	return strings.Join(out, " ")
}

// Matches uses only explicit title grades. Missing grades cannot satisfy a
// selected level; descriptions often mention colleagues of unrelated grades.
func (o Options) Matches(title, description string, remote bool) bool {
	if o.RemoteOnly && !remote {
		return false
	}
	if len(o.Levels) > 0 {
		match := false
		for _, level := range o.Levels {
			for _, alias := range grades[level] {
				if contains(title, alias) {
					match = true
				}
			}
		}
		if !match {
			return false
		}
	}
	if o.Direction != "profile" && o.Direction != "any" {
		match := false
		for _, alias := range directions[o.Direction] {
			if contains(title, alias) {
				match = true
			}
		}
		// Backend titles commonly name a server language instead of the role.
		if o.Direction == "backend" {
			for _, alias := range []string{"golang", "go developer", "go engineer", "java разработчик", "python разработчик", "php разработчик", ".net разработчик", "go разработчик", "backend developer", "бэкенд разработчик"} {
				if contains(title, alias) {
					match = true
				}
			}
		}
		if o.Direction == "frontend" {
			for _, alias := range []string{"react разработчик", "vue разработчик", "angular разработчик"} {
				if contains(title, alias) {
					match = true
				}
			}
			if contains(title, "react native") {
				match = false
			}
		}
		if !match {
			return false
		}
	}
	for _, word := range strings.Fields(o.Query) {
		if !contains(title+" "+description, word) {
			return false
		}
	}
	return true
}
