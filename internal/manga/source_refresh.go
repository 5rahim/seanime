package manga

import (
	"context"
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/events"
	"seanime/internal/extension"
	manga_providers "seanime/internal/manga/providers"
	"seanime/internal/util"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

type MangaSourceRefreshMode string

const (
	MangaSourceRefreshSelected        MangaSourceRefreshMode = "refresh_selected"
	MangaSourceRefreshMissing         MangaSourceRefreshMode = "find_missing"
	MangaSourceRefreshSelectedMissing MangaSourceRefreshMode = "refresh_and_find"
	MangaSourceRefreshAll             MangaSourceRefreshMode = "reevaluate_all"
)

type MangaSourceRefreshStatus string

const (
	MangaSourceRefreshRunning   MangaSourceRefreshStatus = "running"
	MangaSourceRefreshStopping  MangaSourceRefreshStatus = "stopping"
	MangaSourceRefreshCompleted MangaSourceRefreshStatus = "completed"
	MangaSourceRefreshCancelled MangaSourceRefreshStatus = "cancelled"
	MangaSourceRefreshFailed    MangaSourceRefreshStatus = "failed"
)

type MangaSourceRefreshStage string

const (
	MangaSourceRefreshRefreshing  MangaSourceRefreshStage = "refreshing"
	MangaSourceRefreshDiscovering MangaSourceRefreshStage = "discovering"
	MangaSourceRefreshDone        MangaSourceRefreshStage = "done"
)

type MangaSourceRefreshIssue struct {
	MediaId   int      `json:"mediaId"`
	Title     string   `json:"title"`
	Kind      string   `json:"kind"`
	Providers []string `json:"providers"`
}

type MangaSourceRefreshChange struct {
	MediaId      int    `json:"mediaId"`
	Title        string `json:"title"`
	FromProvider string `json:"fromProvider"`
	ToProvider   string `json:"toProvider"`
	Kind         string `json:"kind"`
}

type MangaSourceRefreshResult struct {
	Refreshed int                        `json:"refreshed"`
	Found     int                        `json:"found"`
	Replaced  int                        `json:"replaced"`
	NotFound  int                        `json:"notFound"`
	Failed    int                        `json:"failed"`
	Changes   []MangaSourceRefreshChange `json:"changes"`
	Issues    []MangaSourceRefreshIssue  `json:"issues"`
}

type MangaSourceRefreshJob struct {
	Id      string                   `json:"id"`
	Mode    MangaSourceRefreshMode   `json:"mode"`
	Status  MangaSourceRefreshStatus `json:"status"`
	Stage   MangaSourceRefreshStage  `json:"stage"`
	Current int                      `json:"current"`
	Total   int                      `json:"total"`
	Result  MangaSourceRefreshResult `json:"result"`
	Error   string                   `json:"error,omitempty"`
}

type mangaSourceRefreshState struct {
	owner  string
	job    MangaSourceRefreshJob
	cancel context.CancelFunc
}

type mangaSourceRefreshCompleted struct {
	job       MangaSourceRefreshJob
	expiresAt time.Time
}

type mangaSourceRefreshPlan struct {
	entry           *anilist.MangaListEntry
	mediaId         int
	title           string
	currentProvider string
	providers       []string
}

type mangaSourceRefreshTask struct {
	plan       *mangaSourceRefreshPlan
	providerId string
}

type mangaSourceRefreshTaskResult struct {
	mediaId    int
	providerId string
	container  *ChapterContainer
	err        error
}

var (
	ErrMangaSourceRefreshConflict = errors.New("another manga source refresh is already running")
	ErrNoMangaProviders           = errors.New("no manga providers are installed")
)

func IsMangaSourceRefreshModeValid(mode MangaSourceRefreshMode) bool {
	switch mode {
	case MangaSourceRefreshSelected, MangaSourceRefreshMissing, MangaSourceRefreshSelectedMissing, MangaSourceRefreshAll:
		return true
	default:
		return false
	}
}

func (r *Repository) StartMangaSourceRefresh(
	clientId string,
	mode MangaSourceRefreshMode,
	collection *anilist.MangaCollection,
	mediaIds ...int,
) (*MangaSourceRefreshJob, error) {
	if !IsMangaSourceRefreshModeValid(mode) {
		return nil, errors.New("invalid manga source refresh mode")
	}
	if clientId == "" {
		clientId = "0"
	}
	if job, err := r.GetActiveMangaSourceRefresh(clientId); job != nil || err != nil {
		return job, err
	}

	preferences, err := r.GetMangaPreferences()
	if err != nil {
		return nil, err
	}
	providerIds := r.getMangaProviderIds()
	if len(providerIds) == 0 && mode != MangaSourceRefreshSelected {
		return nil, ErrNoMangaProviders
	}
	phases := buildMangaSourceRefreshPhases(collection, preferences, providerIds, mode, mediaIds...)
	total := 0
	for _, phase := range phases {
		total += len(phase.plans)
	}

	r.sourceRefreshMu.Lock()
	r.cleanupSourceRefreshLog(time.Now())
	if r.sourceRefresh != nil {
		if r.sourceRefresh.owner == clientId {
			job := cloneSourceRefreshJob(r.sourceRefresh.job)
			r.sourceRefreshMu.Unlock()
			return &job, nil
		}
		r.sourceRefreshMu.Unlock()
		return nil, ErrMangaSourceRefreshConflict
	}

	ctx, cancel := context.WithCancel(context.Background())
	state := &mangaSourceRefreshState{
		owner:  clientId,
		cancel: cancel,
		job: MangaSourceRefreshJob{
			Id:     uuid.NewString(),
			Mode:   mode,
			Status: MangaSourceRefreshRunning,
			Stage:  firstMangaSourceRefreshStage(phases),
			Total:  total,
			Result: MangaSourceRefreshResult{
				Changes: make([]MangaSourceRefreshChange, 0),
				Issues:  make([]MangaSourceRefreshIssue, 0),
			},
		},
	}
	r.sourceRefresh = state
	delete(r.sourceRefreshLog, clientId)
	job := cloneSourceRefreshJob(state.job)
	r.sourceRefreshMu.Unlock()

	r.sendSourceRefreshJob(clientId, job)
	go r.runSourceRefresh(ctx, clientId, phases, providerIds)
	return &job, nil
}

func (r *Repository) GetActiveMangaSourceRefresh(clientId string) (*MangaSourceRefreshJob, error) {
	if clientId == "" {
		clientId = "0"
	}
	r.sourceRefreshMu.Lock()
	defer r.sourceRefreshMu.Unlock()
	r.cleanupSourceRefreshLog(time.Now())
	if r.sourceRefresh == nil {
		return nil, nil
	}
	if r.sourceRefresh.owner != clientId {
		return nil, ErrMangaSourceRefreshConflict
	}
	return new(cloneSourceRefreshJob(r.sourceRefresh.job)), nil
}

func (r *Repository) GetMangaSourceRefresh(clientId string) *MangaSourceRefreshJob {
	if clientId == "" {
		clientId = "0"
	}
	r.sourceRefreshMu.Lock()
	defer r.sourceRefreshMu.Unlock()
	r.cleanupSourceRefreshLog(time.Now())
	if r.sourceRefresh != nil && r.sourceRefresh.owner == clientId {
		return new(cloneSourceRefreshJob(r.sourceRefresh.job))
	}
	if completed, found := r.sourceRefreshLog[clientId]; found {
		return new(cloneSourceRefreshJob(completed.job))
	}
	return nil
}

func (r *Repository) StopMangaSourceRefresh(clientId string) (*MangaSourceRefreshJob, error) {
	if clientId == "" {
		clientId = "0"
	}
	r.sourceRefreshMu.Lock()
	if r.sourceRefresh != nil {
		if r.sourceRefresh.owner != clientId {
			if _, found := r.sourceRefreshLog[clientId]; found {
				delete(r.sourceRefreshLog, clientId)
				r.sourceRefreshMu.Unlock()
				return nil, nil
			}
			r.sourceRefreshMu.Unlock()
			return nil, ErrMangaSourceRefreshConflict
		}
		if r.sourceRefresh.job.Status == MangaSourceRefreshRunning {
			r.sourceRefresh.job.Status = MangaSourceRefreshStopping
			r.sourceRefresh.cancel()
		}
		job := cloneSourceRefreshJob(r.sourceRefresh.job)
		r.sourceRefreshMu.Unlock()
		r.sendSourceRefreshJob(clientId, job)
		return &job, nil
	}
	delete(r.sourceRefreshLog, clientId)
	r.sourceRefreshMu.Unlock()
	return nil, nil
}

type mangaSourceRefreshPhase struct {
	stage MangaSourceRefreshStage
	plans []*mangaSourceRefreshPlan
}

func buildMangaSourceRefreshPhases(
	collection *anilist.MangaCollection,
	preferences *MangaPreferences,
	providerIds []string,
	mode MangaSourceRefreshMode,
	mediaIds ...int,
) []mangaSourceRefreshPhase {
	entries := getRefreshableMangaEntries(collection, mediaIds...)
	selected := make([]*mangaSourceRefreshPlan, 0)
	missing := make([]*mangaSourceRefreshPlan, 0)
	all := make([]*mangaSourceRefreshPlan, 0)

	for _, entry := range entries {
		media := entry.GetMedia()
		if media == nil {
			continue
		}
		preference := preferences.Entries[media.ID]
		base := &mangaSourceRefreshPlan{
			entry:           entry,
			mediaId:         media.ID,
			title:           media.GetPreferredTitle(),
			currentProvider: preference.Provider,
		}
		if preference.Provider == "" {
			plan := *base
			plan.providers = append([]string(nil), providerIds...)
			missing = append(missing, &plan)
		} else {
			plan := *base
			plan.providers = []string{preference.Provider}
			selected = append(selected, &plan)
		}
		plan := *base
		plan.providers = append([]string(nil), providerIds...)
		all = append(all, &plan)
	}

	switch mode {
	case MangaSourceRefreshSelected:
		return []mangaSourceRefreshPhase{{stage: MangaSourceRefreshRefreshing, plans: selected}}
	case MangaSourceRefreshMissing:
		return []mangaSourceRefreshPhase{{stage: MangaSourceRefreshDiscovering, plans: missing}}
	case MangaSourceRefreshSelectedMissing:
		return []mangaSourceRefreshPhase{
			{stage: MangaSourceRefreshRefreshing, plans: selected},
			{stage: MangaSourceRefreshDiscovering, plans: missing},
		}
	default:
		return []mangaSourceRefreshPhase{{stage: MangaSourceRefreshDiscovering, plans: all}}
	}
}

func getRefreshableMangaEntries(collection *anilist.MangaCollection, mediaIds ...int) []*anilist.MangaListEntry {
	if collection == nil || collection.MediaListCollection == nil {
		return nil
	}
	targets := make(map[int]struct{}, len(mediaIds))
	for _, mediaId := range mediaIds {
		targets[mediaId] = struct{}{}
	}
	entries := make(map[int]*anilist.MangaListEntry)
	for _, list := range collection.MediaListCollection.Lists {
		for _, entry := range list.GetEntries() {
			if entry == nil || entry.GetMedia() == nil || entry.GetStatus() == nil {
				continue
			}
			status := *entry.GetStatus()
			if status != anilist.MediaListStatusCurrent && status != anilist.MediaListStatusRepeating {
				continue
			}
			if len(targets) > 0 {
				if _, ok := targets[entry.GetMedia().ID]; !ok {
					continue
				}
			}
			entries[entry.GetMedia().ID] = entry
		}
	}
	sortedMediaIds := make([]int, 0, len(entries))
	for mediaId := range entries {
		sortedMediaIds = append(sortedMediaIds, mediaId)
	}
	sort.Ints(sortedMediaIds)
	ret := make([]*anilist.MangaListEntry, 0, len(sortedMediaIds))
	for _, mediaId := range sortedMediaIds {
		ret = append(ret, entries[mediaId])
	}
	return ret
}

func (r *Repository) getMangaProviderIds() []string {
	providerIds := make([]string, 0)
	extension.RangeExtensions[extension.MangaProviderExtension](r.extensionBankRef.Get(), func(id string, _ extension.MangaProviderExtension) bool {
		providerIds = append(providerIds, id)
		return true
	})
	sort.Strings(providerIds)
	return providerIds
}

func firstMangaSourceRefreshStage(phases []mangaSourceRefreshPhase) MangaSourceRefreshStage {
	for _, phase := range phases {
		if len(phase.plans) > 0 {
			return phase.stage
		}
	}
	return MangaSourceRefreshDone
}

func (r *Repository) runSourceRefresh(
	ctx context.Context,
	clientId string,
	phases []mangaSourceRefreshPhase,
	providerIds []string,
) {
	result := MangaSourceRefreshResult{
		Changes: make([]MangaSourceRefreshChange, 0),
		Issues:  make([]MangaSourceRefreshIssue, 0),
	}
	defer util.HandlePanicInModuleThen("manga/runSourceRefresh", func() {
		r.finishSourceRefresh(clientId, MangaSourceRefreshFailed, result, "Source refresh stopped unexpectedly")
	})
	changedMediaIds := make([]int, 0)
	limiters := make(map[string]*rate.Limiter, len(providerIds))
	for _, providerId := range providerIds {
		if providerId != "local-manga" {
			limiters[providerId] = rate.NewLimiter(rate.Limit(2), 1)
		}
	}

	for _, phase := range phases {
		if len(phase.plans) == 0 {
			continue
		}
		if ctx.Err() != nil {
			break
		}
		r.updateMangaSourceRefresh(clientId, func(job *MangaSourceRefreshJob) {
			job.Stage = phase.stage
			job.Result = result
		})
		phaseChanges := r.runMangaSourceRefreshPhase(ctx, clientId, phase.plans, limiters, &result)
		changedMediaIds = append(changedMediaIds, phaseChanges...)
	}

	if len(changedMediaIds) > 0 {
		r.NotifyPreferencesUpdated(changedMediaIds)
	}
	status := MangaSourceRefreshCompleted
	if ctx.Err() != nil {
		status = MangaSourceRefreshCancelled
	}
	r.finishSourceRefresh(clientId, status, result, "")
}

func (r *Repository) runMangaSourceRefreshPhase(
	ctx context.Context,
	clientId string,
	plans []*mangaSourceRefreshPlan,
	limiters map[string]*rate.Limiter,
	jobResult *MangaSourceRefreshResult,
) []int {
	results := make(chan mangaSourceRefreshTaskResult, countSourceRefreshTasks(plans))
	tasksByProvider := make(map[string]chan mangaSourceRefreshTask)
	var workers sync.WaitGroup

	for _, plan := range plans {
		for _, providerId := range plan.providers {
			if _, exists := tasksByProvider[providerId]; exists {
				continue
			}
			if _, found := r.extensionBankRef.Get().Get(providerId); !found {
				continue
			}
			queue := make(chan mangaSourceRefreshTask, len(plans))
			tasksByProvider[providerId] = queue
			workers.Add(1)
			go func(providerId string, tasks <-chan mangaSourceRefreshTask) {
				defer workers.Done()
				for task := range tasks {
					results <- r.fetchSourceRefreshTask(ctx, task, limiters[providerId])
				}
			}(providerId, queue)
		}
	}

	for _, plan := range plans {
		for _, providerId := range plan.providers {
			queue, found := tasksByProvider[providerId]
			if !found {
				results <- mangaSourceRefreshTaskResult{
					mediaId:    plan.mediaId,
					providerId: providerId,
					err:        errors.New("provider is unavailable"),
				}
				continue
			}
			queue <- mangaSourceRefreshTask{plan: plan, providerId: providerId}
		}
	}
	for _, queue := range tasksByProvider {
		close(queue)
	}
	go func() {
		workers.Wait()
		close(results)
	}()

	planByMediaId := make(map[int]*mangaSourceRefreshPlan, len(plans))
	resultsByMediaId := make(map[int][]mangaSourceRefreshTaskResult, len(plans))
	expectedByMediaId := make(map[int]int, len(plans))
	for _, plan := range plans {
		planByMediaId[plan.mediaId] = plan
		expectedByMediaId[plan.mediaId] = len(plan.providers)
	}

	changedMediaIds := make([]int, 0)
	for taskResult := range results {
		mediaResults := append(resultsByMediaId[taskResult.mediaId], taskResult)
		resultsByMediaId[taskResult.mediaId] = mediaResults
		if len(mediaResults) != expectedByMediaId[taskResult.mediaId] {
			continue
		}
		if hasCancelledMangaSourceRefreshResult(mediaResults) {
			continue
		}
		changed := r.finalizeMangaSourceRefreshPlan(planByMediaId[taskResult.mediaId], mediaResults, jobResult)
		if changed {
			changedMediaIds = append(changedMediaIds, taskResult.mediaId)
		}
		r.updateMangaSourceRefresh(clientId, func(job *MangaSourceRefreshJob) {
			job.Current++
			job.Result = *jobResult
		})
	}
	return changedMediaIds
}

func countSourceRefreshTasks(plans []*mangaSourceRefreshPlan) int {
	total := 0
	for _, plan := range plans {
		total += len(plan.providers)
	}
	return total
}

func (r *Repository) fetchSourceRefreshTask(
	ctx context.Context,
	task mangaSourceRefreshTask,
	limiter *rate.Limiter,
) mangaSourceRefreshTaskResult {
	ret := mangaSourceRefreshTaskResult{mediaId: task.plan.mediaId, providerId: task.providerId}
	if ctx.Err() != nil {
		ret.err = ctx.Err()
		return ret
	}
	wait := func() error {
		if limiter == nil {
			return ctx.Err()
		}
		return limiter.Wait(ctx)
	}
	media := task.plan.entry.GetMedia()
	container, err := r.GetMangaChapterContainer(&GetMangaChapterContainerOptions{
		Provider:           task.providerId,
		MediaId:            task.plan.mediaId,
		Titles:             media.GetAllTitles(),
		Year:               media.GetStartYearSafe(),
		skipCache:          true,
		beforeProviderCall: wait,
	})
	if err != nil {
		ret.err = err
		return ret
	}
	if container == nil || len(container.Chapters) == 0 {
		ret.err = ErrNoChapters
		return ret
	}
	ret.container = container
	pageBucket := r.getFcProviderBucket(task.providerId, task.plan.mediaId, bucketTypePage)
	dimensionsBucket := r.getFcProviderBucket(task.providerId, task.plan.mediaId, bucketTypePageDimensions)
	_ = r.fileCacher.Remove(pageBucket.Name())
	_ = r.fileCacher.Remove(dimensionsBucket.Name())
	return ret
}

func hasCancelledMangaSourceRefreshResult(results []mangaSourceRefreshTaskResult) bool {
	for _, result := range results {
		if errors.Is(result.err, context.Canceled) || errors.Is(result.err, context.DeadlineExceeded) {
			return true
		}
	}
	return false
}

func (r *Repository) finalizeMangaSourceRefreshPlan(
	plan *mangaSourceRefreshPlan,
	results []mangaSourceRefreshTaskResult,
	jobResult *MangaSourceRefreshResult,
) bool {
	candidates := make([]mangaSourceRefreshTaskResult, 0, len(results))
	failedProviders := make([]string, 0)
	for _, result := range results {
		if result.err != nil || result.container == nil || len(result.container.Chapters) == 0 {
			if isSourceRefreshProviderError(result.err) {
				failedProviders = append(failedProviders, result.providerId)
			}
			continue
		}
		candidates = append(candidates, result)
	}

	if len(candidates) == 0 {
		kind := "not_found"
		if len(failedProviders) > 0 {
			kind = "provider_error"
			jobResult.Failed++
		} else {
			jobResult.NotFound++
		}
		jobResult.Issues = append(jobResult.Issues, MangaSourceRefreshIssue{
			MediaId:   plan.mediaId,
			Title:     plan.title,
			Kind:      kind,
			Providers: failedProviders,
		})
		return false
	}

	chosen := chooseMangaSourceRefreshCandidate(candidates, plan.currentProvider, r.defaultMangaProvider())
	if plan.currentProvider == "" {
		_, err := r.PatchPreference(plan.mediaId, &MangaPreferencePatch{Provider: new(chosen.providerId)}, false)
		if err != nil {
			r.logger.Error().Err(err).Int("mediaId", plan.mediaId).Msg("manga: Failed to save discovered source")
			jobResult.Failed++
			jobResult.Issues = append(jobResult.Issues, MangaSourceRefreshIssue{
				MediaId: plan.mediaId, Title: plan.title, Kind: "provider_error", Providers: []string{chosen.providerId},
			})
			return false
		}
		jobResult.Found++
		jobResult.Changes = append(jobResult.Changes, MangaSourceRefreshChange{
			MediaId: plan.mediaId, Title: plan.title, ToProvider: chosen.providerId, Kind: "found",
		})
		return true
	}

	if chosen.providerId == plan.currentProvider {
		jobResult.Refreshed++
		return false
	}
	provider := chosen.providerId
	_, err := r.PatchPreference(plan.mediaId, &MangaPreferencePatch{Provider: &provider}, false)
	if err != nil {
		r.logger.Error().Err(err).Int("mediaId", plan.mediaId).Msg("manga: Failed to save replacement source")
		jobResult.Failed++
		jobResult.Issues = append(jobResult.Issues, MangaSourceRefreshIssue{
			MediaId: plan.mediaId, Title: plan.title, Kind: "provider_error", Providers: []string{chosen.providerId},
		})
		return false
	}
	jobResult.Replaced++
	jobResult.Changes = append(jobResult.Changes, MangaSourceRefreshChange{
		MediaId: plan.mediaId, Title: plan.title, FromProvider: plan.currentProvider, ToProvider: chosen.providerId, Kind: "replaced",
	})
	return true
}

func isSourceRefreshProviderError(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	// devnote: intentionally using == to avoid matching wrapped errors
	if err == ErrNoResults || err == ErrNoChapters {
		return false
	}
	return true
}

func chooseMangaSourceRefreshCandidate(
	candidates []mangaSourceRefreshTaskResult,
	currentProvider string,
	defaultProvider string,
) mangaSourceRefreshTaskResult {
	sort.SliceStable(candidates, func(i, j int) bool {
		iScore := mangaSourceRefreshChapterScore(candidates[i].container)
		jScore := mangaSourceRefreshChapterScore(candidates[j].container)
		if iScore != jScore {
			return iScore > jScore
		}
		iCurrent := candidates[i].providerId == currentProvider
		jCurrent := candidates[j].providerId == currentProvider
		if iCurrent != jCurrent {
			return iCurrent
		}
		iDefault := candidates[i].providerId == defaultProvider
		jDefault := candidates[j].providerId == defaultProvider
		if iDefault != jDefault {
			return iDefault
		}
		return candidates[i].providerId < candidates[j].providerId
	})
	return candidates[0]
}

func mangaSourceRefreshChapterScore(container *ChapterContainer) int {
	chapters := make(map[string]struct{})
	for _, chapter := range container.Chapters {
		if chapter == nil {
			continue
		}
		chapters[manga_providers.GetNormalizedChapter(chapter.Chapter)] = struct{}{}
	}
	return len(chapters)
}

func (r *Repository) defaultMangaProvider() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.settings == nil || r.settings.Manga == nil {
		return ""
	}
	return r.settings.Manga.DefaultProvider
}

