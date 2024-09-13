package tasks

import (
	"context"
	"log/slog"
	"sync"
)

type Task = func()

type BackgroudTasks struct {
	log *slog.Logger
	tasks chan Task
	maxWorkers int
	wg *sync.WaitGroup
}

func (t *BackgroudTasks) Run() {
	for i := 0; i < t.maxWorkers; i++ {
		go func() {
			log := t.log.With("worker", i)
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic", "err", err)
				}
				t.wg.Done()
			}()
			for task := range t.tasks {
				task()
				t.log.Info("task done", "task", task)
			}
		}()
	}
}

func (t *BackgroudTasks) Shutdown(ctx context.Context) error {
	const op = "tasks.BackgroudTasks.Shutdown"
	log := t.log.With("op", op)
	log.Info("shutting down background tasks")
	close(t.tasks)
	shutdownCh := make(chan bool, 1)
	go func() {
		t.wg.Wait()
		shutdownCh <- true
	}()
	select {
	case <-ctx.Done():
		log.Warn("graceful shutdown timed out.. forcing exit", "timeout", ctx.Err())
		return ctx.Err()
	case <-shutdownCh:
		log.Info("Background tasks succesfully stopped")
		return nil
	}
}

func New(log *slog.Logger, maxWorkers int, maxTasksQueueSize int) *BackgroudTasks {
	wg := &sync.WaitGroup{}
	wg.Add(maxWorkers)
	tasks := make(chan Task, maxTasksQueueSize)
	return &BackgroudTasks{
		log: log,
		maxWorkers: maxWorkers,
		wg: wg,
		tasks: tasks,
	}
}

func (t *BackgroudTasks) Add(task Task) {
	t.tasks <- task
}

// func (t *BackgroudTasks) Run(task Task) {
// 	t.wg.Add(1)
// 	go func() {
// 		defer func() {
// 			if err := recover(); err != nil {
// 				t.log.Error("panic", "err", err)
// 			}
// 			t.wg.Done()
// 		}()
// 		task()
// 	}()
// }

func (t *BackgroudTasks) IsEmpty() bool {
	return len(t.tasks) == 0
}