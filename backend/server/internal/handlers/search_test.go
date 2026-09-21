package handlers

import (
	"context"

	"errors"
	"job-search/fetch/searchoptions"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestCommandRunnerKillsProcessGroup(t *testing.T) {
	if _, err := exec.LookPath("timeout"); err != nil {
		t.Skip("GNU timeout unavailable")
	}
	root := t.TempDir()
	// This fixture replaces run-search.sh; no fetcher or Claude is invoked.
	script := "#!/bin/bash\ntimeout --foreground 60 bash -c 'echo $$ > child.pid; exec sleep 60' &\nwait\n"
	if err := os.WriteFile(filepath.Join(root, "run-search.sh"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- CommandRunner(root)(ctx, searchoptions.Options{}) }()
	var pid int
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(filepath.Join(root, "child.pid"))
		if err == nil {
			pid, _ = strconv.Atoi(strings.TrimSpace(string(data)))
			if pid > 0 {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	if pid == 0 {
		t.Fatal("child did not start")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected cancellation error")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("runner failed to stop")
	}
	deadline = time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); err == syscall.ESRCH {
			return
		}
		// A killed orphan may briefly remain a zombie until PID 1 reaps it.
		data, _ := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
		if strings.Contains(string(data), ") Z ") {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("search descendant still running")
}

func waitFinished(t *testing.T, s *Search) SearchState {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		state := s.Status()
		if state.Status != "running" {
			return state
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("search did not finish")
	return SearchState{}
}

func TestSearchPassesIndependentFilters(t *testing.T) {
	release := make(chan struct{})
	received := make(chan searchoptions.Options, 1)
	s := NewSearch(time.Second, func(_ context.Context, o searchoptions.Options) error { <-release; received <- o; return nil })
	defer s.Close()
	o := searchoptions.Options{Direction: "backend", Levels: []string{"junior"}, Query: "Go", RemoteOnly: true}
	state, err := s.Start(o)
	if err != nil {
		t.Fatal(err)
	}
	o.Levels[0] = "senior"
	state.Filters.Levels[0] = "lead"
	snapshot := s.Status()
	snapshot.Filters.Levels[0] = "middle"
	close(release)
	got := <-received
	if got.Direction != "backend" || got.Query != "Go" || !got.RemoteOnly || got.Levels[0] != "junior" {
		t.Fatalf("runner options: %+v", got)
	}
	state = waitFinished(t, s)
	if state.Filters.Levels[0] != "junior" {
		t.Fatalf("state mutated: %+v", state)
	}
}
func TestSingleActiveSearch(t *testing.T) {
	release := make(chan struct{})
	var calls atomic.Int32
	s := NewSearch(time.Second, func(context.Context, searchoptions.Options) error { calls.Add(1); <-release; return nil })
	defer s.Close()
	var wg sync.WaitGroup
	var accepted atomic.Int32
	for range 30 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.Start(searchoptions.Options{}); err == nil {
				accepted.Add(1)
			} else if !errors.Is(err, ErrRunning) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	close(release)
	state := waitFinished(t, s)
	if accepted.Load() != 1 || calls.Load() != 1 || state.Status != "succeeded" || state.FinishedAt == nil || state.StartedAt == nil {
		t.Fatalf("unexpected state %+v accepted %d calls %d", state, accepted.Load(), calls.Load())
	}
	if _, err := s.Start(searchoptions.Options{}); err != nil {
		t.Fatal(err)
	}
	waitFinished(t, s)
}
func TestSearchFailureAndCancellation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		runner  Runner
		timeout time.Duration
		close   bool
		message string
	}{
		{"failure", func(context.Context, searchoptions.Options) error { return errors.New("private detail") }, time.Second, false, "Не удалось завершить поиск. Подробности: logs/web-search.log."},
		{"timeout", func(ctx context.Context, _ searchoptions.Options) error { <-ctx.Done(); return ctx.Err() }, time.Millisecond, false, "Превышено время ожидания поиска."},
		{"shutdown", func(ctx context.Context, _ searchoptions.Options) error { <-ctx.Done(); return ctx.Err() }, time.Second, true, "Поиск остановлен вместе с сервером."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSearch(tc.timeout, tc.runner)
			defer s.Close()
			if _, err := s.Start(searchoptions.Options{}); err != nil {
				t.Fatal(err)
			}
			if tc.close {
				s.Close()
				if _, err := s.Start(searchoptions.Options{}); !errors.Is(err, ErrClosed) {
					t.Fatal("start after close", err)
				}
			}
			state := waitFinished(t, s)
			if state.Status != "failed" || state.Error != tc.message {
				t.Fatalf("unexpected %+v", state)
			}
		})
	}
}
