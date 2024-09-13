package tasks

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
    bgTasks := New(slog.Default(), 3, 10)
	bgTasks.Run()
	taskRunned := false
	task := func() {
		t.Log("task")
		taskRunned = true
	}
	bgTasks.Add(task)
	bgTasks.Shutdown(context.Background())
	assert.True(t, taskRunned)
}