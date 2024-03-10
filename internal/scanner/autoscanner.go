package scanner

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/autodownloader"
	"github.com/seanime-app/seanime/internal/db"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/summary"
	"github.com/seanime-app/seanime/internal/util"
	"sync"
	"time"
)

type (
	AutoScanner struct {
		fileActionCh         chan struct{} // Used to notify the scanner that a file action has occurred.
		waiting              bool          // Used to prevent multiple scans from occurring at the same time.
		missedAction         bool          // Used to indicate that a file action was missed while scanning.
		mu                   sync.Mutex
		scannedCh            chan struct{}
		waitTime             time.Duration // Wait time to listen to additional changes before triggering a scan.
		enabled              bool
		AnilistClientWrapper anilist.ClientWrapperInterface
		Logger               *zerolog.Logger
		WSEventManager       events.IWSEventManager
		Database             *db.Database                   // Database instance is required to update the local files.
		AutoDownloader       *autodownloader.AutoDownloader // AutoDownloader instance is required to refresh queue.
	}
	NewAutoScannerOptions struct {
		Database             *db.Database
		AnilistClientWrapper anilist.ClientWrapperInterface
		Logger               *zerolog.Logger
		WSEventManager       events.IWSEventManager
		Enabled              bool
		AutoDownloader       *autodownloader.AutoDownloader
		WaitTime             time.Duration
	}
)

func NewAutoScanner(opts *NewAutoScannerOptions) *AutoScanner {
	wt := time.Second * 15 // Default wait time is 15 seconds.
	if opts.WaitTime > 0 {
		wt = opts.WaitTime
	}

	return &AutoScanner{
		fileActionCh:         make(chan struct{}, 1),
		scannedCh:            make(chan struct{}, 1),
		waiting:              false,
		missedAction:         false,
		waitTime:             wt,
		enabled:              opts.Enabled,
		AutoDownloader:       opts.AutoDownloader,
		AnilistClientWrapper: opts.AnilistClientWrapper,
		Logger:               opts.Logger,
		WSEventManager:       opts.WSEventManager,
		Database:             opts.Database,
	}
}

// Notify is used to notify the AutoScanner that a file action has occurred.
func (as *AutoScanner) Notify() {
	if as == nil {
		return
	}

	defer util.HandlePanicInModuleThen("scanner/autoscanner/Notify", func() {
		as.Logger.Error().Msg("autoscanner: recovered from panic")
	})

	as.mu.Lock()
	defer as.mu.Unlock()

	// If we are currently scanning, we will set the missedAction flag to true.
	if as.waiting {
		as.missedAction = true
		return
	}

	if as.enabled {
		go func() {
			// Otherwise, we will send a signal to the fileActionCh.
			as.fileActionCh <- struct{}{}
		}()
	}
}

// Start starts the AutoScanner in a goroutine.
func (as *AutoScanner) Start() {
	go func() {
		if as.enabled {
			as.Logger.Info().Msg("autoscanner: Module started")
		}

		as.watch()
	}()
}

// SetEnabled should be called after the settings are fetched and updated from the database.
func (as *AutoScanner) SetEnabled(enabled bool) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.enabled = enabled
}

// watch is used to watch for file actions and trigger a scan.
// When a file action occurs, it will wait 30 seconds before triggering a scan.
// If another file action occurs within that 30 seconds, it will reset the timer.
// After the 30 seconds have passed, it will trigger a scan.
// When a scan is complete, it will check the missedAction flag and trigger another scan if necessary.
func (as *AutoScanner) watch() {

	defer util.HandlePanicInModuleThen("scanner/autoscanner/watch", func() {
		as.Logger.Error().Msg("autoscanner: recovered from panic")
	})

	for {
		// Block until the file action channel is ready to receive a signal.
		<-as.fileActionCh
		as.waitAndScan()
	}

}

// waitAndScan is used to wait for additional file actions before triggering a scan.
func (as *AutoScanner) waitAndScan() {
	as.Logger.Trace().Msgf("autoscanner: File action occurred, waiting %v seconds before triggering a scan.", as.waitTime.Seconds())
	as.mu.Lock()
	as.waiting = true       // Set the scanning flag to true.
	as.missedAction = false // Reset the missedAction flag.
	as.mu.Unlock()

	// Wait 30 seconds before triggering a scan.
	// During this time, if another file action occurs, it will reset the timer after it has expired.
	<-time.After(as.waitTime)

	as.mu.Lock()
	// If a file action occurred while we were waiting, we will trigger another scan.
	if as.missedAction {
		as.Logger.Trace().Msg("autoscanner: Missed file action")
		as.mu.Unlock()
		as.waitAndScan()
		return
	}

	as.waiting = false
	as.mu.Unlock()

	// Trigger a scan.
	as.scan()
}

// scan is used to trigger a scan.
func (as *AutoScanner) scan() {

	defer util.HandlePanicInModuleThen("scanner/autoscanner/scan", func() {
		as.Logger.Error().Msg("autoscanner: recovered from panic")
	})

	// Create scan summary logger
	scanSummaryLogger := summary.NewScanSummaryLogger()

	as.Logger.Trace().Msg("autoscanner: Starting scanner")
	as.WSEventManager.SendEvent(events.AutoScanStarted, nil)
	defer as.WSEventManager.SendEvent(events.AutoScanCompleted, nil)

	settings, err := as.Database.GetSettings()
	if err != nil || settings == nil {
		as.Logger.Error().Err(err).Msg("autoscanner: failed to get settings")
		return
	}

	if settings.Library.LibraryPath == "" {
		as.Logger.Error().Msg("autoscanner: library path is not set")
		return
	}

	// Get the user's account
	acc, err := as.Database.GetAccount()
	if err != nil {
		as.Logger.Error().Err(err).Msg("autoscanner: failed to get account")
		return
	}

	// Get existing local files
	existingLfs, _, err := as.Database.GetLocalFiles()
	if err != nil {
		as.Logger.Error().Err(err).Msg("autoscanner: failed to get existing local files")
		return
	}

	// Create a new scanner
	sc := Scanner{
		DirPath:              settings.Library.LibraryPath,
		Username:             acc.Username,
		Enhanced:             false, // Do not use enhanced mode for auto scanner.
		AnilistClientWrapper: as.AnilistClientWrapper,
		Logger:               as.Logger,
		WSEventManager:       as.WSEventManager,
		ExistingLocalFiles:   existingLfs,
		SkipLockedFiles:      true, // Skip locked files by default.
		SkipIgnoredFiles:     true,
		ScanSummaryLogger:    scanSummaryLogger,
	}

	allLfs, err := sc.Scan()
	if err != nil {
		if errors.Is(err, ErrNoLocalFiles) {
			return
		} else {
			as.Logger.Error().Err(err).Msg("autoscanner: failed to scan library")
			return
		}
	}

	if as.Database != nil && len(allLfs) > 0 {
		as.Logger.Trace().Msg("autoscanner: Updating local files")

		// Insert the local files
		_, err = as.Database.InsertLocalFiles(allLfs)
		if err != nil {
			as.Logger.Error().Err(err).Msg("failed to insert local files")
			return
		}

		// Save the scan summary
		err = as.Database.InsertScanSummary(scanSummaryLogger.GenerateSummary())
		if err != nil {
			as.Logger.Error().Err(err).Msg("failed to insert scan summary")
		}
	}

	// Refresh the queue
	go as.AutoDownloader.CleanUpDownloadedItems()

	return
}
