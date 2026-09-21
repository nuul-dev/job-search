package searchoptions

import (
	"strings"
	"testing"
)

func TestCustomSearchFilters(t *testing.T) {
	all, err := Decode([]byte(`{"direction":"backend","levels":[],"query":"","remote_only":false}`))
	if err != nil {
		t.Fatal(err)
	}
	queries := strings.Join(all.Queries([]string{"middle frontend react"}), " ")
	if strings.Contains(queries, "middle") || strings.Contains(queries, "react") {
		t.Fatalf("profile leaked: %s", queries)
	}
	if !all.Matches("Junior Backend developer", "", false) || !all.Matches("Backend developer", "", false) {
		t.Fatal("all levels excluded")
	}
	selected, err := Decode([]byte(`{"direction":"backend","levels":["middle"],"query":"Go","remote_only":true}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		title        string
		remote, want bool
	}{
		{"Middle+ Golang developer", true, true},
		{"Middle Go-разработчик", true, true},
		{"Junior Go developer", true, false},
		{"Golang developer", true, false},
		{"Middle Go developer", false, false},
		{"Middle backend developer", true, false},
	} {
		if got := selected.Matches(tc.title, "", tc.remote); got != tc.want {
			t.Errorf("%s: got %v", tc.title, got)
		}
	}
	profile := Options{Direction: "profile", Levels: []string{"junior"}}
	for _, q := range profile.Queries([]string{"middle+ Go developer"}) {
		if strings.Contains(q, "middle") || !strings.Contains(q, "junior") {
			t.Fatal(q)
		}
	}
}
