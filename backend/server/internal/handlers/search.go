package handlers

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"job-search/fetch/searchoptions"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var ErrRunning = errors.New("search already running")
var ErrClosed = errors.New("server shutting down")

type SearchState struct {
	Filters    searchoptions.Options `json:"filters"`
	Status     string                `json:"status"`
	StartedAt  *time.Time            `json:"started_at"`
	FinishedAt *time.Time            `json:"finished_at"`
	Error      string                `json:"error"`
}
type Runner func(context.Context, searchoptions.Options) error
type Search struct {
	mu      sync.Mutex
	state   SearchState
	runner  Runner
	timeout time.Duration
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	closed  bool
}

func NewSearch(timeout time.Duration, runner Runner) *Search {
	ctx, cancel := context.WithCancel(context.Background())
	return &Search{state: SearchState{Status: "idle", Filters: searchoptions.Options{Direction: "profile", Levels: []string{}}}, runner: runner, timeout: timeout, ctx: ctx, cancel: cancel}
}
func (s *Search) snapshot() SearchState {
	state := s.state
	state.Filters = state.Filters.Clone()
	return state
}
func (s *Search) Status() SearchState { s.mu.Lock(); defer s.mu.Unlock(); return s.snapshot() }
func (s *Search) Start(options searchoptions.Options) (SearchState, error) {
	options, err := options.Normalize()
	if err != nil {
		return s.Status(), err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return s.snapshot(), ErrClosed
	}
	if s.state.Status == "running" {
		return s.snapshot(), ErrRunning
	}
	now := time.Now().UTC()
	s.state = SearchState{Status: "running", StartedAt: &now, Filters: options.Clone()}
	s.wg.Add(1)
	go s.run(options.Clone())
	return s.snapshot(), nil
}
func (s *Search) run(options searchoptions.Options) {
	defer s.wg.Done()
	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()
	err := s.runner(ctx, options)
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	s.state.FinishedAt = &now
	s.state.Status = "succeeded"
	if err != nil || ctx.Err() != nil {
		s.state.Status = "failed"
		s.state.Error = "Не удалось завершить поиск. Подробности: logs/web-search.log."
		if ctx.Err() == context.DeadlineExceeded {
			s.state.Error = "Превышено время ожидания поиска."
		}
		if ctx.Err() == context.Canceled {
			s.state.Error = "Поиск остановлен вместе с сервером."
		}
		logrus.WithError(err).Warn("search failed")
	}
}
func (s *Search) Close() { s.mu.Lock(); s.closed = true; s.cancel(); s.mu.Unlock(); s.wg.Wait() }

// CommandRunner executes only the workspace script. Linux process groups ensure
// timeout and shutdown also stop fetchers and Claude spawned by the script.
func CommandRunner(root string) Runner {
	return func(ctx context.Context, options searchoptions.Options) error {
		if err := os.MkdirAll(filepath.Join(root, "logs"), 0700); err != nil {
			return errors.Wrap(err, "create logs")
		}
		log, err := os.OpenFile(filepath.Join(root, "logs", "web-search.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return errors.Wrap(err, "open search log")
		}
		defer log.Close()
		cmd := exec.CommandContext(ctx, "bash", filepath.Join(root, "run-search.sh"))
		cmd.Dir = root
		data, err := json.Marshal(options)
		if err != nil {
			return errors.Wrap(err, "encode search filters")
		}
		for _, env := range os.Environ() {
			if !strings.HasPrefix(env, searchoptions.Env+"=") {
				cmd.Env = append(cmd.Env, env)
			}
		}
		cmd.Env = append(cmd.Env, searchoptions.Env+"="+string(data))
		cmd.Stdout, cmd.Stderr = log, log
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Cancel = func() error {
			err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			if err == syscall.ESRCH {
				return os.ErrProcessDone
			}
			return err
		}
		cmd.WaitDelay = 5 * time.Second
		return errors.Wrap(cmd.Run(), "run search script")
	}
}