func (r *Repository) updateMangaSourceRefresh(clientId string, update func(job *MangaSourceRefreshJob)) {
	r.sourceRefreshMu.Lock()
	if r.sourceRefresh == nil || r.sourceRefresh.owner != clientId {
		r.sourceRefreshMu.Unlock()
		return
	}
	update(&r.sourceRefresh.job)
	r.sourceRefresh.job = cloneSourceRefreshJob(r.sourceRefresh.job)
	job := cloneSourceRefreshJob(r.sourceRefresh.job)
	r.sourceRefreshMu.Unlock()
	r.sendSourceRefreshJob(clientId, job)
}

func (r *Repository) finishSourceRefresh(
	clientId string,
	status MangaSourceRefreshStatus,
	result MangaSourceRefreshResult,
	errorMessage string,
) {
	r.sourceRefreshMu.Lock()
	if r.sourceRefresh == nil || r.sourceRefresh.owner != clientId {
		r.sourceRefreshMu.Unlock()
		return
	}
	r.sourceRefresh.job.Status = status
	r.sourceRefresh.job.Stage = MangaSourceRefreshDone
	r.sourceRefresh.job.Result = result
	r.sourceRefresh.job.Error = errorMessage
	job := cloneSourceRefreshJob(r.sourceRefresh.job)
	r.sourceRefreshLog[clientId] = mangaSourceRefreshCompleted{job: job, expiresAt: time.Now().Add(24 * time.Hour)}
	r.sourceRefresh = nil
	r.sourceRefreshMu.Unlock()
	r.sendSourceRefreshJob(clientId, job)
}

func (r *Repository) sendSourceRefreshJob(clientId string, job MangaSourceRefreshJob) {
	r.wsEventManager.SendEventTo(clientId, events.MangaSourceRefreshUpdated, job)
}

func (r *Repository) cleanupSourceRefreshLog(now time.Time) {
	for clientId, completed := range r.sourceRefreshLog {
		if now.After(completed.expiresAt) {
			delete(r.sourceRefreshLog, clientId)
		}
	}
}

func cloneSourceRefreshJob(job MangaSourceRefreshJob) MangaSourceRefreshJob {
	job.Result.Changes = append([]MangaSourceRefreshChange(nil), job.Result.Changes...)
	job.Result.Issues = append([]MangaSourceRefreshIssue(nil), job.Result.Issues...)
	for i := range job.Result.Issues {
		job.Result.Issues[i].Providers = append([]string(nil), job.Result.Issues[i].Providers...)
	}
	return job
}
