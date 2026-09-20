package plugin

import (
	"context"
	"seanime/internal/extension"
	"seanime/internal/extension_repo/prompt"
	"seanime/internal/util/result"
	"strings"
)

func newPromptCache() *result.Cache[string, bool] {
	return result.NewCache[string, bool]()
}

func promptKey(parts ...string) string {
	ret := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		ret = append(ret, part)
	}

	return strings.Join(ret, ":")
}

// Ask prompts the user for permission before an action is performed.
func (a *AppContextImpl) Ask(ext *extension.Extension, opts prompt.Options) error {
	return a.ask(ext, opts)
}

func (a *AppContextImpl) ask(ext *extension.Extension, opts prompt.Options) error {
	manager, ok := a.promptManager.Get()
	if !ok || manager == nil {
		return prompt.ErrUnavailable
	}
	if opts.AllowLabel == "" {
		opts.AllowLabel = "Allow"
	}
	if opts.DenyLabel == "" {
		opts.DenyLabel = "Don't Allow"
	}

	return manager.Ask(context.Background(), ext, opts)
}
