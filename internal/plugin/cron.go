// Package cron implements a crontab-like service to execute and schedule
// repeative tasks/jobs.
//
// Example:
//
//	c := cron.New()
//	c.MustAdd("dailyReport", "0 0 * * *", func() { ... })
//	c.Start()
package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"seanime/internal/extension"
	goja_util "seanime/internal/util/goja"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

// Cron is a crontab-like struct for tasks/jobs scheduling.
type Cron struct {
	timezone   *time.Location
	ticker     *time.Ticker
	startTimer *time.Timer
	tickerDone chan bool
	jobs       []*CronJob
	interval   time.Duration
	mux        sync.RWMutex
	scheduler  *goja_util.Scheduler
}

// New create a new Cron struct with default tick interval of 1 minute
// and timezone in UTC.
//
// You can change the default tick interval with Cron.SetInterval().
// You can change the default timezone with Cron.SetTimezone().
func New(scheduler *goja_util.Scheduler) *Cron {
	return &Cron{
		interval:   1 * time.Minute,
		timezone:   time.UTC,
		jobs:       []*CronJob{},
		tickerDone: make(chan bool),
		scheduler:  scheduler,
	}
}

func (a *AppContextImpl) BindCronToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) *Cron {
	cron := New(scheduler)
	cronObj := vm.NewObject()
	_ = cronObj.Set("add", cron.Add)
	_ = cronObj.Set("remove", cron.Remove)
	_ = cronObj.Set("removeAll", cron.RemoveAll)
	_ = cronObj.Set("total", cron.Total)
	_ = cronObj.Set("stop", cron.Stop)
	_ = cronObj.Set("start", cron.Start)
	_ = cronObj.Set("hasStarted", cron.HasStarted)
	_ = obj.Set("cron", cronObj)

	return cron
}

////////////////////////////////////////////////////////////////////////////

// SetInterval changes the current cron tick interval
// (it usually should be >= 1 minute).
func (c *Cron) SetInterval(d time.Duration) {
	// update interval
	c.mux.Lock()
	wasStarted := c.ticker != nil
	c.interval = d
	c.mux.Unlock()

	// restart the ticker
	if wasStarted {
		c.Start()
	}
}

// SetTimezone changes the current cron tick timezone.
func (c *Cron) SetTimezone(l *time.Location) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.timezone = l
}

// MustAdd is similar to Add() but panic on failure.
func (c *Cron) MustAdd(jobId string, cronExpr string, run func()) {
	if err := c.Add(jobId, cronExpr, run); err != nil {
		panic(err)
	}
}

// Add registers a single cron job.
//
// If there is already a job with the provided id, then the old job
// will be replaced with the new one.
//
// cronExpr is a regular cron expression, eg. "0 */3 * * *" (aka. at minute 0 past every 3rd hour).
// Check cron.NewSchedule() for the supported tokens.
func (c *Cron) Add(jobId string, cronExpr string, fn func()) error {
	if fn == nil {
		return errors.New("failed to add new cron job: fn must be non-nil function")
	}

	schedule, err := NewSchedule(cronExpr)
	if err != nil {
		return fmt.Errorf("failed to add new cron job: %w", err)
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	// remove previous (if any)
	c.jobs = slices.DeleteFunc(c.jobs, func(j *CronJob) bool {
		return j.Id() == jobId
	})

	// add new
	c.jobs = append(c.jobs, &CronJob{
		id:        jobId,
		fn:        fn,
		schedule:  schedule,
		scheduler: c.scheduler,
	})

	return nil
}

// Remove removes a single cron job by its id.
func (c *Cron) Remove(jobId string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.jobs == nil {
		return // nothing to remove
	}

	c.jobs = slices.DeleteFunc(c.jobs, func(j *CronJob) bool {
		return j.Id() == jobId
	})
}

// RemoveAll removes all registered cron jobs.
func (c *Cron) RemoveAll() {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.jobs = []*CronJob{}
}

// Total returns the current total number of registered cron jobs.
func (c *Cron) Total() int {
	c.mux.RLock()
	defer c.mux.RUnlock()

	return len(c.jobs)
}

// Jobs returns a shallow copy of the currently registered cron jobs.
func (c *Cron) Jobs() []*CronJob {
	c.mux.RLock()
	defer c.mux.RUnlock()

	copy := make([]*CronJob, len(c.jobs))
	for i, j := range c.jobs {
		copy[i] = j
	}

	return copy
}

// Stop stops the current cron ticker (if not already).
//
// You can resume the ticker by calling Start().
func (c *Cron) Stop() {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.startTimer != nil {
		c.startTimer.Stop()
		c.startTimer = nil
	}

	if c.ticker == nil {
		return // already stopped
	}

	c.tickerDone <- true
	c.ticker.Stop()
	c.ticker = nil
}

// Start starts the cron ticker.
//
// Calling Start() on already started cron will restart the ticker.
func (c *Cron) Start() {
	c.Stop()

	// delay the ticker to start at 00 of 1 c.interval duration
	now := time.Now()
	next := now.Add(c.interval).Truncate(c.interval)
	delay := next.Sub(now)

	c.mux.Lock()
	c.startTimer = time.AfterFunc(delay, func() {
		c.mux.Lock()
		c.ticker = time.NewTicker(c.interval)
		c.mux.Unlock()

		// run immediately at 00
		c.runDue(time.Now())

		// run after each tick
		go func() {
			for {
				select {
				case <-c.tickerDone:
					return
				case t := <-c.ticker.C:
					c.runDue(t)
				}
			}
		}()
	})
	c.mux.Unlock()
}

// HasStarted checks whether the current Cron ticker has been started.
func (c *Cron) HasStarted() bool {
	c.mux.RLock()
	defer c.mux.RUnlock()

	return c.ticker != nil
}

// runDue runs all registered jobs that are scheduled for the provided time.
func (c *Cron) runDue(t time.Time) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	moment := NewMoment(t.In(c.timezone))

	for _, j := range c.jobs {
		if j.schedule.IsDue(moment) {
			go j.Run()
		}
	}
}

