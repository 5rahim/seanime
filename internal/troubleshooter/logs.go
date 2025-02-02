package troubleshooter

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"seanime/internal/util"
	"sort"
	"strings"
	"time"
)

type AnalysisResult struct {
	Items []AnalysisResultItem `json:"items"`
}

type AnalysisResultItem struct {
	Observation    string   `json:"observation"`
	Recommendation string   `json:"recommendation"`
	Severity       string   `json:"severity"`
	Errors         []string `json:"errors"`
	Warnings       []string `json:"warnings"`
	Logs           []string `json:"logs"`
}

// RuleBuilder provides a fluent interface for building rules
type RuleBuilder struct {
	name          string
	description   string
	conditions    []condition
	platforms     []string
	branches      []branch
	defaultBranch *branch
	state         *AppState
}

type condition struct {
	check   func(LogLine) bool
	message string // For debugging/logging
}

type branch struct {
	conditions     []condition
	observation    string
	recommendation string
	severity       string
}

// NewRule starts building a new rule
func NewRule(name string) *RuleBuilder {
	return &RuleBuilder{
		name:       name,
		conditions: []condition{},
		branches:   []branch{},
		defaultBranch: &branch{
			severity: "info",
		},
	}
}

// Desc adds a description to the rule
func (r *RuleBuilder) Desc(desc string) *RuleBuilder {
	r.description = desc
	return r
}

// When adds a base condition that must be met
func (r *RuleBuilder) When(check func(LogLine) bool, message string) *RuleBuilder {
	r.conditions = append(r.conditions, condition{check: check, message: message})
	return r
}

// ModuleIs adds a module condition
func (r *RuleBuilder) ModuleIs(module Module) *RuleBuilder {
	return r.When(func(l LogLine) bool {
		return l.Module == string(module)
	}, "module is "+string(module))
}

// LevelIs adds a level condition
func (r *RuleBuilder) LevelIs(level Level) *RuleBuilder {
	return r.When(func(l LogLine) bool {
		return l.Level == string(level)
	}, "level is "+string(level))
}

// MessageContains adds a message contains condition
func (r *RuleBuilder) MessageContains(substr string) *RuleBuilder {
	return r.When(func(l LogLine) bool {
		return strings.Contains(l.Message, substr)
	}, "message contains "+substr)
}

// MessageMatches adds a message regex condition
func (r *RuleBuilder) MessageMatches(pattern string) *RuleBuilder {
	return r.When(func(l LogLine) bool {
		matched, err := util.MatchesRegex(l.Message, pattern)
		return err == nil && matched
	}, "message matches "+pattern)
}

// OnPlatforms restricts the rule to specific platforms
func (r *RuleBuilder) OnPlatforms(platforms ...string) *RuleBuilder {
	r.platforms = platforms
	return r
}

// Branch adds a new branch with additional conditions
func (r *RuleBuilder) Branch() *BranchBuilder {
	return &BranchBuilder{
		rule: r,
		branch: branch{
			conditions: []condition{},
		},
	}
}

// Then sets the default observation and recommendation
func (r *RuleBuilder) Then(observation, recommendation string) *RuleBuilder {
	r.defaultBranch.observation = observation
	r.defaultBranch.recommendation = recommendation
	return r
}

// WithSeverity sets the default severity
func (r *RuleBuilder) WithSeverity(severity string) *RuleBuilder {
	r.defaultBranch.severity = severity
	return r
}

// BranchBuilder helps build conditional branches
type BranchBuilder struct {
	rule   *RuleBuilder
	branch branch
}

// When adds a condition to the branch
func (b *BranchBuilder) When(check func(LogLine) bool, message string) *BranchBuilder {
	b.branch.conditions = append(b.branch.conditions, condition{check: check, message: message})
	return b
}

// Then sets the branch observation and recommendation
func (b *BranchBuilder) Then(observation, recommendation string) *RuleBuilder {
	b.branch.observation = observation
	b.branch.recommendation = recommendation
	b.rule.branches = append(b.rule.branches, b.branch)
	return b.rule
}

// WithSeverity sets the branch severity
func (b *BranchBuilder) WithSeverity(severity string) *BranchBuilder {
	b.branch.severity = severity
	return b
}

// matches checks if a log line matches the rule and returns the matching branch
func (r *RuleBuilder) matches(line LogLine, platform string) (bool, *branch) {
	// Check platform restrictions
	if len(r.platforms) > 0 && !util.Contains(r.platforms, platform) {
		return false, nil
	}

	// Check base conditions
	for _, cond := range r.conditions {
		if !cond.check(line) {
			return false, nil
		}
	}

	// Check branches in order
	for _, branch := range r.branches {
		matches := true
		for _, cond := range branch.conditions {
			if !cond.check(line) {
				matches = false
				break
			}
		}
		if matches {
			return true, &branch
		}
	}

	// If no branches match but base conditions do, use default branch
	return true, r.defaultBranch
}

// NewAnalyzer creates a new analyzer with the default rule groups
func NewAnalyzer(opts NewTroubleshooterOptions) *Troubleshooter {
	a := &Troubleshooter{
		logsDir: opts.LogsDir,
		logger:  opts.Logger,
		state:   opts.State,
		rules:   defaultRules(),
	}
	return a
}

