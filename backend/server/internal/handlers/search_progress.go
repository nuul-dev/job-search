package handlers

import "context"

type progressContextKey struct{}

func reportSearchProgress(ctx context.Context, stage string) {
	if report, ok := ctx.Value(progressContextKey{}).(func(string)); ok {
		report(stage)
	}
}

var searchStages = map[string]int{"preparing": 0, "hh": 1, "hirify": 2, "habr": 3, "ranking": 4}

// progressWriter accepts only exact, bounded marker lines. Oversized log lines
// are discarded until newline; arbitrary process output never reaches the API.
type progressWriter struct {
	ctx     context.Context
	line    []byte
	discard bool
}

func (w *progressWriter) Write(data []byte) (int, error) {
	for _, b := range data {
		if b == '\n' {
			if !w.discard {
				for stage := range searchStages {
					if string(w.line) == "@@JOB_PROGRESS "+stage {
						reportSearchProgress(w.ctx, stage)
						break
					}
				}
			}
			w.line = w.line[:0]
			w.discard = false
		} else if !w.discard {
			if len(w.line) >= 64 {
				w.discard = true
				w.line = w.line[:0]
			} else {
				w.line = append(w.line, b)
			}
		}
	}
	return len(data), nil
}