////////////////////////////////////////////////

// CronJob defines a single registered cron job.
type CronJob struct {
	fn        func()
	schedule  *Schedule
	id        string
	scheduler *goja_util.Scheduler
}

// Id returns the cron job id.
func (j *CronJob) Id() string {
	return j.id
}

// Expression returns the plain cron job schedule expression.
func (j *CronJob) Expression() string {
	return j.schedule.rawExpr
}

// Run runs the cron job function.
func (j *CronJob) Run() {
	if j.fn != nil {
		j.scheduler.ScheduleAsync(func() error {
			j.fn()
			return nil
		})
	}
}

// MarshalJSON implements [json.Marshaler] and export the current
// jobs data into valid JSON.
func (j CronJob) MarshalJSON() ([]byte, error) {
	plain := struct {
		Id         string `json:"id"`
		Expression string `json:"expression"`
	}{
		Id:         j.Id(),
		Expression: j.Expression(),
	}

	return json.Marshal(plain)
}

////////////////////////////////////////////////

// Moment represents a parsed single time moment.
type Moment struct {
	Minute    int `json:"minute"`
	Hour      int `json:"hour"`
	Day       int `json:"day"`
	Month     int `json:"month"`
	DayOfWeek int `json:"dayOfWeek"`
}

// NewMoment creates a new Moment from the specified time.
func NewMoment(t time.Time) *Moment {
	return &Moment{
		Minute:    t.Minute(),
		Hour:      t.Hour(),
		Day:       t.Day(),
		Month:     int(t.Month()),
		DayOfWeek: int(t.Weekday()),
	}
}

// Schedule stores parsed information for each time component when a cron job should run.
type Schedule struct {
	Minutes    map[int]struct{} `json:"minutes"`
	Hours      map[int]struct{} `json:"hours"`
	Days       map[int]struct{} `json:"days"`
	Months     map[int]struct{} `json:"months"`
	DaysOfWeek map[int]struct{} `json:"daysOfWeek"`

	rawExpr string
}

// IsDue checks whether the provided Moment satisfies the current Schedule.
func (s *Schedule) IsDue(m *Moment) bool {
	if _, ok := s.Minutes[m.Minute]; !ok {
		return false
	}

	if _, ok := s.Hours[m.Hour]; !ok {
		return false
	}

	if _, ok := s.Days[m.Day]; !ok {
		return false
	}

	if _, ok := s.DaysOfWeek[m.DayOfWeek]; !ok {
		return false
	}

	if _, ok := s.Months[m.Month]; !ok {
		return false
	}

	return true
}

