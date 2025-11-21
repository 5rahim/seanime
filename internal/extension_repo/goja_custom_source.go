package extension_repo

import (
	"context"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/events"
	"seanime/internal/extension"
	hibikecustomsource "seanime/internal/extension/hibike/customsource"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/util"

	"github.com/rs/zerolog"
)

type GojaCustomSource struct {
	*gojaProviderBase
	extId               string
	extensionIdentifier int
}

func NewGojaCustomSource(ext *extension.Extension, language extension.Language, logger *zerolog.Logger, runtimeManager *goja_runtime.Manager, wsEventManager events.WSEventManagerInterface) (hibikecustomsource.Provider, *GojaCustomSource, error) {
	base, err := initializeProviderBase(ext, language, logger, runtimeManager, wsEventManager)
	if err != nil {
		return nil, nil, err
	}

	provider := &GojaCustomSource{
		extId:            ext.ID,
		gojaProviderBase: base,
	}
	return provider, provider, nil
}

func (g *GojaCustomSource) GetExtensionIdentifier() int {
	return g.extensionIdentifier
}

func (g *GojaCustomSource) ListAnime(ctx context.Context, search string, page int, perPage int) (ret *hibikecustomsource.ListAnimeResponse, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".ListAnime", &err)

	g.logger.Debug().Str("extension", g.extId).Str("search", search).Msg("custom source: Fetching anime")

	method, err := g.callClassMethod(ctx, "listAnime", search, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to call search method: %w", err)
	}

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for promise: %w", err)
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return ret, nil
}

func (g *GojaCustomSource) ListManga(ctx context.Context, search string, page int, perPage int) (ret *hibikecustomsource.ListMangaResponse, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".ListManga", &err)

	g.logger.Debug().Str("extension", g.extId).Str("search", search).Msg("custom source: Fetching manga")

	method, err := g.callClassMethod(ctx, "listManga", search, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to call search method: %w", err)
	}

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for promise: %w", err)
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return ret, nil
}

func (g *GojaCustomSource) GetSettings() (ret hibikecustomsource.Settings) {
	defer util.HandlePanicInModuleThen(g.ext.ID+".GetSettings", func() {
		ret = hibikecustomsource.Settings{}
	})

	method, err := g.callClassMethod(context.Background(), "getSettings")
	if err != nil {
		return
	}

	err = g.unmarshalValue(method, &ret)
	if err != nil {
		return
	}

	return
}

func (g *GojaCustomSource) GetAnime(ctx context.Context, id []int) (ret []*anilist.BaseAnime, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".GetAnime", &err)

	g.logger.Debug().Str("extension", g.extId).Ints("ids", id).Msg("custom source: Getting anime")

	method, err := g.callClassMethod(ctx, "getAnime", id)
	if err != nil {
		return nil, fmt.Errorf("failed to call search method: %w", err)
	}

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for promise: %w", err)
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return ret, nil
}

func (g *GojaCustomSource) GetAnimeWithRelations(ctx context.Context, id int) (ret *anilist.CompleteAnime, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".GetAnimeWithRelations", &err)

	g.logger.Debug().Str("extension", g.extId).Int("id", id).Msg("custom source: Getting anime with relations")

	method, err := g.callClassMethod(ctx, "getAnimeWithRelations", id)
	if err != nil {
		return nil, fmt.Errorf("failed to call search method: %w", err)
	}

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for promise: %w", err)
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	if ret.Relations == nil {
		return nil, fmt.Errorf("relations not found")
	}

	return ret, nil
}

func (g *GojaCustomSource) GetAnimeMetadata(ctx context.Context, id int) (ret *metadata.AnimeMetadata, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".GetAnimeMetadata", &err)

	g.logger.Debug().Str("extension", g.extId).Int("id", id).Msg("custom source: Getting anime metadata")

	method, err := g.callClassMethod(ctx, "getAnimeMetadata", id)
	if err != nil {
		return nil, fmt.Errorf("failed to call search method: %w", err)
	}

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for promise: %w", err)
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return ret, nil
}

func (g *GojaCustomSource) GetAnimeDetails(ctx context.Context, id int) (ret *anilist.AnimeDetailsById_Media, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".GetAnimeDetails", &err)

	g.logger.Debug().Str("extension", g.extId).Int("id", id).Msg("custom source: Getting anime details")

	method, err := g.callClassMethod(ctx, "getAnimeDetails", id)
	if err != nil {
		return nil, fmt.Errorf("failed to call search method: %w", err)
	}

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for promise: %w", err)
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return &anilist.AnimeDetailsById_Media{}, nil
	}

	return ret, nil
}

func (g *GojaCustomSource) GetManga(ctx context.Context, id []int) (ret []*anilist.BaseManga, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".GetManga", &err)

	g.logger.Debug().Str("extension", g.extId).Ints("ids", id).Msg("custom source: Getting manga")

	method, err := g.callClassMethod(ctx, "getManga", id)
	if err != nil {
		return nil, fmt.Errorf("failed to call search method: %w", err)
	}

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for promise: %w", err)
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return ret, nil
}

func (g *GojaCustomSource) GetMangaDetails(ctx context.Context, id int) (ret *anilist.MangaDetailsById_Media, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".GetMangaDetails", &err)

	g.logger.Debug().Str("extension", g.extId).Int("id", id).Msg("custom source: Getting manga details")

	method, err := g.callClassMethod(ctx, "getMangaDetails", id)
	if err != nil {
		return nil, fmt.Errorf("failed to call search method: %w", err)
	}

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for promise: %w", err)
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return &anilist.MangaDetailsById_Media{}, nil
	}

	return ret, nil
}

//----------------------------------------------------------------------------------------------------------------------------------------------------
