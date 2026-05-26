package main

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