var macros = map[string]string{
	"@yearly":   "0 0 1 1 *",
	"@annually": "0 0 1 1 *",
	"@monthly":  "0 0 1 * *",
	"@weekly":   "0 0 * * 0",
	"@daily":    "0 0 * * *",
	"@midnight": "0 0 * * *",
	"@hourly":   "0 * * * *",
	"@30min":    "*/30 * * * *",
	"@15min":    "*/15 * * * *",
	"@10min":    "*/10 * * * *",
	"@5min":     "*/5 * * * *",
}

// NewSchedule creates a new Schedule from a cron expression.
//
// A cron expression could be a macro OR 5 segments separated by space,
// representing: minute, hour, day of the month, month and day of the week.
//
// The following segment formats are supported:
//   - wildcard: *
//   - range:    1-30
//   - step:     */n or 1-30/n
//   - list:     1,2,3,10-20/n
//
// The following macros are supported:
//   - @yearly (or @annually)
//   - @monthly
//   - @weekly
//   - @daily (or @midnight)
//   - @hourly
func NewSchedule(cronExpr string) (*Schedule, error) {
	if v, ok := macros[cronExpr]; ok {
		cronExpr = v
	}

	segments := strings.Split(cronExpr, " ")
	if len(segments) != 5 {
		return nil, errors.New("invalid cron expression - must be a valid macro or to have exactly 5 space separated segments")
	}

	minutes, err := parseCronSegment(segments[0], 0, 59)
	if err != nil {
		return nil, err
	}

	hours, err := parseCronSegment(segments[1], 0, 23)
	if err != nil {
		return nil, err
	}

	days, err := parseCronSegment(segments[2], 1, 31)
	if err != nil {
		return nil, err
	}

	months, err := parseCronSegment(segments[3], 1, 12)
	if err != nil {
		return nil, err
	}

	daysOfWeek, err := parseCronSegment(segments[4], 0, 6)
	if err != nil {
		return nil, err
	}

	return &Schedule{
		Minutes:    minutes,
		Hours:      hours,
		Days:       days,
		Months:     months,
		DaysOfWeek: daysOfWeek,
		rawExpr:    cronExpr,
	}, nil
}

// parseCronSegment parses a single cron expression segment and
// returns its time schedule slots.
func parseCronSegment(segment string, min int, max int) (map[int]struct{}, error) {
	slots := map[int]struct{}{}

	list := strings.Split(segment, ",")
	for _, p := range list {
		stepParts := strings.Split(p, "/")

		// step (*/n, 1-30/n)
		var step int
		switch len(stepParts) {
		case 1:
			step = 1
		case 2:
			parsedStep, err := strconv.Atoi(stepParts[1])
			if err != nil {
				return nil, err
			}
			if parsedStep < 1 || parsedStep > max {
				return nil, fmt.Errorf("invalid segment step boundary - the step must be between 1 and the %d", max)
			}
			step = parsedStep
		default:
			return nil, errors.New("invalid segment step format - must be in the format */n or 1-30/n")
		}

		// find the min and max range of the segment part
		var rangeMin, rangeMax int
		if stepParts[0] == "*" {
			rangeMin = min
			rangeMax = max
		} else {
			// single digit (1) or range (1-30)
			rangeParts := strings.Split(stepParts[0], "-")
			switch len(rangeParts) {
			case 1:
				if step != 1 {
					return nil, errors.New("invalid segement step - step > 1 could be used only with the wildcard or range format")
				}
				parsed, err := strconv.Atoi(rangeParts[0])
				if err != nil {
					return nil, err
				}
				if parsed < min || parsed > max {
					return nil, errors.New("invalid segment value - must be between the min and max of the segment")
				}
				rangeMin = parsed
				rangeMax = rangeMin
			case 2:
				parsedMin, err := strconv.Atoi(rangeParts[0])
				if err != nil {
					return nil, err
				}
				if parsedMin < min || parsedMin > max {
					return nil, fmt.Errorf("invalid segment range minimum - must be between %d and %d", min, max)
				}
				rangeMin = parsedMin

				parsedMax, err := strconv.Atoi(rangeParts[1])
				if err != nil {
					return nil, err
				}
				if parsedMax < parsedMin || parsedMax > max {
					return nil, fmt.Errorf("invalid segment range maximum - must be between %d and %d", rangeMin, max)
				}
				rangeMax = parsedMax
			default:
				return nil, errors.New("invalid segment range format - the range must have 1 or 2 parts")
			}
		}

		// fill the slots
		for i := rangeMin; i <= rangeMax; i += step {
			slots[i] = struct{}{}
		}
	}

	return slots, nil
}
