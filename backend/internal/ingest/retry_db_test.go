package ingest

import (
	"errors"
	"testing"
)

func TestIsRetryableDBError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"connection reset", errors.New("connection reset by peer"), true},
		{"connection refused", errors.New("dial tcp: connection refused"), true},
		{"deadlock", errors.New("pq: deadlock detected"), true},
		{"serialization", errors.New("could not serialize access"), true},
		{"server closed", errors.New("server closed the connection unexpectedly"), true},
		{"i/o timeout", errors.New("read tcp: i/o timeout"), true},
		{"context deadline", errors.New("context deadline exceeded"), true},
		{"unique violation (not retryable)", errors.New("pq: duplicate key value violates unique constraint"), false},
		{"bad input (not retryable)", errors.New("invalid input syntax for type uuid"), false},
		{"random error (not retryable)", errors.New("something else went wrong"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryableDBError(tt.err); got != tt.want {
				t.Errorf("isRetryableDBError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