// defaultRules returns the default set of rules
func defaultRules() []RuleBuilder {
	return []RuleBuilder{
		*mpvRules(),
	}
}

// Analyze analyzes the logs in the logs directory and returns an AnalysisResult
// App.OnFlushLogs should be called before this function
func (t *Troubleshooter) Analyze() (AnalysisResult, error) {

	files, err := os.ReadDir(t.logsDir)
	if err != nil {
		return AnalysisResult{}, err
	}

	if len(files) == 0 {
		return AnalysisResult{}, errors.New("no logs found")
	}

	// Get the latest server log file
	// name: seanime-<timestamp>.log
	// e.g., seanime-2025-01-21-12-00-00.log
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() > files[j].Name()
	})

	latestFile := files[0]

	return analyzeLogFile(filepath.Join(t.logsDir, latestFile.Name()))
}

// LogLine represents a parsed log line
type LogLine struct {
	Timestamp time.Time
	Line      string
	Module    string
	Level     string
	Message   string
}

// analyzeLogFile analyzes a log file and returns an AnalysisResult
func analyzeLogFile(filepath string) (res AnalysisResult, err error) {
	platform := runtime.GOOS

	// Read the log file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return res, err
	}

	lines := strings.Split(string(content), "\n")

	// Get lines no older than 1 hour
	_lines := []string{}
	for _, line := range lines {
		timestamp, ok := util.SliceStrTo(line, len(time.DateTime))
		if !ok {
			continue
		}
		timestampTime, err := time.Parse(time.DateTime, timestamp)
		if err != nil {
			continue
		}
		if time.Since(timestampTime) < 1*time.Hour {
			_lines = append(_lines, line)
		}
	}
	lines = _lines

	// Parse lines into LogLine
	logLines := []LogLine{}
	for _, line := range lines {
		logLine, err := parseLogLine(line)
		if err != nil {
			continue
		}
		logLines = append(logLines, logLine)
	}

	// Group log lines by rule group
	type matchGroup struct {
		ruleGroup *RuleBuilder
		branch    *branch
		logLines  []LogLine
	}
	matches := make(map[string]*matchGroup) // key is rule group name

	// For each log line, check against all rules
	for _, logLine := range logLines {
		for _, ruleGroup := range defaultRules() {
			if matched, branch := ruleGroup.matches(logLine, platform); matched {
				if _, ok := matches[ruleGroup.name]; !ok {
					matches[ruleGroup.name] = &matchGroup{
						ruleGroup: &ruleGroup,
						branch:    branch,
						logLines:  []LogLine{},
					}
				}
				matches[ruleGroup.name].logLines = append(matches[ruleGroup.name].logLines, logLine)
				break // Stop checking other rules in this group once we find a match
			}
		}
	}

	// Convert matches to analysis result items
	for _, group := range matches {
		item := AnalysisResultItem{
			Observation:    group.branch.observation,
			Recommendation: group.branch.recommendation,
			Severity:       group.branch.severity,
		}

		// Add log lines based on their level
		for _, logLine := range group.logLines {
			switch logLine.Level {
			case "error":
				item.Errors = append(item.Errors, logLine.Line)
			case "warning":
				item.Warnings = append(item.Warnings, logLine.Line)
			default:
				item.Logs = append(item.Logs, logLine.Line)
			}
		}

		res.Items = append(res.Items, item)
	}

	return res, nil
}

func parseLogLine(line string) (ret LogLine, err error) {

	ret.Line = line

	timestamp, ok := util.SliceStrTo(line, len(time.DateTime))
	if !ok {
		return LogLine{}, errors.New("failed to parse timestamp")
	}
	timestampTime, err := time.Parse(time.DateTime, timestamp)
	if err != nil {
		return LogLine{}, errors.New("failed to parse timestamp")
	}
	ret.Timestamp = timestampTime

	rest, ok := util.SliceStrFrom(line, len(timestamp))
	if !ok {
		return LogLine{}, errors.New("failed to parse rest")
	}
	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, "|ERR|") {
		ret.Level = "error"
	} else if strings.HasPrefix(rest, "|WRN|") {
		ret.Level = "warning"
	} else if strings.HasPrefix(rest, "|INF|") {
		ret.Level = "info"
	} else if strings.HasPrefix(rest, "|DBG|") {
		ret.Level = "debug"
	} else if strings.HasPrefix(rest, "|TRC|") {
		ret.Level = "trace"
	} else if strings.HasPrefix(rest, "|PNC|") {
		ret.Level = "panic"
	}

	// Remove the level prefix
	rest, ok = util.SliceStrFrom(rest, 6)
	if !ok {
		return LogLine{}, errors.New("failed to parse rest")
	}

	// Get the module (string before `>`)
	moduleCaretIndex := strings.Index(rest, ">")
	if moduleCaretIndex != -1 {
		ret.Module = strings.TrimSpace(rest[:moduleCaretIndex])
		rest = strings.TrimSpace(rest[moduleCaretIndex+1:])
	}

	ret.Message = rest

	return
}
