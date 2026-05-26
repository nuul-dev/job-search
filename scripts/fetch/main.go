// fetch — scrapes vacancies from job boards and writes raw JSON to jobs/raw/.
// Run from the repo root: go run ./scripts/fetch
// Outputs the path to the written file on the last line of stdout.
package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Vacancy struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Company     string `json:"company"`
	Salary      string `json:"salary"`
	Source      string `json:"source"`
	Description string `json:"description"`
	Remote      bool   `json:"remote"`
}

type Output struct {
	FetchedAt  string    `json:"fetched_at"`
	QueryCount int       `json:"query_count"`
	Vacancies  []Vacancy `json:"vacancies"`
}

type rssFeed struct {
	Items []struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		PubDate     string `xml:"pubDate"`
		Description string `xml:"description"`
	} `xml:"channel>item"`
}

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

// ── HTTP ──────────────────────────────────────────────────────────────────────

var httpClient = &http.Client{Timeout: 15 * time.Second}

func get(rawURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; job-search-bot/1.0)")
	req.Header.Set("Accept", "application/xml,application/json,text/html;q=0.9")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 2<<20))
}

// ── hh.ru RSS ─────────────────────────────────────────────────────────────────

var (
	reCDATA   = regexp.MustCompile(`(?s)<!\[CDATA\[(.*?)]]>`)
	reCompany = regexp.MustCompile(`Вакансия компании:\s*([^\n<]+)`)
	reSalary  = regexp.MustCompile(`уровень[^:]*:\s*([^\n<]+)`)
)

var pubDateFormats = []string{
	"Mon, 02 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05 -0700",
	time.RFC1123Z,
	time.RFC1123,
}

func parsePubDate(s string) (time.Time, bool) {
	for _, f := range pubDateFormats {
		if t, err := time.Parse(f, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func stripCDATA(s string) string {
	if m := reCDATA.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return s
}

func fetchHH(queries []string, cutoff time.Time) []Vacancy {
	var result []Vacancy
	dedup := map[string]bool{}

	for _, q := range queries {
		enc := url.QueryEscape(q)
		for _, suffix := range []string{"&schedule=remote", ""} {
			u := "https://hh.ru/search/vacancy/rss?text=" + enc + "&area=1&sort_by=publication_time" + suffix
			body, err := get(u)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  warn  hh.ru %q: %v\n", q, err)
				continue
			}
			var feed rssFeed
			if err := xml.Unmarshal(body, &feed); err != nil {
				fmt.Fprintf(os.Stderr, "  warn  hh.ru xml: %v\n", err)
				continue
			}
			for _, item := range feed.Items {
				if dedup[item.Link] {
					continue
				}
				if t, ok := parsePubDate(item.PubDate); ok && t.Before(cutoff) {
					continue
				}
				desc := stripCDATA(item.Description)
				company, salary := "", ""
				if m := reCompany.FindStringSubmatch(desc); m != nil {
					company = strings.TrimSpace(m[1])
				}
				if m := reSalary.FindStringSubmatch(desc); m != nil {
					salary = strings.TrimSpace(m[1])
				}
				dedup[item.Link] = true
				result = append(result, Vacancy{
					Title:   strings.TrimSpace(item.Title),
					URL:     item.Link,
					Company: company,
					Salary:  salary,
					Source:  "hh.ru",
					Remote:  suffix != "",
				})
			}
		}
	}
	return result
}

// ── Hirify ────────────────────────────────────────────────────────────────────

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

		// API may return a list or a wrapped object
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
			// remote filter
			fmtVal := fmt.Sprintf("%v", v["work_format"])
			if !strings.Contains(strings.ToLower(fmtVal), "remote") &&
				!strings.Contains(fmtVal, "удал") {
				continue
			}
			// date filter
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

// ── Habr Career ───────────────────────────────────────────────────────────────

func fetchHabr(queries []string, cutoff time.Time) []Vacancy {
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

// ── dedup + blacklist filter ──────────────────────────────────────────────────

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

func main() {
	repoDir, _ := os.Getwd()

	queries, blacklist, err := loadProfile(repoDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	seen := loadSeen(repoDir)
	cutoff := time.Now().UTC().Add(-24 * time.Hour)

	fmt.Printf("queries: %d  blacklist: %d  seen: %d\n", len(queries), len(blacklist), len(seen))

	var all []Vacancy

	fmt.Print("hh.ru...       ")
	hh := fetchHH(queries, cutoff)
	fmt.Printf("%d\n", len(hh))
	all = append(all, hh...)

	fmt.Print("Hirify...      ")
	hi := fetchHirify(queries, cutoff)
	fmt.Printf("%d\n", len(hi))
	all = append(all, hi...)

	fmt.Print("Habr Career... ")
	habr := fetchHabr(queries, cutoff)
	fmt.Printf("%d\n", len(habr))
	all = append(all, habr...)

	all = filter(all, seen, blacklist)
	fmt.Printf("total after dedup/filter: %d\n", len(all))

	now := time.Now().UTC()
	outDir := filepath.Join(repoDir, "jobs", "raw")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	outPath := filepath.Join(outDir, now.Format("2006-01-02-1504")+".json")

	data, _ := json.MarshalIndent(Output{
		FetchedAt:  now.Format(time.RFC3339),
		QueryCount: len(queries),
		Vacancies:  all,
	}, "", "  ")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	fmt.Println(outPath) // last line captured by run-search.sh
}
