package main

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

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

func fetchHH(queries []string, cutoff time.Time, remoteOnly bool) []Vacancy {
	var result []Vacancy
	dedup := map[string]bool{}
	suffixes := []string{"&schedule=remote", ""}
	if remoteOnly {
		suffixes = suffixes[:1]
	}

	for _, q := range queries {
		enc := url.QueryEscape(q)
		for _, suffix := range suffixes {
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
					Title:       strings.TrimSpace(item.Title),
					URL:         item.Link,
					Company:     company,
					Salary:      salary,
					Source:      "hh.ru",
					Description: desc,
					Remote:      suffix != "",
				})
			}
		}
	}
	return result
}
