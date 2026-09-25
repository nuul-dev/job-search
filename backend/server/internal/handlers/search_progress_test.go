package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"job-search/fetch/searchoptions"
)

func TestSearchProgressMarkers(t *testing.T) {
	ready := make(chan context.Context, 2)
	finish := make(chan error, 2)
	s := NewSearch(time.Second, func(ctx context.Context, _ searchoptions.Options) error {
		writer := &progressWriter{ctx: ctx}
		writer.Write([]byte("@@JOB_PRO"))
		writer.Write([]byte("GRESS hh\n" + strings.Repeat("x", 1000) + "@@JOB_PROGRESS ranking\n@@JOB_PROGRESS private-data\n"))
		ready <- ctx
		return <-finish
	})
	defer s.Close()
	if _, err := s.Start(searchoptions.Options{}); err != nil {
		t.Fatal(err)
	}
	oldCtx := <-ready
	state := s.Status()
	if state.Stage != "hh" || state.CompletedSteps != 1 || state.TotalSteps != 6 {
		t.Fatal(state)
	}
	reportSearchProgress(oldCtx, "ranking")
	finish <- errors.New("fixture failure")
	state = waitFinished(t, s)
	if state.Status != "failed" || state.CompletedSteps != 5 {
		t.Fatal(state)
	}
	if _, err := s.Start(searchoptions.Options{}); err != nil {
		t.Fatal(err)
	}
	<-ready
	reportSearchProgress(oldCtx, "ranking")
	if state := s.Status(); state.Stage != "hh" || state.CompletedSteps != 1 {
		t.Fatal("stale update", state)
	}
	finish <- nil
	state = waitFinished(t, s)
	if state.Status != "succeeded" || state.Stage != "finished" || state.CompletedSteps != 6 {
		t.Fatal(state)
	}
}
