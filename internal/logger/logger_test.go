package logger

import "testing"

func TestInitCreatesLogger(t *testing.T) {
	t.Parallel()

	Init("development")
	if Log == nil {
		t.Fatal("expected logger to be initialized")
	}
}
