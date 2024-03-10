package anilist

import (
	"context"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/seanime-app/seanime/internal/result"
)

type (
	BaseMediaRelationTree struct {
		*result.Map[int, *BaseMedia]
	}

	FetchMediaTreeRelation = string
)

const (
	FetchMediaTreeSequels  FetchMediaTreeRelation = "sequels"
	FetchMediaTreePrequels FetchMediaTreeRelation = "prequels"
	FetchMediaTreeAll      FetchMediaTreeRelation = "all"
)

// NewBaseMediaRelationTree returns a new result.Map[int, *BaseMedia].
// It is used to store the results of FetchMediaTree or FetchMediaTree calls.
func NewBaseMediaRelationTree() *BaseMediaRelationTree {
	return &BaseMediaRelationTree{result.NewResultMap[int, *BaseMedia]()}
}

func (m *BasicMedia) FetchMediaTree(rel FetchMediaTreeRelation, acw *ClientWrapper, rl *limiter.Limiter, tree *BaseMediaRelationTree, cache *BaseMediaCache) error {
	rl.Wait()
	res, err := acw.Client.BaseMediaByID(context.Background(), &m.ID)
	if err != nil {
		return err
	}
	return res.GetMedia().FetchMediaTree(rel, acw, rl, tree, cache)
}

// FetchMediaTree populates the BaseMediaRelationTree with the given media's sequels and prequels.
// It also takes a BaseMediaCache to store the fetched media in and avoid duplicate fetches.
// It also takes a limiter.Limiter to limit the number of requests made to the AniList API.
func (m *BaseMedia) FetchMediaTree(rel FetchMediaTreeRelation, acw *ClientWrapper, rl *limiter.Limiter, tree *BaseMediaRelationTree, cache *BaseMediaCache) error {
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
	edges = lo.Filter(edges, func(_edge *BaseMedia_Relations_Edges, _ int) bool {
		return (*_edge.RelationType == MediaRelationSequel || *_edge.RelationType == MediaRelationPrequel) &&
			*_edge.GetNode().Status != MediaStatusNotYetReleased &&
			_edge.IsBroadRelationFormat() && !tree.Has(_edge.GetNode().ID)
	})

	if len(edges) == 0 {
		return nil
	}

	doneCh := make(chan struct{})
	processEdges(edges, rel, acw, rl, tree, cache, doneCh)

	for {
		select {
		case <-doneCh:
			return nil
		default:
		}
	}
}

// processEdges fetches the next node(s) for each edge in parallel.
func processEdges(edges []*BaseMedia_Relations_Edges, rel FetchMediaTreeRelation, acw *ClientWrapper, rl *limiter.Limiter, tree *BaseMediaRelationTree, cache *BaseMediaCache, doneCh chan struct{}) {
	lop.ForEach(edges, func(edge *BaseMedia_Relations_Edges, _ int) {
		processEdge(edge, rel, acw, rl, tree, cache)
	})
	go func() {
		close(doneCh)
	}()
}

func processEdge(edge *BaseMedia_Relations_Edges, rel FetchMediaTreeRelation, acw *ClientWrapper, rl *limiter.Limiter, tree *BaseMediaRelationTree, cache *BaseMediaCache) {
	cacheV, ok := cache.Get(edge.GetNode().ID)
	edgeBaseMedia := cacheV
	if !ok {
		rl.Wait()
		// Fetch the next node
		res, err := acw.Client.BaseMediaByID(context.Background(), &edge.GetNode().ID)
		if err == nil {
			edgeBaseMedia = res.GetMedia()
			cache.Set(edgeBaseMedia.ID, edgeBaseMedia)
		}
	}
	// Get the relation type to fetch for the next node
	edgeRel := getEdgeRelation(edge, rel)
	// Fetch the next node(s)
	err := edgeBaseMedia.FetchMediaTree(edgeRel, acw, rl, tree, cache)
	if err != nil {
		return
	}
}

// getEdgeRelation returns the relation to fetch for the next node based on the current edge and the relation to fetch.
// If the relation to fetch is FetchMediaTreeAll, it will return FetchMediaTreePrequels for prequels and FetchMediaTreeSequels for sequels.
//
// For example, if the current node is a sequel and the relation to fetch is FetchMediaTreeAll, it will return FetchMediaTreeSequels so that
// only sequels are fetched for the next node.
func getEdgeRelation(edge *BaseMedia_Relations_Edges, rel FetchMediaTreeRelation) FetchMediaTreeRelation {
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
