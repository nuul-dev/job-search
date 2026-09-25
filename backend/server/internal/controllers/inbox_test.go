package controllers

import (
	"context"

	"encoding/json"
	"job-search/fetch/searchoptions"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"job-search/server/internal/handlers"
	"job-search/server/internal/repository"

	"github.com/gofiber/fiber/v2"
)

func TestInbox(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "jobs/raw"), 0700); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{"03.json": `{"vacancies":[{"id":"a"},null,3]}`, "02.json": `{"vacancies":null}`, "01.json": `{"vacancies":false}`} {
		if err := os.WriteFile(filepath.Join(root, "jobs/raw", name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	search := handlers.NewSearch(time.Minute, func(ctx context.Context, _ searchoptions.Options) error { <-ctx.Done(); return ctx.Err() })
	defer search.Close()
	app := fiber.New()
	NewInbox(handlers.NewRuns(repository.Files{Root: root}), search, root, 8080).RegisterRoutes(app)
	t.Run("exports", func(t *testing.T) {
		res, err := app.Test(httptest.NewRequest("GET", "http://127.0.0.1:8080/api/runs", nil))
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var body repository.Runs
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body.Runs) != 2 || len(body.Errors) != 1 || body.Runs[0].ID != "03" || len(body.Runs[0].Vacancies) != 1 || body.Runs[1].Vacancies == nil {
			t.Fatalf("unexpected: %+v", body)
		}
	})
	for _, tc := range []struct {
		name, host, origin, media, body string
		status                          int
	}{
		{"foreign host", "evil.test:8080", "", "application/json", "{}", 403},
		{"foreign origin", "127.0.0.1:8080", "https://evil.test", "application/json", "{}", 403},
		{"form", "127.0.0.1:8080", "", "text/plain", "{}", 415},
		{"invalid body", "127.0.0.1:8080", "", "application/json", "null", 400},
		{"command injection", "127.0.0.1:8080", "", "application/json", `{"command":"anything"}`, 400},
		{"invalid level", "127.0.0.1:8080", "", "application/json", `{"direction":"backend","levels":["boss"]}`, 400},
		{"invalid direction", "127.0.0.1:8080", "", "application/json", `{"direction":"../../"}`, 400},
		{"invalid remote", "127.0.0.1:8080", "", "application/json", `{"remote_only":"true"}`, 400},
		{"accepted", "127.0.0.1:8080", "http://127.0.0.1:8080", "application/json", "{}", 202},
		{"duplicate", "127.0.0.1:8080", "", "application/json", "{}", 409},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "http://"+tc.host+"/api/search", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.media)
			req.Header.Set("Origin", tc.origin)
			res, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != tc.status {
				t.Fatalf("got %d want %d", res.StatusCode, tc.status)
			}
			if tc.status >= 400 && tc.status != 409 {
				req := httptest.NewRequest("POST", "http://"+tc.host+"/api/search/cancel", strings.NewReader(tc.body))
				req.Header.Set("Content-Type", tc.media)
				req.Header.Set("Origin", tc.origin)
				res, err := app.Test(req)
				if err != nil {
					t.Fatal(err)
				}
				defer res.Body.Close()
				if res.StatusCode != tc.status {
					t.Fatalf("cancel: got %d want %d", res.StatusCode, tc.status)
				}
			}
		})
	}
	req := httptest.NewRequest("POST", "http://127.0.0.1:8080/api/search/cancel", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var state handlers.SearchState
	if err := json.NewDecoder(res.Body).Decode(&state); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 || state.Status != "canceling" {
		t.Fatalf("cancel: %d %+v", res.StatusCode, state)
	}
	for _, path := range []string{"/run-search.sh", "/config/user-profile.md", "/logs/web-search.log", "/../config/user-profile.md"} {
		res, err := app.Test(httptest.NewRequest("GET", "http://127.0.0.1:8080"+path, nil))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != 404 {
			t.Errorf("%s exposed: %d", path, res.StatusCode)
		}
	}
}
