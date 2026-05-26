package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

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
	habr := fetchHabr(queries)
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

	fmt.Println(outPath)
}
