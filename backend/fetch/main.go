package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"job-search/fetch/searchoptions"
)

func main() {
	repoDir, _ := os.Getwd()
	var options *searchoptions.Options
	if raw, ok := os.LookupEnv(searchoptions.Env); ok {
		parsed, err := searchoptions.Decode([]byte(raw))
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR:", err)
			os.Exit(1)
		}
		options = &parsed
	}

	queries, blacklist, err := loadProfile(repoDir)
	if err != nil && (options == nil || options.Direction == "profile") {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	if options != nil {
		queries = options.Queries(queries)
	}
	seen := loadSeen(repoDir)
	cutoff := time.Now().UTC().Add(-24 * time.Hour)

	fmt.Printf("queries: %d  blacklist: %d  seen: %d\n", len(queries), len(blacklist), len(seen))

	var all []Vacancy

	fmt.Println("@@JOB_PROGRESS hh")
	fmt.Print("hh.ru...       ")
	hh := fetchHH(queries, cutoff, options != nil && options.RemoteOnly)
	fmt.Printf("%d\n", len(hh))
	all = append(all, hh...)

	fmt.Println("@@JOB_PROGRESS hirify")
	fmt.Print("Hirify...      ")
	hi := fetchHirify(queries, cutoff, options == nil || options.RemoteOnly)
	fmt.Printf("%d\n", len(hi))
	all = append(all, hi...)

	fmt.Println("@@JOB_PROGRESS habr")
	fmt.Print("Habr Career... ")
	habr := fetchHabr(queries)
	fmt.Printf("%d\n", len(habr))
	all = append(all, habr...)

	all = filter(all, seen, blacklist)
	if options != nil {
		matched := make([]Vacancy, 0, len(all))
		for _, v := range all {
			if options.Matches(v.Title, v.Description, v.Remote) {
				matched = append(matched, v)
			}
		}
		all = matched
	}
	fmt.Printf("total after dedup/filter: %d\n", len(all))

	now := time.Now().UTC()
	outDir := filepath.Join(repoDir, "jobs", "raw")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	outPath := filepath.Join(outDir, now.Format("2006-01-02-150405.000000000")+".json")

	data, _ := json.MarshalIndent(Output{
		SearchFilters: options,
		FetchedAt:     now.Format(time.RFC3339),
		QueryCount:    len(queries),
		Vacancies:     all,
	}, "", "  ")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	fmt.Println(outPath)
}
