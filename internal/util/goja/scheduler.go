package goja_util

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/samber/mo"
)

// Job represents a task to be executed in the VM
type Job struct {
	fn       func() error
	resultCh chan error
	async    bool // Flag to indicate if the job is async (doesn't need to wait for result)
}

// Scheduler handles all VM operations added concurrently in a single goroutine
// Any goroutine that needs to execute a VM operation must schedule it because the UI VM isn't thread safe
type Scheduler struct {
	jobQueue chan *Job
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	// Track the currently executing job to detect nested scheduling
	currentJob     *Job
	currentJobLock sync.Mutex

	onException mo.Option[func(err error)]
}

func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		jobQueue:    make(chan *Job, 9999),
		ctx:         ctx,
		onException: mo.None[func(err error)](),
		cancel:      cancel,
	}

	s.start()
	return s
}

func (s *Scheduler) SetOnException(onException func(err error)) {
	s.onException = mo.Some(onException)
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
				// Set the current job before execution
				s.currentJobLock.Lock()
				s.currentJob = job
				s.currentJobLock.Unlock()

				err := job.fn()

				// Clear the current job after execution
				s.currentJobLock.Lock()
				s.currentJob = nil
				s.currentJobLock.Unlock()

				// Only send result if the job is not async
				if !job.async {
					job.resultCh <- err
				}

				if err != nil {
					if onException, ok := s.onException.Get(); ok {
						onException(err)
					}
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
		async:    false,
	}

	// Check if we're already in a job execution context
	s.currentJobLock.Lock()
	isNestedCall := s.currentJob != nil && !s.currentJob.async
	s.currentJobLock.Unlock()

	// If this is a nested call from a synchronous job, we need to be careful
	// We can't execute directly because the VM isn't thread-safe
	// Instead, we'll queue it and use a separate goroutine to wait for the result
	if isNestedCall {
		// Queue the job
		select {
		case <-s.ctx.Done():
			return fmt.Errorf("scheduler stopped")
		case s.jobQueue <- job:
			// Create a separate goroutine to wait for the result
			// This prevents deadlock while still ensuring the job runs in the scheduler
			resultCh2 := make(chan error, 1)
			go func() {
				resultCh2 <- <-resultCh
			}()
			return <-resultCh2
		}
	}

	// Otherwise, queue the job normally
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("scheduler stopped")
	case s.jobQueue <- job:
		return <-resultCh
	}
}

// ScheduleAsync adds a job to the queue without waiting for completion
// This is useful for fire-and-forget operations or when a job needs to schedule another job
func (s *Scheduler) ScheduleAsync(fn func() error) {
	job := &Job{
		fn: func() error {
			defer func() {
				if r := recover(); r != nil {
					if onException, ok := s.onException.Get(); ok {
						onException(fmt.Errorf("panic in async job: %v", r))
					}
				}
			}()
			return fn()
		},
		resultCh: nil, // No result channel needed
		async:    true,
	}

	// Queue the job without blocking
	select {
	case <-s.ctx.Done():
		// Scheduler is stopped, just ignore
		return
	case s.jobQueue <- job:
		// Job queued successfully
		// fmt.Printf("job queued successfully, length: %d\n", len(s.jobQueue))
		return
	default:
		// Queue is full, log an error
		if onException, ok := s.onException.Get(); ok {
			onException(fmt.Errorf("async job queue is full"))
		}
	}
}

// ScheduleWithTimeout schedules a job with a timeout
func (s *Scheduler) ScheduleWithTimeout(fn func() error, timeout time.Duration) error {
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
		async:    false,
	}

	// Check if we're already in a job execution context
	s.currentJobLock.Lock()
	isNestedCall := s.currentJob != nil && !s.currentJob.async
	s.currentJobLock.Unlock()

	// If this is a nested call from a synchronous job, handle it specially
	if isNestedCall {
		// Queue the job
		select {
		case <-s.ctx.Done():
			return fmt.Errorf("scheduler stopped")
		case s.jobQueue <- job:
			// Create a separate goroutine to wait for the result with timeout
			resultCh2 := make(chan error, 1)
			go func() {
				select {
				case err := <-resultCh:
					resultCh2 <- err
				case <-time.After(timeout):
					resultCh2 <- fmt.Errorf("operation timed out")
				}
			}()
			return <-resultCh2
		}
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
