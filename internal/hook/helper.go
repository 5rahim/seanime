package hook

import "seanime/internal/hook_resolver"

type HookTriggerOptions[T hook_resolver.Resolver] struct {
	Event              T
	Hook               func() *Hook[hook_resolver.Resolver]
	OnError            func(error) error
	OnDefaultPrevented func() error
	OnSuccess          func() error
}

// TriggerHook triggers the given hook with the provided event and handles
//
//	Example:
//	ok, err := hook.TriggerHook(&hook.HookTriggerOptions[*MissingEpisodesRequestedEvent]{
//		Hook: hook.GlobalHookManager.OnMissingEpisodesRequested,
//		Event: reqEvent,
//		OnDefaultPrevented: func() error {
//			event := new(MissingEpisodesEvent)
//			event.MissingEpisodes = missing
//			err = hook.GlobalHookManager.OnMissingEpisodes().Trigger(event)
//			if err != nil {
//				return nil
//			}
//			missing = event.MissingEpisodes
//			return nil
//		},
//		OnSuccess: func() error {
//			opts.AnimeCollection = reqEvent.AnimeCollection   // Override the anime collection
//			opts.LocalFiles = reqEvent.LocalFiles             // Override the local files
//			opts.SilencedMediaIds = reqEvent.SilencedMediaIds // Override the silenced media IDs
//			missing = reqEvent.MissingEpisodes
//			return nil
//		},
//	})
//	if err != nil {
//		return nil
//	}
//	if !ok {
//		return missing
//	}
func TriggerHook[T hook_resolver.Resolver](opts *HookTriggerOptions[T]) (cont bool, _ error) {
	if err := opts.Hook().Trigger(opts.Event); err != nil {
		// The hook errored out
		if opts.OnError != nil {
			return false, opts.OnError(err)
		}
		return false, err
	}

	// Default prevented
	if opts.Event.IsDefaultPrevented() {
		if opts.OnDefaultPrevented != nil {
			return false, opts.OnDefaultPrevented()
		}
		// No error but don't continue
		return false, nil
	}

	if opts.OnSuccess != nil {
		return true, opts.OnSuccess()
	}
	return true, nil
}
