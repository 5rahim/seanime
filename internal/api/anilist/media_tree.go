package anilist

import (
	"context"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"seanime/internal/util/result"
	"sync"

	"github.com/samber/lo"
)

type (
	CompleteAnimeRelationTree struct {
		*result.Map[int, *CompleteAnime]
	}

	FetchMediaTreeRelation = string
)

const (
	FetchMediaTreeSequels  FetchMediaTreeRelation = "sequels"
	FetchMediaTreePrequels FetchMediaTreeRelation = "prequels"
	FetchMediaTreeAll      FetchMediaTreeRelation = "all"
)

// NewCompleteAnimeRelationTree returns a new result.Map[int, *CompleteAnime].
// It is used to store the results of FetchMediaTree or FetchMediaTree calls.
func NewCompleteAnimeRelationTree() *CompleteAnimeRelationTree {
	return &CompleteAnimeRelationTree{result.NewMap[int, *CompleteAnime]()}
}

func (m *BaseAnime) FetchMediaTree(rel FetchMediaTreeRelation, anilistClient AnilistClient, rl *limiter.Limiter, tree *CompleteAnimeRelationTree, cache *CompleteAnimeCache) (err error) {
	if m == nil {
		return nil
	}

	defer util.HandlePanicInModuleWithError("anilist/BaseAnime.FetchMediaTree", &err)

	rl.Wait()
	res, err := anilistClient.CompleteAnimeByID(context.Background(), &m.ID)
	if err != nil {
		return err
	}
	return res.GetMedia().FetchMediaTree(rel, anilistClient, rl, tree, cache)
}

// FetchMediaTree populates the CompleteAnimeRelationTree with the given media's sequels and prequels.
// It also takes a CompleteAnimeCache to store the fetched media in and avoid duplicate fetches.
// It also takes a limiter.Limiter to limit the number of requests made to the AniList API.
func (m *CompleteAnime) FetchMediaTree(rel FetchMediaTreeRelation, anilistClient AnilistClient, rl *limiter.Limiter, tree *CompleteAnimeRelationTree, cache *CompleteAnimeCache) (err error) {
	if m == nil {
		return nil
	}

	defer util.HandlePanicInModuleWithError("anilist/CompleteAnime.FetchMediaTree", &err)

	if tree.Has(m.ID) {
		cache.Set(m.ID, m)
		return nil
	}
	cache.Set(m.ID, m)
	tree.Set(m.ID, m)

	if m.Relations == nil {
		return nil
	}

	// Get all edges
	edges := m.GetRelations().GetEdges()
	// Filter edges
	edges = lo.Filter(edges, func(_edge *CompleteAnime_Relations_Edges, _ int) bool {
		return (*_edge.RelationType == MediaRelationSequel || *_edge.RelationType == MediaRelationPrequel) &&
			*_edge.GetNode().Status != MediaStatusNotYetReleased &&
			_edge.IsBroadRelationFormat() && !tree.Has(_edge.GetNode().ID)
	})

	if len(edges) == 0 {
		return nil
	}

	doneCh := make(chan struct{})
	processEdges(edges, rel, anilistClient, rl, tree, cache, doneCh)

	for {
		select {
		case <-doneCh:
			return nil
		default:
		}
	}
}

// processEdges fetches the next node(s) for each edge in parallel.
func processEdges(edges []*CompleteAnime_Relations_Edges, rel FetchMediaTreeRelation, anilistClient AnilistClient, rl *limiter.Limiter, tree *CompleteAnimeRelationTree, cache *CompleteAnimeCache, doneCh chan struct{}) {
	var wg sync.WaitGroup
	wg.Add(len(edges))

	for i, item := range edges {
		go func(edge *CompleteAnime_Relations_Edges, _ int) {
			defer wg.Done()
			if edge == nil {
				return
			}
			processEdge(edge, rel, anilistClient, rl, tree, cache)
		}(item, i)
	}

	wg.Wait()

	go func() {
		close(doneCh)
	}()
}

func processEdge(edge *CompleteAnime_Relations_Edges, rel FetchMediaTreeRelation, anilistClient AnilistClient, rl *limiter.Limiter, tree *CompleteAnimeRelationTree, cache *CompleteAnimeCache) {
	defer util.HandlePanicInModuleThen("anilist/processEdge", func() {})
	cacheV, ok := cache.Get(edge.GetNode().ID)
	edgeCompleteAnime := cacheV
	if !ok {
		rl.Wait()
		// Fetch the next node
		res, err := anilistClient.CompleteAnimeByID(context.Background(), &edge.GetNode().ID)
		if err == nil {
			edgeCompleteAnime = res.GetMedia()
			cache.Set(edgeCompleteAnime.ID, edgeCompleteAnime)
		}
	}
	if edgeCompleteAnime == nil {
		return
	}
	// Get the relation type to fetch for the next node
	edgeRel := getEdgeRelation(edge, rel)
	// Fetch the next node(s)
	err := edgeCompleteAnime.FetchMediaTree(edgeRel, anilistClient, rl, tree, cache)
	if err != nil {
		return
	}
}

// getEdgeRelation returns the relation to fetch for the next node based on the current edge and the relation to fetch.
// If the relation to fetch is FetchMediaTreeAll, it will return FetchMediaTreePrequels for prequels and FetchMediaTreeSequels for sequels.
//
// For example, if the current node is a sequel and the relation to fetch is FetchMediaTreeAll, it will return FetchMediaTreeSequels so that
// only sequels are fetched for the next node.
func getEdgeRelation(edge *CompleteAnime_Relations_Edges, rel FetchMediaTreeRelation) FetchMediaTreeRelation {
	if rel == FetchMediaTreeAll {
		if *edge.RelationType == MediaRelationPrequel {
			return FetchMediaTreePrequels
		}
		if *edge.RelationType == MediaRelationSequel {
			return FetchMediaTreeSequels
		}
	}
	return rel
}
