package plugin_ui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// Job represents a task to be executed in the VM
type Job struct {
	fn       func() error
	resultCh chan error
}

// Scheduler handles all VM operations in a single goroutine
type Scheduler struct {
	jobQueue chan *Job
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		jobQueue: make(chan *Job, 100),
		ctx:      ctx,
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
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
}

// Schedule adds a job to the queue and waits for its completion
func (s *Scheduler) Schedule(fn func() error, isFast bool) error {
	resultCh := make(chan error, 1)
	job := &Job{
		fn:       fn,
		resultCh: resultCh,
	}

	select {
	case <-s.ctx.Done():
		return fmt.Errorf("scheduler stopped")
	case s.jobQueue <- job:
		return <-resultCh
	}
}

// ScheduleCallback schedules a Goja function call
// Automatically detects if it's a fast job based on presence of async operations
func (s *Scheduler) ScheduleCallback(fn *goja.Callable) error {
	// For now, treat all callbacks as fast jobs since we can't reliably detect
	// if they contain async operations without parsing the AST
	return s.Schedule(func() error {
		_, err := (*fn)(goja.Undefined())
		return err
	}, true)
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
