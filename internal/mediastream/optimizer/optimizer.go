package optimizer

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"seanime/internal/events"
	"seanime/internal/mediastream/videofile"
	"seanime/internal/util"
)

const (
	QualityLow    Quality = "low"
	QualityMedium Quality = "medium"
	QualityHigh   Quality = "high"
	QualityMax    Quality = "max"
)

type (
	Quality string

	Optimizer struct {
		wsEventManager  events.WSEventManagerInterface
		logger          *zerolog.Logger
		libraryDir      mo.Option[string]
		concurrentTasks int
	}

	NewOptimizerOptions struct {
		Logger         *zerolog.Logger
		WSEventManager events.WSEventManagerInterface
	}
)

func NewOptimizer(opts *NewOptimizerOptions) *Optimizer {
	ret := &Optimizer{
		logger:          opts.Logger,
		wsEventManager:  opts.WSEventManager,
		libraryDir:      mo.None[string](),
		concurrentTasks: 2,
	}
	return ret
}

func (o *Optimizer) SetLibraryDir(libraryDir string) {
	o.libraryDir = mo.Some[string](libraryDir)
}

/////////////

type StartMediaOptimizationOptions struct {
	Filepath          string
	Quality           Quality
	AudioChannelIndex int
	MediaInfo         *videofile.MediaInfo
}

func (o *Optimizer) StartMediaOptimization(opts *StartMediaOptimizationOptions) (err error) {
	defer util.HandlePanicInModuleWithError("mediastream/optimizer/StartMediaOptimization", &err)

	o.logger.Debug().Any("opts", opts).Msg("mediastream: Starting media optimization")

	if !o.libraryDir.IsPresent() {
		return fmt.Errorf("library directory not set")
	}

	if opts.Filepath == "" {
		return fmt.Errorf("no filepath")
	}

	return
}

func qualityToPreset(quality Quality) string {
	switch quality {
	case QualityLow:
		return "ultrafast"
	case QualityMedium:
		return "veryfast"
	case QualityHigh:
		return "fast"
	case QualityMax:
		return "medium"
	default:
		return "veryfast"
	}
}
