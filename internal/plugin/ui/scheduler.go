package plugin_ui

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Job represents a task to be executed in the VM
type Job struct {
	fn       func() error
	resultCh chan error
}

// Scheduler handles all VM operations added concurrently in a single goroutine
// Any goroutine that needs to execute a VM operation must schedule it because the UI VM isn't thread safe
type Scheduler struct {
	jobQueue chan *Job
	ctx      context.Context
	context  *Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewScheduler(uiCtx *Context) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		jobQueue: make(chan *Job, 100),
		ctx:      ctx,
		context:  uiCtx,
		cancel:   cancel,
	}

	s.start()
	return s
}

func (s *Scheduler) start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case job := <-s.jobQueue:
				err := job.fn()
				job.resultCh <- err
				if err != nil {
					s.context.HandleException(err)
				}
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
}

// Schedule adds a job to the queue and waits for its completion
func (s *Scheduler) Schedule(fn func() error) error {
	resultCh := make(chan error, 1)
	job := &Job{
		fn: func() error {
			defer func() {
				if r := recover(); r != nil {
					resultCh <- fmt.Errorf("panic: %v", r)
				}
			}()
			return fn()
		},
		resultCh: resultCh,
	}

	select {
	case <-s.ctx.Done():
		return fmt.Errorf("scheduler stopped")
	case s.jobQueue <- job:
		return <-resultCh
	}
}

// ScheduleWithTimeout schedules a job with a timeout
func (s *Scheduler) ScheduleWithTimeout(fn func() error, timeout time.Duration) error {
	resultCh := make(chan error, 1)
	job := &Job{
		fn:       fn,
		resultCh: resultCh,
	}

	select {
	case <-s.ctx.Done():
		return fmt.Errorf("scheduler stopped")
	case s.jobQueue <- job:
		select {
		case err := <-resultCh:
			return err
		case <-time.After(timeout):
			return fmt.Errorf("operation timed out")
		}
	}
}
